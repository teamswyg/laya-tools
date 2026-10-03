package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type cappedWriter struct {
	w      io.Writer
	cap    int64
	n      int64
	failed bool
}

func (w *cappedWriter) Write(b []byte) (int, error) {
	if w.failed || int64(len(b)) > w.cap-w.n {
		w.failed = true
		return 0, errCap
	}
	n, e := w.w.Write(b)
	w.n += int64(n)
	if e != nil || n != len(b) {
		w.failed = true
		if e == nil {
			e = io.ErrShortWrite
		}
	}
	return n, e
}

type cappedReader struct {
	r    io.Reader
	left int64
}

func (r *cappedReader) Read(b []byte) (int, error) {
	if r.left == 0 {
		var one [1]byte
		n, e := r.r.Read(one[:])
		if n != 0 {
			return 0, errCap
		}
		return 0, e
	}
	if int64(len(b)) > r.left {
		b = b[:r.left]
	}
	n, e := r.r.Read(b)
	r.left -= int64(n)
	return n, e
}

func archiveTarget(source, destination string, t target, capBytes int64) (pin, error) {
	if e := checkTree(source, t, false, false); e != nil {
		return pin{}, e
	}
	f, e := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return pin{}, e
	}
	// Failure never removes this file. Sync/Close are best effort on failure,
	// while the successful path explicitly requires both Sync calls to return.
	defer func() { f.Sync(); f.Close() }()
	cw := &cappedWriter{w: f, cap: capBytes}
	gz := gzip.NewWriter(cw)
	tw := tar.NewWriter(gz)
	for _, d := range orderedDirectories(t, false) {
		dm, e := modeOf(d.Mode)
		if e != nil {
			return pin{}, e
		}
		h := &tar.Header{Name: d.Relative + "/", Mode: int64(dm), Typeflag: tar.TypeDir, Format: tar.FormatPAX, ModTime: time.Unix(0, 0)}
		if e := tw.WriteHeader(h); e != nil {
			return pin{}, e
		}
	}
	for _, x := range t.Files {
		p := filepath.Join(source, filepath.FromSlash(x.Relative))
		in, before, e := openReadNoFollow(p)
		if e != nil {
			return pin{}, e
		}
		if !sameMeta(before, x.meta, true) || before.Size() != x.Bytes {
			in.Close()
			return pin{}, errChanged
		}
		fm, e := modeOf(x.Mode)
		if e != nil {
			in.Close()
			return pin{}, e
		}
		h := &tar.Header{Name: x.Relative, Mode: int64(fm), Size: x.Bytes, Typeflag: tar.TypeReg, Format: tar.FormatPAX, ModTime: time.Unix(0, 0)}
		if e := tw.WriteHeader(h); e != nil {
			in.Close()
			return pin{}, e
		}
		hash := sha256.New()
		n, e := io.CopyN(io.MultiWriter(tw, hash), in, x.Bytes)
		var one [1]byte
		more, tailErr := in.Read(one[:])
		after, statErr := in.Stat()
		in.Close()
		last, lstatErr := os.Lstat(p)
		if e != nil || n != x.Bytes || more != 0 || tailErr != io.EOF || statErr != nil || lstatErr != nil || !sameMeta(after, x.meta, true) || !sameMeta(last, x.meta, true) || hex.EncodeToString(hash.Sum(nil)) != x.SHA256 {
			return pin{}, errChanged
		}
	}
	if e := tw.Close(); e != nil {
		return pin{}, e
	}
	if e := gz.Close(); e != nil {
		return pin{}, e
	}
	if e := f.Sync(); e != nil {
		return pin{}, e
	}
	if e := f.Close(); e != nil {
		return pin{}, e
	}
	if e := syncDir(filepath.Dir(destination)); e != nil {
		return pin{}, e
	}
	// Detect file additions, deletions or metadata/content changes during packing.
	if e := checkTree(source, t, false, false); e != nil {
		return pin{}, e
	}
	got, _, e := hashRegular(destination, capBytes)
	return got, e
}

type zeroTail struct{}

func (zeroTail) Write(b []byte) (int, error) {
	for _, v := range b {
		if v != 0 {
			return 0, errArtifact
		}
	}
	return len(b), nil
}

// The decoder writes directly to final restored files: no decompressed tar/temp
// copy exists. The full restored regular-file payload peak is <=1 MiB, and
// inflated headers/padding separately stop at 2 MiB.
func restoreArchive(ap pin, destination string, t target) error {
	if e := os.Mkdir(destination, 0700); e != nil {
		return e
	}
	if e := syncDir(filepath.Dir(destination)); e != nil {
		return e
	}
	in, before, e := openReadNoFollow(ap.Path)
	if e != nil {
		return e
	}
	defer in.Close()
	if before.Size() != ap.Bytes || ap.Bytes > maxArchiveCap {
		return errChanged
	}
	br := bufio.NewReaderSize(in, 4096)
	gz, e := gzip.NewReader(br)
	if e != nil {
		return e
	}
	defer gz.Close()
	gz.Multistream(false)
	cr := &cappedReader{r: gz, left: inflateCap}
	tr := tar.NewReader(cr)
	fileMap := map[string]entry{}
	dirMap := map[string]directory{}
	for _, x := range t.Files {
		fileMap[x.Relative] = x
	}
	for _, x := range t.Directories {
		dirMap[x.Relative] = x
	}
	seen := map[string]bool{}
	payload := int64(0)
	for {
		h, e := tr.Next()
		if e == io.EOF {
			break
		}
		if e != nil {
			return e
		}
		isDir := h.Typeflag == tar.TypeDir
		rel := h.Name
		if isDir {
			rel = strings.TrimSuffix(rel, "/")
		}
		if !cleanRelative(rel) || seen[rel] || (h.Format != tar.FormatPAX && h.Format != tar.FormatUSTAR) || h.Uid != 0 || h.Gid != 0 || h.Linkname != "" || len(h.Xattrs) != 0 || h.ModTime.Unix() != 0 {
			return errArtifact
		}
		for k, v := range h.PAXRecords {
			if !((k == "path" && v == h.Name) || (k == "mtime" && v == "0")) {
				return errArtifact
			}
		}
		seen[rel] = true
		p := filepath.Join(destination, filepath.FromSlash(rel))
		if isDir {
			d, ok := dirMap[rel]
			dm, er := modeOf(d.Mode)
			if !ok || er != nil || h.Mode != int64(dm) || h.Size != 0 {
				return errArtifact
			}
			// Archive emits explicit parents before children; implicit extra
			// directories would hide a malformed archive and are not created.
			if e := noSymlinkAbsolute(filepath.Dir(p)); e != nil {
				return e
			}
			if e := os.Mkdir(p, 0700); e != nil {
				return e
			}
			if e := syncDir(filepath.Dir(p)); e != nil {
				return e
			}
			continue
		}
		x, ok := fileMap[rel]
		fm, er := modeOf(x.Mode)
		if !ok || er != nil || (h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeRegA) || h.Size != x.Bytes || h.Mode != int64(fm) || payload+x.Bytes > restoreCap {
			return errArtifact
		}
		if e := noSymlinkAbsolute(filepath.Dir(p)); e != nil {
			return e
		}
		f, e := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if e != nil {
			return e
		}
		hash := sha256.New()
		n, copyErr := io.CopyN(io.MultiWriter(f, hash), tr, x.Bytes)
		payload += n
		if copyErr == nil && n == x.Bytes && hex.EncodeToString(hash.Sum(nil)) == x.SHA256 {
			copyErr = f.Chmod(fm)
			if copyErr == nil {
				copyErr = f.Sync()
			}
		} else if copyErr == nil {
			copyErr = errChanged
		}
		closeErr := f.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if e := syncDir(filepath.Dir(p)); e != nil {
			return e
		}
	}
	if len(seen) != len(t.Files)+len(t.Directories) {
		return errArtifact
	}
	// tar EOF alone is insufficient. Drain through gzip EOF to validate CRC
	// and ISIZE, accept only bounded all-zero tar trailer, reject extra members.
	if _, e := io.Copy(zeroTail{}, cr); e != nil {
		return e
	}
	if _, e := br.ReadByte(); e != io.EOF {
		return errArtifact
	}
	if e := gz.Close(); e != nil {
		return e
	}
	after, e := in.Stat()
	m, me := infoMeta(before)
	last, le := os.Lstat(ap.Path)
	if e != nil || me != nil || le != nil || !sameMeta(after, m, true) || !sameMeta(last, m, true) {
		return errChanged
	}
	actual, _, e := hashRegular(ap.Path, maxArchiveCap)
	if e != nil || actual != ap {
		return errChanged
	}
	for _, d := range orderedDirectories(t, true) {
		dm, _ := modeOf(d.Mode)
		p := filepath.Join(destination, filepath.FromSlash(d.Relative))
		if e := os.Chmod(p, dm); e != nil {
			return e
		}
		if e := syncDir(p); e != nil {
			return e
		}
		if e := syncDir(filepath.Dir(p)); e != nil {
			return e
		}
	}
	rm, _ := modeOf(t.Root.Mode)
	if e := os.Chmod(destination, rm); e != nil {
		return e
	}
	if e := syncDir(destination); e != nil {
		return e
	}
	if e := syncDir(filepath.Dir(destination)); e != nil {
		return e
	}
	return checkTree(destination, t, true, false)
}
