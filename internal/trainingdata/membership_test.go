package trainingdata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

func member(id, role, repo string) sweaudit.TaskRole {
	return sweaudit.TaskRole{Role: role, Task: sweaudit.Selection{Source: "train", ID: id, Repository: repo, BaseCommit: strings.Repeat("a", 40), ComponentSHA256: strings.Repeat("b", 64)}}
}
func TestDevelopmentExcludesFinalAndSamplesByIdentity(t *testing.T) {
	rows := []sweaudit.TaskRole{member("z", "train", "a/a"), member("final", "final", "c/c"), member("a", "train", "a/a"), member("v", "validation", "b/b")}
	before := slices.Clone(rows)
	dev, e := development(rows, [3]int{2, 1, 1})
	if e != nil || len(dev) != 3 {
		t.Fatal(e)
	}
	for _, r := range dev {
		if r.Role == "final" {
			t.Fatal("final identity escaped")
		}
	}
	sample, e := Samples(dev)
	if e != nil || len(sample) != 2 || sample[0].Task.ID != "a" {
		t.Fatal("sample choice")
	}
	slices.Reverse(dev)
	again, e := Samples(dev)
	if e != nil || !reflect.DeepEqual(sample, again) {
		t.Fatal("order-dependent sample")
	}
	if !reflect.DeepEqual(rows, before) {
		t.Fatal("input mutated")
	}
	if _, e = Samples(rows); e == nil {
		t.Fatal("final sampling accepted")
	}
	conflict := []sweaudit.TaskRole{member("a", "train", "a/a"), member("b", "validation", "a/a")}
	if _, e = Samples(conflict); e == nil {
		t.Fatal("cross-role repository")
	}
}
func TestMembershipHashSchemaAndCountGuards(t *testing.T) {
	rows := []sweaudit.TaskRole{member("a", "train", "a/a"), member("b", "validation", "b/b"), member("c", "final", "c/c")}
	b, _ := json.Marshal(rows)
	h := sha256.Sum256(b)
	sha := hex.EncodeToString(h[:])
	p := filepath.Join(t.TempDir(), "membership.json")
	if e := os.WriteFile(p, b, 0600); e != nil {
		t.Fatal(e)
	}
	if out, e := readDevelopment(p, sha, [3]int{1, 1, 1}); e != nil || len(out) != 2 {
		t.Fatal(e)
	}
	if _, e := readDevelopment(p, strings.Repeat("0", 64), [3]int{1, 1, 1}); e == nil {
		t.Fatal("bad hash")
	}
	if _, e := readDevelopment(p, sha, [3]int{2, 0, 1}); e == nil {
		t.Fatal("bad role counts")
	}
	rows[2].Task.ID = "a"
	if _, e := development(rows, [3]int{1, 1, 1}); e == nil {
		t.Fatal("duplicate ID")
	}
	if _, e := ReadDevelopment(p); e == nil {
		t.Fatal("synthetic file passed production pin")
	}
}
