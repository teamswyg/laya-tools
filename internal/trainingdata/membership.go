// Package trainingdata exposes only fixed development-role identities to data
// preparation. Final identities are validated but never returned for collection.
package trainingdata

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

const MembershipSHA256 = "5ceb4aa4681784431606f2ff4410c9225b9639d88a18f618b9de70fa2c5da1fd"
const PartitionSHA256 = "0f41c1dafc2268f4c60544d2fd386168172b1c6f68bbbe2bd4e1517c96f9cb64"

var repoPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
var groupPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func ReadDevelopment(path string) ([]sweaudit.TaskRole, error) {
	return readDevelopment(path, MembershipSHA256, [3]int{7335, 5686, 2402})
}
func readDevelopment(path, expected string, counts [3]int) ([]sweaudit.TaskRole, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, (8<<20)+1))
	if e != nil {
		return nil, e
	}
	h := sha256.Sum256(b)
	if len(b) > 8<<20 || hex.EncodeToString(h[:]) != expected {
		return nil, fmt.Errorf("membership size or hash mismatch")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	var rows []sweaudit.TaskRole
	if e = d.Decode(&rows); e != nil {
		return nil, fmt.Errorf("invalid membership schema")
	}
	if e = d.Decode(new(any)); e != io.EOF {
		return nil, fmt.Errorf("extra membership content")
	}
	return development(rows, counts)
}
func development(rows []sweaudit.TaskRole, want [3]int) ([]sweaudit.TaskRole, error) {
	if len(rows) == 0 || len(rows) > 32768 {
		return nil, fmt.Errorf("membership row bound")
	}
	var got [3]int
	var ids []string
	var out []sweaudit.TaskRole
	for _, r := range rows {
		t := r.Task
		if t.Source != "train" || strings.TrimSpace(t.ID) == "" || len(t.ID) > 4096 || !repoPattern.MatchString(t.Repository) || !commitPattern.MatchString(t.BaseCommit) || !groupPattern.MatchString(t.ComponentSHA256) {
			return nil, fmt.Errorf("invalid membership identity")
		}
		ids = append(ids, t.ID)
		switch r.Role {
		case "train":
			got[0]++
		case "validation":
			got[1]++
		case "final":
			got[2]++
		default:
			return nil, fmt.Errorf("invalid role")
		}
		if r.Role != "final" {
			out = append(out, r)
		}
	}
	slices.Sort(ids)
	if len(slices.Compact(ids)) != len(rows) || got != want {
		return nil, fmt.Errorf("membership counts or duplicate ID")
	}
	slices.SortFunc(out, func(a, b sweaudit.TaskRole) int {
		if a.Role != b.Role {
			return strings.Compare(a.Role, b.Role)
		}
		if a.Task.Repository != b.Task.Repository {
			return strings.Compare(a.Task.Repository, b.Task.Repository)
		}
		return strings.Compare(a.Task.ID, b.Task.ID)
	})
	return out, nil
}

// Samples chooses one development task per repository by lexicographic ID,
// without consulting query length, tree size, license or outcome.
func Samples(rows []sweaudit.TaskRole) ([]sweaudit.TaskRole, error) {
	out := slices.Clone(rows)
	for _, r := range out {
		if r.Role != "train" && r.Role != "validation" {
			return nil, fmt.Errorf("nondevelopment role")
		}
	}
	slices.SortFunc(out, func(a, b sweaudit.TaskRole) int {
		if a.Task.Repository != b.Task.Repository {
			return strings.Compare(a.Task.Repository, b.Task.Repository)
		}
		return strings.Compare(a.Task.ID, b.Task.ID)
	})
	n := 0
	for _, r := range out {
		if n == 0 || out[n-1].Task.Repository != r.Task.Repository {
			out[n] = r
			n++
		} else if out[n-1].Role != r.Role {
			return nil, fmt.Errorf("repository spans development roles")
		}
	}
	return out[:n], nil
}
