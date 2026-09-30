package githubmeta

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestRootCacheAndValidation(t *testing.T) {
	revision := strings.Repeat("a", 40)
	good := []byte(`{"sha":"` + revision + `","truncated":false,"tree":[{"path":"LICENSE","sha":"` + revision + `","mode":"100644","type":"blob"}]}`)
	cache := t.TempDir()
	calls := 0
	fetch := func(_ context.Context, path string) ([]byte, error) {
		calls++
		if path != "repos/a/b/git/trees/"+revision {
			t.Fatal("wrong immutable endpoint")
		}
		return good, nil
	}
	for i := 0; i < 2; i++ {
		b, e := Root(context.Background(), "a/b", revision, cache, fetch)
		if e != nil || string(b) != string(good) {
			t.Fatal("cache failure", e)
		}
	}
	if calls != 1 {
		t.Fatal("cache refetched")
	}
	entries, _ := os.ReadDir(cache)
	if len(entries) != 1 {
		t.Fatal("temporary files remained")
	}
	badCache := t.TempDir()
	bad := func(context.Context, string) ([]byte, error) { return []byte(`{"message":"not found"}`), nil }
	if _, e := Root(context.Background(), "a/b", revision, badCache, bad); e == nil {
		t.Fatal("bad API response accepted")
	}
	entries, _ = os.ReadDir(badCache)
	if len(entries) != 0 {
		t.Fatal("failed response cached")
	}
	failed := func(context.Context, string) ([]byte, error) { return nil, errors.New("failure") }
	if _, e := Root(context.Background(), "a/b", revision, badCache, failed); e == nil {
		t.Fatal("failure ignored")
	}
	if _, e := Root(context.Background(), "../a/b", revision, badCache, fetch); e == nil {
		t.Fatal("unsafe identity accepted")
	}
}
func TestResponseBound(t *testing.T) {
	b := boundedBuffer{limit: 3}
	if _, e := b.Write([]byte("abc")); e != nil {
		t.Fatal(e)
	}
	if _, e := b.Write([]byte("d")); e == nil || string(b.b) != "abc" {
		t.Fatal("unbounded API capture")
	}
}
