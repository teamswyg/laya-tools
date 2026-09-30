package alignment

import (
	"github.com/teamswyg/laya-tools/internal/lexicalhint"
	"github.com/teamswyg/laya-tools/internal/paireval"
	"github.com/teamswyg/laya-tools/internal/staticembed"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestLexicalCompatibility(t *testing.T) {
	q, c := "parse HTTP response", "def parse_response(x): return x"
	x, e := Features(nil, "lexical", q, c)
	if e != nil {
		t.Fatal(e)
	}
	want := lexicalhint.Features(q, c, false)
	if !reflect.DeepEqual(x, want[:]) {
		t.Fatal("control changed")
	}
	if lexicalhint.NormalizeText("parseHTTPResponse snake_case") != "parse http response snake case" {
		t.Fatal("normalization")
	}
}
func TestBoundary(t *testing.T) {
	if _, e := Features(nil, "normalized_alignment", "a", "b"); e == nil {
		t.Fatal("missing encoder")
	}
	if _, e := Features(nil, "lexical", strings.Repeat("a ", 65), "b"); e == nil {
		t.Fatal("scope exceeded")
	}
	if _, e := Prepare(nil, paireval.Split{}, "reserve1", "lexical", nil); e == nil {
		t.Fatal("holdout accepted")
	}
}
func TestHoldoutIsolation(t *testing.T) {
	one := 1
	rows := []paireval.Row{{Query: "a", Code: "a", Label: &one}, {Query: "b", Code: "b", Label: &one}}
	s := paireval.Split{Rows: []paireval.Membership{{Group: 1, Split: "development"}, {Group: 2, Split: "reserve1"}}}
	x, e := Prepare(rows, s, "development", "lexical", nil)
	if e != nil {
		t.Fatal(e)
	}
	rows[1] = paireval.Row{}
	y, e := Prepare(rows, s, "development", "lexical", nil)
	if e != nil || !reflect.DeepEqual(x, y) {
		t.Fatal("holdout changed training", e)
	}
}

func TestActualEmbeddingFeatureContract(t *testing.T) {
	dir := os.Getenv("RIIDO_STATIC_MODEL_DIR")
	if dir == "" {
		t.Skip("real pinned asset required")
	}
	m, e := staticembed.Load(dir, "ternary")
	if e != nil {
		t.Fatal(e)
	}
	x, e := Features(m, "normalized_alignment", "parse HTTP", "def parse_http(x): return x")
	if e != nil {
		t.Fatal(e)
	}
	if len(x) != 145 {
		t.Fatal("width")
	}
	sum := 0.
	for _, v := range x[17:81] {
		sum += v / 8
	}
	if math.Abs(sum-x[16]) > 1e-12 {
		t.Fatal("cosine and alignment products disagree")
	}
	y, e := Features(m, "normalized_cosine", "parse HTTP", "def parse_http(x): return x")
	if e != nil || !reflect.DeepEqual(x[:17], y) {
		t.Fatal("variant prefix mismatch", e)
	}
	if _, e = Features(m, "normalized_alignment", strings.Repeat("aB", 2048), "x"); e == nil {
		t.Fatal("normalized input silently truncated")
	}
}
