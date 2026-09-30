package hintlearn

import (
	"hash/fnv"
	"math"
	"reflect"
	"testing"
)

func TestFeatureOrderDistinguishesOpposites(t *testing.T) {
	a := Features("retain active entries", "cache keeps active entries and removes inactive entries")
	b := Features("retain active entries", "cache removes active entries and keeps inactive entries")
	if reflect.DeepEqual(a, b) {
		t.Fatal("opposing order collapsed")
	}
}
func TestPackedRoundtripAndCorruption(t *testing.T) {
	w := make([]float64, Dimension)
	for i := range w {
		w[i] = math.Sin(float64(i))
	}
	for _, mode := range []string{"fp32", "int8", "ternary_ptq", "ternary_ste"} {
		q := Quantize(w, mode)
		b, e := Encode(q, mode)
		if e != nil {
			t.Fatal(e)
		}
		r, e := Decode(b)
		if e != nil || !reflect.DeepEqual(q, r) {
			t.Fatalf("%s roundtrip: %v", mode, e)
		}
		if _, e = Decode(b[:len(b)-1]); e == nil {
			t.Fatal("truncation accepted")
		}
		b[13] = 1
		if _, e = Decode(b); e == nil {
			t.Fatal("reserved header accepted")
		}
	}
	b, _ := Encode(make([]float64, Dimension), "ternary_ptq")
	if _, e := Decode(b); e != nil {
		t.Fatal(e)
	}
	b[24] = 1
	if _, e := Decode(b); e == nil {
		t.Fatal("bad presence bitmap accepted")
	}
}
func TestFamilyLeakAndTrainingImprovement(t *testing.T) {
	train := Dataset{Split: "train", Documents: []Document{{"a", "cache", "cache keeps active entries"}, {"b", "cache", "cache removes active entries"}}, Queries: []Query{{"qa", "cache", "cache retain active entries", "a"}, {"qb", "cache", "cache discard active entries", "b"}}}
	val := Dataset{Split: "validation", Documents: []Document{{"c", "session", "session keeps active sessions"}, {"d", "session", "session removes active sessions"}}, Queries: []Query{{"qc", "session", "session retain active sessions", "c"}, {"qd", "session", "session discard active sessions", "d"}}}
	before := Evaluate(make([]float64, Dimension), val)
	fit, e := Fit(train, val, Config{1729, .2, "fp32", 60})
	if e != nil {
		t.Fatal(e)
	}
	if fit.Validation.NLL >= before.NLL || fit.Train.Recall1 != 1 {
		t.Fatal("training did not learn", fit.Validation, fit.Train)
	}
	val.Documents[0].Family = "cache"
	if _, e = Fit(train, val, Config{1729, .2, "fp32", 1}); e == nil {
		t.Fatal("family leak allowed")
	}
	val.Split = "final"
	if _, e = Fit(train, val, Config{1729, .2, "fp32", 1}); e == nil {
		t.Fatal("final used for selection")
	}
}

func TestFeatureHashIncludesUTF8Bytes(t *testing.T) {
	for _, text := range []string{"ascii", "한글", "query\x00document"} {
		h := fnv.New64a()
		h.Write([]byte(text))
		if hash(text) != h.Sum64() {
			t.Fatal("UTF-8 bytes omitted")
		}
	}
}
