package githubmeta

import (
	"bytes"
	"context"
	"crypto/sha1" // Git's existing object identity, not a new security primitive.
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

const MaxLicenseBytes = 256 << 10

type BlobInfo struct {
	GitSHA, SHA256 string
	Bytes          int
}

func verifyText(b []byte, expected string) (BlobInfo, error) {
	r := BlobInfo{GitSHA: expected, Bytes: len(b)}
	if !revisionPattern.MatchString(expected) || len(b) == 0 || len(b) > MaxLicenseBytes || !utf8.Valid(b) || bytes.IndexByte(b, 0) >= 0 {
		return r, fmt.Errorf("invalid license text")
	}
	h := sha1.New()
	fmt.Fprintf(h, "blob %d%c", len(b), 0)
	h.Write(b)
	if hex.EncodeToString(h.Sum(nil)) != expected {
		return r, fmt.Errorf("Git blob identity mismatch")
	}
	s := sha256.Sum256(b)
	r.SHA256 = hex.EncodeToString(s[:])
	return r, nil
}
func decodeBlob(b []byte, expected string) ([]byte, BlobInfo, error) {
	var r struct {
		SHA, Encoding, Content string
		Size                   int
	}
	if len(b) > MaxResponseBytes {
		return nil, BlobInfo{}, fmt.Errorf("blob response too large")
	}
	if e := json.Unmarshal(b, &r); e != nil {
		return nil, BlobInfo{}, fmt.Errorf("invalid blob response")
	}
	if r.SHA != expected || r.Encoding != "base64" || r.Size < 1 || r.Size > MaxLicenseBytes {
		return nil, BlobInfo{}, fmt.Errorf("blob response contract")
	}
	content, e := base64.StdEncoding.DecodeString(r.Content)
	if e != nil || len(content) != r.Size {
		return nil, BlobInfo{}, fmt.Errorf("blob encoding or size")
	}
	info, e := verifyText(content, expected)
	return content, info, e
}
func Blob(ctx context.Context, repo, sha, cache string, fetch func(context.Context, string) ([]byte, error)) (BlobInfo, error) {
	if !repositoryPattern.MatchString(repo) || !revisionPattern.MatchString(sha) {
		return BlobInfo{}, fmt.Errorf("invalid blob identity")
	}
	path := filepath.Join(cache, sha+".txt")
	f, e := os.Open(path)
	if e == nil {
		defer f.Close()
		b, e := io.ReadAll(io.LimitReader(f, MaxLicenseBytes+1))
		if e != nil {
			return BlobInfo{}, e
		}
		return verifyText(b, sha)
	}
	if !os.IsNotExist(e) {
		return BlobInfo{}, e
	}
	response, e := fetch(ctx, "repos/"+repo+"/git/blobs/"+sha)
	if e != nil {
		return BlobInfo{}, e
	}
	content, info, e := decodeBlob(response, sha)
	if e != nil {
		return BlobInfo{}, e
	}
	if e = os.MkdirAll(cache, 0700); e != nil {
		return BlobInfo{}, e
	}
	f, e = os.CreateTemp(cache, ".blob-*.tmp")
	if e != nil {
		return BlobInfo{}, e
	}
	name := f.Name()
	defer os.Remove(name)
	if _, e = f.Write(content); e != nil {
		f.Close()
		return BlobInfo{}, e
	}
	if e = f.Close(); e != nil {
		return BlobInfo{}, e
	}
	if e = os.Rename(name, path); e != nil {
		return BlobInfo{}, e
	}
	return info, nil
}
func Tree(ctx context.Context, repo, sha, cache string, fetch func(context.Context, string) ([]byte, error)) ([]sweaudit.TreeEntry, error) {
	checked := func(ctx context.Context, endpoint string) ([]byte, error) {
		b, e := fetch(ctx, endpoint)
		if e != nil {
			return nil, e
		}
		got, _, e := sweaudit.TreeEntries(repo, b)
		if e != nil {
			return nil, e
		}
		if got != sha {
			return nil, fmt.Errorf("subtree identity mismatch")
		}
		return b, nil
	}
	b, e := Root(ctx, repo, sha, cache, checked)
	if e != nil {
		return nil, e
	}
	got, entries, e := sweaudit.TreeEntries(repo, b)
	if e != nil {
		return nil, e
	}
	if got != sha {
		return nil, fmt.Errorf("cached subtree identity mismatch")
	}
	return entries, nil
}
