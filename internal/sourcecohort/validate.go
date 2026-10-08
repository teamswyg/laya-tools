package sourcecohort

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

type failure string

func (e failure) Error() string { return string(e) }
func errCode(s string) error    { return failure(s) }
func require(ok bool, code string) {
	if !ok {
		panic(failure(code))
	}
}
func guarded(fn func()) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = errCode("internal_failure")
			if f, ok := p.(failure); ok {
				err = f
			}
		}
	}()
	fn()
	return
}
func Hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func identifier(s string) bool {
	if len(s) < 1 || len(s) > 128 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}
func validFile(f File) bool {
	b, e := hex.DecodeString(f.SHA256)
	return f.Path != "" && len(f.Path) <= 4096 && e == nil && len(b) == 32 && strings.ToLower(f.SHA256) == f.SHA256 && f.Bytes >= 0 && (f.Bytes == 0) == (f.SHA256 == Hash(nil))
}
func timeUTC(s string) time.Time {
	t, e := time.Parse(time.RFC3339Nano, s)
	require(e == nil && strings.HasSuffix(s, "Z") && t.Year() > 1970, "chronology")
	return t
}
func sortedIDs(ids []string) []string {
	out := append([]string{}, ids...)
	sort.Strings(out)
	for i, id := range out {
		require(identifier(id) && (i == 0 || out[i-1] != id), "dependency_id")
	}
	return out
}
func rowIndex(rows []Row, id string) int {
	i := sort.Search(len(rows), func(i int) bool { return rows[i].ID >= id })
	if i == len(rows) || rows[i].ID != id {
		return -1
	}
	return i
}

func assignedReviewers(p Plan, rows []Row) []string {
	if len(p.ReviewerAssignments) == 0 {
		return nil
	}
	require(len(p.ReviewerAssignments) <= len(rows), "reviewer_assignment_count")
	assignments := append([]ReviewerAssignment{}, p.ReviewerAssignments...)
	sort.Slice(assignments, func(i, j int) bool { return assignments[i].ReviewerID < assignments[j].ReviewerID })
	assigned := make([]string, len(rows))
	primary := false
	for i, a := range assignments {
		require(identifier(a.ReviewerID) && a.ReviewerID != p.AuthorID && (i == 0 || assignments[i-1].ReviewerID != a.ReviewerID), "reviewer_assignment_reviewer")
		primary = primary || a.ReviewerID == p.CheckerID
		require(len(a.SourceIDs) > 0 && len(a.SourceIDs) <= len(rows), "reviewer_assignment_count")
		ids := append([]string{}, a.SourceIDs...)
		sort.Strings(ids)
		for _, id := range ids {
			require(identifier(id), "reviewer_assignment_source")
			j := rowIndex(rows, id)
			require(j >= 0, "reviewer_assignment_source")
			require(assigned[j] == "", "reviewer_assignment_duplicate")
			assigned[j] = a.ReviewerID
		}
	}
	require(primary, "reviewer_assignment_primary")
	for _, reviewer := range assigned {
		require(reviewer != "", "reviewer_assignment_missing")
	}
	return assigned
}
func counts(names []string) []Count {
	out := make([]Count, len(names))
	for i, n := range names {
		out[i] = Count{n, 0}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
func countAt(cs []Count, name string) int {
	i := sort.Search(len(cs), func(i int) bool { return cs[i].Name >= name })
	if i == len(cs) || cs[i].Name != name {
		return -1
	}
	return i
}
func initial() Summary {
	return Summary{Schema: "riido-sourcecohort-summary-v1", State: "blocked", Expected: 400, Unresolved: 400, Holds: counts([]string{"completeness", "structural", "source_scope", "semantic", "rights", "dependency", "technical", "declared"}), ProviderIndependence: "unknown", Limits: "Verifies structural provenance joins and declared checks only. Source meaning, reviewer identity/provider independence, human Gold and S3-S11 remain unproved."}
}
func recipeCount(r Recipe) int {
	sum := 0
	for _, s := range r.Splits {
		sum += s.PerStratum
	}
	return len(r.Strata) * sum
}
func validateRegistry(reg Registry, r Recipe, s *Summary) []Row {
	require(r.Schema == "riido-sourcecohort-recipe-v1" && len(r.Strata) == 8 && len(r.Splits) == 4 && r.DEVShort == 30 && r.DEVGeneral == 10, "whole400_recipe")
	expected := DefaultRecipe()
	require(sameIDs(r.Strata, expected.Strata), "whole400_strata")
	for _, got := range r.Splits {
		found := false
		for _, want := range expected.Splits {
			if got == want {
				found = true
			}
		}
		require(found, "whole400_splits")
	}
	require(r.Schema == "riido-sourcecohort-recipe-v1" && len(r.Strata) > 0 && len(r.Strata) <= 64 && len(r.Splits) == 4, "recipe")
	strata := sortedIDs(r.Strata)
	s.Strata = counts(strata)
	s.Splits = counts([]string{"train", "dev", "cal", "test"})
	s.DEVStyles = counts([]string{"short", "general"})
	per := make([]int, 4)
	for _, split := range r.Splits {
		i := countAt(s.Splits, split.Name)
		require(i >= 0 && split.PerStratum > 0 && per[i] == 0, "recipe_split")
		per[i] = split.PerStratum
	}
	s.Expected = recipeCount(r)
	s.Unresolved = s.Expected
	s.Registered = len(reg.Rows)
	require(s.Expected > 0 && s.Expected <= 4096 && r.DEVShort >= 0 && r.DEVGeneral >= 0 && r.DEVShort+r.DEVGeneral == len(strata)*per[countAt(s.Splits, "dev")], "recipe_count")
	require(reg.Schema == "riido-sourcecohort-registry-v1" && identifier(reg.CohortID) && !reg.IntendedLabels && len(reg.Rows) == s.Expected, "registry_count")
	rows := append([]Row{}, reg.Rows...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
	grid := make([]int, len(strata)*4)
	for i, row := range rows {
		require(identifier(row.ID) && (i == 0 || rows[i-1].ID != row.ID), "registry_id")
		a, b := countAt(s.Strata, row.Stratum), countAt(s.Splits, row.Split)
		require(a >= 0 && b >= 0, "registry_allocation")
		s.Strata[a].Rows++
		s.Splits[b].Rows++
		grid[a*4+b]++
		if row.Split == "dev" {
			d := countAt(s.DEVStyles, row.DEVStyle)
			require(d >= 0, "dev_style")
			s.DEVStyles[d].Rows++
		} else {
			require(row.DEVStyle == "", "dev_style")
		}
		if row.Source != nil {
			require(validFile(*row.Source), "source_pin")
		}
	}
	for a := range strata {
		for b := range per {
			require(grid[a*4+b] == per[b], "stratum_split_count")
		}
	}
	require(s.DEVStyles[countAt(s.DEVStyles, "short")].Rows == r.DEVShort && s.DEVStyles[countAt(s.DEVStyles, "general")].Rows == r.DEVGeneral, "dev_style_count")
	for i, row := range rows {
		checkDependencies(rows, i, row.Dependencies)
	}
	return rows
}
func checkDependencies(rows []Row, i int, ids []string) []int {
	out := []int{}
	for _, id := range sortedIDs(ids) {
		j := rowIndex(rows, id)
		require(j >= 0 && j != i && rows[j].Split == rows[i].Split, "dependency_containment")
		out = append(out, j)
	}
	return out
}

// ValidateAllocation needs no source files or generated material.
func ValidateAllocation(reg Registry, r Recipe) (s Summary, err error) {
	s = initial()
	err = guarded(func() { validateRegistry(reg, r, &s); s.State = "allocation_ready"; s.Unresolved = 0 })
	if err != nil {
		s.Code = err.Error()
	}
	return
}
func reviewHolds(reviews []Review, s *Summary) {
	for _, r := range reviews {
		held := len(r.Holds) > 0 || !r.Complete || !r.Structural || r.SourceScope != "complete_source_situation" || r.SemanticVerdict != "declared_pass"
		for _, h := range r.Holds {
			i := countAt(s.Holds, h.Category)
			if i < 0 {
				i = countAt(s.Holds, "declared")
			}
			s.Holds[i].Rows++
		}
		for _, item := range []struct {
			bad      bool
			category string
		}{{!r.Complete, "completeness"}, {!r.Structural, "structural"}, {r.SourceScope != "complete_source_situation", "source_scope"}, {r.SemanticVerdict != "declared_pass", "semantic"}} {
			if item.bad {
				s.Holds[countAt(s.Holds, item.category)].Rows++
			}
		}
		if held {
			s.Held++
		}
	}
}
func checkSpans(data []byte, spans []Span) {
	require(utf8.Valid(data) && strings.TrimSpace(string(data)) != "" && len(spans) > 0, "source_evidence")
	whole := false
	for _, p := range spans {
		require(p.Start >= 0 && p.End > p.Start && p.End <= len(data) && (p.Start == 0 || utf8.RuneStart(data[p.Start])) && (p.End == len(data) || utf8.RuneStart(data[p.End])), "evidence_boundary")
		switch p.Kind {
		case "source_scope", "proposition", "opposition", "event_anchor", "communication_time", "absence", "not_applicable":
		default:
			require(false, "evidence_kind")
		}
		if p.Kind == "source_scope" && p.Start == 0 && p.End == len(data) {
			whole = true
		}
	}
	require(whole, "whole_source_evidence")
}
