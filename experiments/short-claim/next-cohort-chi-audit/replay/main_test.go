// SPDX-License-Identifier: Apache-2.0
package main

import (
	"io"
	"strings"
	"testing"
)

// Hiding WriterTo exercises io.Copy's destination path. Embedding bytes.Buffer
// previously exposed ReaderFrom there and bypassed the writer's output limit.
func TestOutputBoundCopyPaths(t *testing.T) {
	for _, writerTo := range []bool{false, true} {
		for _, size := range []int{8, 9, 32768} {
			var src io.Reader = strings.NewReader(strings.Repeat("x", size))
			if !writerTo {
				src = struct{ io.Reader }{src}
			}
			dst := &bounded{max: 8}
			if _, ok := any(dst).(io.ReaderFrom); ok {
				t.Fatal("bounded output unexpectedly implements ReaderFrom")
			}
			n, err := io.Copy(dst, src)
			if n > 8 || len(dst.Bytes()) > 8 {
				t.Fatalf("output cap bypassed: accepted=%d retained=%d", n, len(dst.Bytes()))
			}
			if size == 8 {
				if err != nil || n != 8 || string(dst.Bytes()) != "xxxxxxxx" {
					t.Fatalf("exact fit failed: accepted=%d error=%v", n, err)
				}
			} else if err == nil {
				t.Fatal("oversize output accepted")
			}
		}
	}
}
