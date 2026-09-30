package retrievalbench

import "fmt"

type InterleaveProbe struct {
	Questions                               int
	Baseline, AuxiliaryFirst, BaselineFirst ClaimMetrics
	BaselineFirstByRepository               map[string]ClaimMetrics
	ProductionReady                         bool
}

func ProbeInterleave(rows []Row) (InterleaveProbe, error) {
	var r InterleaveProbe
	a, e := PrepareClaims(rows)
	if e != nil {
		return r, e
	}
	b, e := PrepareBaselineFirstClaims(rows)
	if e != nil {
		return r, e
	}
	if len(a) != 2948 || len(b) != len(a) {
		return r, fmt.Errorf("unexpected question count")
	}
	for i := range a {
		if a[i].Group != b[i].Group || a[i].BaselineRank != b[i].BaselineRank || a[i].AuxiliaryRank != b[i].AuxiliaryRank {
			return r, fmt.Errorf("baseline changed")
		}
	}
	r.Questions = len(a)
	r.Baseline = measureClaims(a, nil, "baseline")
	r.AuxiliaryFirst = measureClaims(a, nil, "always_interleave")
	r.BaselineFirst = measureClaims(b, nil, "always_interleave")
	r.BaselineFirstByRepository = map[string]ClaimMetrics{}
	for _, repo := range claimRepositories {
		var xs []ClaimExample
		for _, x := range b {
			if x.Repository == repo {
				xs = append(xs, x)
			}
		}
		r.BaselineFirstByRepository[repo] = measureClaims(xs, nil, "always_interleave")
	}
	return r, nil
}
