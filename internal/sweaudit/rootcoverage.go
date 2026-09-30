package sweaudit

import (
	"cmp"
	"encoding/json"
	"slices"
)

const RootCoveragePlanSHA256 = "4e538bbf28f5a7d1fcbbbc39f2b18ea8a410e9642b3b328d190139f94538a0d4"
const FrozenSelectionSHA256 = "fa58c4e51ae307963eb587813e49adf70d987dcb4893f6ee1aa6050bc6b80a5a"

// Candidate identifies a public source object, not an evaluation answer. Keep
// this per-object inventory local until a separate publication decision.
type LicenseCandidate struct {
	Repository, Path, BlobSHA string
	Occurrences               int
}
type RepositoryCoverage struct {
	Repository                                                                                          string
	Snapshots, Available, ResponseBytes, RootCandidateSnapshots, MissingCandidateSnapshots, Unavailable int
}
type RootCoverageReport struct {
	Schema, PlanSHA256, SelectedSHA256                                                                                          string
	Snapshots, Available, ResponseBytes, RootCandidateSnapshots, MissingCandidateSnapshots, Unavailable, UniqueCandidateObjects int
	Repositories                                                                                                                []RepositoryCoverage
	AllSelectedRootsChecked, LicensesReviewed, ProductionReady                                                                  bool
}

func RootCandidates(repo string, b []byte) ([]LicenseCandidate, error) {
	observation, e := InspectRoot(repo, b)
	if e != nil {
		return nil, e
	}
	var tree rootTree
	if e = json.Unmarshal(b, &tree); e != nil {
		return nil, e
	}
	var out []LicenseCandidate
	for _, entry := range tree.Tree {
		if slices.Contains(observation.LicenseCandidates, entry.Path) {
			out = append(out, LicenseCandidate{repo, entry.Path, entry.SHA, 1})
		}
	}
	return out, nil
}
func CompactCandidates(xs []LicenseCandidate) []LicenseCandidate {
	out := slices.Clone(xs)
	slices.SortFunc(out, func(a, b LicenseCandidate) int {
		if a.Repository != b.Repository {
			return cmp.Compare(a.Repository, b.Repository)
		}
		if a.Path != b.Path {
			return cmp.Compare(a.Path, b.Path)
		}
		return cmp.Compare(a.BlobSHA, b.BlobSHA)
	})
	n := 0
	for _, x := range out {
		if n > 0 && out[n-1].Repository == x.Repository && out[n-1].Path == x.Path && out[n-1].BlobSHA == x.BlobSHA {
			out[n-1].Occurrences += x.Occurrences
		} else {
			out[n] = x
			n++
		}
	}
	return out[:n]
}
func SummarizeRoots(observations []RootObservation, candidates []LicenseCandidate) RootCoverageReport {
	r := RootCoverageReport{Schema: "riido-historical-root-coverage-v1", PlanSHA256: RootCoveragePlanSHA256, SelectedSHA256: FrozenSelectionSHA256, Snapshots: len(observations), UniqueCandidateObjects: len(CompactCandidates(candidates))}
	obs := slices.Clone(observations)
	slices.SortFunc(obs, func(a, b RootObservation) int { return cmp.Compare(a.Repository, b.Repository) })
	for _, o := range obs {
		if len(r.Repositories) == 0 || r.Repositories[len(r.Repositories)-1].Repository != o.Repository {
			r.Repositories = append(r.Repositories, RepositoryCoverage{Repository: o.Repository})
		}
		rr := &r.Repositories[len(r.Repositories)-1]
		rr.Snapshots++
		if !o.Available {
			r.Unavailable++
			rr.Unavailable++
			continue
		}
		r.Available++
		rr.Available++
		r.ResponseBytes += o.ResponseBytes
		rr.ResponseBytes += o.ResponseBytes
		if len(o.LicenseCandidates) > 0 {
			r.RootCandidateSnapshots++
			rr.RootCandidateSnapshots++
		} else {
			r.MissingCandidateSnapshots++
			rr.MissingCandidateSnapshots++
		}
	}
	r.AllSelectedRootsChecked = r.Snapshots == 2400 && r.Available == 2400
	return r
}
