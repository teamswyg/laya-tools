package reviewpacket

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const publicFixture = "package fixture\n"
const publicFixturePin = "c9f09e0a76fbbf83847013f7d79aa09c1566778489cc499e76e3e5d72b1c8db6"

func fixtureConfig(t *testing.T) Config {
	t.Helper()
	dir := t.TempDir()
	input := filepath.Join(dir, "original-public-fixture.go")
	if err := os.WriteFile(input, []byte(publicFixture), 0600); err != nil {
		t.Fatal(err)
	}
	return Config{InputPath: input, ReceiptPath: filepath.Join(dir, "intent.json"), ExpectedSHA256: publicFixturePin, MaxBytes: DefaultMaxBytes}
}

func diskOps() operations { return operations{now: time.Now, open: openInput} }

func writeNewReceipt(path string, data []byte) (os.FileInfo, bool, error) {
	writer, err := newReceiptWriter(path)
	if err != nil {
		return nil, false, err
	}
	info, created, writeErr := writer.writeNew(path, data)
	return info, created, errors.Join(writeErr, writer.close())
}

func resultFrom(t *testing.T, path string) resultRecord {
	t.Helper()
	data, err := os.ReadFile(path + ".result")
	if err != nil {
		t.Fatal(err)
	}
	var result resultRecord
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func assertFailed(t *testing.T, payload []byte, outcome Outcome, err error, cfg Config, code string) resultRecord {
	t.Helper()
	if len(payload) != 0 || err == nil || !strings.Contains(err.Error(), code) || !outcome.StartDurable || !outcome.ResultDurable {
		t.Fatalf("payload length=%d outcome=%+v error=%v; expected durable failure %s", len(payload), outcome, err, code)
	}
	result := resultFrom(t, cfg.ReceiptPath)
	if result.State != "failed" || len(result.Errors) == 0 {
		t.Fatalf("not a preserved failure: %+v", result)
	}
	for _, path := range []string{cfg.ReceiptPath, cfg.ReceiptPath + ".result"} {
		data, err := os.ReadFile(path)
		if err != nil || bytes.Contains(data, []byte(cfg.InputPath)) || bytes.Contains(data, []byte(publicFixture)) {
			t.Fatalf("unsafe receipt content or unreadable receipt: %v", err)
		}
	}
	return result
}

type wrappedInput struct {
	inputFile
	read  func([]byte) (int, error)
	close func() error
}

func (f wrappedInput) Read(b []byte) (int, error) {
	if f.read != nil {
		return f.read(b)
	}
	return f.inputFile.Read(b)
}

func (f wrappedInput) Close() error {
	if f.close != nil {
		return f.close()
	}
	return f.inputFile.Close()
}

func TestDurableStartPrecedesOpenAndTerminalPrecedesPayload(t *testing.T) {
	cfg := fixtureConfig(t)
	cfg.Actor = "public-test-actor"
	cfg.MaxBytes = int64(len(publicFixture)) // An input exactly at the bound succeeds.
	ops := diskOps()
	tick := 0
	ops.now = func() time.Time {
		tick++
		return time.Date(2026, 10, 8, 0, 0, tick, 0, time.UTC)
	}
	startDurable, inputClosed, resultDurable := false, false, false
	var startBytes []byte
	ops.writeNew = func(path string, data []byte) (os.FileInfo, bool, error) {
		if path == cfg.ReceiptPath {
			if tick != 1 {
				t.Fatal("intent UTC must be captured before start write")
			}
			startBytes = append([]byte(nil), data...)
			var start startRecord
			if err := json.Unmarshal(data, &start); err != nil || start.State != "started" || start.ExpectedSHA256 != publicFixturePin || start.Actor != cfg.Actor {
				t.Fatalf("start intent missing independent pin: %+v %v", start, err)
			}
		} else {
			if !inputClosed || tick != 3 {
				t.Fatal("terminal receipt attempted before completed read and close")
			}
		}
		info, created, err := writeNewReceipt(path, data)
		if err == nil {
			if path == cfg.ReceiptPath {
				startDurable = true
			} else {
				resultDurable = true
			}
		}
		return info, created, err
	}
	ops.open = func(path string) (inputFile, error) {
		if !startDurable {
			t.Fatal("input opened before start receipt durability")
		}
		saved, err := os.ReadFile(cfg.ReceiptPath)
		if err != nil || !bytes.Equal(saved, startBytes) {
			t.Fatal("source open cannot observe original persisted start")
		}
		file, err := openInput(path)
		return wrappedInput{inputFile: file, close: func() error { inputClosed = true; return file.Close() }}, err
	}
	payload, outcome, err := capture(cfg, ops)
	if err != nil || string(payload) != publicFixture || !outcome.StartDurable || !outcome.ResultDurable || !resultDurable {
		t.Fatalf("capture failed: %v %+v", err, outcome)
	}
	result := resultFrom(t, cfg.ReceiptPath)
	hash := sha256.Sum256(startBytes)
	if result.State != "verified" || result.StartReceiptSHA256 != hex.EncodeToString(hash[:]) || result.ActualSHA256 != publicFixturePin || result.OpenAttempts != 1 || result.BytesRead != int64(len(publicFixture)) {
		t.Fatalf("result does not bind exact start and input: %+v", result)
	}
	if result.ReadCompletedUTC != "2026-10-08T00:00:02Z" || result.DecidedUTC != "2026-10-08T00:00:03Z" {
		t.Fatalf("read completion must follow intent: %+v", result)
	}
	for _, path := range []string{cfg.ReceiptPath, cfg.ReceiptPath + ".result"} {
		info, err := os.Stat(path)
		if err != nil || info.Mode().Perm() != 0600 {
			t.Fatalf("new receipt permissions: %v %v", info, err)
		}
	}
}

func TestPinMismatchOversizeMissingAndNonregularPreserveFailure(t *testing.T) {
	for _, kind := range []string{"mismatch", "oversize", "missing", "directory"} {
		t.Run(kind, func(t *testing.T) {
			cfg := fixtureConfig(t)
			code := "pin_mismatch"
			switch kind {
			case "mismatch":
				cfg.ExpectedSHA256 = strings.Repeat("0", 64)
			case "oversize":
				cfg.MaxBytes = 3
				code = "input_oversized"
			case "missing":
				cfg.InputPath = filepath.Join(t.TempDir(), "missing-parent", "missing.go")
				code = "input_stat_failed"
			case "directory":
				cfg.InputPath = t.TempDir()
				code = "input_type_denied"
			}
			payload, outcome, err := Capture(cfg)
			result := assertFailed(t, payload, outcome, err, cfg, code)
			if kind == "mismatch" && (result.ActualSHA256 != publicFixturePin || result.ReadCompletedUTC == "") {
				t.Fatal("complete mismatched input must retain its actual digest and read completion")
			}
			if kind != "mismatch" && (result.ActualSHA256 != "" || result.ReadCompletedUTC != "") {
				t.Fatal("incomplete input must not claim complete digest")
			}
			if kind == "oversize" && result.BytesRead != cfg.MaxBytes+1 {
				t.Fatal("oversized input must stop at one byte beyond the bound")
			}
		})
	}
}

func TestInputFailuresAreRedactedAndAllReadErrorsPreserved(t *testing.T) {
	for _, kind := range []string{"permission", "read-and-close"} {
		t.Run(kind, func(t *testing.T) {
			cfg := fixtureConfig(t)
			ops := diskOps()
			ops.open = func(path string) (inputFile, error) {
				if kind == "permission" {
					return nil, &os.PathError{Op: "open", Path: cfg.InputPath, Err: os.ErrPermission}
				}
				file, err := openInput(path)
				return wrappedInput{inputFile: file,
					read:  func(b []byte) (int, error) { return copy(b, "public"), errors.New(publicFixture + cfg.InputPath) },
					close: func() error { _ = file.Close(); return os.ErrPermission }}, err
			}
			payload, outcome, err := capture(cfg, ops)
			code := "input_open_failed"
			if kind == "read-and-close" {
				code = "input_read_failed"
			}
			result := assertFailed(t, payload, outcome, err, cfg, code)
			if strings.Contains(err.Error(), cfg.InputPath) || strings.Contains(err.Error(), publicFixture) || result.OpenAttempts != 1 {
				t.Fatal("errors leaked source/path or retried the input")
			}
			if kind == "read-and-close" && (len(result.Errors) != 2 || result.Errors[1].Code != "input_close_failed") {
				t.Fatal("secondary close error was discarded")
			}
		})
	}
}

func TestReceiptFailuresSuppressPayloadAndPreserveArtifacts(t *testing.T) {
	for _, phase := range []string{"start-create", "start-sync", "result-create", "result-sync", "result-race"} {
		t.Run(phase, func(t *testing.T) {
			cfg := fixtureConfig(t)
			ops := diskOps()
			opens, writes := 0, 0
			ops.open = func(path string) (inputFile, error) { opens++; return openInput(path) }
			ops.writeNew = func(path string, data []byte) (os.FileInfo, bool, error) {
				writes++
				selected := strings.HasPrefix(phase, "start") == (path == cfg.ReceiptPath)
				if !selected {
					return writeNewReceipt(path, data)
				}
				if strings.HasSuffix(phase, "create") {
					return nil, false, os.ErrPermission
				}
				if phase == "result-race" {
					if err := os.WriteFile(path, []byte("original public sentinel\n"), 0600); err != nil {
						t.Fatal(err)
					}
					return writeNewReceipt(path, data)
				}
				info, created, err := writeNewReceipt(path, data)
				if err != nil {
					t.Fatal(err)
				}
				return info, created, os.ErrPermission // Inject failure of the durability acknowledgement.
			}
			payload, outcome, err := capture(cfg, ops)
			if len(payload) != 0 || err == nil {
				t.Fatalf("receipt failure released payload: %v %+v", err, outcome)
			}
			if strings.HasPrefix(phase, "start") && (opens != 0 || writes != 1 || outcome.StartDurable) {
				t.Fatal("failed start receipt permitted input open or hidden retry")
			}
			if strings.HasPrefix(phase, "result") && (opens != 1 || writes != 2 || !outcome.StartDurable || outcome.ResultDurable) {
				t.Fatal("failed terminal receipt was treated as durable or retried")
			}
			if outcome.StartCreated {
				var start startRecord
				data, readErr := os.ReadFile(cfg.ReceiptPath)
				if readErr != nil || json.Unmarshal(data, &start) != nil || start.State != "started" {
					t.Fatal("immutable start receipt was discarded or overwritten")
				}
			}
			if phase == "result-race" {
				data, _ := os.ReadFile(cfg.ReceiptPath + ".result")
				if string(data) != "original public sentinel\n" {
					t.Fatal("racing existing terminal receipt was overwritten")
				}
			}
		})
	}
}

func TestOptionsAndReceiptAliasesCannotOverwriteInput(t *testing.T) {
	for _, kind := range []string{"input-alias", "result-alias", "parent-symlink-alias", "existing-start", "existing-result", "input-symlink", "dangling-input-symlink", "receipt-symlink", "receipt-hardlink", "invalid-pin", "invalid-bound", "actor", "parent-traversal"} {
		t.Run(kind, func(t *testing.T) {
			cfg := fixtureConfig(t)
			originalInput := cfg.InputPath
			switch kind {
			case "input-alias":
				cfg.ReceiptPath = cfg.InputPath
			case "result-alias":
				cfg.InputPath = cfg.ReceiptPath + ".result"
				if err := os.WriteFile(cfg.InputPath, []byte(publicFixture), 0600); err != nil {
					t.Fatal(err)
				}
			case "parent-symlink-alias":
				link := filepath.Join(t.TempDir(), "public-parent-link")
				if err := os.Symlink(filepath.Dir(cfg.InputPath), link); err != nil {
					t.Fatal(err)
				}
				cfg.ReceiptPath = filepath.Join(link, filepath.Base(cfg.InputPath))
			case "existing-start", "existing-result":
				path := cfg.ReceiptPath
				if kind == "existing-result" {
					path += ".result"
				}
				if err := os.WriteFile(path, []byte("original public sentinel\n"), 0600); err != nil {
					t.Fatal(err)
				}
			case "input-symlink", "dangling-input-symlink":
				cfg.InputPath = filepath.Join(filepath.Dir(cfg.ReceiptPath), "public-link")
				target := originalInput
				if kind == "dangling-input-symlink" {
					target = cfg.ReceiptPath
				}
				if err := os.Symlink(target, cfg.InputPath); err != nil {
					t.Fatal(err)
				}
			case "receipt-symlink":
				if err := os.Symlink(cfg.InputPath, cfg.ReceiptPath); err != nil {
					t.Fatal(err)
				}
			case "receipt-hardlink":
				if err := os.Link(cfg.InputPath, cfg.ReceiptPath); err != nil {
					t.Fatal(err)
				}
			case "invalid-pin":
				cfg.ExpectedSHA256 = "invalid"
			case "invalid-bound":
				cfg.MaxBytes = MaximumMaxBytes + 1
			case "actor":
				cfg.Actor = "  "
			case "parent-traversal":
				cfg.InputPath = filepath.Join(filepath.Dir(cfg.InputPath), "unused") + "/../original-public-fixture.go"
			}
			ops := diskOps()
			ops.open = func(string) (inputFile, error) { t.Fatal("invalid request opened input"); return nil, nil }
			payload, outcome, err := capture(cfg, ops)
			if err == nil || len(payload) != 0 || outcome.StartCreated || outcome.ResultCreated {
				t.Fatalf("unsafe request accepted: %v %+v", err, outcome)
			}
			data, err := os.ReadFile(originalInput)
			if err != nil || string(data) != publicFixture {
				t.Fatal("public input was overwritten")
			}
			if kind == "existing-start" || kind == "existing-result" {
				path := cfg.ReceiptPath
				if kind == "existing-result" {
					path += ".result"
				}
				data, _ := os.ReadFile(path)
				if string(data) != "original public sentinel\n" {
					t.Fatal("existing receipt was overwritten")
				}
			}
		})
	}
}

func TestOpenedIdentityMustMatchCheckedFile(t *testing.T) {
	cfg := fixtureConfig(t)
	ops := diskOps()
	ops.open = func(path string) (inputFile, error) {
		replacement := filepath.Join(filepath.Dir(path), "replacement-public-fixture.go")
		if err := os.WriteFile(replacement, []byte(publicFixture), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(replacement, path); err != nil {
			t.Fatal(err)
		}
		file, err := openInput(path)
		return wrappedInput{inputFile: file, read: func([]byte) (int, error) { t.Fatal("changed identity was read"); return 0, io.EOF }}, err
	}
	payload, outcome, err := capture(cfg, ops)
	result := assertFailed(t, payload, outcome, err, cfg, "input_identity_changed")
	if result.BytesRead != 0 || result.ActualSHA256 != "" {
		t.Fatal("replaced input was read or hashed")
	}
}

func TestReceiptPairRemainsInAnchoredParentAfterPathReplacement(t *testing.T) {
	cfg := fixtureConfig(t)
	parent, replacement, links := t.TempDir(), t.TempDir(), t.TempDir()
	alias := filepath.Join(links, "public-parent-link")
	if err := os.Symlink(parent, alias); err != nil {
		t.Fatal(err)
	}
	cfg.ReceiptPath = filepath.Join(alias, "intent.json")
	before, err := os.Stat(parent)
	if err != nil {
		t.Fatal(err)
	}
	ops := diskOps()
	ops.open = func(path string) (inputFile, error) {
		if err := os.Remove(alias); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(replacement, alias); err != nil {
			t.Fatal(err)
		}
		return openInput(path)
	}
	payload, outcome, err := capture(cfg, ops)
	if err != nil || !outcome.StartDurable || !outcome.ResultDurable || string(payload) != publicFixture {
		t.Fatalf("anchored capture failed: %+v %v", outcome, err)
	}
	for _, name := range []string{"intent.json", "intent.json.result"} {
		if _, err := os.Stat(filepath.Join(parent, name)); err != nil {
			t.Fatal("receipt pair separated from the anchored parent:", err)
		}
		if _, err := os.Stat(filepath.Join(replacement, name)); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("receipt was created in the replacement parent")
		}
	}
	after, err := os.Stat(parent)
	if err != nil || before.Mode().Perm() != after.Mode().Perm() {
		t.Fatal("existing parent permissions changed")
	}
}

func TestCreatedStartReceiptCannotBeReadThroughHardLink(t *testing.T) {
	cfg := fixtureConfig(t)
	ops := diskOps()
	ops.writeNew = func(path string, data []byte) (os.FileInfo, bool, error) {
		info, created, err := writeNewReceipt(path, data)
		if err == nil && path == cfg.ReceiptPath {
			if err := os.Remove(cfg.InputPath); err != nil {
				t.Fatal(err)
			}
			if err := os.Link(cfg.ReceiptPath, cfg.InputPath); err != nil {
				t.Fatal(err)
			}
		}
		return info, created, err
	}
	ops.open = func(string) (inputFile, error) {
		t.Fatal("start receipt alias was opened as input")
		return nil, nil
	}
	payload, outcome, err := capture(cfg, ops)
	result := assertFailed(t, payload, outcome, err, cfg, "path_alias")
	if result.OpenAttempts != 0 {
		t.Fatal("receipt alias recorded an input-open attempt")
	}
}

func TestFailedTerminalWriteRetainsReadError(t *testing.T) {
	cfg := fixtureConfig(t)
	cfg.ExpectedSHA256 = strings.Repeat("0", 64)
	ops := diskOps()
	ops.writeNew = func(path string, data []byte) (os.FileInfo, bool, error) {
		if path != cfg.ReceiptPath {
			return nil, false, os.ErrPermission
		}
		return writeNewReceipt(path, data)
	}
	payload, outcome, err := capture(cfg, ops)
	if len(payload) != 0 || !outcome.StartDurable || outcome.ResultDurable || err == nil || !strings.Contains(err.Error(), "pin_mismatch") || !strings.Contains(err.Error(), "result_receipt_failed") {
		t.Fatalf("terminal failure discarded read failure: %+v %v", outcome, err)
	}
}
