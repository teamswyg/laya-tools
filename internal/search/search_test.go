package search

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func fixture(t testing.TB) string {
	t.Helper()
	d := t.TempDir()
	if err := exec.Command("git", "init", "-q", d).Run(); err != nil {
		t.Fatal(err)
	}
	for name, s := range map[string]string{"http.go": "package demo\n// Strip authorization on cross-host redirects.\nfunc redirectHeaders() {}\n", "other.go": "package demo\nfunc renderPage() {}\n", ".env": "SECRET=fixture\n", "credentials.json": "{\"fixture\":true}", "ignored.go": "package ignored\nfunc redirectHeaders() {}\n", ".gitignore": "ignored.go\n"} {
		if err := os.WriteFile(filepath.Join(d, name), []byte(s), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(d, "http.go"), filepath.Join(d, "linked.go")); err != nil {
		t.Fatal(err)
	}
	return d
}
func TestSearchBoundaries(t *testing.T) {
	idx, err := Load(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if idx.Files != 2 {
		t.Fatalf("included ignored/private/link file: %d", idx.Files)
	}
	r, err := idx.Search("cross host authorization", "", 8, 3, nil)
	if err != nil || len(r.Results) == 0 || r.Results[0].Path != "http.go" || r.Results[0].Start != 1 || r.Results[0].End != 3 {
		t.Fatalf("%+v %v", r, err)
	}
	r, err = idx.Search("존재하지않음", "", 8, 3, nil)
	if err != nil || len(r.Results) != 0 || len(r.Warnings) == 0 {
		t.Fatal("missing honest no-candidate result")
	}
}
func TestBudgets(t *testing.T) {
	idx, err := Load(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range [][2]int{{0, 1}, {65, 1}, {4, 5}, {4, -1}} {
		if _, err := idx.Search("redirect", "", p[0], p[1], nil); err == nil {
			t.Fatal("invalid budget accepted")
		}
	}
}
func BenchmarkLexicalSearch(b *testing.B) {
	idx, err := Load(fixture(b))
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := idx.Search("authorization redirects", "", 8, 3, nil); err != nil {
			b.Fatal(err)
		}
	}
}
