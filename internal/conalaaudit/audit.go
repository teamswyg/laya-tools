// Package conalaaudit counts source groups without executing or scoring code.
package conalaaudit

import (
	"bufio"
	"bytes"
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

const PlanSHA256 = "5b38b7c1fce429301eb6a518e2106abff7934ea21254b11c205845d5f70f5872"
const TrainSHA256 = "f25078347c2738c3308e2a1e6aa90dc0c944febfee083dd073e26a68d3747bb7"
const TestSHA256 = "3a7e5eea6deeccb5e7c9557534af860854fd2f0ae870752b42c296ed30e53cb7"

type Row struct {
	ID      int    `json:"question_id"`
	Intent  string `json:"intent"`
	Rewrite string `json:"rewritten_intent"`
	Code    string `json:"snippet"`
}
type Counts struct{ Rows, QuestionIDs, OriginalIntents, RewrittenIntents, Snippets, EmptyOriginal, EmptyRewrite, EmptySnippet, ChangedRewrite int }
type Report struct {
	Schema, SourceRevision, PlanSHA256, TrainSHA256, TestSHA256                      string
	Train, Test, Combined                                                            Counts
	SharedQuestionIDs, SharedOriginalIntents, SharedRewrittenIntents, SharedSnippets int
	ConnectedGroups, LargestGroup, CrossSplitGroups, RowsInCrossSplitGroups          int
	Meets2400OriginalIntents, ProductionReady                                        bool
}

func normalize(s string) string { return strings.Join(strings.Fields(strings.ToLower(s)), " ") }
func hash(b []byte) string      { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func Read(path, expected string) ([]Row, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, (2<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(b) > 2<<20 || hash(b) != expected {
		return nil, fmt.Errorf("source size or hash mismatch")
	}
	return parse(b)
}
func parse(b []byte) ([]Row, error) {
	s := bufio.NewScanner(bytes.NewReader(b))
	s.Buffer(make([]byte, 4096), 256<<10)
	var rows []Row
	for s.Scan() {
		var r Row
		d := json.NewDecoder(bytes.NewReader(s.Bytes()))
		d.DisallowUnknownFields()
		if err := d.Decode(&r); err != nil {
			return nil, fmt.Errorf("invalid source row %d", len(rows)+1)
		}
		if err := d.Decode(new(any)); err != io.EOF || r.ID <= 0 {
			return nil, fmt.Errorf("invalid row framing or question ID")
		}
		rows = append(rows, r)
		if len(rows) > 10000 {
			return nil, fmt.Errorf("too many rows")
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("empty source")
	}
	return rows, nil
}
func ids(rows []Row) []int {
	out := make([]int, len(rows))
	for i, r := range rows {
		out[i] = r.ID
	}
	slices.Sort(out)
	return slices.Compact(out)
}
func keys(rows []Row, field int) []string {
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		var s string
		switch field {
		case 0:
			s = normalize(r.Intent)
		case 1:
			s = normalize(r.Rewrite)
		case 2:
			s = strings.TrimSpace(r.Code)
		}
		if s != "" {
			out = append(out, s)
		}
	}
	slices.Sort(out)
	return slices.Compact(out)
}
func intersection[T cmp.Ordered](a, b []T) int {
	n := 0
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] < b[j] {
			i++
		} else if a[i] > b[j] {
			j++
		} else {
			n++
			i++
			j++
		}
	}
	return n
}
func count(rows []Row) Counts {
	c := Counts{Rows: len(rows), QuestionIDs: len(ids(rows)), OriginalIntents: len(keys(rows, 0)), RewrittenIntents: len(keys(rows, 1)), Snippets: len(keys(rows, 2))}
	for _, r := range rows {
		a, b := normalize(r.Intent), normalize(r.Rewrite)
		if a == "" {
			c.EmptyOriginal++
		}
		if b == "" {
			c.EmptyRewrite++
		}
		if strings.TrimSpace(r.Code) == "" {
			c.EmptySnippet++
		}
		if a != "" && b != "" && a != b {
			c.ChangedRewrite++
		}
	}
	return c
}
func Audit(train, test []Row) Report {
	all := append(slices.Clone(train), test...)
	r := Report{Schema: "riido-conala-audit-v1", SourceRevision: "fbc749f1c537e5c3834e93b15784302e331debe2", PlanSHA256: PlanSHA256, TrainSHA256: TrainSHA256, TestSHA256: TestSHA256, Train: count(train), Test: count(test), Combined: count(all)}
	r.SharedQuestionIDs = intersection(ids(train), ids(test))
	r.SharedOriginalIntents = intersection(keys(train, 0), keys(test, 0))
	r.SharedRewrittenIntents = intersection(keys(train, 1), keys(test, 1))
	r.SharedSnippets = intersection(keys(train, 2), keys(test, 2))
	parent := make([]int, len(all))
	for i := range parent {
		parent[i] = i
	}
	find := func(i int) int {
		for parent[i] != i {
			parent[i] = parent[parent[i]]
			i = parent[i]
		}
		return i
	}
	join := func(a, b int) {
		a, b = find(a), find(b)
		if a != b {
			parent[max(a, b)] = min(a, b)
		}
	}
	type idEntry struct{ key, row int }
	idEntries := make([]idEntry, len(all))
	for i, x := range all {
		idEntries[i] = idEntry{x.ID, i}
	}
	slices.SortFunc(idEntries, func(a, b idEntry) int { return cmp.Compare(a.key, b.key) })
	for i := 1; i < len(idEntries); i++ {
		if idEntries[i].key == idEntries[i-1].key {
			join(idEntries[i].row, idEntries[i-1].row)
		}
	}
	type textEntry struct {
		key string
		row int
	}
	for _, field := range []int{0, 2} {
		entries := make([]textEntry, 0, len(all))
		for i, x := range all {
			s := normalize(x.Intent)
			if field == 2 {
				s = strings.TrimSpace(x.Code)
			}
			if s != "" {
				entries = append(entries, textEntry{s, i})
			}
		}
		slices.SortFunc(entries, func(a, b textEntry) int { return cmp.Compare(a.key, b.key) })
		for i := 1; i < len(entries); i++ {
			if entries[i].key == entries[i-1].key {
				join(entries[i].row, entries[i-1].row)
			}
		}
	}
	sizes := make([]int, len(all))
	split := make([]uint8, len(all))
	for i := range all {
		root := find(i)
		sizes[root]++
		if i < len(train) {
			split[root] |= 1
		} else {
			split[root] |= 2
		}
	}
	for i, size := range sizes {
		if size == 0 {
			continue
		}
		r.ConnectedGroups++
		r.LargestGroup = max(r.LargestGroup, size)
		if split[i] == 3 {
			r.CrossSplitGroups++
			r.RowsInCrossSplitGroups += size
		}
	}
	r.Meets2400OriginalIntents = r.Combined.OriginalIntents >= 2400
	return r
}
