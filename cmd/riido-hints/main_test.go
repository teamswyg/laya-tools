package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

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
