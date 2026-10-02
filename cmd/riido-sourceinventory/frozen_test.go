// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/laya-tools/internal/sourceinventory"
)

// Added after the original observation. This checks archived bytes and raw
// spans only: no Rebind, AST, formatter, source-pin API or candidate execution.
func TestFrozenOriginalInventory60Integrity(t *testing.T) {
	repo := filepath.Join("..", "..")
	read := func(path string) []byte {
		t.Helper()
		b, e := os.ReadFile(filepath.Join(repo, path))
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	raw := read("experiments/short-claim/source-inventory-60.json")
	if len(raw) != 174651 || sha(raw) != "d99d635929555d4123bbaa84aac99d0fb29db345f8b813ed70074dd3e09cdbd4" {
		t.Fatal("archived original inventory changed")
	}
	var r record
	if decodeCanonical(raw, &r) != nil {
		t.Fatal("noncanonical original record")
	}
	planRaw := read(planPath)
	var p plan
	if len(planRaw) != 5643 || sha(planRaw) != "555dd524cfa5da5b0bc6eecbd9e13a21dfab2a7683978fa4b949765077779a4c" || decodeCanonical(planRaw, &p) != nil {
		t.Fatal("frozen plan changed")
	}
	if r.Schema != recordSchema || r.State != "metadata_rebound_content_review_pending" || r.FailureCode != "" || r.FailureCause != "" || r.PlanSHA != sha(planRaw) || r.InputCommit != "c0f279da4c2410699977b502ae4ca2fddd3d3ca8" || r.Invocation != (invocation{1, 1, true}) || r.VerifiedBlobs != 27 || r.ContentReview != "pending" || r.TrainingReady {
		t.Fatal("original execution receipt or pending status changed")
	}
	if p.SourceCommit != "cc8502f872b047e02961322ff1fcdce4af46e59c" || p.Binary.SHA256 != "831b73221b06dd757572521979e4f6fe0e15f1f5ccb4801dc0eddecaf05390c2" || p.Binary.Bytes != 5432642 || p.GOOS != "darwin" || p.GOARCH != "arm64" || p.Go != "go1.27.1" || p.Attempts != 1 || p.Retries != 0 {
		t.Fatal("original binary/source/platform provenance changed")
	}
	for _, pins := range [][]artifact{p.Inputs, p.CompileSource, p.Support} {
		for _, a := range pins {
			b := read(a.Path)
			if len(b) != a.Bytes || sha(b) != a.SHA256 {
				t.Fatal("frozen artifact changed", a.Path)
			}
		}
	}
	stored := read(inputPins()[4].Path)
	roots, e := historicalRoots(stored)
	if e != nil {
		t.Fatal(e)
	}
	in := sourceinventory.Input{GoVersion: p.Go, Roots: roots}
	pins := inputPins()
	for i := range in.Files {
		in.Files[i] = sourceinventory.File{Path: pins[i].Path, Raw: read(pins[i].Path), ExpectedSHA256: pins[i].SHA256}
	}
	if !completeScope(in, r.Inventory) || r.Inventory.ObjectBindingPerformed || r.Inventory.ClosureDiscoveryPerformed || r.Inventory.ContentReview != "pending" {
		t.Fatal("stored correspondence scope changed")
	}
	spans := 0
	check := func(ref sourceinventory.DeclarationReference) {
		t.Helper()
		var b []byte
		for _, f := range in.Files {
			if f.Path == ref.Path {
				b = f.Raw
				break
			}
		}
		if b == nil || ref.StartByte < 0 || ref.EndByte <= ref.StartByte || ref.EndByte > len(b) || sha(b[ref.StartByte:ref.EndByte]) != ref.RawSHA256 {
			t.Fatal("invalid raw declaration span", ref.ID, ref.Symbol)
		}
		if bytes.Count(b[:ref.StartByte], []byte{'\n'})+1 != ref.StartLine || bytes.Count(b[:ref.EndByte-1], []byte{'\n'})+1 != ref.EndLine {
			t.Fatal("physical line receipt changed", ref.Symbol)
		}
		spans++
	}
	for _, root := range r.Inventory.Roots {
		check(root.Root)
		for _, c := range root.Components {
			check(c)
		}
	}
	if spans != 264 {
		t.Fatal("relation count changed", spans)
	}
}
