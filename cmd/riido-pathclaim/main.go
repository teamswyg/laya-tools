// riido-pathclaim is a bounded maintainer runner, not an automatic router.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/teamswyg/laya-tools/internal/pairlearn"
	"github.com/teamswyg/laya-tools/internal/pathclaim"
	"github.com/teamswyg/laya-tools/internal/pathinput"
)

type config struct {
	phase, out, seal, sealSHA string
	files                     pathinput.Files
}
type report struct {
	Schema, Status, Failure, InputSealSHA256                                                   string
	Input                                                                                      pathinput.Seal
	Coverage                                                                                   [2]pathinput.Coverage
	Repositories                                                                               []pathinput.Coverage
	Fits                                                                                       pathclaim.Fits
	Epochs                                                                                     [10]pairlearn.Trace
	Selections                                                                                 [2]pathclaim.ValidationSelection
	Controls                                                                                   [2]pathclaim.Controls
	SelectedAllDevelopment                                                                     [2]pathclaim.Evaluation
	Usefulness                                                                                 pathclaim.PilotGates
	ModelArtifacts                                                                             []modelArtifact
	CorpusFittingAttempted, TrainingPerformed, FixedFitAndPolicyReplayComplete                 bool
	ActualSelectedSearchExecuted, FinalScored, GPUUsed, PublicReleaseApproved, ProductionReady bool
}

func main() {
	if e := run(os.Args[1:]); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func parse(args []string) (config, error) {
	var c config
	f := flag.NewFlagSet("riido-pathclaim", flag.ContinueOnError)
	f.StringVar(&c.phase, "phase", "", "prepare or fit; neither approves release")
	f.StringVar(&c.out, "out", "", "new private directory under .cache")
	f.StringVar(&c.files.Source.Path, "source", "", "private reviewed source evidence")
	f.StringVar(&c.files.Source.SHA256, "source-sha256", "", "reviewed source evidence SHA256")
	f.StringVar(&c.files.Costs.Path, "costs", "", "private measured paired numeric costs")
	f.StringVar(&c.files.Costs.SHA256, "costs-sha256", "", "paired numeric costs SHA256")
	f.StringVar(&c.files.Checkpoint.Path, "checkpoint", "", "private frozen availability/target evidence")
	f.StringVar(&c.files.Checkpoint.SHA256, "checkpoint-sha256", "", "availability/target evidence SHA256")
	f.StringVar(&c.seal, "seal", "", "committed experiments/path-cost-claim/input-46.json; fit only")
	f.StringVar(&c.sealSHA, "seal-sha256", "", "committed input seal SHA256; fit only")
	if e := f.Parse(args); e != nil {
		return c, e
	}
	if f.NArg() != 0 || (c.phase != "prepare" && c.phase != "fit") || !privatePath(c.out) {
		return c, fmt.Errorf("require prepare/fit phase and new private output")
	}
	for _, input := range []pathinput.File{c.files.Source, c.files.Costs, c.files.Checkpoint} {
		if input.Path == "" || !digest(input.SHA256, 64) {
			return c, fmt.Errorf("require all pinned private inputs")
		}
	}
	if c.phase == "fit" {
		if c.seal != "experiments/path-cost-claim/input-46.json" || !digest(c.sealSHA, 64) {
			return c, fmt.Errorf("fit requires the pinned committed experiment46 input seal")
		}
	} else if c.seal != "" || c.sealSHA != "" {
		return c, fmt.Errorf("prepare does not consume a fitting seal")
	}
	return c, nil
}
func digest(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'f' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}
func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func privatePath(p string) bool {
	return !filepath.IsAbs(p) && filepath.Clean(p) == p && strings.HasPrefix(p, ".cache/") && p != ".cache/"
}
func runnerRevision() (string, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", fmt.Errorf("runner build metadata missing")
	}
	rev, modified := "", ""
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			modified = s.Value
		}
	}
	if !digest(rev, 40) || modified != "false" {
		return "", fmt.Errorf("build the runner from a clean committed revision before preparing or fitting")
	}
	return rev, nil
}
func committedSeal(file pathinput.File) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "show", "HEAD:"+file.Path)
	b := boundedBuffer{limit: 64 << 10}
	cmd.Stdout = &b
	if e := cmd.Run(); e != nil || hash(b.Bytes()) != file.SHA256 {
		return fmt.Errorf("input seal must match committed HEAD bytes before fitting")
	}
	return nil
}

type boundedBuffer struct {
	buffer bytes.Buffer
	limit  int
}

func (b *boundedBuffer) Len() int      { return b.buffer.Len() }
func (b *boundedBuffer) Bytes() []byte { return b.buffer.Bytes() }

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.Len() {
		return 0, fmt.Errorf("committed input exceeds byte limit")
	}
	return b.buffer.Write(p)
}

func fixedPlan() error {
	f, e := os.Open("experiments/path-cost-claim/plan-46.json")
	if e != nil {
		return fmt.Errorf("frozen experiment46 plan unavailable")
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, (64<<10)+1))
	if e != nil || len(b) > 64<<10 || hash(b) != pathclaim.PlanSHA256 {
		return fmt.Errorf("frozen experiment46 plan mismatch")
	}
	return nil
}

func run(args []string) error {
	c, e := parse(args)
	if e != nil {
		return e
	}
	if _, e = os.Lstat(c.out); !os.IsNotExist(e) {
		return fmt.Errorf("output must be new")
	}
	if e = fixedPlan(); e != nil {
		return e
	}
	rev, e := runnerRevision()
	if e != nil {
		return e
	}
	p, e := pathinput.Load(c.files, pathinput.Caches{Roots: ".cache/training-roots-38", Blobs: ".cache/training-license-blobs-38", Subtrees: ".cache/training-subtrees-38"}, ".cache/training-partition-37/membership.json")
	if e != nil {
		return e
	}
	seal, e := pathinput.MakeSeal(p, c.files, rev)
	if e != nil {
		return e
	}
	r := report{Schema: "riido-path-claim-experiment46-v1", Status: "prepared_no_fit", Input: seal, Coverage: p.Coverage, Repositories: p.Repositories}
	if c.phase == "prepare" {
		b, e := marshal(seal)
		if e != nil {
			return e
		}
		r.InputSealSHA256 = hash(b)
		if _, e = seal.Readiness(r.InputSealSHA256); e != nil {
			r.Status = "prepared_insufficient_readiness"
			r.Failure = e.Error()
		}
		if e = writeOutput(c.out, r, []artifact{{"seal.json", b}}); e != nil {
			return e
		}
		return printSummary(r)
	}
	file := pathinput.File{Path: c.seal, SHA256: c.sealSHA}
	if e = committedSeal(file); e != nil {
		return e
	}
	committed, e := pathinput.ReadSeal(file)
	if e != nil {
		return e
	}
	if !reflect.DeepEqual(committed, seal) {
		return fmt.Errorf("actual verified inputs or runner differ from committed seal")
	}
	ready, e := seal.Readiness(file.SHA256)
	if e != nil {
		return e
	}
	if e = pathclaim.VerifyReadiness(p.Rows, ready); e != nil {
		return e
	}
	r.InputSealSHA256 = file.SHA256
	r.Status = "fitting"
	r.CorpusFittingAttempted = true
	r.Fits, r.Epochs, e = pathclaim.FitWithTrace(p.Rows, ready)
	for _, t := range r.Fits.Trials {
		if t.Fitted {
			r.TrainingPerformed = true
		}
	}
	if e != nil {
		r.Status = "registered_fit_failure"
		r.Failure = e.Error()
		if w := writeOutput(c.out, r, nil); w != nil {
			return w
		}
		return e
	}
	artifacts, e := evaluate(p, ready, &r)
	if e != nil {
		r.Status = "registered_evaluation_failure"
		r.Failure = e.Error()
		if w := writeOutput(c.out, r, artifacts); w != nil {
			return w
		}
		return e
	}
	r.FixedFitAndPolicyReplayComplete = true
	r.Status = "pilot_gate_failed"
	if r.Usefulness.BothPass {
		r.Status = "pilot_gate_passed_confirmation_pending"
	}
	if e = writeOutput(c.out, r, artifacts); e != nil {
		return e
	}
	return printSummary(r)
}

func evaluationRows(p pathinput.Prepared) ([]pathclaim.EvaluationRow, []pathclaim.EvaluationRow) {
	all := make([]pathclaim.EvaluationRow, len(p.Rows))
	var validation []pathclaim.EvaluationRow
	for i, row := range p.Rows {
		c := p.Costs[i]
		all[i] = pathclaim.EvaluationRow{Row: row, Baseline: c.Baseline, Helper: c.Helper, WorkKnown: true, Work: c.Work}
		if row.Role == "validation" {
			validation = append(validation, all[i])
		}
	}
	return all, validation
}
func scores(rows []pathclaim.EvaluationRow, trial pathclaim.Trial, ready pathclaim.Readiness, placeholder pathclaim.Threshold) ([]float64, error) {
	model, e := pathclaim.NewModel(trial, ready, placeholder)
	if e != nil {
		return nil, e
	}
	out := make([]float64, len(rows))
	for i, row := range rows {
		out[i], e = model.Score(row.Row.Features)
		if e != nil {
			return nil, e
		}
	}
	return out, nil
}
func evaluate(p pathinput.Prepared, ready pathclaim.Readiness, r *report) ([]artifact, error) {
	all, validation := evaluationRows(p)
	baseline, e := pathclaim.Evaluate(validation, make([]bool, len(validation)))
	if e != nil {
		return nil, e
	}
	placeholder := pathclaim.Threshold{Disabled: true, BudgetPercent: 90, ValidationQuestions: len(validation), ValidationPages: baseline.Total.Pages}
	var artifacts []artifact
	var winners [2]pathclaim.CandidateEvaluation
	for si, seed := range pathclaim.Seeds() {
		var candidates [5]pathclaim.ScoredCandidate
		for pi, lambda := range pathclaim.Penalties() {
			trial := r.Fits.Trials[pi*2+si]
			if !trial.Fitted || trial.Seed != seed || trial.Lambda != lambda {
				return artifacts, fmt.Errorf("registered trial order or success mismatch")
			}
			ss, e := scores(validation, trial, ready, placeholder)
			if e != nil {
				return artifacts, e
			}
			candidates[pi] = pathclaim.ScoredCandidate{Lambda: lambda, Seed: seed, Epoch: trial.Epoch, Scores: ss}
		}
		r.Selections[si], e = pathclaim.SelectValidation(validation, ready, candidates)
		if e != nil {
			return artifacts, e
		}
		r.Controls[si], e = pathclaim.ControlsForValidation(validation, ready, candidates[0])
		if e != nil {
			return artifacts, e
		}
		selected := r.Selections[si].Selected
		winners[si] = r.Selections[si].Candidates[selected]
		for pi := range candidates {
			trial := r.Fits.Trials[pi*2+si]
			threshold := r.Selections[si].Candidates[pi].Threshold
			model, e := pathclaim.NewModel(trial, ready, threshold)
			if e != nil {
				return artifacts, e
			}
			b, e := marshal(model)
			if e != nil {
				return artifacts, e
			}
			if _, e = pathclaim.DecodeModel(b); e != nil {
				return artifacts, e
			}
			artifacts = append(artifacts, artifact{fmt.Sprintf("model-penalty%d-seed%d.json", pi, seed), b})
			r.ModelArtifacts = append(r.ModelArtifacts, modelArtifact{Name: artifacts[len(artifacts)-1].Name, SHA256: hash(b), FileBytes: len(b), FP32CoefficientPayloadBytes: 16 * 4})
		}
		trial := r.Fits.Trials[selected*2+si]
		ss, e := scores(all, trial, ready, placeholder)
		if e != nil {
			return artifacts, e
		}
		r.SelectedAllDevelopment[si], e = pathclaim.EvaluateThreshold(all, ss, winners[si].Threshold)
		if e != nil {
			return artifacts, e
		}
	}
	r.Usefulness, e = pathclaim.PilotGate(ready, r.Controls[0].Baseline, r.Controls[0].AlwaysHelper, winners)
	return artifacts, e
}

type artifact struct {
	Name  string
	Bytes []byte
}

// Coefficient payload excludes the decoded model wrapper, feature buffers,
// allocator overhead and the complete preparation/training pipeline.
type modelArtifact struct {
	Name, SHA256                           string
	FileBytes, FP32CoefficientPayloadBytes int
}

func marshal(v any) ([]byte, error) {
	b, e := json.MarshalIndent(v, "", "  ")
	return append(b, '\n'), e
}
func writeOutput(out string, r report, artifacts []artifact) error {
	b, e := marshal(r)
	if e != nil {
		return e
	}
	for _, a := range artifacts {
		if filepath.Base(a.Name) != a.Name || a.Name == "." || a.Name == "results.json" {
			return fmt.Errorf("artifact file name invalid")
		}
	}
	// Validate every output before creating the destination. Every parent must
	// be an ordinary directory so a symlink cannot publish private coefficients.
	parent := filepath.Dir(out)
	parts := strings.Split(parent, string(filepath.Separator))
	current := ""
	for _, part := range parts {
		current = filepath.Join(current, part)
		st, e := os.Lstat(current)
		if os.IsNotExist(e) {
			if e = os.Mkdir(current, 0700); e != nil {
				return e
			}
			continue
		}
		if e != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("private output parent is not an ordinary directory")
		}
	}
	if e = os.Mkdir(out, 0700); e != nil {
		return e
	}
	for _, a := range artifacts {
		if e = os.WriteFile(filepath.Join(out, a.Name), a.Bytes, 0600); e != nil {
			return e
		}
	}
	return os.WriteFile(filepath.Join(out, "results.json"), b, 0600)
}
func printSummary(r report) error {
	summary := struct {
		Status                                                                                  string
		Training, Validation                                                                    int
		FittingAttempted, FixedFitAndPolicyReplayComplete, PilotBothPass, PublicReleaseApproved bool
	}{r.Status, r.Input.TrainingEligible, r.Input.ValidationEligible, r.CorpusFittingAttempted, r.FixedFitAndPolicyReplayComplete, r.Usefulness.BothPass, false}
	b, e := marshal(summary)
	if e != nil {
		return e
	}
	_, e = os.Stdout.Write(bytes.TrimSpace(b))
	if e == nil {
		_, e = os.Stdout.Write([]byte{'\n'})
	}
	return e
}
