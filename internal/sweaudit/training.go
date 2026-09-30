package sweaudit

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
)

// ReadTraining explicitly widens identity-only audit bounds, not runtime/model
// or evaluation limits. The caller must pin the expected projection digest.
func ReadTraining(path, expected string) ([]Row, error) {
	return readBounded(path, expected, 96<<20, 32768)
}

type TrainingReport struct {
	Schema, MembershipSHA256                                                                                             string
	Training, Evaluation, Candidates                                                                                     Counts
	DirectSharedIDs, DirectSharedRequests, DirectSharedSnapshots                                                         int
	SharedRepositories                                                                                                   []string
	Components, ExcludedComponents, ExcludedTrainingRows, EligibleComponents, EligibleTrainingRows, DuplicateSiblingRows int
	TrainingApproved, SplitEstablished, PerformanceEstablished, ProductionReady                                          bool
}

// TrainingCandidates excludes transitively related evaluation examples and
// whole evaluation/prior repositories before choosing any representatives.
// This audits identity only; it neither scores tasks nor approves licenses.
func TrainingCandidates(train, evaluation []Row, priorRepos []string) (TrainingReport, []Selection, error) {
	report := TrainingReport{Schema: "riido-training-source-audit-v1"}
	if len(train) == 0 || len(evaluation) == 0 || len(train)+len(evaluation) > 32768 {
		return report, nil, fmt.Errorf("invalid audit size")
	}
	rows := append(slices.Clone(train), evaluation...)
	for _, r := range rows {
		if strings.TrimSpace(r.ID) == "" || normalized(r.Request) == "" || !repoPattern.MatchString(r.Repository) || !commitPattern.MatchString(r.BaseCommit) {
			return report, nil, fmt.Errorf("invalid identity row")
		}
	}
	report.Training, report.Evaluation = count(train), count(evaluation)
	if report.Training.DistinctIDs != len(train) {
		return report, nil, fmt.Errorf("duplicate training ID")
	}
	report.DirectSharedIDs = len(intersect(keys(train, 0), keys(evaluation, 0)))
	report.DirectSharedRequests = len(intersect(keys(train, 1), keys(evaluation, 1)))
	report.DirectSharedSnapshots = len(intersect(keys(train, 3), keys(evaluation, 3)))
	report.SharedRepositories = intersect(keys(train, 2), keys(evaluation, 2))
	forbidden := distinct(append(keys(evaluation, 2), priorRepos...))
	parent := make([]int, len(rows))
	for i := range parent {
		parent[i] = i
	}
	root := func(i int) int {
		for parent[i] != i {
			parent[i] = parent[parent[i]]
			i = parent[i]
		}
		return i
	}
	type edge struct {
		key   string
		index int
	}
	edges := make([]edge, len(rows))
	for field := range 3 {
		for i, r := range rows {
			var key string
			switch field {
			case 0:
				key = r.ID
			case 1:
				key = normalized(r.Request)
			case 2:
				key = r.Repository + "\x00" + r.BaseCommit
			}
			edges[i] = edge{key, i}
		}
		slices.SortFunc(edges, func(a, b edge) int { return cmp.Compare(a.key, b.key) })
		for i := 1; i < len(edges); i++ {
			if edges[i].key == edges[i-1].key {
				a, b := root(edges[i].index), root(edges[i-1].index)
				if a > b {
					a, b = b, a
				}
				parent[b] = a
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
		return strings.Compare(rows[a.index].ID, rows[b.index].ID)
	})
	var selected []Selection
	var candidates []Row
	for i := 0; i < len(members); {
		j := i + 1
		for j < len(members) && members[j].root == members[i].root {
			j++
		}
		blocked := false
		trainingRows := 0
		representative := -1
		var names []string
		for _, m := range members[i:j] {
			x := rows[m.index]
			_, forbiddenRepo := slices.BinarySearch(forbidden, x.Repository)
			blocked = blocked || m.index >= len(train) || forbiddenRepo
			if m.index < len(train) {
				trainingRows++
				names = append(names, "train\x00"+x.ID)
				if representative < 0 || x.ID < rows[representative].ID {
					representative = m.index
				}
			}
		}
		if trainingRows > 0 {
			report.Components++
			if blocked {
				report.ExcludedComponents++
				report.ExcludedTrainingRows += trainingRows
			} else {
				report.EligibleComponents++
				report.EligibleTrainingRows += trainingRows
				slices.Sort(names)
				x := rows[representative]
				sig := hashList(names)
				selected = append(selected, Selection{"train", x.ID, x.Repository, x.BaseCommit, sig})
				candidates = append(candidates, x)
			}
		}
		i = j
	}
	report.DuplicateSiblingRows = report.EligibleTrainingRows - report.EligibleComponents
	report.Candidates = count(candidates)
	slices.SortFunc(selected, func(a, b Selection) int {
		if a.ComponentSHA256 != b.ComponentSHA256 {
			return strings.Compare(a.ComponentSHA256, b.ComponentSHA256)
		}
		return strings.Compare(a.ID, b.ID)
	})
	slices.SortFunc(candidates, func(a, b Row) int { return strings.Compare(a.ID, b.ID) })
	var names []string
	for _, s := range selected {
		i, _ := slices.BinarySearchFunc(candidates, s.ID, func(a Row, b string) int { return strings.Compare(a.ID, b) })
		names = append(names, strings.Join([]string{s.ID, s.Repository, s.BaseCommit, s.ComponentSHA256, digest([]byte(normalized(candidates[i].Request)))}, "\x00"))
	}
	report.MembershipSHA256 = hashList(names)
	return report, selected, nil
}
