// SPDX-License-Identifier: Apache-2.0
package roleplan

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/lexicalhint"
)

// Preserve preparation history, upstream bytes and full notices. This invokes
// no original behavior, graph assignment, inference or training API.
func TestFrozenPreparationArchives62And63(t *testing.T) {
	type pin struct {
		path  string
		bytes int
		sha   string
	}
	root := filepath.Join("..", "..")
	pins := [...]pin{
		{"role-preparation/PORT-LEDGER-62.v2.json", 5793, "c73fcdd8b9dcd0f66705cf35b51048cc07b63f1c488b4439c49d699c15a32d8a"},
		{"role-preparation/independent-findings-62.v2.json", 2854, "58a1dd7f51eaa90e4dbb56713fcb83f637aefe1c34e584c5deff6db79c46428f"},
		{"role-preparation/independent-ledger-62.v2.json", 1705, "e3efdd11b19f26a8192b661f6d0eb1e08fd564acf0fdc7882777176901b06a8e"},
		{"upstream-wording/quote-catalog-63.json", 204333, "b8253c07943dda0984c5befd32ef3233cf2b8426f5beb6324aea1c6520ecb9ec"},
		{"upstream-wording/quote-catalog-63.attempt1.json", 203944, "c65c6f5d59056694085f68e4b82ab9b311df054699c59ef73e709af372e1fd5d"},
		{"upstream-wording/PREPARATION-LEDGER-63.v2.json", 6842, "26ec8518451dc58c969ff7c455b849395d9053ddd32b8c0bdb738a6cc09ca92d"},
		{"upstream-wording/INDEPENDENT-QA-63.json", 5778, "9c8237341aa7a998639c3ba5718da968a7d29256969068d510a312c41b6163dc"},
		{"upstream-wording/INDEPENDENT-QA-LEDGER-63.json", 1286, "ef9f908f7863ebfdfecfe10935f358978f83480207faecb8abf77f06d13549ad"},
	}
	hash := func(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
	read := func(path string, size int, want string) []byte {
		t.Helper()
		f, err := os.Open(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		b, readErr := io.ReadAll(io.LimitReader(f, (1<<20)+1))
		closeErr := f.Close()
		if readErr != nil || closeErr != nil || len(b) > 1<<20 || len(b) != size || hash(b) != want {
			t.Fatal("frozen asset changed", path, readErr, closeErr)
		}
		return b
	}
	var catalogBytes []byte
	for _, p := range pins {
		b := read(filepath.Join("experiments", "short-claim", p.path), p.bytes, p.sha)
		if p.path == "upstream-wording/quote-catalog-63.json" {
			catalogBytes = b
		}
	}
	var catalog struct {
		Ready     bool `json:"training_ready"`
		Diversity bool `json:"authoring_diversity_cleared"`
		Quotes    []struct {
			ID        string `json:"id"`
			Path      string `json:"retained_path"`
			FileSHA   string `json:"file_sha256"`
			FileBytes int    `json:"file_bytes"`
			Text      string `json:"raw_text"`
			Words     int    `json:"normalized_words"`
			Human     bool   `json:"human_only_authorship_verified"`
			Input     bool   `json:"new_model_input"`
			Span      struct {
				Start int    `json:"byte_start"`
				End   int    `json:"byte_end"`
				SHA   string `json:"sha256"`
			} `json:"raw_quote_span"`
		} `json:"quotes"`
		Licenses []struct {
			Text  string `json:"raw_text"`
			SHA   string `json:"sha256"`
			Bytes int    `json:"bytes"`
		} `json:"original_full_mit_licenses"`
		Originals []struct {
			Path  string `json:"path"`
			SHA   string `json:"sha256"`
			Bytes int    `json:"bytes"`
		} `json:"retained_original_artifacts"`
	}
	if err := json.Unmarshal(catalogBytes, &catalog); err != nil {
		t.Fatal(err)
	}
	if catalog.Ready || catalog.Diversity || len(catalog.Quotes) != 25 || len(catalog.Licenses) != 2 || len(catalog.Originals) != 15 {
		t.Fatal("scope/readiness changed")
	}
	totalSource, totalQuotes, maxWords := 0, 0, 0
	for _, o := range catalog.Originals {
		read(o.Path, o.Bytes, o.SHA)
		totalSource += o.Bytes
	}
	for i, q := range catalog.Quotes {
		b := read(q.Path, q.FileBytes, q.FileSHA)
		if q.Span.Start < 0 || q.Span.End < q.Span.Start || q.Span.End > len(b) {
			t.Fatal("quote span outside source", q.ID)
		}
		raw := b[q.Span.Start:q.Span.End]
		words := len(strings.Fields(lexicalhint.NormalizeText(q.Text)))
		if string(raw) != q.Text || hash(raw) != q.Span.SHA || len(raw) > 512 || words != q.Words || words > 32 || q.Human || q.Input {
			t.Fatal("quote or origin claim changed", q.ID)
		}
		for _, prev := range catalog.Quotes[:i] {
			if prev.ID == q.ID {
				t.Fatal("duplicate quote identity")
			}
		}
		totalQuotes += len(raw)
		maxWords = max(maxWords, words)
	}
	for _, l := range catalog.Licenses {
		if len(l.Text) != l.Bytes || hash([]byte(l.Text)) != l.SHA {
			t.Fatal("full license notice changed")
		}
	}
	if totalSource != 99035 || totalQuotes != 1818 || maxWords != 26 {
		t.Fatal("original asset cardinality changed")
	}
}
