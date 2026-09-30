package conalaaudit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConnectedOverlapAndDistinctQuestions(t *testing.T) {
	train := []Row{{1, " Read  File ", "read path x", "a"}, {1, "read file", "read path y", "b"}, {2, "different", "other", "c"}}
	test := []Row{{3, "READ FILE", "new", "d"}, {4, "unrelated", "new", "c"}, {5, "", "", " "}}
	r := Audit(train, test)
	if r.Combined.Rows != 6 || r.Combined.QuestionIDs != 5 || r.Combined.OriginalIntents != 3 || r.ConnectedGroups != 3 || r.LargestGroup != 3 || r.CrossSplitGroups != 2 || r.RowsInCrossSplitGroups != 5 || r.SharedQuestionIDs != 0 || r.SharedOriginalIntents != 1 || r.SharedSnippets != 1 {
		t.Fatalf("wrong audit: %+v", r)
	}
	if r.Combined.EmptyOriginal != 1 || r.Combined.EmptyRewrite != 1 || r.Combined.EmptySnippet != 1 || r.Meets2400OriginalIntents {
		t.Fatal("empty/count handling")
	}
}
func TestInternalIndentationPreserved(t *testing.T) {
	r := Audit([]Row{{1, "a", "a", "if x:\n a()"}}, []Row{{2, "b", "b", "if x:\n  a()"}})
	if r.SharedSnippets != 0 || r.ConnectedGroups != 2 {
		t.Fatal("collapsed Python indentation")
	}
}
func TestParserRejectsMalformed(t *testing.T) {
	for _, s := range []string{"", `{"question_id":0}`, `{"question_id":1,"extra":"x"}`, `{"question_id":1} {}`, strings.Repeat("x", 256<<10)} {
		if _, err := parse([]byte(s)); err == nil {
			t.Fatal("accepted malformed row")
		}
	}
	rows, err := parse([]byte("{\"question_id\":1,\"intent\":\"a\",\"rewritten_intent\":null,\"snippet\":\"x\"}\n"))
	if err != nil || len(rows) != 1 || rows[0].Rewrite != "" {
		t.Fatal("nullable rewrite", err)
	}
}

func TestTransitiveLinksAcrossDifferentKeys(t *testing.T) {
	r := Audit([]Row{{1, "a", "", "x"}, {1, "b", "", "y"}}, []Row{{2, "c", "", "y"}})
	if r.ConnectedGroups != 1 || r.LargestGroup != 3 || r.RowsInCrossSplitGroups != 3 {
		t.Fatal("transitive links lost", r)
	}
}

func TestReadRequiresPinnedHashAndBound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.jsonl")
	data := []byte(`{"question_id":1,"intent":"public synthetic","rewritten_intent":"","snippet":"x"}`)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(path, hash(data)); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(path, strings.Repeat("0", 64)); err == nil {
		t.Fatal("accepted hash mismatch")
	}
	large := []byte(strings.Repeat("x", (2<<20)+1))
	if err := os.WriteFile(path, large, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(path, hash(large)); err == nil {
		t.Fatal("accepted oversized file")
	}
}
