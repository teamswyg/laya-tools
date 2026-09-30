package staticembed

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type reference struct {
	Schema       string `json:"schema"`
	ModelSHA     string `json:"model_sha256"`
	TokenizerSHA string `json:"tokenizer_sha256"`
	Cases        []struct {
		Text   string    `json:"text"`
		IDs    []int     `json:"ids"`
		Vector []float32 `json:"vector"`
	} `json:"cases"`
}

func TestPinnedReference(t *testing.T) {
	dir := os.Getenv("RIIDO_STATIC_MODEL_DIR")
	if dir == "" {
		t.Skip("set RIIDO_STATIC_MODEL_DIR for real asset parity")
	}
	m, e := Load(dir, "fp32")
	if e != nil {
		t.Fatal(e)
	}
	referencePath := os.Getenv("RIIDO_STATIC_REFERENCE")
	if referencePath == "" {
		referencePath = "../../experiments/static-embedding/reference.json"
	}
	b, e := os.ReadFile(referencePath)
	if e != nil {
		t.Fatal(e)
	}
	var r reference
	if e = json.Unmarshal(b, &r); e != nil {
		t.Fatal(e)
	}
	if r.Schema != "riido-static-reference-v1" || r.ModelSHA != ModelSHA || r.TokenizerSHA != VocabularySHA256 || len(r.Cases) < 30 {
		t.Fatal("invalid oracle")
	}
	largest := 0.
	for _, c := range r.Cases {
		v, ids, e := m.Encode(c.Text)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(ids, c.IDs) {
			t.Errorf("IDs %q: got %v want %v", c.Text, ids, c.IDs)
		}
		if len(c.Vector) != Dimension {
			t.Fatal("oracle dimension")
		}
		for j, want := range c.Vector {
			delta := math.Abs(float64(v[j] - want))
			largest = max(largest, delta)
			if delta > 2e-6 {
				t.Errorf("vector %q dimension%d error%g", c.Text, j, delta)
				break
			}
		}
	}
	t.Logf("%d cases max absolute error %.9g", len(r.Cases), largest)
	for _, mode := range []string{"int8", "ternary"} {
		q, e := Load(dir, mode)
		if e != nil {
			t.Fatal(e)
		}
		want := Rows*Dimension + Rows*4
		if mode == "ternary" {
			want = Rows*Dimension/4 + Rows*4
		}
		if q.TableBytes() != want {
			t.Fatal("unpacked storage")
		}
		var wg sync.WaitGroup
		for range 4 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				v, ids, e := q.Encode("parse HTTP response")
				if e != nil || len(ids) == 0 || math.Abs(Cosine(v, v)-1) > 1e-5 {
					t.Error("quantized encoding", e)
				}
			}()
		}
		wg.Wait()
	}
}
func TestBudget(t *testing.T) {
	m := &Model{}
	for _, s := range []string{strings.Repeat("a", MaxBytes+1), string([]byte{0xff})} {
		if _, e := m.Tokenize(s); e == nil {
			t.Fatal("unbounded input")
		}
	}
}
func TestVerifiedAssetRejectsCorruption(t *testing.T) {
	p := t.TempDir() + "/asset"
	if e := os.WriteFile(p, []byte("bad"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := readVerified(p, ModelSHA, 8); e == nil {
		t.Fatal("corrupt asset accepted")
	}
}
func TestWordPieceRollback(t *testing.T) {
	tok := tokenizer{unknown: 0, vocab: []token{{"##b", 2}, {"a", 1}}}
	if got := tok.segment("ab ax", nil); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatal(got)
	}
}

func TestPackedArithmetic(t *testing.T) {
	tok := tokenizer{unknown: -1, vocab: []token{{"x", 0}}}
	q := &Model{tokens: tok, mode: "int8", scales: []float32{.5}, int8s: make([]int8, Dimension)}
	q.int8s[0] = 3
	q.int8s[1] = -4
	v, _, e := q.Encode("x")
	if e != nil || math.Abs(float64(v[0])-.6) > 1e-6 || math.Abs(float64(v[1])+.8) > 1e-6 {
		t.Fatal(v, e)
	}
	q = &Model{tokens: tok, mode: "ternary", scales: []float32{.5}, ternary: make([]byte, Dimension/4)}
	q.ternary[0] = 9
	v, _, e = q.Encode("x")
	if e != nil || math.Abs(float64(v[0])-1/math.Sqrt(2)) > 1e-6 || math.Abs(float64(v[1])+1/math.Sqrt(2)) > 1e-6 {
		t.Fatal(v, e)
	}
	for _, x := range v[2:] {
		if x != 0 {
			t.Fatal("nonzero packed zero")
		}
	}
}
