// riido-behavioraudit evaluates public authored development controls. It never
// partitions, trains, invokes an external model, or activates a policy.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/teamswyg/laya-tools/internal/behaviorprobe"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

type auditPlan struct {
	Schema                  string      `json:"schema"`
	State                   string      `json:"state"`
	InputSHA256             string      `json:"input_sha256"`
	SourceArtifactSHA256    string      `json:"source_artifact_sha256"`
	PreparationPlanSHA256   string      `json:"preparation_plan_sha256"`
	Parents                 int         `json:"parents"`
	MinConnectedGroups      int         `json:"min_connected_groups"`
	MinPossibleRelativeGain float64     `json:"min_possible_relative_gain"`
	CPUThreads              int         `json:"cpu_threads"`
	BenchmarkInputSHA256    string      `json:"benchmark_input_sha256"`
	BenchmarkIterations     int         `json:"benchmark_iterations"`
	BenchmarkColdRepeats    int         `json:"benchmark_cold_repeats"`
	BenchmarkBinarySHA256   string      `json:"benchmark_binary_sha256"`
	BenchmarkBuildGo        string      `json:"benchmark_build_go"`
	BenchmarkCanonicalSHA   string      `json:"benchmark_canonical_input_sha256"`
	ProfileIterations       int         `json:"instrumented_profile_iterations"`
	ProfileReplaysMax       int         `json:"instrumented_profile_replays_max"`
	InferenceRSSMax         int64       `json:"inference_rss_target_bytes"`
	InferenceP95MaxNS       int64       `json:"inference_p95_target_ns"`
	Fits                    int         `json:"fits_authorized_in_this_plan"`
	Policy                  string      `json:"policy"`
	ImplementationFiles     []sourcePin `json:"implementation_files"`
}

type sourcePin struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type metrics struct {
	Kind               string `json:"kind"`
	KnownRequests      int    `json:"known_including_no_answer_requests"`
	AnswerableRequests int    `json:"answerable_requests"`
	NoAnswerRequests   int    `json:"no_answer_requests"`
	Checks             int    `json:"known_including_no_answer_checks"`
	AnswerableChecks   int    `json:"answerable_checks"`
	Top1Correct        int    `json:"answerable_top1_correct"`
	Top3Correct        int    `json:"answerable_top3_correct"`
	RuleFallbacks      int    `json:"rule_fallbacks_all_original_requests"`
}

type parentRanking struct {
	ParentID string                `json:"parent_id"`
	Rankings [4]shortclaim.Ranking `json:"rankings"`
}

type report struct {
	Schema                 string               `json:"schema"`
	Stage                  string               `json:"stage"`
	PlanSHA256             string               `json:"plan_sha256"`
	InputSHA256            string               `json:"input_sha256"`
	Truth                  behaviorprobe.Report `json:"truth_and_groups"`
	Metrics                [4]metrics           `json:"baselines"`
	ParentRankings         []parentRanking      `json:"parent_rankings"`
	OracleChecks           int                  `json:"oracle_known_checks"`
	OracleAnswerableChecks int                  `json:"oracle_answerable_checks"`
	BestBaseline           string               `json:"best_nonlearned_baseline"`
	PossibleRelativeGain   float64              `json:"possible_relative_gain_known"`
	AnswerableRelativeGain float64              `json:"possible_relative_gain_answerable"`
	GroupGatePass          bool                 `json:"connected_group_gate_pass"`
	UtilityUpperBoundPass  bool                 `json:"utility_upper_bound_gate_pass"`
	StopReasons            []string             `json:"stop_reasons"`
	Partitions             int                  `json:"partitions"`
	Fits                   int                  `json:"fits"`
	Weights                int                  `json:"weight_artifacts"`
	ModelCalls             int                  `json:"model_calls"`
	TrainingExecutionReady bool                 `json:"training_execution_ready"`
	ProductionActivation   bool                 `json:"production_activation"`
	FinalEligible          bool                 `json:"final_eligible"`
	ProtectedFinalRead     bool                 `json:"protected_final_read"`
	Scope                  string               `json:"scope"`
}

func hash(raw []byte) string {
	x := sha256.Sum256(raw)
	return hex.EncodeToString(x[:])
}

func readBounded(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.New("file_read_failed")
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || int64(len(b)) > limit {
		return nil, errors.New("file_read_bounds_or_io")
	}
	return b, nil
}

func evaluate(d behaviorprobe.Dataset, truth behaviorprobe.Report, plan auditPlan) (report, error) {
	if len(d.Parents) != plan.Parents || len(truth.Outcomes) != len(d.Parents) {
		return report{}, errors.New("parent_truth_count_mismatch")
	}
	r := report{Schema: "riido-shortclaim-development-result-v1", Stage: "unpartitioned_nonlearned_diagnostics_only", Truth: truth, Scope: "Original authored short finite contracts, one author; code truth and natural-language fidelity are separate. No held-out inference, learned performance, actual LLM savings or final eligibility."}
	for i, parent := range d.Parents {
		outcome := truth.Outcomes[i]
		if outcome.ParentID != parent.ID {
			return report{}, errors.New("parent_truth_identity_mismatch")
		}
		input := shortclaim.Input{Schema: shortclaim.Schema, Request: parent.Request, Provenance: "original-development-56"}
		for j, c := range parent.Candidates {
			// The predictable local ID is metadata only. Provenance and original
			// source IDs, prototype, labels, truth vectors and evaluation mode
			// never enter the request/candidate text used for ranking.
			input.Candidates = append(input.Candidates, shortclaim.Candidate{ID: fmt.Sprintf("candidate-%d", j), Text: c.Text})
		}
		p, err := shortclaim.Validate(input)
		if err != nil {
			return report{}, err
		}
		all, err := shortclaim.Baselines(p)
		if err != nil {
			return report{}, err
		}
		r.ParentRankings = append(r.ParentRankings, parentRanking{parent.ID, all})
		var allowed [shortclaim.MaxCandidates]bool
		for _, index := range outcome.Acceptable {
			if index < 0 || index >= p.Count || allowed[index] {
				return report{}, errors.New("invalid_acceptable_set")
			}
			allowed[index] = true
		}
		switch outcome.State {
		case "known":
			if len(outcome.Acceptable) == 0 {
				return report{}, errors.New("known_without_answer")
			}
			r.OracleChecks++
			r.OracleAnswerableChecks++
		case "no_answer":
			if len(outcome.Acceptable) != 0 {
				return report{}, errors.New("no_answer_with_answer")
			}
			r.OracleChecks += p.Count
		case "unknown":
			if len(outcome.Acceptable) != 0 {
				return report{}, errors.New("unknown_with_answer")
			}
		default:
			return report{}, errors.New("unrecognized_truth_state")
		}
		for j, ranking := range all {
			m := &r.Metrics[j]
			m.Kind = ranking.Kind
			if ranking.FallbackReason != "" {
				m.RuleFallbacks++
			}
			if outcome.State == "unknown" {
				continue
			}
			m.KnownRequests++
			checks := p.Count
			if outcome.State == "known" {
				m.AnswerableRequests++
				for k := 0; k < ranking.Count; k++ {
					if allowed[ranking.Order[k]] {
						checks = k + 1
						break
					}
				}
				m.AnswerableChecks += checks
				if allowed[ranking.Order[0]] {
					m.Top1Correct++
				}
				for k := 0; k < min(3, ranking.Count); k++ {
					if allowed[ranking.Order[k]] {
						m.Top3Correct++
						break
					}
				}
			} else {
				m.NoAnswerRequests++
			}
			m.Checks += checks
		}
	}
	best := 0
	for i := 1; i < len(r.Metrics); i++ {
		if r.Metrics[i].Checks < r.Metrics[best].Checks {
			best = i
		}
	}
	m := r.Metrics[best]
	if m.Checks < 1 || m.AnswerableChecks < 1 {
		return report{}, errors.New("no_valid_utility_denominator")
	}
	r.BestBaseline = m.Kind
	r.PossibleRelativeGain = float64(m.Checks-r.OracleChecks) / float64(m.Checks)
	r.AnswerableRelativeGain = float64(m.AnswerableChecks-r.OracleAnswerableChecks) / float64(m.AnswerableChecks)
	r.GroupGatePass = truth.ConnectedGroups >= plan.MinConnectedGroups
	r.UtilityUpperBoundPass = r.PossibleRelativeGain >= plan.MinPossibleRelativeGain
	if !r.GroupGatePass {
		r.StopReasons = append(r.StopReasons, "insufficient_connected_groups")
	}
	if !r.UtilityUpperBoundPass {
		r.StopReasons = append(r.StopReasons, "insufficient_oracle_utility_upper_bound")
	}
	// Passing these diagnostic checks alone cannot activate fitting. This plan
	// authorizes zero fits and freezes neither roles nor a training execution.
	r.StopReasons = append(r.StopReasons, "no_partition_or_training_execution_plan")
	return r, nil
}

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("riido-behavioraudit", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	input := fs.String("input", "experiments/short-claim/probes-56.json", "pinned original public development fixture")
	planFile := fs.String("plan", "experiments/short-claim/execution-plan-56a.json", "pre-outcome nonlearned diagnostic plan")
	planSHA := fs.String("plan-sha256", "", "required exact frozen plan SHA-256")
	out := fs.String("out", "", "new local result directory, never overwritten")
	if err := fs.Parse(args); err != nil {
		return errors.New("invalid_arguments")
	}
	if fs.NArg() != 0 || len(*planSHA) != 64 || *out == "" {
		return errors.New("require_plan_digest_and_new_output_directory")
	}
	planBytes, err := readBounded(*planFile, 64<<10)
	if err != nil {
		return err
	}
	if hash(planBytes) != *planSHA {
		return errors.New("plan_hash_mismatch")
	}
	var plan auditPlan
	dec := json.NewDecoder(bytes.NewReader(planBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&plan); err != nil {
		return errors.New("invalid_plan_json")
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return errors.New("extra_plan_json")
	}
	if plan.Schema != "riido-shortclaim-nonlearned-execution-plan-v1" || plan.State != "nonlearned_diagnostics_only" || plan.Fits != 0 || plan.MinConnectedGroups != 15 || plan.MinPossibleRelativeGain != .05 || plan.CPUThreads != 1 {
		return errors.New("unsupported_execution_plan")
	}
	if len(plan.ImplementationFiles) == 0 || len(plan.ImplementationFiles) > 32 {
		return errors.New("implementation_manifest_bounds")
	}
	for _, pin := range plan.ImplementationFiles {
		if filepath.IsAbs(pin.Path) || filepath.Clean(pin.Path) != pin.Path || len(pin.Path) > 128 || strings.HasPrefix(pin.Path, "../") || !strings.HasSuffix(pin.Path, ".go") {
			return errors.New("invalid_implementation_path")
		}
		info, e := os.Lstat(pin.Path)
		if e != nil || !info.Mode().IsRegular() || info.Size() > 1<<20 {
			return errors.New("invalid_implementation_file")
		}
		b, e := readBounded(pin.Path, 1<<20)
		if e != nil || hash(b) != pin.SHA256 {
			return errors.New("implementation_hash_mismatch")
		}
	}
	preparation, err := readBounded("experiments/short-claim/plan-56.json", 64<<10)
	if err != nil || hash(preparation) != plan.PreparationPlanSHA256 {
		return errors.New("preparation_plan_hash_mismatch")
	}
	raw, err := readBounded(*input, behaviorprobe.MaxDataBytes)
	if err != nil || hash(raw) != plan.InputSHA256 {
		return errors.New("input_hash_mismatch")
	}
	// Decode precisely the bytes whose digest was checked; never reopen input.
	var d behaviorprobe.Dataset
	dd := json.NewDecoder(bytes.NewReader(raw))
	dd.DisallowUnknownFields()
	if dd.Decode(&d) != nil {
		return errors.New("invalid_dataset_json")
	}
	if err := dd.Decode(new(any)); err != io.EOF {
		return errors.New("extra_dataset_json")
	}
	runtime.GOMAXPROCS(1)
	truth, err := behaviorprobe.Evaluate(d)
	if err != nil {
		return err
	}
	if truth.SourceArtifactSHA256 != plan.SourceArtifactSHA256 {
		return errors.New("embedded_source_hash_mismatch")
	}
	r, err := evaluate(d, truth, plan)
	if err != nil {
		return err
	}
	r.PlanSHA256, r.InputSHA256 = *planSHA, hash(raw)
	encoded, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err := os.Mkdir(*out, 0700); err != nil {
		return errors.New("output_directory_create_failed")
	}
	if err := os.WriteFile(filepath.Join(*out, "results.json"), append(encoded, '\n'), 0600); err != nil {
		return errors.New("output_write_failed")
	}
	return json.NewEncoder(stdout).Encode(struct {
		Parents int `json:"parents"`
		Groups  int `json:"connected_groups"`
		Fits    int `json:"fits"`
	}{len(d.Parents), truth.ConnectedGroups, 0})
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
