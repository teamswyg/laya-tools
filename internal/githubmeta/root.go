// Package githubmeta fetches bounded public GitHub metadata via existing gh auth.
package githubmeta

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

const MaxResponseBytes = 2 << 20

var ErrResponseLimit = errors.New("GitHub metadata response exceeds bound")

var repositoryPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
var revisionPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

type boundedBuffer struct {
	b        []byte
	limit    int
	exceeded bool
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if len(b.b)+len(p) > b.limit {
		b.exceeded = true
		return 0, fmt.Errorf("API output limit")
	}
	b.b = append(b.b, p...)
	return len(p), nil
}
func Fetch(ctx context.Context, endpoint string) ([]byte, error) {
	return fetchBounded(ctx, endpoint, MaxResponseBytes)
}

func fetchBounded(ctx context.Context, endpoint string, limit int) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", "api", endpoint)
	stdout, stderr := &boundedBuffer{limit: limit}, &boundedBuffer{limit: 4096}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if e := cmd.Run(); e != nil {
		if stdout.exceeded {
			return nil, ErrResponseLimit
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("GitHub metadata request failed")
	}
	return stdout.b, nil
}

// Root never stores failed/malformed responses. Fetch can be injected for tests.
func Root(ctx context.Context, repo, revision, cache string, fetch func(context.Context, string) ([]byte, error)) ([]byte, error) {
	if !repositoryPattern.MatchString(repo) || !revisionPattern.MatchString(revision) {
		return nil, fmt.Errorf("invalid root identity")
	}
	h := sha256.Sum256([]byte(repo + "\x00" + revision))
	path := filepath.Join(cache, hex.EncodeToString(h[:])+".json")
	f, e := os.Open(path)
	if e == nil {
		defer f.Close()
		b, e := io.ReadAll(io.LimitReader(f, MaxResponseBytes+1))
		if e != nil {
			return nil, e
		}
		if _, e = sweaudit.InspectRoot(repo, b); e != nil {
			return nil, e
		}
		return b, nil
	}
	if !os.IsNotExist(e) {
		return nil, e
	}
	b, e := fetch(ctx, "repos/"+repo+"/git/trees/"+revision)
	if e != nil {
		return nil, e
	}
	if _, e = sweaudit.InspectRoot(repo, b); e != nil {
		return nil, e
	}
	if e = os.MkdirAll(cache, 0700); e != nil {
		return nil, e
	}
	tmp, e := os.CreateTemp(cache, ".root-*.tmp")
	if e != nil {
		return nil, e
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, e = tmp.Write(b); e != nil {
		tmp.Close()
		return nil, e
	}
	if e = tmp.Close(); e != nil {
		return nil, e
	}
	if e = os.Rename(name, path); e != nil {
		return nil, e
	}
	return b, nil
}
