package sweaudit

import (
	"reflect"
	"testing"
)

func TestRootCoverageAccounting(t *testing.T) {
	candidates := []LicenseCandidate{{"a/b", "LICENSE", "one", 1}, {"a/b", "LICENSE", "one", 1}, {"a/b", "NOTICE", "two", 1}}
	compact := CompactCandidates(candidates)
	if len(compact) != 2 || compact[0].Occurrences != 2 || candidates[0].Occurrences != 1 {
		t.Fatal("candidate dedup or mutation")
	}
	r := SummarizeRoots([]RootObservation{{Repository: "a/b", ResponseBytes: 10, Available: true, LicenseCandidates: []string{"LICENSE"}}, {Repository: "a/b", ResponseBytes: 20, Available: true}, {Repository: "c/d"}}, candidates)
	if r.Snapshots != 3 || r.Available != 2 || r.Unavailable != 1 || r.RootCandidateSnapshots != 1 || r.MissingCandidateSnapshots != 1 || r.ResponseBytes != 30 || r.UniqueCandidateObjects != 2 || r.AllSelectedRootsChecked || r.LicensesReviewed || r.ProductionReady {
		t.Fatalf("incorrect coverage %+v", r)
	}
	if !reflect.DeepEqual(r.Repositories, []RepositoryCoverage{{"a/b", 2, 2, 30, 1, 1, 0}, {"c/d", 1, 0, 0, 0, 0, 1}}) {
		t.Fatal("repository counts")
	}
}
