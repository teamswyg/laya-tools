package sweaudit

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
)

type RepositoryIdentity struct {
	Requested, Canonical        string
	ID, NetworkID               int64
	Fork                        bool
	LicenseHint, ResponseSHA256 string
}
type RoleCount struct {
	Role            string
	Representatives int
	Repositories    []string
}
type RepositoryUnit struct {
	SHA256          string
	Repositories    []string
	Representatives int
	Role            string
}
type TaskRole struct {
	Role string
	Task Selection
}
type PartitionReport struct {
	Schema, CandidatesSHA256, MembershipSHA256                                   string
	Candidates, Units                                                            int
	Counts                                                                       []RoleCount
	RepositoryUnits                                                              []RepositoryUnit
	Complete, SemanticIndependenceEstablished, TrainingApproved, ProductionReady bool
}

// PartitionTraining uses identities and requests only; labels cannot enter its
// signature. Whole repository units and all their siblings receive one role.
func PartitionTraining(train, evaluation []Row, prior []string, identities []RepositoryIdentity) (PartitionReport, []TaskRole, error) {
	return partitionTraining(train, evaluation, prior, identities, 2400, 2400, 4800)
}
func partitionTraining(train, evaluation []Row, prior []string, identities []RepositoryIdentity, finalMin, valMin, trainMin int) (PartitionReport, []TaskRole, error) {
	report := PartitionReport{Schema: "riido-training-partition-v1"}
	if finalMin < 1 || valMin < 1 || trainMin < 1 || len(identities) > 256 {
		return report, nil, fmt.Errorf("invalid partition bounds")
	}
	ids := slices.Clone(identities)
	slices.SortFunc(ids, func(a, b RepositoryIdentity) int {
		return strings.Compare(strings.ToLower(a.Requested), strings.ToLower(b.Requested))
	})
	for i, x := range ids {
		if !repoPattern.MatchString(x.Requested) || !repoPattern.MatchString(x.Canonical) || x.ID <= 0 || x.NetworkID <= 0 || x.Canonical != strings.ToLower(x.Canonical) {
			return report, nil, fmt.Errorf("invalid canonical identity")
		}
		if i > 0 && strings.EqualFold(x.Requested, ids[i-1].Requested) {
			return report, nil, fmt.Errorf("duplicate requested identity")
		}
		for _, y := range ids[:i] {
			if (x.ID == y.ID && (x.Canonical != y.Canonical || x.NetworkID != y.NetworkID)) || (x.Canonical == y.Canonical && x.ID != y.ID) {
				return report, nil, fmt.Errorf("inconsistent repository identity")
			}
		}
	}
	lookup := func(name string) (RepositoryIdentity, error) {
		i, ok := slices.BinarySearchFunc(ids, strings.ToLower(name), func(a RepositoryIdentity, b string) int { return strings.Compare(strings.ToLower(a.Requested), b) })
		if !ok {
			return RepositoryIdentity{}, fmt.Errorf("missing canonical identity")
		}
		return ids[i], nil
	}
	canon := func(rows []Row) ([]Row, error) {
		out := slices.Clone(rows)
		for i := range out {
			x, e := lookup(out[i].Repository)
			if e != nil {
				return nil, e
			}
			out[i].Repository = x.Canonical
		}
		return out, nil
	}
	ct, e := canon(train)
	if e != nil {
		return report, nil, e
	}
	ce, e := canon(evaluation)
	if e != nil {
		return report, nil, e
	}
	var protectedNetworks []int64
	for _, name := range append(keys(evaluation, 2), prior...) {
		x, e := lookup(name)
		if e != nil {
			return report, nil, e
		}
		protectedNetworks = append(protectedNetworks, x.NetworkID)
	}
	slices.Sort(protectedNetworks)
	protectedNetworks = slices.Compact(protectedNetworks)
	var forbidden []string
	for _, x := range ids {
		if _, ok := slices.BinarySearch(protectedNetworks, x.NetworkID); ok {
			forbidden = append(forbidden, x.Canonical)
		}
	}
	audited, candidates, e := TrainingCandidates(ct, ce, forbidden)
	if e != nil {
		return report, nil, e
	}
	report.CandidatesSHA256 = audited.MembershipSHA256
	report.Candidates = len(candidates)
	repos := keys(ct, 2)
	parent := make([]int, len(repos))
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
	join := func(a, b int) {
		a, b = root(a), root(b)
		if a > b {
			a, b = b, a
		}
		parent[b] = a
	}
	index := func(name string) int { i, _ := slices.BinarySearch(repos, name); return i }
	for i, a := range ids {
		for _, b := range ids[:i] {
			if a.NetworkID == b.NetworkID {
				ai, af := slices.BinarySearch(repos, a.Canonical)
				bi, bf := slices.BinarySearch(repos, b.Canonical)
				if af && bf {
					join(ai, bi)
				}
			}
		}
	}
	type queryRepo struct {
		query string
		repo  int
	}
	qs := make([]queryRepo, len(ct))
	for i, r := range ct {
		qs[i] = queryRepo{normalized(r.Request), index(r.Repository)}
	}
	slices.SortFunc(qs, func(a, b queryRepo) int { return strings.Compare(a.query, b.query) })
	for i := 1; i < len(qs); i++ {
		if qs[i].query == qs[i-1].query {
			join(qs[i].repo, qs[i-1].repo)
		}
	}
	// Include all original repositories and nonrepresentative rows in the union,
	// even when a repository contributes zero surviving representatives.
	for i := range repos {
		if root(i) != i {
			continue
		}
		var names []string
		count := 0
		for j, name := range repos {
			if root(j) == i {
				names = append(names, name)
			}
		}
		for _, c := range candidates {
			if root(index(c.Repository)) == i {
				count++
			}
		}
		if count > 0 {
			report.RepositoryUnits = append(report.RepositoryUnits, RepositoryUnit{SHA256: hashList(names), Repositories: names, Representatives: count})
		}
	}
	slices.SortFunc(report.RepositoryUnits, func(a, b RepositoryUnit) int {
		if a.SHA256 != b.SHA256 {
			return strings.Compare(a.SHA256, b.SHA256)
		}
		return strings.Compare(a.Repositories[0], b.Repositories[0])
	})
	report.Units = len(report.RepositoryUnits)
	counts := [3]RoleCount{{Role: "final"}, {Role: "validation"}, {Role: "train"}}
	for i := range report.RepositoryUnits {
		unit := &report.RepositoryUnits[i]
		role := 2
		if counts[0].Representatives < finalMin {
			role = 0
		} else if counts[1].Representatives < valMin {
			role = 1
		}
		unit.Role = counts[role].Role
		counts[role].Representatives += unit.Representatives
		counts[role].Repositories = append(counts[role].Repositories, unit.Repositories...)
	}
	for i := range counts {
		slices.Sort(counts[i].Repositories)
	}
	report.Counts = counts[:]
	if counts[0].Representatives < finalMin || counts[1].Representatives < valMin || counts[2].Representatives < trainMin {
		return report, nil, nil
	}
	original := slices.Clone(train)
	slices.SortFunc(original, func(a, b Row) int { return strings.Compare(a.ID, b.ID) })
	var assigned []TaskRole
	for _, c := range candidates {
		role := ""
		for _, unit := range report.RepositoryUnits {
			if _, ok := slices.BinarySearch(unit.Repositories, c.Repository); ok {
				role = unit.Role
				break
			}
		}
		if role == "" {
			return report, nil, fmt.Errorf("unassigned candidate")
		}
		i, ok := slices.BinarySearchFunc(original, c.ID, func(a Row, b string) int { return strings.Compare(a.ID, b) })
		if !ok {
			return report, nil, fmt.Errorf("missing original identity")
		}
		c.Repository = original[i].Repository
		assigned = append(assigned, TaskRole{role, c})
	}
	slices.SortFunc(assigned, func(a, b TaskRole) int {
		if a.Role != b.Role {
			return cmp.Compare(a.Role, b.Role)
		}
		return cmp.Compare(a.Task.ID, b.Task.ID)
	})
	var framed []string
	for _, x := range assigned {
		framed = append(framed, strings.Join([]string{x.Role, x.Task.ID, x.Task.Repository, x.Task.BaseCommit, x.Task.ComponentSHA256}, "\x00"))
	}
	if e := verifyRoles(ct, ids, report, assigned); e != nil {
		return report, nil, e
	}
	report.MembershipSHA256 = hashList(framed)
	report.Complete = true
	return report, assigned, nil
}

// verifyRoles independently checks the completed assignment with sorted keys,
// rather than trusting the union-find paths that constructed it.
func verifyRoles(canonicalTrain []Row, identities []RepositoryIdentity, r PartitionReport, tasks []TaskRole) error {
	type keyRole struct{ key, role string }
	check := func(xs []keyRole) error {
		slices.SortFunc(xs, func(a, b keyRole) int { return strings.Compare(a.key, b.key) })
		for i := 1; i < len(xs); i++ {
			if xs[i].key == xs[i-1].key && xs[i].role != xs[i-1].role {
				return fmt.Errorf("cross-role identity overlap")
			}
		}
		return nil
	}
	var repoRoles []keyRole
	for _, u := range r.RepositoryUnits {
		if u.Role != "train" && u.Role != "validation" && u.Role != "final" {
			return fmt.Errorf("unknown role")
		}
		for _, name := range u.Repositories {
			repoRoles = append(repoRoles, keyRole{name, u.Role})
		}
	}
	if e := check(repoRoles); e != nil {
		return e
	}
	roleFor := func(repo string) string {
		i, ok := slices.BinarySearchFunc(repoRoles, repo, func(a keyRole, b string) int { return strings.Compare(a.key, b) })
		if ok {
			return repoRoles[i].role
		}
		return ""
	}
	var networks, queries []keyRole
	for _, x := range identities {
		if role := roleFor(x.Canonical); role != "" {
			networks = append(networks, keyRole{fmt.Sprint(x.NetworkID), role})
		}
	}
	if e := check(networks); e != nil {
		return e
	}
	for _, x := range canonicalTrain {
		if role := roleFor(x.Repository); role != "" {
			queries = append(queries, keyRole{normalized(x.Request), role})
		}
	}
	if e := check(queries); e != nil {
		return e
	}
	if len(tasks) != r.Candidates {
		return fmt.Errorf("candidate conservation failed")
	}
	seen := make([]string, 0, len(tasks))
	counts := [3]int{}
	rows := slices.Clone(canonicalTrain)
	slices.SortFunc(rows, func(a, b Row) int { return strings.Compare(a.ID, b.ID) })
	for _, x := range tasks {
		i, ok := slices.BinarySearchFunc(rows, x.Task.ID, func(a Row, b string) int { return strings.Compare(a.ID, b) })
		if !ok || x.Task.BaseCommit != rows[i].BaseCommit || x.Role != roleFor(rows[i].Repository) {
			return fmt.Errorf("assignment does not match repository role")
		}
		seen = append(seen, x.Task.ID)
		switch x.Role {
		case "final":
			counts[0]++
		case "validation":
			counts[1]++
		case "train":
			counts[2]++
		default:
			return fmt.Errorf("unknown task role")
		}
	}
	slices.Sort(seen)
	if len(slices.Compact(seen)) != len(tasks) {
		return fmt.Errorf("duplicate assigned task")
	}
	if len(r.Counts) != 3 {
		return fmt.Errorf("role count missing")
	}
	for i, name := range []string{"final", "validation", "train"} {
		if r.Counts[i].Role != name || r.Counts[i].Representatives != counts[i] {
			return fmt.Errorf("role count mismatch")
		}
	}
	return nil
}
