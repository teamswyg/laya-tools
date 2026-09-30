package githubmeta

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

type repositoryObject struct {
	ID             int64  `json:"id"`
	Name           string `json:"full_name"`
	Private        *bool  `json:"private"`
	Visibility     string `json:"visibility"`
	Fork           *bool  `json:"fork"`
	Parent, Source *repositoryObject
	License        *struct {
		SPDX string `json:"spdx_id"`
	} `json:"license"`
}

func publicRepository(r *repositoryObject) bool {
	return r != nil && r.ID > 0 && repositoryPattern.MatchString(r.Name) && r.Private != nil && !*r.Private && (r.Visibility == "" || r.Visibility == "public")
}
func decodeRepository(requested string, b []byte) (sweaudit.RepositoryIdentity, error) {
	var out sweaudit.RepositoryIdentity
	if !repositoryPattern.MatchString(requested) || len(b) == 0 || len(b) > MaxResponseBytes {
		return out, fmt.Errorf("invalid repository metadata bounds")
	}
	var r repositoryObject
	if e := json.Unmarshal(b, &r); e != nil || !publicRepository(&r) || r.Visibility != "public" || r.Fork == nil {
		return out, fmt.Errorf("require explicit public repository identity")
	}
	network := r.ID
	if *r.Fork {
		if !publicRepository(r.Parent) || !publicRepository(r.Source) {
			return out, fmt.Errorf("require public fork ancestry")
		}
		network = r.Source.ID
	}
	spdx := ""
	if r.License != nil {
		spdx = r.License.SPDX
		if len(spdx) > 128 || strings.ContainsAny(spdx, "\x00\r\n") {
			return out, fmt.Errorf("invalid license hint")
		}
	}
	h := sha256.Sum256(b)
	return sweaudit.RepositoryIdentity{Requested: requested, Canonical: strings.ToLower(r.Name), ID: r.ID, NetworkID: network, Fork: *r.Fork, LicenseHint: spdx, ResponseSHA256: hex.EncodeToString(h[:])}, nil
}

// Repository captures public identity and current license metadata. A license
// hint is not historical license text or blanket permission for training.
func Repository(ctx context.Context, repo, cache string, fetch func(context.Context, string) ([]byte, error)) (sweaudit.RepositoryIdentity, error) {
	var empty sweaudit.RepositoryIdentity
	if !repositoryPattern.MatchString(repo) {
		return empty, fmt.Errorf("invalid repository name")
	}
	h := sha256.Sum256([]byte(strings.ToLower(repo)))
	p := filepath.Join(cache, hex.EncodeToString(h[:])+".json")
	f, e := os.Open(p)
	if e == nil {
		defer f.Close()
		b, e := io.ReadAll(io.LimitReader(f, MaxResponseBytes+1))
		if e != nil {
			return empty, e
		}
		return decodeRepository(repo, b)
	}
	if !os.IsNotExist(e) {
		return empty, e
	}
	b, e := fetch(ctx, "repos/"+repo)
	if e != nil {
		return empty, e
	}
	result, e := decodeRepository(repo, b)
	if e != nil {
		return empty, e
	}
	if e = os.MkdirAll(cache, 0700); e != nil {
		return empty, e
	}
	tmp, e := os.CreateTemp(cache, ".repository-*.tmp")
	if e != nil {
		return empty, e
	}
	defer os.Remove(tmp.Name())
	if _, e = tmp.Write(b); e != nil {
		tmp.Close()
		return empty, e
	}
	if e = tmp.Close(); e != nil {
		return empty, e
	}
	if e = os.Rename(tmp.Name(), p); e != nil {
		return empty, e
	}
	return result, nil
}
