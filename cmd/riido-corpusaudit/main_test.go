package main

import (
	"strings"
	"testing"
)

func TestConnectedLeakageAndConflicts(t *testing.T) {
	s := `[{"code":"a","doc":" Q one ","idx":"1","label":1},{"code":"b","doc":"q ONE","idx":"2","label":0},{"code":"b","doc":"q two","idx":"3","label":1},{"code":"a","doc":"q one","idx":"4","label":0}]`
	r, e := audit(strings.NewReader(s))
	if e != nil {
		t.Fatal(e)
	}
	if r.Rows != 4 || r.UniqueQueries != 2 || r.UniqueCodes != 2 || r.ConnectedGroups != 1 || r.LargestGroup != 4 || r.DuplicatePairs != 1 || r.ConflictingPairs != 1 {
		t.Fatal(r)
	}
}
func TestInvalidRows(t *testing.T) {
	for _, s := range []string{`{}`, `[{"code":"a","doc":"q","label":2}]`, `[] {}`, `[{"code":"a","doc":"q"}]`, `[{"code":"a","doc":"q","label":1,"repo":"unknown"}]`} {
		if _, e := audit(strings.NewReader(s)); e == nil {
			t.Fatal("invalid accepted")
		}
	}
}
