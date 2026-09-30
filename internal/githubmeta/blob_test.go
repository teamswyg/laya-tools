package githubmeta

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func blobFixture(content []byte) (string, []byte) {
	h := sha1.New()
	fmt.Fprintf(h, "blob %d%c", len(content), 0)
	h.Write(content)
	sha := hex.EncodeToString(h.Sum(nil))
	b, _ := json.Marshal(struct {
		SHA, Encoding, Content string
		Size                   int
	}{sha, "base64", base64.StdEncoding.EncodeToString(content), len(content)})
	return sha, b
}
func TestVerifiedBlobCache(t *testing.T) {
	sha, response := blobFixture([]byte("Original license fixture\n"))
	cache := t.TempDir()
	calls := 0
	fetch := func(context.Context, string) ([]byte, error) { calls++; return response, nil }
	for i := 0; i < 2; i++ {
		r, e := Blob(context.Background(), "a/b", sha, cache, fetch)
		if e != nil || r.GitSHA != sha || r.Bytes != 25 || len(r.SHA256) != 64 {
			t.Fatalf("bad blob %+v %v", r, e)
		}
	}
	if calls != 1 {
		t.Fatal("cache fetched again")
	}
	if e := os.WriteFile(filepath.Join(cache, sha+".txt"), []byte("corrupt"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := Blob(context.Background(), "a/b", sha, cache, fetch); e == nil {
		t.Fatal("corrupt cache accepted")
	}
	if _, _, e := decodeBlob(response, strings.Repeat("0", 40)); e == nil {
		t.Fatal("wrong expected identity accepted")
	}
	var altered map[string]any
	json.Unmarshal(response, &altered)
	altered["Size"] = 1
	b, _ := json.Marshal(altered)
	if _, _, e := decodeBlob(b, sha); e == nil {
		t.Fatal("wrong declared size accepted")
	}
	for _, content := range [][]byte{{0}, {0xff}, []byte(strings.Repeat("x", MaxLicenseBytes+1))} {
		badSHA, bad := blobFixture(content)
		if _, _, e := decodeBlob(bad, badSHA); e == nil {
			t.Fatal("invalid text accepted")
		}
	}
}
func TestSubtreeIdentity(t *testing.T) {
	sha := strings.Repeat("a", 40)
	b := []byte(`{"sha":"` + strings.Repeat("b", 40) + `","tree":[{"path":"LICENSE","mode":"100644","type":"blob","sha":"` + sha + `"}]}`)
	cache := t.TempDir()
	fetch := func(context.Context, string) ([]byte, error) { return b, nil }
	if _, e := Tree(context.Background(), "a/b", sha, cache, fetch); e == nil {
		t.Fatal("subtree identity mismatch accepted")
	}
	files, _ := os.ReadDir(cache)
	if len(files) != 0 {
		t.Fatal("wrong subtree cached")
	}
}
