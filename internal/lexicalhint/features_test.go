package lexicalhint

import (
	"math"
	"reflect"
	"testing"
)

func TestIdentifiers(t *testing.T) {
	if got := tokens("parseHTTPResponse snake_case 한국어"); !reflect.DeepEqual(got, []string{"parse", "http", "response", "snake", "case", "한국어"}) {
		t.Fatal(got)
	}
}
func TestMatchAndAblation(t *testing.T) {
	q := "parse response"
	code := "def parse_response(data): return data"
	x := Features(q, code, false)
	if x[1] != 1 || x[7] != 1 {
		t.Fatal(x)
	}
	for i := 12; i < Dimension; i++ {
		if x[i] != 0 {
			t.Fatal("length ablation leaked")
		}
	}
	y := Features(q, code, true)
	for i := 0; i < 12; i++ {
		if x[i] != y[i] {
			t.Fatal("matching features changed")
		}
	}
	if y[12] <= 0 || y[13] <= 0 {
		t.Fatal("length features missing")
	}
	z := Features("zebra", code, false)
	if z[1] != 0 || z[7] != 0 {
		t.Fatal(z)
	}
}
func TestEmptyAndUnicode(t *testing.T) {
	for _, q := range []string{"", "!!", "한국어"} {
		for _, v := range Features(q, "", true) {
			if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
				t.Fatal(v)
			}
		}
	}
}

func BenchmarkFeatures(b *testing.B) {
	for b.Loop() {
		_ = Features("parse HTTP response headers", "func parseHTTPResponse(data string) Response { return decodeResponse(data) }", true)
	}
}
