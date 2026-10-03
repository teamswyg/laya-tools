// Copyright 2026 teamswyg. Licensed under the Apache License, Version 2.0.
package compare

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

func Digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func hashShape(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range []byte(s) {
		if !(c >= 'a' && c <= 'f' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}
func canonicalPath(p string) bool { return filepath.IsAbs(p) && filepath.Clean(p) == p }
func compact(b []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		return ErrShape
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return ErrShape
	}
	got, e := json.Marshal(v)
	var want bytes.Buffer
	if e != nil || json.Compact(&want, b) != nil || !bytes.Equal(got, want.Bytes()) {
		return ErrShape
	}
	return nil
}
func Canonical(b []byte, v any) error {
	if len(b) < 2 || b[len(b)-1] != '\n' || compact(b, v) != nil {
		return ErrShape
	}
	want, e := json.Marshal(v)
	if e != nil || !bytes.Equal(b, append(want, '\n')) {
		return ErrShape
	}
	return nil
}
func readPin(p Pin, cap int64, private bool) ([]byte, error) {
	if p.ID == "" || !canonicalPath(p.Path) || p.Bytes <= 0 || p.Bytes > cap || !hashShape(p.SHA256) {
		return nil, ErrPin
	}
	real, e := filepath.EvalSymlinks(p.Path)
	if e != nil || real != p.Path {
		return nil, ErrIO
	}
	s, e := os.Lstat(p.Path)
	if e != nil || !s.Mode().IsRegular() || s.Size() != p.Bytes {
		return nil, ErrPin
	}
	if private {
		native, ok := s.Sys().(*syscall.Stat_t)
		if s.Mode().Perm() != 0600 || !ok || native.Nlink != 1 {
			return nil, ErrShape
		}
	}
	f, e := os.OpenFile(p.Path, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if e != nil {
		return nil, ErrIO
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil || !os.SameFile(s, st) {
		return nil, ErrIO
	}
	b, e := io.ReadAll(io.LimitReader(f, cap+1))
	after, x := os.Lstat(p.Path)
	if e != nil || x != nil || len(b) != int(p.Bytes) || Digest(b) != p.SHA256 || !os.SameFile(s, after) || s.Size() != after.Size() || !s.ModTime().Equal(after.ModTime()) {
		return nil, ErrPin
	}
	return b, nil
}
func LoadConfig(path, sha string) (c Config, e error) {
	if !hashShape(sha) || !canonicalPath(path) {
		return c, ErrPin
	}
	s, e := os.Lstat(path)
	if e != nil || !s.Mode().IsRegular() || s.Size() > 65536 {
		return c, ErrIO
	}
	b, e := readPin(Pin{ID: "Root_saved_config", Path: path, Bytes: s.Size(), SHA256: sha}, 65536, false)
	if e != nil {
		return
	}
	e = Canonical(b, &c)
	return
}

// Reserve before comparing. A failure preserves this exclusive empty/partial
// output as attempt evidence; callers must not retry into the same pathname.
func ReserveExclusive(path string) (*os.File, error) {
	if !canonicalPath(path) {
		return nil, ErrBounds
	}
	parent := filepath.Dir(path)
	real, e := filepath.EvalSymlinks(parent)
	if e != nil || real != parent {
		return nil, ErrIO
	}
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return nil, ErrIO
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return nil, ErrIO
	}
	d, e := os.Open(parent)
	if e != nil {
		f.Close()
		return nil, ErrIO
	}
	e = d.Sync()
	x := d.Close()
	if e != nil || x != nil {
		f.Close()
		return nil, ErrIO
	}
	return f, nil
}
func WriteReserved(path string, f *os.File, b []byte) error {
	if f == nil || !canonicalPath(path) || len(b) > ReportCap {
		return ErrBounds
	}
	named, e := os.Lstat(path)
	opened, x := f.Stat()
	if e != nil || x != nil || !named.Mode().IsRegular() || named.Mode().Perm() != 0600 || !os.SameFile(named, opened) || opened.Size() != 0 {
		return ErrIO
	}
	n, e := f.Write(b)
	if e == nil && n != len(b) {
		e = ErrIO
	}
	if e == nil {
		e = f.Sync()
	}
	x = f.Close()
	if e != nil || x != nil {
		return ErrIO
	}
	d, e := os.Open(filepath.Dir(path))
	if e != nil {
		return ErrIO
	}
	e = d.Sync()
	x = d.Close()
	if e != nil || x != nil {
		return ErrIO
	}
	return nil
}
