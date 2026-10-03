package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const globalCap int64 = 512 << 20
const controlCap int64 = 128 << 10
const restoreCap int64 = 1 << 20
const inflateCap int64 = 2 << 20
const maxArchiveCap int64 = 2 << 20
const metadataCap int64 = 256 << 10
const markerName = ".RECOVERABLE-ARCHIVE.v1.json"

var targetNames = [2]string{"HF33-delta-public-v1", "HF33-pinned-delta-readback-v1"}
var fileCounts = [2]int{82, 252}
var directoryCounts = [2]int{29, 62}
var payloadSizes = [2]int64{842356, 962496}

var errContract = errors.New("archive_contract")
var errChanged = errors.New("archive_source_changed")
var errCap = errors.New("archive_cap")
var errArtifact = errors.New("archive_artifact")

type pin struct {
	Path   string `json:"path"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

type meta struct {
	Mode    string `json:"mode_octal"`
	Device  string `json:"device"`
	Inode   string `json:"inode"`
	Nlink   uint64 `json:"nlink"`
	MtimeNS string `json:"mtime_ns"`
}

type entry struct {
	Relative string `json:"relative_path"`
	Bytes    int64  `json:"bytes"`
	SHA256   string `json:"sha256"`
	meta
}

type directory struct {
	Relative string `json:"relative_path"`
	meta
}

type target struct {
	ID          string      `json:"id"`
	Root        meta        `json:"root_metadata"`
	Files       []entry     `json:"files"`
	Directories []directory `json:"directories"`
}

type selection struct {
	Schema                              string   `json:"schema"`
	PublicationCommit                   string   `json:"publication_commit"`
	Targets                             []target `json:"targets"`
	CanonicalScienceExcluded            bool     `json:"canonical_science_excluded"`
	ModelsTokensRawObservationsExcluded bool     `json:"models_tokens_raw_observations_excluded"`
}

type admission struct {
	Schema                string  `json:"schema"`
	Selection             pin     `json:"selection"`
	PriorAccount          pin     `json:"prior_account"`
	Census                pin     `json:"census"`
	HelperSources         []pin   `json:"helper_sources"`
	CapBytes              int64   `json:"cap_bytes"`
	CurrentRetainedBytes  int64   `json:"current_retained_bytes"`
	ArchiveCaps           []int64 `json:"archive_caps"`
	RestoreCap            int64   `json:"restore_payload_cap_bytes"`
	ControlCap            int64   `json:"new_control_cap_bytes"`
	InflateCap            int64   `json:"tar_inflate_cap_bytes"`
	WritersQuiescent      bool    `json:"writers_quiescent"`
	FixedPublicCopiesOnly bool    `json:"fixed_public_copies_only"`
}

type config struct {
	Schema       string   `json:"schema"`
	Root         string   `json:"root"`
	OutputBase   string   `json:"output_base"`
	RestoreBases []string `json:"restore_bases"`
	Selection    pin      `json:"selection"`
	Admission    pin      `json:"admission"`
	PriorAccount pin      `json:"prior_account"`
}

type outcome struct {
	Schema                       string    `json:"schema"`
	Complete                     bool      `json:"complete"`
	CompletedTargets             int       `json:"completed_targets"`
	RemovedACKFiles              int       `json:"removed_ack_files"`
	RemovedACKLogicalBytes       int64     `json:"removed_ack_logical_bytes"`
	UnacknowledgedRemovePossible bool      `json:"unacknowledged_remove_possible"`
	ArchiveBytes                 [2]*int64 `json:"archive_bytes"`
	ReclaimedLogicalBytes        *int64    `json:"reclaimed_logical_bytes"`
	Failure                      string    `json:"failure"`
	AutomaticRetries             int       `json:"automatic_retries"`
}

type commit struct {
	Schema                             string `json:"schema"`
	TargetID                           string `json:"target_id"`
	Selection                          pin    `json:"selection"`
	Admission                          pin    `json:"admission"`
	PriorAccount                       pin    `json:"prior_account"`
	InventorySHA                       string `json:"inventory_sha256"`
	Archive                            pin    `json:"archive"`
	RestoredRoot                       string `json:"restored_root"`
	PayloadBytes                       int64  `json:"payload_bytes"`
	RegularFiles                       int    `json:"regular_files"`
	FullEOFCRCVerified                 bool   `json:"full_eof_crc_verified"`
	RestoredFilesBytesSHAAndModesMatch bool   `json:"restored_files_bytes_sha_and_modes_match"`
	FileAndDirectorySyncReturned       bool   `json:"file_and_directory_sync_returned"`
	BeforeAnyOriginalUnlink            bool   `json:"before_any_original_unlink"`
}

func digest(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func validSHA(s string) bool {
	if len(s) != 64 || strings.ToLower(s) != s {
		return false
	}
	_, e := hex.DecodeString(s)
	return e == nil
}
func canonical(v any) ([]byte, error) {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return nil, e
	}
	return append(b, '\n'), nil
}

// Byte equality to the typed canonical encoding rejects duplicate keys, case aliases,
// unknown fields, alternate ordering and omitted fields. Raw keys are never trusted.
func decodeCanonical(b []byte, v any) error {
	d := json.NewDecoder(strings.NewReader(string(b)))
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		return errContract
	}
	var trailing any
	if d.Decode(&trailing) != io.EOF {
		return errContract
	}
	c, e := canonical(v)
	if e != nil || string(c) != string(b) {
		return errContract
	}
	return nil
}
func cleanAbsolute(p string) bool {
	return filepath.IsAbs(p) && filepath.Clean(p) == p && !strings.ContainsRune(p, 0)
}
func cleanRelative(p string) bool {
	if p == "" || len(p) > 256 || strings.ContainsAny(p, "\\\x00") || strings.HasPrefix(p, "/") {
		return false
	}
	return filepath.ToSlash(filepath.Clean(p)) == p && p != "." && p != ".." && !strings.HasPrefix(p, "../")
}
func freshBase(s string) bool {
	if len(s) < 3 || len(s) > 96 || strings.HasPrefix(s, ".") {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
			return false
		}
	}
	return true
}
func modeOf(s string) (os.FileMode, error) {
	n, e := strconv.ParseUint(s, 8, 12)
	if e != nil || n > 0777 || strconv.FormatUint(n, 8) != strings.TrimPrefix(s, "0") {
		return 0, errContract
	}
	// Only plain owner-readable/writable permission bits are admissible.
	if n&0600 != 0600 {
		return 0, errContract
	}
	return os.FileMode(n), nil
}
func plainPermissions(f os.FileInfo) bool {
	return f.Mode()&(os.ModeSetuid|os.ModeSetgid|os.ModeSticky) == 0
}
func infoMeta(f os.FileInfo) (meta, error) {
	if !plainPermissions(f) {
		return meta{}, errContract
	}
	x, ok := f.Sys().(*syscall.Stat_t)
	if !ok {
		return meta{}, errContract
	}
	return meta{Mode: "0" + strconv.FormatUint(uint64(f.Mode().Perm()), 8), Device: strconv.FormatUint(uint64(x.Dev), 10), Inode: strconv.FormatUint(x.Ino, 10), Nlink: uint64(x.Nlink), MtimeNS: strconv.FormatInt(f.ModTime().UnixNano(), 10)}, nil
}
func sameMeta(f os.FileInfo, m meta, checkTime bool) bool {
	got, e := infoMeta(f)
	return e == nil && got.Mode == m.Mode && got.Device == m.Device && got.Inode == m.Inode && got.Nlink == m.Nlink && (!checkTime || got.MtimeNS == m.MtimeNS)
}
func noSymlinkAbsolute(p string) error {
	if !cleanAbsolute(p) {
		return errContract
	}
	cur := string(filepath.Separator)
	for _, part := range strings.Split(strings.TrimPrefix(p, cur), cur) {
		if part == "" {
			continue
		}
		cur = filepath.Join(cur, part)
		f, e := os.Lstat(cur)
		if e != nil || f.Mode()&os.ModeSymlink != 0 {
			return errArtifact
		}
		if cur != p && !f.IsDir() {
			return errArtifact
		}
	}
	return nil
}
func syncDir(p string) error {
	f, e := os.Open(p)
	if e != nil {
		return e
	}
	e = f.Sync()
	ce := f.Close()
	if e != nil {
		return e
	}
	return ce
}
func openReadNoFollow(p string) (*os.File, os.FileInfo, error) {
	if e := noSymlinkAbsolute(p); e != nil {
		return nil, nil, e
	}
	before, e := os.Lstat(p)
	if e != nil || !before.Mode().IsRegular() {
		return nil, nil, errArtifact
	}
	m, e := infoMeta(before)
	if e != nil || m.Nlink != 1 {
		return nil, nil, errArtifact
	}
	fd, e := syscall.Open(p, syscall.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if e != nil {
		return nil, nil, e
	}
	f := os.NewFile(uintptr(fd), p)
	current, e := f.Stat()
	if e != nil || !os.SameFile(before, current) || !sameMeta(current, m, true) {
		f.Close()
		return nil, nil, errChanged
	}
	return f, before, nil
}
func hashRegular(p string, capBytes int64) (pin, os.FileInfo, error) {
	f, before, e := openReadNoFollow(p)
	if e != nil {
		return pin{}, nil, e
	}
	defer f.Close()
	if before.Size() < 0 || before.Size() > capBytes {
		return pin{}, nil, errCap
	}
	h := sha256.New()
	n, e := io.Copy(h, io.LimitReader(f, capBytes+1))
	after, ae := f.Stat()
	last, le := os.Lstat(p)
	m, me := infoMeta(before)
	if e != nil || ae != nil || le != nil || me != nil || n != before.Size() || !sameMeta(after, m, true) || !sameMeta(last, m, true) {
		return pin{}, nil, errChanged
	}
	return pin{p, n, hex.EncodeToString(h.Sum(nil))}, before, nil
}
func readPin(p pin, capBytes int64) ([]byte, error) {
	if !cleanAbsolute(p.Path) || p.Bytes < 0 || p.Bytes > capBytes || !validSHA(p.SHA256) {
		return nil, errContract
	}
	got, _, e := hashRegular(p.Path, capBytes)
	if e != nil || got != p {
		return nil, errChanged
	}
	f, _, e := openReadNoFollow(p.Path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, capBytes+1))
	if e != nil || int64(len(b)) != p.Bytes || digest(b) != p.SHA256 {
		return nil, errChanged
	}
	return b, nil
}
