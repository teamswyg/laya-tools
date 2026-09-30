package retrievalbench

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/internal/pairlearn"
	"github.com/teamswyg/laya-tools/internal/searchclaim"
)

const PageClaimPlanSHA256 = "d77b7680340c6c3fec5476509d983efc54911c1e6cd1bcf92e67c3ebf1101211"

type PageClaimMetrics struct {
	Quality                  ClaimMetrics
	Pages, EmittedCandidates int
}
type PageClaimModel struct {
	File, SHA256, Mode, EvaluationRepository string
	Bytes                                    int
	Seed                                     uint64
	Epoch                                    int
	ValidationWeightedNLL                    float64
	Metrics                                  PageClaimMetrics
}
type PageClaimAggregate struct {
	Mode                  string
	Seed                  uint64
	Metrics               PageClaimMetrics
	PassesExploratoryGate bool
}
type PageClaimReport struct {
	Schema, PlanSHA256                                                              string
	Questions, Dimension, PageSize                                                  int
	Baseline, AlwaysHelper                                                          PageClaimMetrics
	Models                                                                          []PageClaimModel
	Aggregates                                                                      []PageClaimAggregate
	PrimaryPassesExploratoryGate, ReplicationPassesExploratoryGate, ProductionReady bool
}

func pageClaimDataset(xs []ClaimExample, repo, split string) pairlearn.Dataset {
	d := claimDataset(xs, repo, split)
	d.SampleWeights = make([]float64, len(d.Labels))
	i := 0
	for _, x := range xs {
		if x.Repository != repo {
			continue
		}
		delta := (x.BaselineRank+19)/20 - (x.InterleavedRank+19)/20
		d.Labels[i] = 0
		if delta > 0 {
			d.Labels[i] = 1
		}
		if delta < 0 {
			delta = -delta
		}
		d.SampleWeights[i] = float64(delta)
		i++
	}
	return d
}
func measurePageClaims(xs []ClaimExample, w []float64, policy string) PageClaimMetrics {
	m := PageClaimMetrics{Quality: measureClaims(xs, w, policy)}
	for _, x := range xs {
		rank := x.BaselineRank
		if policy == "always_interleave" || (policy == "learned" && searchclaim.Score(w, x.Features) >= 0) {
			rank = x.InterleavedRank
		}
		pages := (rank + 19) / 20
		m.Pages += pages
		m.EmittedCandidates += min(3009, pages*20)
	}
	return m
}
func TrainPageClaims(xs []ClaimExample, out, plan string) (PageClaimReport, error) {
	r := PageClaimReport{Schema: "riido-page-claim-report-v1", PlanSHA256: plan, Questions: len(xs), Dimension: searchclaim.Dimension, PageSize: 20}
	if plan != PageClaimPlanSHA256 || len(xs) != 2948 {
		return r, fmt.Errorf("plan or question count mismatch")
	}
	seen := make([]bool, len(xs))
	counts := [3]int{}
	for _, x := range xs {
		repo := slices.Index(claimRepositories, x.Repository)
		if repo < 0 || x.Group < 0 || x.Group >= len(xs) || seen[x.Group] || x.BaselineRank < 1 || x.BaselineRank > 3009 || x.InterleavedRank < 1 || x.InterleavedRank > 3009 {
			return r, fmt.Errorf("invalid frozen question membership")
		}
		seen[x.Group] = true
		counts[repo]++
	}
	if counts != [3]int{806, 1152, 990} {
		return r, fmt.Errorf("repository counts mismatch")
	}
	if err := os.Mkdir(out, 0700); err != nil {
		return r, err
	}
	r.Baseline = measurePageClaims(xs, nil, "baseline")
	r.AlwaysHelper = measurePageClaims(xs, nil, "always_interleave")
	modes := []string{"fp32", "int8", "ternary_ptq", "ternary_ste"}
	seeds := []uint64{1729, 2718}
	var aggregates [8][]ClaimExample
	for fold, testRepo := range claimRepositories {
		trainRepo, valRepo := claimRepositories[(fold+1)%3], claimRepositories[(fold+2)%3]
		train, validation := pageClaimDataset(xs, trainRepo, "development"), pageClaimDataset(xs, valRepo, "validation")
		var test []ClaimExample
		for _, x := range xs {
			if x.Repository == testRepo {
				test = append(test, x)
			}
		}
		for si, seed := range seeds {
			for _, fitMode := range []string{"fp32", "ternary_ste"} {
				fit, err := pairlearn.Fit(train, validation, pairlearn.Config{Seed: seed, Mode: fitMode, LearningRate: .1, L2: .0001, Epochs: 100, Batch: 64, Dimension: searchclaim.Dimension})
				if err != nil {
					return r, err
				}
				derived := []string{fitMode}
				if fitMode == "fp32" {
					derived = modes[:3]
				}
				for _, mode := range derived {
					w := slices.Clone(fit.Weights)
					if mode != fitMode {
						w = hintlearn.Quantize(w, mode)
					}
					nll := pairlearn.NLL(validation, w)
					head := ClaimHead{"riido-page-claim-v1", mode, trainRepo, valRepo, testRepo, plan, seed, fit.Epoch, nll, w}
					b, err := json.MarshalIndent(head, "", "  ")
					if err != nil {
						return r, err
					}
					b = append(b, '\n')
					name := fmt.Sprintf("fold%d-%s-%d.claim.json", fold, mode, seed)
					if err := os.WriteFile(filepath.Join(out, name), b, 0600); err != nil {
						return r, err
					}
					h := sha256.Sum256(b)
					r.Models = append(r.Models, PageClaimModel{name, hex.EncodeToString(h[:]), mode, testRepo, len(b), seed, fit.Epoch, nll, measurePageClaims(test, w, "learned")})
					slot := slices.Index(modes, mode)*2 + si
					for _, x := range test {
						x.AuxiliaryRank = 1
						if searchclaim.Score(w, x.Features) < 0 {
							x.InterleavedRank = x.BaselineRank
							x.AuxiliaryRank = 0
						}
						aggregates[slot] = append(aggregates[slot], x)
					}
				}
			}
		}
	}
	for mi, mode := range modes {
		for si, seed := range seeds {
			selected := aggregates[mi*2+si]
			m := measurePageClaims(selected, nil, "always_interleave")
			m.Quality.AuxiliaryCalls = 0
			for _, x := range selected {
				m.Quality.AuxiliaryCalls += x.AuxiliaryRank
			}
			pass := m.Pages <= r.AlwaysHelper.Pages && m.Quality.AuxiliaryCalls*10 <= r.Questions*9 && m.Quality.Recall1 >= r.Baseline.Quality.Recall1 && m.Quality.Recall10 >= r.Baseline.Quality.Recall10
			r.Aggregates = append(r.Aggregates, PageClaimAggregate{mode, seed, m, pass})
			if mode == "fp32" {
				if si == 0 {
					r.PrimaryPassesExploratoryGate = pass
				} else {
					r.ReplicationPassesExploratoryGate = pass
				}
			}
		}
	}
	return r, nil
}
