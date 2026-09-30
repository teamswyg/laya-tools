package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func pageRequest(t *testing.T, req request, identifiers bool, p pageOptions) (response, error) {
	t.Helper()
	b, e := json.Marshal(req)
	if e != nil {
		t.Fatal(e)
	}
	var out bytes.Buffer
	e = runPage(bytes.NewReader(b), &out, nil, "", identifiers, p)
	if e != nil {
		if out.Len() != 0 {
			t.Fatal("output before cursor validation")
		}
		return response{}, e
	}
	var r response
	if e = json.Unmarshal(out.Bytes(), &r); e != nil {
		t.Fatal(e)
	}
	return r, nil
}
func TestPagesReconstructCompleteOrdering(t *testing.T) {
	for _, identifiers := range []bool{false, true} {
		req := request{Snapshot: "s", Query: "read config", Documents: []document{{"a", "read"}, {"b", "config"}, {"c", "func ReadConfig() {}"}, {"d", "unknown"}, {"e", "!!!"}}}
		whole, e := pageRequest(t, req, identifiers, pageOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if whole.Page != nil {
			t.Fatal("default gained paging metadata")
		}
		cursor := ""
		var got []candidate
		digest := ""
		limit := 1
		for steps := 0; steps < 10; steps++ {
			page, e := pageRequest(t, req, identifiers, pageOptions{limit, cursor})
			if e != nil {
				t.Fatal(e)
			}
			if page.Page.Total != len(whole.Candidates) || page.Page.Offset != len(got) || page.Page.Returned != len(page.Candidates) {
				t.Fatal("bad page metadata")
			}
			if digest == "" {
				digest = page.Page.RankingSHA256
			}
			if digest != page.Page.RankingSHA256 {
				t.Fatal("digest changed")
			}
			got = append(got, page.Candidates...)
			cursor = page.Page.NextCursor
			limit = 2
			if cursor == "" {
				break
			}
		}
		if !reflect.DeepEqual(got, whole.Candidates) {
			t.Fatalf("lost/repeated/reordered candidates: %+v", got)
		}
	}
}
func TestCursorRejectsChangedRequestOrPolicy(t *testing.T) {
	req := request{Snapshot: "s", Query: "read", Documents: []document{{"a", "read"}, {"b", "other"}}}
	page, e := pageRequest(t, req, false, pageOptions{1, ""})
	if e != nil {
		t.Fatal(e)
	}
	for _, mutation := range []string{"snapshot", "query", "content", "ids", "order", "policy"} {
		changed := req
		changed.Documents = append([]document(nil), req.Documents...)
		identifiers := false
		switch mutation {
		case "snapshot":
			changed.Snapshot = "s2"
		case "query":
			changed.Query = "other"
		case "content":
			changed.Documents[1].Text = "OTHER"
		case "ids":
			changed.Documents[1].ID = "c"
		case "order":
			changed.Documents[0], changed.Documents[1] = changed.Documents[1], changed.Documents[0]
		case "policy":
			identifiers = true
		}
		if _, e := pageRequest(t, changed, identifiers, pageOptions{1, page.Page.NextCursor}); e == nil {
			t.Fatalf("accepted changed %s", mutation)
		}
	}
}
func TestInvalidPaging(t *testing.T) {
	req := request{Snapshot: "s", Query: "read", Documents: []document{{"a", "read"}, {"b", "other"}}}
	for _, p := range []pageOptions{{-1, ""}, {4097, ""}, {0, "anything"}, {1, strings.Repeat("a", 257)}, {1, "!notbase64"}, {1, base64.RawURLEncoding.EncodeToString([]byte(`{"v":1,"o":1,"h":"x","extra":true}`))}} {
		if _, e := pageRequest(t, req, false, p); e == nil {
			t.Fatalf("accepted %+v", p)
		}
	}
	page, e := pageRequest(t, req, false, pageOptions{1, ""})
	if e != nil {
		t.Fatal(e)
	}
	for _, offset := range []int{-1, 0, 2, 9999} {
		b, _ := json.Marshal(pageCursor{1, offset, page.Page.RankingSHA256})
		if _, e := pageRequest(t, req, false, pageOptions{1, base64.RawURLEncoding.EncodeToString(b)}); e == nil {
			t.Fatal("accepted invalid offset")
		}
	}
}
