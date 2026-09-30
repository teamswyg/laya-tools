package trainingdata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

func queryFixture(t *testing.T, data []byte) (string, string) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "queries.jsonl")
	if err := os.WriteFile(p, data, 0600); err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(data)
	return p, hex.EncodeToString(h[:])
}
func queryRows(t *testing.T, ids ...string) []byte {
	t.Helper()
	var out []byte
	for _, id := range ids {
		b, err := json.Marshal(sweaudit.Row{ID: id, Repository: "Owner/Repo", BaseCommit: strings.Repeat("a", 40), Request: "공개 합성 질문 " + id})
		if err != nil {
			t.Fatal(err)
		}
		out = append(append(out, b...), '\n')
	}
	return out
}
func TestQueryVisitorOnlyDevelopmentRepresentatives(t *testing.T) {
	p, h := queryFixture(t, queryRows(t, "final", "a", "sibling", "v"))
	members := []sweaudit.TaskRole{member("v", "validation", "Owner/Repo"), member("a", "train", "Owner/Repo")}
	before := slices.Clone(members)
	var got []string
	err := visitQueries(p, h, 4, members, func(m sweaudit.TaskRole, r sweaudit.Row) error {
		if r.ID != m.Task.ID || r.Request != "공개 합성 질문 "+r.ID {
			t.Fatal("callback lost identity or query")
		}
		got = append(got, r.ID)
		return nil
	})
	if err != nil || !slices.Equal(got, []string{"a", "v"}) || !reflect.DeepEqual(members, before) {
		t.Fatal("visitor role selection or ownership", got, err)
	}
}
func TestQueryVisitorGuards(t *testing.T) {
	good := queryRows(t, "a", "other")
	for _, tc := range []struct {
		name  string
		data  []byte
		count int
		m     sweaudit.TaskRole
	}{
		{"selected duplicate", queryRows(t, "a", "a"), 2, member("a", "train", "Owner/Repo")},
		{"unselected duplicate", queryRows(t, "a", "other", "other"), 3, member("a", "train", "Owner/Repo")},
		{"missing", good, 2, member("absent", "train", "Owner/Repo")},
		{"case identity", good, 2, member("a", "train", "owner/repo")},
		{"final", good, 2, member("a", "final", "Owner/Repo")},
		{"too few", good, 3, member("a", "train", "Owner/Repo")},
		{"too many", good, 1, member("a", "train", "Owner/Repo")},
		{"invalid utf8", append(slices.Clone(good), 0xff), 3, member("a", "train", "Owner/Repo")},
		{"extra json", []byte(strings.TrimSpace(string(queryRows(t, "a"))) + " {}\n"), 1, member("a", "train", "Owner/Repo")},
		{"oversize line", []byte(strings.Repeat(" ", 1<<20) + "\n"), 1, member("a", "train", "Owner/Repo")},
		{"blank", []byte("\n"), 1, member("a", "train", "Owner/Repo")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, h := queryFixture(t, tc.data)
			if err := visitQueries(p, h, tc.count, []sweaudit.TaskRole{tc.m}, func(sweaudit.TaskRole, sweaudit.Row) error { return nil }); err == nil {
				t.Fatal("invalid projection accepted")
			}
		})
	}
}
func TestQueryVisitorHashBeforeCallbacksAndErrorPropagation(t *testing.T) {
	p, h := queryFixture(t, queryRows(t, "a"))
	m := []sweaudit.TaskRole{member("a", "train", "Owner/Repo")}
	calls := 0
	stop := errors.New("stop")
	visit := func(sweaudit.TaskRole, sweaudit.Row) error { calls++; return stop }
	if err := visitQueries(p, strings.Repeat("0", 64), 1, m, visit); err == nil || calls != 0 {
		t.Fatal("unverified source reached callback")
	}
	if err := visitQueries(p, h, 1, m, visit); !errors.Is(err, stop) || calls != 1 {
		t.Fatal("visitor failure not propagated", err)
	}
	if err := visitQueries(p, h, 1, m, nil); err == nil {
		t.Fatal("nil visitor")
	}
}
func TestQueryVisitorDetectsChangeAfterCallback(t *testing.T) {
	// Exceed the scanner's initial buffer so the callback can change bytes not
	// yet read. The pre-pass hash alone would accept this mutated projection.
	rows := queryRows(t, "a")
	for i := 0; i < 500; i++ {
		rows = append(rows, queryRows(t, "unselected"+strings.Repeat("x", i))...)
	}
	p, h := queryFixture(t, rows)
	err := visitQueries(p, h, 501, []sweaudit.TaskRole{member("a", "train", "Owner/Repo")}, func(sweaudit.TaskRole, sweaudit.Row) error {
		f, err := os.OpenFile(p, os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		// Change a query character while preserving valid JSON and identities.
		_, err = f.WriteAt([]byte("y"), int64(len(rows)-4))
		return err
	})
	if err == nil || !strings.Contains(err.Error(), "changed during read") {
		t.Fatal("second-pass digest did not detect mutation", err)
	}
}
