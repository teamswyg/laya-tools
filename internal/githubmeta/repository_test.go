package githubmeta

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func repoFixture(extra string) []byte {
	return []byte(`{"id":123,"full_name":"New/Name","private":false,"visibility":"public","fork":false` + extra + `}`)
}
func TestRepositoryAliasesAndCache(t *testing.T) {
	cache := t.TempDir()
	calls := 0
	fetch := func(_ context.Context, endpoint string) ([]byte, error) {
		calls++
		if endpoint != "repos/Old/Name" {
			t.Fatal(endpoint)
		}
		return repoFixture(`,"license":{"spdx_id":"Apache-2.0"}`), nil
	}
	r, e := Repository(context.Background(), "Old/Name", cache, fetch)
	if e != nil || r.Canonical != "new/name" || r.ID != 123 || r.NetworkID != 123 || r.LicenseHint != "Apache-2.0" {
		t.Fatalf("%+v %v", r, e)
	}
	again, e := Repository(context.Background(), "old/name", cache, func(context.Context, string) ([]byte, error) { return nil, fmt.Errorf("offline") })
	if e != nil || again.ResponseSHA256 != r.ResponseSHA256 || calls != 1 {
		t.Fatal("cache replay")
	}
	files, e := os.ReadDir(cache)
	if e != nil || len(files) != 1 {
		t.Fatal(e)
	}
	p := filepath.Join(cache, files[0].Name())
	info, _ := os.Stat(p)
	if info.Mode().Perm() != 0600 {
		t.Fatal("permissions")
	}
	if e = os.WriteFile(p, []byte(`{}`), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = Repository(context.Background(), "old/name", cache, fetch); e == nil || calls != 1 {
		t.Fatal("corrupt cache silently refreshed")
	}
}
func TestRepositoryRejectsPrivateMalformedAndForkWithoutAncestry(t *testing.T) {
	valid := string(repoFixture(""))
	for _, s := range []string{
		`{}`, strings.Replace(valid, `"private":false`, `"private":true`, 1),
		strings.Replace(valid, `"visibility":"public"`, `"visibility":"internal"`, 1),
		strings.Replace(valid, `"id":123`, `"id":0`, 1),
		strings.Replace(valid, `"fork":false`, `"fork":true`, 1),
		strings.Repeat("x", MaxResponseBytes+1),
	} {
		if _, e := decodeRepository("old/name", []byte(s)); e == nil {
			t.Fatal("invalid public identity accepted")
		}
	}
	fork := []byte(`{"id":2,"full_name":"copy/repo","private":false,"visibility":"public","fork":true,"parent":{"id":3,"full_name":"parent/repo","private":false},"source":{"id":1,"full_name":"origin/repo","private":false}}`)
	r, e := decodeRepository("copy/repo", fork)
	if e != nil || r.NetworkID != 1 || !r.Fork {
		t.Fatal(e)
	}
	bad := strings.Replace(string(fork), `"full_name":"origin/repo","private":false`, `"full_name":"origin/repo","private":true`, 1)
	if _, e = decodeRepository("copy/repo", []byte(bad)); e == nil {
		t.Fatal("private fork source")
	}
}
