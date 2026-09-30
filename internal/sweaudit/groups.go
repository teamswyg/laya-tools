package sweaudit

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
)

const GroupsPlanSHA256 = "cbfcaa906a184678525ecfee878a67870e724df38b36535105d7608979e1074f"

type Selection struct{ Source, ID, Repository, BaseCommit, ComponentSHA256 string }
type ComponentSize struct{ Rows, Components int }
type SourceCount struct {
	Source   string
	Selected int
}
type GroupReport struct {
	Schema, PlanSHA256, FullProjectionSHA256, MultilingualProjectionSHA256, AllGroupsSHA256, SelectedSHA256 string
	Rows, Components, NonSingletonComponents, LargestComponent, Required, SelectedComponents, Shortfall     int
	Sizes                                                                                                   []ComponentSize
	SelectedCounts                                                                                          Counts
	Sources                                                                                                 []SourceCount
	SelectedRequestBytes, MaxRequestBytes, RequestsOver8192                                                 int
	EvaluationOnly, ProductionReady                                                                         bool
}
type groupRow struct {
	row         Row
	source, tag string
}
type component struct {
	signature      string
	representative int
	size           int
}

func hashList(values []string) string {
	h := sha256.New()
	for _, s := range values {
		fmt.Fprintf(h, "%d:%s\n", len(s), s)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// GroupEvaluation freezes representatives without consulting any outcome.
// All source rows, including unselected group siblings, stay evaluation-only.
func GroupEvaluation(full, multi []Row) (GroupReport, []Selection, error) {
	return groupEvaluation(full, multi, 2400)
}
func groupEvaluation(full, multi []Row, required int) (GroupReport, []Selection, error) {
	r := GroupReport{Schema: "riido-evaluation-groups-v1", PlanSHA256: GroupsPlanSHA256, FullProjectionSHA256: FullProjectionSHA256, MultilingualProjectionSHA256: MultilingualProjectionSHA256, Required: required, EvaluationOnly: true}
	if len(full)+len(multi) == 0 || len(full)+len(multi) > 4096 || required < 1 {
		return r, nil, fmt.Errorf("invalid grouping size")
	}
	var rows []groupRow
	for si, input := range [][]Row{full, multi} {
		source := []string{"full", "multilingual"}[si]
		for _, x := range input {
			if strings.TrimSpace(x.ID) == "" || normalized(x.Request) == "" || !repoPattern.MatchString(x.Repository) || !commitPattern.MatchString(x.BaseCommit) {
				return r, nil, fmt.Errorf("invalid grouping row")
			}
			rows = append(rows, groupRow{x, source, source + "\x00" + x.ID})
		}
	}
	// Stable positions prevent input order from affecting component roots/ties.
	slices.SortFunc(rows, func(a, b groupRow) int { return cmp.Compare(a.tag, b.tag) })
	for i := 1; i < len(rows); i++ {
		if rows[i].tag == rows[i-1].tag {
			return r, nil, fmt.Errorf("duplicate source-tagged identity")
		}
	}
	r.Rows = len(rows)
	parent := make([]int, len(rows))
	for i := range parent {
		parent[i] = i
	}
	var root func(int) int
	root = func(i int) int {
		for parent[i] != i {
			parent[i] = parent[parent[i]]
			i = parent[i]
		}
		return i
	}
	join := func(a, b int) {
		a, b = root(a), root(b)
		if a > b {
			a, b = b, a
		}
		parent[b] = a
	}
	type edge struct {
		key   string
		index int
	}
	edges := make([]edge, len(rows))
	for field := 0; field < 3; field++ {
		for i, x := range rows {
			var key string
			switch field {
			case 0:
				key = x.row.ID
			case 1:
				key = normalized(x.row.Request)
			case 2:
				key = x.row.Repository + "\x00" + x.row.BaseCommit
			}
			edges[i] = edge{key, i}
		}
		slices.SortFunc(edges, func(a, b edge) int { return cmp.Compare(a.key, b.key) })
		for i := 1; i < len(edges); i++ {
			if edges[i].key == edges[i-1].key {
				join(edges[i].index, edges[i-1].index)
			}
		}
	}
	type member struct{ root, index int }
	members := make([]member, len(rows))
	for i := range rows {
		members[i] = member{root(i), i}
	}
	slices.SortFunc(members, func(a, b member) int {
		if a.root != b.root {
			return cmp.Compare(a.root, b.root)
		}
		return cmp.Compare(a.index, b.index)
	})
	var groups []component
	var sizes []int
	for i := 0; i < len(members); {
		j := i + 1
		for j < len(members) && members[j].root == members[i].root {
			j++
		}
		names := make([]string, j-i)
		for k := i; k < j; k++ {
			names[k-i] = rows[members[k].index].tag
		}
		groups = append(groups, component{hashList(names), members[i].index, j - i})
		sizes = append(sizes, j-i)
		r.LargestComponent = max(r.LargestComponent, j-i)
		if j-i > 1 {
			r.NonSingletonComponents++
		}
		i = j
	}
	slices.SortFunc(groups, func(a, b component) int {
		if a.signature != b.signature {
			return cmp.Compare(a.signature, b.signature)
		}
		return cmp.Compare(rows[a.representative].tag, rows[b.representative].tag)
	})
	r.Components = len(groups)
	slices.Sort(sizes)
	for i := 0; i < len(sizes); {
		j := i + 1
		for j < len(sizes) && sizes[j] == sizes[i] {
			j++
		}
		r.Sizes = append(r.Sizes, ComponentSize{sizes[i], j - i})
		i = j
	}
	signatures := make([]string, len(groups))
	for i, g := range groups {
		signatures[i] = g.signature
	}
	r.AllGroupsSHA256 = hashList(signatures)
	if len(groups) < required {
		r.Shortfall = required - len(groups)
		return r, nil, nil
	}
	var selected []Selection
	var selectedRows []Row
	var selectionKeys []string
	sourceCounts := [2]int{}
	for _, g := range groups[:required] {
		x := rows[g.representative]
		selected = append(selected, Selection{x.source, x.row.ID, x.row.Repository, x.row.BaseCommit, g.signature})
		selectedRows = append(selectedRows, x.row)
		selectionKeys = append(selectionKeys, strings.Join([]string{x.tag, x.row.Repository, x.row.BaseCommit, g.signature, digest([]byte(normalized(x.row.Request)))}, "\x00"))
		if x.source == "full" {
			sourceCounts[0]++
		} else {
			sourceCounts[1]++
		}
		size := len(x.row.Request)
		r.SelectedRequestBytes += size
		r.MaxRequestBytes = max(r.MaxRequestBytes, size)
		if size > 8192 {
			r.RequestsOver8192++
		}
	}
	r.SelectedComponents = len(selected)
	r.SelectedCounts = count(selectedRows)
	r.SelectedSHA256 = hashList(selectionKeys)
	r.Sources = []SourceCount{{"full", sourceCounts[0]}, {"multilingual", sourceCounts[1]}}
	if r.SelectedCounts.DistinctIDs != required || r.SelectedCounts.DistinctRequests != required || r.SelectedCounts.DistinctSnapshots != required {
		return r, nil, fmt.Errorf("selected component overlap")
	}
	return r, selected, nil
}
