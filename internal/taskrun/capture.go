package taskrun

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"os"
)

// Each capture has exactly one os/exec copy goroutine. State is read only after
// Wait finishes those copy goroutines, so there is no hot-path mutex.
type captureWriter struct {
	file        *os.File
	hash        hash.Hash
	limit       int64
	bytes       int64
	retained    int64
	exceeded    bool
	writeFailed bool
	cancel      func()
}

func newCapture(path string, limit int64, cancel func()) (*captureWriter, error) {
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return nil, Error("capture_stage_failed")
	}
	return &captureWriter{file: f, hash: sha256.New(), limit: limit, cancel: cancel}, nil
}
func (w *captureWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.hash.Write(p)
	w.bytes += int64(n)
	retain := min(int64(n), max(int64(0), w.limit-w.retained))
	if retain > 0 {
		written, e := w.file.Write(p[:retain])
		w.retained += int64(written)
		if e != nil || int64(written) != retain {
			w.writeFailed = true
			w.cancel()
		}
	}
	if w.bytes > w.limit && !w.exceeded {
		w.exceeded = true
		w.cancel()
	}
	// Continue draining already-written pipe bytes after cancellation: the hash
	// covers the full stream observed by this executor, including its rejected
	// suffix. Only the bounded prefix is retained on disk.
	return n, nil
}
func (w *captureWriter) report() Capture {
	status := "captured"
	if w.writeFailed {
		status = "write_failed"
	} else if w.exceeded {
		status = "byte_limit"
	}
	return Capture{SHA256: hex.EncodeToString(w.hash.Sum(nil)), Bytes: w.bytes, RetainedBytes: w.retained, Complete: !w.exceeded && !w.writeFailed, Status: status}
}
