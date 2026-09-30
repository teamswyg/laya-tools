package trainingdata

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

const QueryProjectionSHA256 = "82419343bf7a06daffe79b2145800648e776f971a0f6f873fd11277602a87f51"

// VisitQueries verifies the pinned projection before streaming development
// representatives to visit. Other source rows are decoded for integrity checks
// but never reach visit. Callers must stage results until this function succeeds:
// a missing member or a concurrent file change can be detected after callbacks.
// The reader retains IDs and one query at a time, not the full query corpus.
func VisitQueries(path string, members []sweaudit.TaskRole, visit func(sweaudit.TaskRole, sweaudit.Row) error) error {
	return visitQueries(path, QueryProjectionSHA256, 19008, members, visit)
}

func visitQueries(path, expected string, count int, members []sweaudit.TaskRole, visit func(sweaudit.TaskRole, sweaudit.Row) error) error {
	if visit == nil || count < 1 || count > 32768 || len(members) == 0 || len(members) > count {
		return fmt.Errorf("invalid query visitor bounds")
	}
	selected := slices.Clone(members)
	slices.SortFunc(selected, func(a, b sweaudit.TaskRole) int { return strings.Compare(a.Task.ID, b.Task.ID) })
	for i, m := range selected {
		t := m.Task
		if (m.Role != "train" && m.Role != "validation") || t.Source != "train" || strings.TrimSpace(t.ID) == "" || len(t.ID) > 4096 || !repoPattern.MatchString(t.Repository) || !commitPattern.MatchString(t.BaseCommit) || !groupPattern.MatchString(t.ComponentSHA256) || (i > 0 && selected[i-1].Task.ID == t.ID) {
			return fmt.Errorf("invalid development query member")
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	const maxBytes = 96 << 20
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, maxBytes+1))
	if err != nil {
		return err
	}
	if n > maxBytes || hex.EncodeToString(h.Sum(nil)) != expected {
		return fmt.Errorf("query projection size or hash mismatch")
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	h.Reset()
	s := bufio.NewScanner(io.TeeReader(io.LimitReader(f, maxBytes+1), h))
	s.Buffer(make([]byte, 64<<10), 1<<20)
	ids := make([]string, 0, count)
	seen := make([]bool, len(selected))
	for s.Scan() {
		if len(ids) >= count || !utf8.Valid(s.Bytes()) {
			return fmt.Errorf("query row count or encoding invalid")
		}
		var row sweaudit.Row
		d := json.NewDecoder(bytes.NewReader(s.Bytes()))
		d.DisallowUnknownFields()
		if d.Decode(&row) != nil || d.Decode(new(any)) != io.EOF || strings.TrimSpace(row.ID) == "" || len(row.ID) > 4096 || !repoPattern.MatchString(row.Repository) || !commitPattern.MatchString(row.BaseCommit) || strings.TrimSpace(row.Request) == "" {
			return fmt.Errorf("invalid query row")
		}
		ids = append(ids, row.ID)
		i, found := slices.BinarySearchFunc(selected, row.ID, func(m sweaudit.TaskRole, id string) int { return strings.Compare(m.Task.ID, id) })
		if !found {
			continue
		}
		m := selected[i]
		if seen[i] || row.Repository != m.Task.Repository || row.BaseCommit != m.Task.BaseCommit {
			return fmt.Errorf("query member identity mismatch or duplicate")
		}
		seen[i] = true
		if err = visit(m, row); err != nil {
			return err
		}
	}
	if s.Err() != nil {
		return fmt.Errorf("query projection read or line bound failure")
	}
	if hex.EncodeToString(h.Sum(nil)) != expected {
		return fmt.Errorf("query projection changed during read")
	}
	slices.Sort(ids)
	if len(ids) != count || len(slices.Compact(ids)) != count || slices.Contains(seen, false) {
		return fmt.Errorf("query projection duplicate, count or missing member")
	}
	return nil
}
