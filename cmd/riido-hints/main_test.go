package main

import (
	"bytes"
	"encoding/json"
	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIdentifierHintsPreserveFirstAndAllCandidates(t *testing.T) {
	req := request{Snapshot: "public-fixture", Query: "read config", Documents: []document{{"a", "read"}, {"b", "config"}, {"c", "func ReadConfig() {}"}}}
	b, _ := json.Marshal(req)
	var out bytes.Buffer
	if e := runPolicy(bytes.NewReader(b), &out, nil, "", true); e != nil {
		t.Fatal(e)
	}
	var r response
	if e := json.Unmarshal(out.Bytes(), &r); e != nil {
		t.Fatal(e)
	}
	if r.Policy != "identifier_bm25_baseline_first" || r.Status != "unverified" || len(r.Candidates) != 3 || r.Candidates[0].ID != "a" || r.Candidates[1].ID != "c" || r.Candidates[2].ID != "b" {
		t.Fatalf("unexpected ranking %+v", r)
	}
	if strings.Contains(out.String(), "ReadConfig") {
		t.Fatal("source text leaked")
	}
}

func TestIdentifierHintsFallbackAndConflict(t *testing.T) {
	req := request{Snapshot: "s", Query: "!!!", Documents: []document{{"a", "one"}, {"b", "two"}}}
	b, _ := json.Marshal(req)
	var out bytes.Buffer
	if e := runPolicy(bytes.NewReader(b), &out, nil, "", true); e != nil {
		t.Fatal(e)
	}
	var r response
	json.Unmarshal(out.Bytes(), &r)
	if r.Policy != "bm25_identifier_input_out_of_scope" || len(r.Candidates) != 2 || r.Candidates[0].ID != "a" {
		t.Fatal(r)
	}
	req.Hints = &hints{Snapshot: "s", Query: "!!!", IDs: []string{"b"}}
	b, _ = json.Marshal(req)
	out.Reset()
	if e := runPolicy(bytes.NewReader(b), &out, nil, "", true); e == nil || out.Len() != 0 {
		t.Fatal("accepted conflicting hints")
	}
	if e := runPolicy(bytes.NewReader(b), &out, []float64{1}, "test", true); e == nil {
		t.Fatal("accepted model and identifiers")
	}
}

func TestUnverifiedCatalog(t *testing.T) {
	req := request{Snapshot: "s1", Query: "payment", Documents: []document{{"a", "payment retry"}, {"b", "cache expiry"}}, Hints: &hints{"s1", "payment", []string{"b"}}}
	data, _ := json.Marshal(req)
	var out bytes.Buffer
	if err := run(bytes.NewReader(data), &out); err != nil {
		t.Fatal(err)
	}
	var res response
	if err := json.Unmarshal(out.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Status != "unverified" || res.Candidates[0].ID != "b" || res.Candidates[1].ID != "a" {
		t.Fatal(res)
	}
	if strings.Contains(out.String(), "payment retry") {
		t.Fatal("document text leaked")
	}
}
func TestBadHints(t *testing.T) {
	for _, h := range []*hints{{"stale", "q", []string{"a"}}, {"s", "other", []string{"a"}}, {"s", "q", []string{"missing"}}, {"s", "q", []string{"a", "a"}}} {
		req := request{"s", "q", []document{{"a", "one"}, {"b", "two"}}, h}
		data, _ := json.Marshal(req)
		var out bytes.Buffer
		if err := run(bytes.NewReader(data), &out); err == nil {
			t.Fatal("bad hints accepted")
		}
		if out.Len() != 0 {
			t.Fatal("output before validation")
		}
	}
}
func TestStrictRequest(t *testing.T) {
	for _, s := range []string{`{} {}`, `{"unknown":1}`, `{"snapshot_id":"s","query":"q","documents":[{"id":"a","text":"a"},{"id":"a","text":"b"}]}`} {
		if err := run(strings.NewReader(s), &bytes.Buffer{}); err == nil {
			t.Fatal("bad request accepted")
		}
	}
}

func TestModelFallbackScope(t *testing.T) {
	req := request{Snapshot: "s", Query: strings.Repeat("word ", 65), Documents: []document{{"a", "word"}, {"b", "other"}}}
	data, _ := json.Marshal(req)
	var out bytes.Buffer
	if err := runModel(bytes.NewReader(data), &out, make([]float64, 8192), strings.Repeat("0", 64)); err != nil {
		t.Fatal(err)
	}
	var res response
	json.Unmarshal(out.Bytes(), &res)
	if res.Policy != "bm25_model_input_out_of_scope" || len(res.Candidates) != 2 {
		t.Fatal(res)
	}
	req.Query = "word"
	req.Hints = &hints{"s", "word", []string{"a"}}
	data, _ = json.Marshal(req)
	if err := runModel(bytes.NewReader(data), &bytes.Buffer{}, make([]float64, 8192), strings.Repeat("0", 64)); err == nil {
		t.Fatal("ambiguous hints allowed")
	}
}

func TestPinnedModelIntegrity(t *testing.T) {
	b, err := hintlearn.Encode(make([]float64, hintlearn.Dimension), "int8")
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "model.hbin")
	if err = os.WriteFile(file, b, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = loadModel(file, hintlearn.Hash(b)); err != nil {
		t.Fatal(err)
	}
	if _, err = loadModel(file, strings.Repeat("0", 64)); err == nil {
		t.Fatal("wrong hash accepted")
	}
	b[0] = 'X'
	os.WriteFile(file, b, 0600)
	if _, err = loadModel(file, hintlearn.Hash(b)); err == nil {
		t.Fatal("corrupt model accepted")
	}
}
