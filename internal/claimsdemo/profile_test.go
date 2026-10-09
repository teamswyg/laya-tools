// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package claimsdemo

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalProfilesProducePrivateGzipFilesAfterFinish(t *testing.T) {
	dir := t.TempDir()
	cpu, heap := filepath.Join(dir, "cpu.pprof"), filepath.Join(dir, "heap.pprof")
	p, err := startProfiles(cpu, heap)
	if err != nil {
		t.Fatal(err)
	}
	keep := make([]byte, 1<<20)
	for i := range keep {
		keep[i] = byte(i)
	}
	if err := p.finish(keep); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{cpu, heap} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0600 || info.Size() <= 0 || info.Size() > maxProfileBytes {
			t.Fatal("profile was not a bounded private file")
		}
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		z, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			t.Fatal("profile was not gzip")
		}
		_, readErr := io.Copy(io.Discard, z)
		z.Close()
		f.Close()
		if readErr != nil {
			t.Fatal("profile was incomplete")
		}
	}
}

func TestProfileCollisionAndSymlinkDoNotClobberExistingData(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "existing")
	want := []byte("keep existing data")
	if err := os.WriteFile(existing, want, 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(existing, link); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{existing, link} {
		if _, err := startProfiles(path, ""); err == nil || strings.Contains(err.Error(), path) {
			t.Fatal("collision accepted or path exposed")
		}
		got, err := os.ReadFile(existing)
		if err != nil || !bytes.Equal(got, want) {
			t.Fatal("existing data changed")
		}
	}
}

func TestProfileWriterStopsAtItsBudget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bounded")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	w := &profileFile{file: f, remaining: 5}
	if n, err := w.Write([]byte("123456789")); n != 5 || err != io.ErrShortWrite {
		t.Fatal("profile budget not enforced")
	}
	if n, err := w.Write([]byte("extra")); n != 0 || err == nil {
		t.Fatal("failed profile continued writing")
	}
	f.Close()
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "12345" {
		t.Fatal("profile exceeded budget")
	}
}
