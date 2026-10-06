package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSourceRoleAndOverlapAreRejected(t *testing.T) {
	p := filepath.Join(t.TempDir(), "source.jsonl")
	body := []byte("{\"id\":\"original-a\",\"group_id\":\"original-a\",\"locale\":\"en\",\"intent\":\"question\",\"split\":\"test\",\"text\":\"Please explain this fact.\",\"origin\":\"original_authored_synthetic_development\"}\n")
	if e := os.WriteFile(p, body, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := rowsFrom(pin{Path: p, SHA: hash(body), Rows: 1}, "train"); e == nil {
		t.Fatal("final case accepted as train")
	}
	if e := disjoint([]row{{ID: "left", Group: "same"}}, []row{{ID: "right", Group: "same"}}); e == nil {
		t.Fatal("same incident allowed across partitions")
	}
}
