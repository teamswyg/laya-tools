// Package reviewpacket records intent before a bounded, pinned source read.
// Its receipts describe only this tool's read, never an agent's first-ever read.
package reviewpacket

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultMaxBytes int64 = 1 << 20
	MaximumMaxBytes int64 = 8 << 20
	version               = "riido-reviewpacket/v1"
)

type Config struct {
	InputPath, ReceiptPath, ExpectedSHA256, Actor string
	MaxBytes                                      int64
}

// Outcome also reports partial receipt creation when storage fails.
type Outcome struct {
	StartCreated, StartDurable, ResultCreated, ResultDurable bool
}

type issue struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

func (e issue) Error() string { return e.Code + ": " + e.Detail }

type startRecord struct {
	Version        string `json:"version"`
	State          string `json:"state"`
	StartedUTC     string `json:"started_utc"`
	ExpectedSHA256 string `json:"expected_sha256"`
	MaxBytes       int64  `json:"max_bytes"`
	InputPolicy    string `json:"input_policy"`
	AttemptLimit   int    `json:"read_attempt_limit"`
	Actor          string `json:"actor,omitempty"`
}

type resultRecord struct {
	Version            string  `json:"version"`
	State              string  `json:"state"`
	StartReceiptSHA256 string  `json:"start_receipt_sha256"`
	ExpectedSHA256     string  `json:"expected_sha256"`
	ActualSHA256       string  `json:"actual_sha256,omitempty"`
	BytesRead          int64   `json:"bytes_read"`
	OpenAttempts       int     `json:"input_open_attempts"`
	ReadCompletedUTC   string  `json:"read_completed_utc,omitempty"`
	DecidedUTC         string  `json:"decided_utc"`
	Errors             []issue `json:"errors,omitempty"`
}

type inputFile interface {
	io.Reader
	Stat() (os.FileInfo, error)
	Close() error
}

type operations struct {
	now      func() time.Time
	open     func(string) (inputFile, error)
	writeNew func(string, []byte) (os.FileInfo, bool, error)
}

// Capture returns payload only after the immutable start and terminal sibling
// receipts have been synced. It makes at most one input-open attempt.
func Capture(cfg Config) ([]byte, Outcome, error) {
	return capture(cfg, operations{now: time.Now, open: openInput})
}

func capture(cfg Config, ops operations) (payload []byte, outcome Outcome, err error) {
	if err := validate(&cfg); err != nil {
		return nil, outcome, err
	}
	if err := validatePaths(cfg); err != nil {
		return nil, outcome, err
	}
	if ops.writeNew == nil {
		writer, openErr := newReceiptWriter(cfg.ReceiptPath)
		if openErr != nil {
			return nil, outcome, problem("receipt_parent_open_failed", openErr)
		}
		defer func() {
			if closeErr := writer.close(); closeErr != nil {
				payload = nil
				err = errors.Join(err, problem("receipt_parent_close_failed", closeErr))
			}
		}()
		ops.writeNew = writer.writeNew
	}
	stamp := func() string { return ops.now().UTC().Format(time.RFC3339Nano) }
	start := startRecord{version, "started", stamp(), cfg.ExpectedSHA256, cfg.MaxBytes,
		"regular-file; no final symlink; nonblocking open; identity checked; bounded read; no retries", 1, cfg.Actor}
	startBytes, err := encode(start)
	if err != nil {
		return nil, outcome, problem("start_encode_failed", err)
	}
	startInfo, created, err := ops.writeNew(cfg.ReceiptPath, startBytes)
	outcome.StartCreated = created
	if err != nil {
		return nil, outcome, problem("start_receipt_failed", err)
	}
	outcome.StartDurable = true
	startHash := sha256.Sum256(startBytes)
	result := resultRecord{Version: version, State: "failed", StartReceiptSHA256: hex.EncodeToString(startHash[:]), ExpectedSHA256: cfg.ExpectedSHA256}
	payload, readErrors := readInput(cfg, startInfo, &result, ops, stamp)
	result.Errors = readErrors
	if len(readErrors) == 0 {
		result.State = "verified"
	}
	result.DecidedUTC = stamp()
	resultBytes, err := encode(result)
	if err != nil {
		return nil, outcome, problem("result_encode_failed", err)
	}
	_, created, err = ops.writeNew(cfg.ReceiptPath+".result", resultBytes)
	outcome.ResultCreated = created
	if err != nil {
		return nil, outcome, errors.Join(issuesError(readErrors), problem("result_receipt_failed", err))
	}
	outcome.ResultDurable = true
	if len(readErrors) != 0 {
		return nil, outcome, issuesError(readErrors)
	}
	return payload, outcome, nil
}

func validate(cfg *Config) error {
	if cfg.InputPath == "" || cfg.ReceiptPath == "" {
		return issue{"path_required", "input and receipt paths are required"}
	}
	if cfg.MaxBytes == 0 {
		cfg.MaxBytes = DefaultMaxBytes
	}
	if cfg.MaxBytes < 1 || cfg.MaxBytes > MaximumMaxBytes {
		return issue{"max_bytes_invalid", "max-bytes must be between 1 and 8388608"}
	}
	pin, err := hex.DecodeString(cfg.ExpectedSHA256)
	if err != nil || len(pin) != sha256.Size {
		return issue{"pin_invalid", "sha256 must be an independently supplied 64-digit hexadecimal digest"}
	}
	cfg.ExpectedSHA256 = strings.ToLower(cfg.ExpectedSHA256)
	if cfg.Actor != "" && (strings.TrimSpace(cfg.Actor) == "" || len(cfg.Actor) > 256) {
		return issue{"actor_invalid", "actor must be a nonempty public label of at most 256 bytes"}
	}
	return nil
}

func validatePaths(cfg Config) error {
	paths := []string{cfg.InputPath, cfg.ReceiptPath, cfg.ReceiptPath + ".result"}
	canonical := make([]string, len(paths))
	for i, path := range paths {
		for _, part := range strings.Split(filepath.ToSlash(path), "/") {
			if part == ".." {
				return issue{"path_invalid", "parent traversal components are not allowed"}
			}
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return problem("path_invalid", err)
		}
		parent, err := filepath.EvalSymlinks(filepath.Dir(abs))
		if err != nil {
			if i != 0 {
				return problem("receipt_parent_invalid", err)
			}
			parent = filepath.Dir(abs) // Missing inputs still receive failure receipts.
		}
		canonical[i] = filepath.Join(parent, filepath.Base(abs))
	}
	if canonical[0] == canonical[1] || canonical[0] == canonical[2] {
		return issue{"path_alias", "input must differ from both receipt paths"}
	}
	if info, err := os.Lstat(cfg.InputPath); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return issue{"input_symlink", "a final input symlink is not allowed"}
	}
	for _, path := range paths[1:] {
		if _, err := os.Lstat(path); err == nil {
			return issue{"receipt_exists", "both receipt paths must be new"}
		} else if !errors.Is(err, os.ErrNotExist) {
			return problem("receipt_path_invalid", err)
		}
	}
	return nil
}

func readInput(cfg Config, startInfo os.FileInfo, result *resultRecord, ops operations, stamp func() string) ([]byte, []issue) {
	before, err := os.Lstat(cfg.InputPath)
	if err != nil {
		return nil, []issue{problem("input_stat_failed", err)}
	}
	if !before.Mode().IsRegular() {
		return nil, []issue{{"input_type_denied", "input must be a regular file without a final symlink"}}
	}
	if startInfo != nil && os.SameFile(before, startInfo) {
		return nil, []issue{{"path_alias", "input aliases the created start receipt"}}
	}
	result.OpenAttempts++
	file, err := ops.open(cfg.InputPath)
	if err != nil {
		return nil, []issue{problem("input_open_failed", err)}
	}
	var issues []issue
	after, err := file.Stat()
	if err != nil {
		issues = append(issues, problem("input_stat_failed", err))
	} else if !after.Mode().IsRegular() || !os.SameFile(before, after) {
		issues = append(issues, issue{"input_identity_changed", "opened input must match the checked regular file"})
	}
	var payload []byte
	if len(issues) == 0 {
		// One extra byte detects growth or inaccurate size metadata without an
		// unbounded allocation. No payload escapes before terminal durability.
		buffer := make([]byte, int(cfg.MaxBytes)+1)
		n, readErr := io.ReadFull(file, buffer)
		result.BytesRead = int64(n)
		if n > int(cfg.MaxBytes) {
			issues = append(issues, issue{"input_oversized", "input exceeds max-bytes"})
		} else if readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
			issues = append(issues, problem("input_read_failed", readErr))
		} else {
			result.ReadCompletedUTC = stamp()
			payload = buffer[:n]
			hash := sha256.Sum256(payload)
			result.ActualSHA256 = hex.EncodeToString(hash[:])
			if result.ActualSHA256 != cfg.ExpectedSHA256 {
				issues = append(issues, issue{"pin_mismatch", "complete input digest differs from expected sha256"})
			}
		}
	}
	if err := file.Close(); err != nil {
		issues = append(issues, problem("input_close_failed", err))
	}
	return payload, issues
}

func encode(value any) ([]byte, error) {
	b, err := json.MarshalIndent(value, "", "  ")
	return append(b, '\n'), err
}

func issuesError(issues []issue) error {
	var errs []error
	for _, entry := range issues {
		errs = append(errs, entry)
	}
	return errors.Join(errs...)
}

func problem(code string, err error) issue {
	detail := "I/O failure"
	switch {
	case errors.Is(err, os.ErrNotExist):
		detail = "not found"
	case errors.Is(err, os.ErrExist):
		detail = "already exists"
	case errors.Is(err, os.ErrPermission):
		detail = "permission denied"
	}
	return issue{code, detail}
}

type receiptWriter struct {
	root *os.Root
	dir  *os.File
}

func newReceiptWriter(path string) (*receiptWriter, error) {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	dir, err := root.Open(".")
	if err != nil {
		return nil, errors.Join(err, root.Close())
	}
	return &receiptWriter{root, dir}, nil
}

func (w *receiptWriter) close() error { return errors.Join(w.dir.Close(), w.root.Close()) }

func (w *receiptWriter) writeNew(path string, data []byte) (os.FileInfo, bool, error) {
	file, err := w.root.OpenFile(filepath.Base(path), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, false, err
	}
	info, statErr := file.Stat()
	chmodErr := file.Chmod(0600) // Only the exclusively created file is changed.
	n, writeErr := file.Write(data)
	if n != len(data) && writeErr == nil {
		writeErr = io.ErrShortWrite
	}
	syncErr := file.Sync()
	closeErr := file.Close()
	if err := errors.Join(statErr, chmodErr, writeErr, syncErr, closeErr); err != nil {
		return info, true, err
	}
	return info, true, w.dir.Sync()
}
