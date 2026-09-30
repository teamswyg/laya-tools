// Package layaprobe evaluates a frozen choice scorer on validation data only.
package layaprobe

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/teamswyg/laya-tools/internal/inference"
	"github.com/teamswyg/laya-tools/internal/paireval"
)

const PlanSHA256 = "c5ba89f7c531f5abf7cd8d13f865e38a5177d6838f9065836e61f335a3fd191d"
const DocumentedPlanSHA256 = "f53e45a434da731202172cbe185224148af6b498be2b8efdfb42195955a80010"
const ModelSHA256 = "0e3665bcf1aa7d8b224af93b5e10d8c86796227f7741da92f6c8a786364a44db"
const Instruction = "Which statement best describes the relationship between the request and the code?"

var choices = [2]string{"The code implements the request.", "The code does not implement the request."}

type Predictor interface {
	Predict(string, string, string, []string) (inference.Prediction, error)
}
type Report struct {
	Protocol                                                                                             string
	AveragedAUC                                                                                          *float64
	Schema, PlanSHA256, SourceSHA256, MembershipSHA256, ModelSHA256, ScoreTraceSHA256                    string
	ValidationCases, ScopeExcluded, TokenizerExcluded, Eligible, NativeCalls, OrderDecisionDisagreements int
	PrimaryAUC, ForwardAUC, ReverseAUC, BM25AUC                                                          *float64
	MeanAbsoluteOrderGap, NativeCallP50MS, NativeCallP95MS, ProbeSeconds                                 float64
	PassesExploratoryGate, ProductionReady                                                               bool
}

func probability(p inference.Prediction, index int) (float64, error) {
	if len(p.Probabilities) != 2 {
		return 0, fmt.Errorf("expected two probabilities")
	}
	sum := 0.
	for _, v := range p.Probabilities {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			return 0, fmt.Errorf("invalid probability")
		}
		sum += v
	}
	if math.Abs(sum-1) > 1e-6 {
		return 0, fmt.Errorf("probabilities do not sum to one")
	}
	return p.Probabilities[index], nil
}

// guard checks resources between cases, not inside a native inference call.
func Run(rows []paireval.Row, split paireval.Split, p Predictor, guard func() error, progress func(int, int)) (Report, error) {
	return RunProtocol(rows, split, p, guard, progress, "choice")
}

func input(x paireval.Row, protocol string) (string, string, [2]string, int) {
	if protocol == "noul" {
		return "File: [unavailable]\n" + x.Code, "Is this source code relevant to the software change: \"" + x.Query + "\"?", [2]string{"false: no, the statement does not hold", "true: yes, the statement holds"}, 1
	}
	return "Request:\n" + x.Query + "\n\nCode:\n" + x.Code, Instruction, choices, 0
}

func RunProtocol(rows []paireval.Row, split paireval.Split, p Predictor, guard func() error, progress func(int, int), protocol string) (Report, error) {
	start := time.Now()
	r := Report{Schema: "riido-laya-relevance-v1", PlanSHA256: PlanSHA256, SourceSHA256: paireval.SourceSHA, MembershipSHA256: split.MembershipSHA256, ModelSHA256: ModelSHA256}
	if protocol != "choice" && protocol != "noul" {
		return r, fmt.Errorf("unknown protocol")
	}
	r.Protocol = protocol
	if protocol == "noul" {
		r.PlanSHA256 = DocumentedPlanSHA256
	}
	if len(rows) != len(split.Rows) {
		return r, fmt.Errorf("membership length mismatch")
	}
	lexical := paireval.NewLexical(rows, split)
	var primary, forward, reverse, baseline []paireval.Prediction
	var times []float64
	trace := sha256.New()
	enc := json.NewEncoder(trace)
	for i, x := range rows {
		if split.Rows[i].Split != "validation" {
			continue
		}
		r.ValidationCases++
		if !paireval.InScope(x) {
			r.ScopeExcluded++
			continue
		}
		if x.Label == nil || (*x.Label != 0 && *x.Label != 1) {
			return r, fmt.Errorf("invalid validation label")
		}
		if err := guard(); err != nil {
			return r, err
		}
		state, instruction, options, positiveIndex := input(x, protocol)
		a, err := p.Predict(state, protocol, instruction, []string{options[0], options[1]})
		if err != nil {
			return r, err
		}
		r.NativeCalls++
		if err := guard(); err != nil {
			return r, err
		}
		b, err := p.Predict(state, protocol, instruction, []string{options[1], options[0]})
		if err != nil {
			return r, err
		}
		r.NativeCalls++
		av, err := probability(a, positiveIndex)
		if err != nil {
			return r, err
		}
		bv, err := probability(b, 1-positiveIndex)
		if err != nil {
			return r, err
		}
		if err := enc.Encode(struct {
			Row              int
			Forward, Reverse float64
			Truncated        bool
		}{i, av, bv, a.Truncated || b.Truncated}); err != nil {
			return r, err
		}
		times = append(times, a.MS, b.MS)
		if a.Truncated || b.Truncated {
			r.TokenizerExcluded++
		} else {
			r.Eligible++
			r.MeanAbsoluteOrderGap += math.Abs(av - bv)
			if (av >= .5) != (bv >= .5) {
				r.OrderDecisionDisagreements++
			}
			pred := paireval.Prediction{Row: i, Group: split.Rows[i].Group, Label: *x.Label, Eligible: true}
			pred.Score = (av + bv) / 2
			primary = append(primary, pred)
			pred.Score = av
			forward = append(forward, pred)
			pred.Score = bv
			reverse = append(reverse, pred)
			pred.Score = lexical.Score(x.Query, x.Code)
			baseline = append(baseline, pred)
		}
		progress(r.ValidationCases, r.NativeCalls)
	}
	if err := guard(); err != nil {
		return r, err
	}
	if r.Eligible > 0 {
		r.MeanAbsoluteOrderGap /= float64(r.Eligible)
	}
	r.PrimaryAUC = paireval.AUC(primary)
	r.AveragedAUC = r.PrimaryAUC
	r.ForwardAUC = paireval.AUC(forward)
	if protocol == "noul" {
		r.PrimaryAUC = r.ForwardAUC
	}
	r.ReverseAUC = paireval.AUC(reverse)
	r.BM25AUC = paireval.AUC(baseline)
	r.PassesExploratoryGate = r.PrimaryAUC != nil && r.BM25AUC != nil && *r.PrimaryAUC >= .60 && *r.PrimaryAUC >= *r.BM25AUC+.03
	slices.Sort(times)
	if len(times) > 0 {
		r.NativeCallP50MS = times[(len(times)-1)/2]
		r.NativeCallP95MS = times[int(float64(len(times)-1)*.95)]
	}
	r.ScoreTraceSHA256 = hex.EncodeToString(trace.Sum(nil))
	r.ProbeSeconds = time.Since(start).Seconds()
	return r, nil
}
