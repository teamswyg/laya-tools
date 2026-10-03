package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func canonicalTemp(t *testing.T) string {
	t.Helper()
	p, e := filepath.EvalSymlinks(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	return p
}

// All controls below use fresh tiny owned fake directories. No HF33 source,
// original packages, models, protected data, helper CLI or network is invoked.
func tinyTarget(t *testing.T) (string, target) {
	t.Helper()
	root := filepath.Join(canonicalTemp(t), "source")
	if e := os.Mkdir(root, 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.Mkdir(filepath.Join(root, "nested"), 0700); e != nil {
		t.Fatal(e)
	}
	for path, data := range map[string][]byte{"empty": {}, "nested/data": []byte("abc")} {
		if e := os.WriteFile(filepath.Join(root, path), data, 0600); e != nil {
			t.Fatal(e)
		}
	}
	dInfo, e := os.Lstat(filepath.Join(root, "nested"))
	if e != nil {
		t.Fatal(e)
	}
	dm, e := infoMeta(dInfo)
	if e != nil {
		t.Fatal(e)
	}
	out := target{ID: "owned-fake", Directories: []directory{{Relative: "nested", meta: dm}}}
	for _, rel := range []string{"empty", "nested/data"} {
		p, fi, e := hashRegular(filepath.Join(root, rel), restoreCap)
		if e != nil {
			t.Fatal(e)
		}
		m, e := infoMeta(fi)
		if e != nil {
			t.Fatal(e)
		}
		out.Files = append(out.Files, entry{Relative: rel, Bytes: p.Bytes, SHA256: p.SHA256, meta: m})
	}
	ri, e := os.Lstat(root)
	if e != nil {
		t.Fatal(e)
	}
	out.Root, e = infoMeta(ri)
	if e != nil {
		t.Fatal(e)
	}
	return root, out
}

func TestCanonicalRejectsAliasDuplicateAndMissing(t *testing.T) {
	type own struct {
		A int  `json:"a"`
		B bool `json:"b"`
	}
	b, e := canonical(own{1, true})
	if e != nil {
		t.Fatal(e)
	}
	var x own
	if e := decodeCanonical(b, &x); e != nil || x.A != 1 {
		t.Fatal("canonical rejected")
	}
	for _, bad := range []string{
		"{\"a\":1,\"a\":1,\"b\":true}\n",
		"{\"A\":1,\"b\":true}\n",
		"{\"a\":1}\n",
		"{\"a\":1,\"b\":true,\"extra\":0}\n",
	} {
		if e := decodeCanonical([]byte(bad), &x); e == nil {
			t.Fatal("noncanonical accepted")
		}
	}
}

func TestBoundedWriterStickyAndExactCap(t *testing.T) {
	var b bytes.Buffer
	w := &cappedWriter{w: &b, cap: 3}
	if n, e := w.Write([]byte("abc")); e != nil || n != 3 {
		t.Fatal(n, e)
	}
	if n, e := w.Write([]byte("d")); e == nil || n != 0 || b.String() != "abc" {
		t.Fatal("overflow wrote")
	}
	if n, e := w.Write(nil); e == nil || n != 0 {
		t.Fatal("failure not sticky")
	}
}

func TestFullPhysicalRestoreIncludesEmptyFileAndMode(t *testing.T) {
	root, x := tinyTarget(t)
	archive := filepath.Join(canonicalTemp(t), "tiny.tar.gz")
	p, e := archiveTarget(root, archive, x, 65536)
	if e != nil {
		t.Fatal(e)
	}
	restore := filepath.Join(canonicalTemp(t), "restored")
	if e := restoreArchive(p, restore, x); e != nil {
		t.Fatal(e)
	}
	empty, e := os.Stat(filepath.Join(restore, "empty"))
	if e != nil || empty.Size() != 0 || empty.Mode().Perm() != 0600 {
		t.Fatal("empty file lost")
	}
	if e := os.Chmod(filepath.Join(restore, "nested/data"), 0644); e != nil {
		t.Fatal(e)
	}
	if e := checkTree(restore, x, true, false); e == nil {
		t.Fatal("mode change accepted")
	}
}

func TestCRCAndExtraGzipMemberRejected(t *testing.T) {
	for _, kind := range []string{"crc", "member"} {
		t.Run(kind, func(t *testing.T) {
			root, x := tinyTarget(t)
			path := filepath.Join(canonicalTemp(t), "tiny.tar.gz")
			p, e := archiveTarget(root, path, x, 65536)
			if e != nil {
				t.Fatal(e)
			}
			b, e := os.ReadFile(path)
			if e != nil {
				t.Fatal(e)
			}
			if kind == "crc" {
				b[len(b)-8] ^= 1
			} else {
				var extra bytes.Buffer
				gz := gzip.NewWriter(&extra)
				if e := gz.Close(); e != nil {
					t.Fatal(e)
				}
				b = append(b, extra.Bytes()...)
			}
			if e := os.WriteFile(path, b, 0600); e != nil {
				t.Fatal(e)
			}
			p, _, e = hashRegular(path, 65536)
			if e != nil {
				t.Fatal(e)
			}
			restore := filepath.Join(canonicalTemp(t), "restored")
			if e := restoreArchive(p, restore, x); e == nil {
				t.Fatal("invalid archive accepted")
			}
			if _, e := os.Stat(filepath.Join(root, "nested/data")); e != nil {
				t.Fatal("original removed")
			}
			if _, e := os.Lstat(restore); e != nil {
				t.Fatal("failed restore discarded")
			}
		})
	}
}

func TestSourceChangeRejectsBeforeArchive(t *testing.T) {
	root, x := tinyTarget(t)
	if e := os.WriteFile(filepath.Join(root, "nested/data"), []byte("xyz"), 0600); e != nil {
		t.Fatal(e)
	}
	destination := filepath.Join(canonicalTemp(t), "new.tar.gz")
	if _, e := archiveTarget(root, destination, x, 65536); e == nil {
		t.Fatal("changed source accepted")
	}
	if _, e := os.Lstat(destination); !os.IsNotExist(e) {
		t.Fatal("archive created before source guard")
	}
}

func TestTraversalDuplicateAndLinksCannotEnterInventory(t *testing.T) {
	for _, s := range []string{"../x", "a/../x", "/x", "a//x", "a\\x", "."} {
		if cleanRelative(s) {
			t.Fatal("unsafe relative path", s)
		}
	}
	root, x := tinyTarget(t)
	if e := os.Symlink("nested/data", filepath.Join(root, "extra")); e != nil {
		t.Fatal(e)
	}
	if e := checkTree(root, x, false, false); e == nil {
		t.Fatal("symlink accepted")
	}
	if e := os.Remove(filepath.Join(root, "extra")); e != nil {
		t.Fatal(e)
	}
	if e := os.Link(filepath.Join(root, "nested/data"), filepath.Join(root, "extra")); e != nil {
		t.Fatal(e)
	}
	if _, _, e := hashRegular(filepath.Join(root, "nested/data"), restoreCap); e == nil {
		t.Fatal("hard link accepted")
	}
}

func TestPartialOriginalCleanupPreservesPrefixAndNoSuccess(t *testing.T) {
	root, x := tinyTarget(t)
	out := canonicalTemp(t)
	j, e := os.OpenFile(filepath.Join(out, "journal"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer j.Close()
	if e := os.WriteFile(filepath.Join(root, markerName), []byte("owned fake marker\n"), 0600); e != nil {
		t.Fatal(e)
	}
	injected := errors.New("owned injected failure")
	s := &session{out: out, journal: j, previous: strings.Repeat("0", 64),
		result: outcome{Schema: "owned-fake"}, fault: func(stage string, ti, fi int) error {
			if stage == "before_remove" && fi == 1 {
				return injected
			}
			return nil
		}}
	if e := removeSelectedTree(root, x, true, s, 0); !errors.Is(e, injected) {
		t.Fatal(e)
	}
	if _, e := os.Lstat(filepath.Join(root, "empty")); !os.IsNotExist(e) {
		t.Fatal("first file not removed")
	}
	if _, e := os.Lstat(filepath.Join(root, "nested/data")); e != nil {
		t.Fatal("remaining original lost")
	}
	if _, e := os.Lstat(filepath.Join(root, markerName)); e != nil {
		t.Fatal("recovery marker lost")
	}
	if s.result.RemovedACKFiles != 1 || s.result.Complete || s.result.CompletedTargets != 0 || s.result.ReclaimedLogicalBytes != nil {
		t.Fatal("partial became complete")
	}
}

func TestPeakBoundUsesRetainedArchiveAndSharedControls(t *testing.T) {
	// This exercises the source gate without creating admission or census data.
	small := int64(512 << 10)
	base := globalCap - small - restoreCap - controlCap
	cases := []struct {
		name    string
		current int64
		caps    []int64
		accept  bool
	}{
		{"exact-first-peak", base, []int64{small, small}, true},
		{"one-byte-over", base + 1, []int64{small, small}, false},
		{"later-archive-retained", base, []int64{small, maxArchiveCap}, false},
		{"missing-pair", base, []int64{small}, false},
		{"two-MiB-with-two-point-five-head", globalCap - int64(5<<19), []int64{maxArchiveCap, small}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if accepted := checkPeak(c.current, c.caps) == nil; accepted != c.accept {
				t.Fatal("peak gate", accepted)
			}
		})
	}
}

func TestLargeFrozenMetadataCap(t *testing.T) {
	p := filepath.Join(canonicalTemp(t), "owned-fake-metadata")
	raw := bytes.Repeat([]byte("x"), 170774)
	if e := os.WriteFile(p, raw, 0600); e != nil {
		t.Fatal(e)
	}
	got, _, e := hashRegular(p, metadataCap)
	if e != nil {
		t.Fatal(e)
	}
	b, e := readPin(got, metadataCap)
	if e != nil || !bytes.Equal(b, raw) {
		t.Fatal("bounded metadata rejected")
	}
	tooLarge := got
	tooLarge.Bytes = metadataCap + 1
	if _, e := readPin(tooLarge, metadataCap); e == nil {
		t.Fatal("metadata cap bypass")
	}
}

type syntheticSpecialInfo struct {
	os.FileInfo
	special os.FileMode
}

func (s syntheticSpecialInfo) Mode() os.FileMode { return s.FileInfo.Mode() | s.special }
func TestSpecialPermissionBitsRejected(t *testing.T) {
	root, x := tinyTarget(t)
	info, e := os.Lstat(root)
	if e != nil {
		t.Fatal(e)
	}
	for _, bit := range []os.FileMode{os.ModeSetuid, os.ModeSetgid, os.ModeSticky} {
		if _, e := infoMeta(syntheticSpecialInfo{info, bit}); e == nil {
			t.Fatal("special bits lost in metadata")
		}
	}
	if e := os.Chmod(root, 0700|os.ModeSticky); e != nil {
		t.Fatal(e)
	}
	if e := checkTree(root, x, true, false); e == nil {
		t.Fatal("restored root special mode accepted")
	}
	if e := os.Chmod(root, 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.Chmod(filepath.Join(root, "nested"), 0700|os.ModeSticky); e != nil {
		t.Fatal(e)
	}
	if e := checkTree(root, x, true, false); e == nil {
		t.Fatal("restored directory special mode accepted")
	}
}

func TestAuthorizedMarkerRootMetadataAllowance(t *testing.T) {
	want := meta{Mode: "0700", Device: "1", Inode: "2", Nlink: 4, MtimeNS: "3"}
	for _, test := range []struct {
		n      uint64
		marker bool
		accept bool
	}{{4, false, true}, {5, false, false}, {4, true, true}, {5, true, true}, {6, true, false}, {3, true, false}} {
		got := want
		got.Nlink = test.n
		if sameRootMeta(got, want, test.marker) != test.accept {
			t.Fatal("marker-specific nlink guard")
		}
	}
	got := want
	got.Inode = "different"
	if sameRootMeta(got, want, true) {
		t.Fatal("marker allowed replacement inode")
	}
	max := want
	max.Nlink = ^uint64(0)
	got = max
	got.Nlink = 0
	if sameRootMeta(got, max, true) {
		t.Fatal("link count overflow")
	}
	root, x := tinyTarget(t)
	if e := checkTree(root, x, false, true); e == nil {
		t.Fatal("marker missing but allowance admitted")
	}
}
