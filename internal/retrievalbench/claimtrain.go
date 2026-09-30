package retrievalbench

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/internal/pairlearn"
	"github.com/teamswyg/laya-tools/internal/searchclaim"
)

var claimRepositories = []string{"etcd-io/etcd", "kubernetes/test-infra", "lxc/lxd"}

type ClaimMetrics struct {
	Questions, AuxiliaryCalls, Wins, Losses, Ties int
	Recall1, Recall10, MRR, MeanRank              float64
}

func measureClaims(examples []ClaimExample, weights []float64, policy string) ClaimMetrics {
	var m ClaimMetrics
	for _, x := range examples {
		rank := x.BaselineRank
		use := policy == "always_interleave" || (policy == "learned" && searchclaim.Score(weights, x.Features) >= 0)
		if use {
			m.AuxiliaryCalls++
			rank = x.InterleavedRank
		}
		m.Questions++
		if rank < x.BaselineRank {
			m.Wins++
		} else if rank > x.BaselineRank {
			m.Losses++
		} else {
			m.Ties++
		}
		if rank == 1 {
			m.Recall1++
		}
		if rank <= 10 {
			m.Recall10++
		}
		m.MRR += 1 / float64(rank)
		m.MeanRank += float64(rank)
	}
	if m.Questions > 0 {
		n := float64(m.Questions)
		m.Recall1 /= n
		m.Recall10 /= n
		m.MRR /= n
		m.MeanRank /= n
	}
	return m
}
func claimDataset(examples []ClaimExample, repo, split string) pairlearn.Dataset {
	d := pairlearn.Dataset{Split: split, Offsets: []int{0}}
	for _, x := range examples {
		if x.Repository != repo {
			continue
		}
		for i, v := range x.Features {
			if v != 0 {
				d.Indices = append(d.Indices, uint16(i))
				d.Values = append(d.Values, v)
			}
		}
		d.Offsets = append(d.Offsets, len(d.Indices))
		label := 0.
		if x.InterleavedRank < x.BaselineRank {
			label = 1
		}
		d.Labels = append(d.Labels, label)
		d.Groups = append(d.Groups, x.Group)
	}
	return d
}

type ClaimHead struct {
	Schema, Mode, TrainingRepository, ValidationRepository, EvaluationRepository, PlanSHA256 string
	Seed                                                                                     uint64
	Epoch                                                                                    int
	ValidationNLL                                                                            float64
	Weights                                                                                  []float64
}
type ClaimModelResult struct {
	File, SHA256, Mode, EvaluationRepository string
	Bytes                                    int
	Seed                                     uint64
	Epoch                                    int
	ValidationNLL                            float64
	Metrics                                  ClaimMetrics
}
type ClaimAggregate struct {
	Mode                  string
	Seed                  uint64
	Metrics               ClaimMetrics
	PassesExploratoryGate bool
}
type ClaimReport struct {
	Schema, PlanSHA256                            string
	Questions, Dimension                          int
	Label                                         string
	Baseline, AlwaysInterleave                    ClaimMetrics
	Models                                        []ClaimModelResult
	Aggregates                                    []ClaimAggregate
	PrimaryPassesExploratoryGate, ProductionReady bool
}

func TrainClaims(examples []ClaimExample, out, planHash string) (ClaimReport, error) {
	r := ClaimReport{Schema: "riido-search-claim-report-v1", PlanSHA256: planHash, Questions: len(examples), Dimension: searchclaim.Dimension, Label: "interleaved known-target rank strictly smaller than raw baseline rank"}
	if len(examples) != 2948 {
		return r, fmt.Errorf("unexpected question count")
	}
	seen := map[int]bool{}
	counts := map[string]int{}
	for _, x := range examples {
		if seen[x.Group] {
			return r, fmt.Errorf("duplicate question group")
		}
		seen[x.Group] = true
		counts[x.Repository]++
		if x.BaselineRank < 1 || x.InterleavedRank < 1 {
			return r, fmt.Errorf("invalid target rank")
		}
	}
	if len(counts) != 3 || counts[claimRepositories[0]] != 806 || counts[claimRepositories[1]] != 1152 || counts[claimRepositories[2]] != 990 {
		return r, fmt.Errorf("repository split mismatch")
	}
	if e := os.Mkdir(out, 0700); e != nil {
		return r, e
	}
	r.Baseline = measureClaims(examples, nil, "baseline")
	r.AlwaysInterleave = measureClaims(examples, nil, "always_interleave")
	type aggregateKey struct {
		mode string
		seed uint64
	}
	aggregateExamples := map[aggregateKey][]ClaimExample{}
	for fold, testRepo := range claimRepositories {
		trainRepo := claimRepositories[(fold+1)%3]
		validationRepo := claimRepositories[(fold+2)%3]
		train := claimDataset(examples, trainRepo, "development")
		validation := claimDataset(examples, validationRepo, "validation")
		var test []ClaimExample
		for _, x := range examples {
			if x.Repository == testRepo {
				test = append(test, x)
			}
		}
		for _, seed := range []uint64{1729, 2718} {
			for _, fitMode := range []string{"fp32", "ternary_ste"} {
				fit, e := pairlearn.Fit(train, validation, pairlearn.Config{Seed: seed, Mode: fitMode, LearningRate: .1, L2: .0001, Epochs: 100, Batch: 64, Dimension: searchclaim.Dimension})
				if e != nil {
					return r, e
				}
				modes := []string{fitMode}
				if fitMode == "fp32" {
					modes = []string{"fp32", "int8", "ternary_ptq"}
				}
				for _, mode := range modes {
					w := append([]float64(nil), fit.Weights...)
					if mode != fitMode {
						w = hintlearn.Quantize(fit.Weights, mode)
					}
					nll := pairlearn.NLL(validation, w)
					head := ClaimHead{searchclaim.Schema, mode, trainRepo, validationRepo, testRepo, planHash, seed, fit.Epoch, nll, w}
					b, e := json.MarshalIndent(head, "", "  ")
					if e != nil {
						return r, e
					}
					b = append(b, '\n')
					name := fmt.Sprintf("fold%d-%s-%d.claim.json", fold, mode, seed)
					if e := os.WriteFile(filepath.Join(out, name), b, 0600); e != nil {
						return r, e
					}
					h := sha256.Sum256(b)
					m := measureClaims(test, w, "learned")
					r.Models = append(r.Models, ClaimModelResult{name, hex.EncodeToString(h[:]), mode, testRepo, len(b), seed, fit.Epoch, nll, m})
					// Store each held-out question's actual selected rank; aggregate
					// by variant across folds without fitting to aggregate outcomes.
					key := aggregateKey{mode, seed}
					for _, x := range test {
						if searchclaim.Score(w, x.Features) < 0 {
							x.InterleavedRank = x.BaselineRank
							x.AuxiliaryRank = 0
						} else {
							x.AuxiliaryRank = 1
						}
						aggregateExamples[key] = append(aggregateExamples[key], x)
					}
				}
			}
		}
	}
	for _, mode := range []string{"fp32", "int8", "ternary_ptq", "ternary_ste"} {
		for _, seed := range []uint64{1729, 2718} {
			xs := aggregateExamples[aggregateKey{mode, seed}]
			m := measureClaims(xs, nil, "always_interleave")
			m.AuxiliaryCalls = 0
			for _, x := range xs {
				m.AuxiliaryCalls += x.AuxiliaryRank
			}
			pass := m.MeanRank <= r.Baseline.MeanRank*.9 && m.Recall1 >= r.Baseline.Recall1 && m.Recall10 >= r.Baseline.Recall10
			r.Aggregates = append(r.Aggregates, ClaimAggregate{mode, seed, m, pass})
			if mode == "fp32" && seed == 1729 {
				r.PrimaryPassesExploratoryGate = pass
			}
		}
	}
	return r, nil
}
