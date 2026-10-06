// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintfit

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/teamswyg/laya-tools/internal/statehintcorpus"
	"github.com/teamswyg/laya-tools/internal/statehintfamily"
	"github.com/teamswyg/laya-tools/pkg/statehint"
)

type Trial struct {
	Arm         Arm                    `json:"arm"`
	Fit         statehint.FitReport    `json:"fit"`
	Validation  statehintfamily.Report `json:"validation"`
	ArtifactSHA string                 `json:"artifact_sha256"`
}

type Calibration struct {
	Temperature float64                `json:"temperature"`
	Report      statehintfamily.Report `json:"report"`
}

type ForwardCounts struct {
	Validation   int `json:"validation_trials"`
	Calibration  int `json:"calibration_grid"`
	ReloadParity int `json:"selected_and_reloaded_validation_calibration"`
	ChildTest    int `json:"child_test"`
	ParentTest   int `json:"parent_test"`
	RuleCalls    int `json:"speechact_rule_calls"`
}

type StudyResult struct {
	Schema             string                 `json:"schema"`
	Status             string                 `json:"status"`
	PlanSHA            string                 `json:"plan_sha256"`
	Trials             []Trial                `json:"trials"`
	Calibration        []Calibration          `json:"calibration"`
	SelectedArm        int                    `json:"selected_arm_index"`
	Temperature        float64                `json:"temperature"`
	ModelSHA           string                 `json:"model_sha256"`
	Counts             ForwardCounts          `json:"forward_calls"`
	TestOpened         bool                   `json:"test_opened"`
	ChildTest          statehintfamily.Report `json:"child_test"`
	ParentTest         statehintfamily.Report `json:"parent_test"`
	RulesTest          statehintfamily.Report `json:"speechact_control_descriptive"`
	ProductGroundTruth bool                   `json:"product_ground_truth"`
	MutationExecuted   bool                   `json:"mutation_executed"`
}

type beforeTestLock struct {
	Schema        string                 `json:"schema"`
	PlanSHA       string                 `json:"plan_sha256"`
	SourceCommit  string                 `json:"source_commit"`
	BinarySHA     string                 `json:"binary_sha256"`
	Sources       []Pin                  `json:"sources"`
	Partitions    [4]Pin                 `json:"partitions_train_validation_calibration_test"`
	Manifest      Pin                    `json:"family_manifest"`
	Audit         Pin                    `json:"authoring_audit"`
	Rubric        Pin                    `json:"rubric_receipt"`
	Definitions   [2]Pin                 `json:"rubric_definitions_ko_en"`
	Parent        Pin                    `json:"parent"`
	Arm           Arm                    `json:"selected_arm"`
	ModelSHA      string                 `json:"selected_artifact_sha256"`
	Temperature   float64                `json:"temperature"`
	TrainingSteps uint64                 `json:"training_steps"`
	Validation    statehintfamily.Report `json:"validation_at_temperature_one"`
	Calibration   statehintfamily.Report `json:"calibration_at_selected_temperature"`
	Gate          [2]float64             `json:"confidence_margin"`
	TestParsed    bool                   `json:"test_parsed_yet"`
}

type prepared struct {
	plan       Plan
	planBytes  []byte
	planName   string
	manifest   Manifest
	metadata   []statehintcorpus.Family
	partitions [3]statehintcorpus.Corpus
	parent     *statehint.Model
}

func readBounded(root *os.Root, name string, limit int64) ([]byte, error) {
	if !localPath(name) {
		return nil, ErrStudy
	}
	before, err := root.Lstat(name)
	if err != nil || !before.Mode().IsRegular() || before.Mode()&os.ModeSymlink != 0 {
		return nil, ErrStudy
	}
	f, err := root.Open(name)
	if err != nil {
		return nil, ErrStudy
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil || !os.SameFile(before, st) || !st.Mode().IsRegular() || st.Size() < 1 || st.Size() > limit {
		return nil, ErrStudy
	}
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || len(b) == 0 || int64(len(b)) > limit {
		return nil, ErrStudy
	}
	return b, nil
}

// This is metadata-only: no dataset bytes, including test bytes, are read.
// Reject all declared input inode aliases before source/data/model reads.
func inputIdentities(root *os.Root, planName string, p Plan) error {
	names := []string{planName, p.Rubric.Path, p.Definitions[0].Path, p.Definitions[1].Path, p.Manifest.Path, p.Audit.Path, p.Parent.Path}
	for _, pin := range p.Partitions {
		names = append(names, pin.Path)
	}
	for _, pin := range p.Sources {
		names = append(names, pin.Path)
	}
	infos := make([]os.FileInfo, 0, len(names))
	for _, name := range names {
		info, err := root.Lstat(name)
		if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return ErrStudy
		}
		for _, prior := range infos {
			if os.SameFile(prior, info) {
				return ErrStudy
			}
		}
		infos = append(infos, info)
	}
	return nil
}

func readPin(root *os.Root, pin Pin, limit int64) ([]byte, error) {
	b, err := readBounded(root, pin.Path, limit)
	if err != nil || digest(b) != pin.SHA256 {
		return nil, ErrStudy
	}
	return b, nil
}

func executableSHA() (string, error) {
	name, err := os.Executable()
	if err != nil {
		return "", ErrStudy
	}
	f, err := os.Open(name)
	if err != nil {
		return "", ErrStudy
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil || !st.Mode().IsRegular() || st.Size() < 1 || st.Size() > 64<<20 {
		return "", ErrStudy
	}
	b, err := io.ReadAll(io.LimitReader(f, (64<<20)+1))
	if err != nil || len(b) > 64<<20 {
		return "", ErrStudy
	}
	return digest(b), nil
}

func validBuild(info *debug.BuildInfo, commit string) bool {
	if info == nil || info.GoVersion != "go1.27.1" || info.Path != "github.com/teamswyg/laya-tools/cmd/riido-statehint-tune-v4" || info.Main.Path != "github.com/teamswyg/laya-tools" {
		return false
	}
	var vcs, revision, modified string
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs":
			if vcs != "" {
				return false
			}
			vcs = setting.Value
		case "vcs.revision":
			if revision != "" {
				return false
			}
			revision = setting.Value
		case "vcs.modified":
			if modified != "" {
				return false
			}
			modified = setting.Value
		}
	}
	return vcs == "git" && revision == commit && modified == "false"
}

func verifySources(root *os.Root, p Plan) error {
	paths := []string{"go.mod", "go.sum"}
	for _, dir := range []string{"cmd/riido-statehint-tune-v4", "internal/statehintfit", "internal/statehintfamily", "internal/statehintcorpus", "pkg/statehint"} {
		f, err := root.Open(dir)
		if err != nil {
			return ErrStudy
		}
		entries, err := f.ReadDir(-1)
		f.Close()
		if err != nil {
			return ErrStudy
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
				paths = append(paths, path.Join(dir, entry.Name()))
			}
		}
	}
	sort.Strings(paths)
	pins := append([]Pin(nil), p.Sources...)
	sort.Slice(pins, func(i, j int) bool { return pins[i].Path < pins[j].Path })
	if len(paths) != len(pins) {
		return ErrStudy
	}
	for i, name := range paths {
		if pins[i].Path != name {
			return ErrStudy
		}
		if _, err := readPin(root, pins[i], 1<<20); err != nil {
			return err
		}
	}
	return nil
}

func prepare(root *os.Root, planName, binarySHA string) (prepared, error) {
	var ready prepared
	b, err := readBounded(root, planName, 1<<20)
	if err != nil || strictJSON(b, &ready.plan) != nil || validPlan(ready.plan) != nil {
		return ready, ErrStudy
	}
	p := ready.plan
	// The supplied plan may not alias a declared input pathname. os.Root also
	// confines symlink traversal to this explicit workspace root.
	for _, pin := range append([]Pin{p.Rubric, p.Definitions[0], p.Definitions[1], p.Manifest, p.Audit, p.Parent}, p.Partitions[:]...) {
		if pin.Path == planName {
			return prepared{}, ErrStudy
		}
	}
	if inputIdentities(root, planName, p) != nil {
		return prepared{}, ErrStudy
	}
	build, ok := debug.ReadBuildInfo()
	if !ok || !validBuild(build, p.SourceCommit) || p.BinarySHA != binarySHA || runtime.Version() != "go1.27.1" || verifySources(root, p) != nil {
		return prepared{}, ErrStudy
	}
	ready.planBytes = b
	ready.planName = planName
	if _, err = readPin(root, p.Rubric, 1<<20); err != nil {
		return prepared{}, err
	}
	for _, pin := range p.Definitions {
		if _, err = readPin(root, pin, 1<<20); err != nil {
			return prepared{}, err
		}
	}
	manifestBytes, err := readPin(root, p.Manifest, 1<<20)
	if err != nil || strictJSON(manifestBytes, &ready.manifest) != nil {
		return prepared{}, ErrStudy
	}
	ready.metadata, err = validateManifest(ready.manifest)
	if err != nil {
		return prepared{}, err
	}
	auditBytes, err := readPin(root, p.Audit, 1<<20)
	var audit Audit
	if err != nil || strictJSON(auditBytes, &audit) != nil || audit.Schema != "riido-statehint-v4-reviewed-original-audit-v1" || audit.ManifestSHA != p.Manifest.SHA256 || audit.RubricSHA != rubricSHA || !audit.OriginalSynthetic || !audit.PrivateExcluded || !audit.ExposedExcluded || !audit.RightsReviewed || !audit.SemanticReviewed || !audit.BilingualReviewed || !audit.AncestryReviewed || !audit.DisagreementsResolved || !audit.PublicReleaseAllowed {
		return prepared{}, ErrStudy
	}
	parentBytes, err := readPin(root, p.Parent, statehint.ArtifactBytes)
	if err != nil {
		return prepared{}, err
	}
	ready.parent, err = statehint.Load(bytes.NewReader(parentBytes))
	if err != nil || ready.parent.TrainingSteps() != 1920 {
		return prepared{}, ErrStudy
	}
	for i, name := range [3]string{"train", "validation", "calibration"} {
		ready.partitions[i], err = loadPartition(root, p.Partitions[i], name, ready.manifest, ready.metadata)
		if err != nil {
			return prepared{}, err
		}
	}
	// Test bytes are neither hashed nor parsed here. Its expected digest and
	// declared paired membership were inspected only as manifest metadata.
	return ready, nil
}

func loadPartition(root *os.Root, pin Pin, partition string, m Manifest, metadata []statehintcorpus.Family) (statehintcorpus.Corpus, error) {
	b, err := readPin(root, pin, statehintcorpus.MaxFileBytes)
	if err != nil {
		return statehintcorpus.Corpus{}, err
	}
	c, _, err := statehintcorpus.Read(bytes.NewReader(b), statehintcorpus.Options{Partition: partition, RubricSHA: rubricSHA})
	if err != nil || matchRows(c, m, metadata, partition) != nil {
		return statehintcorpus.Corpus{}, ErrStudy
	}
	return c, nil
}

type journal struct {
	root *os.Root
	used int
}

func (j *journal) write(name string, b []byte) error {
	if name == "" || len(name) > 128 || path.Base(name) != name || strings.Contains(name, "..") || len(b) == 0 || j.used+len(b) > 8<<20 {
		return ErrStudy
	}
	for _, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.') {
			return ErrStudy
		}
	}
	f, err := j.root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return ErrStudy
	}
	// Reserve before the first byte: a partial failed file also consumes budget.
	j.used += len(b)
	n, err := f.Write(b)
	if err == nil && n != len(b) {
		err = io.ErrShortWrite
	}
	if err == nil {
		err = f.Sync()
	}
	closed := f.Close()
	if err != nil || closed != nil {
		return ErrStudy
	}
	dir, err := j.root.Open(".")
	if err != nil {
		return ErrStudy
	}
	err = dir.Sync()
	dir.Close()
	if err != nil {
		return ErrStudy
	}
	return nil
}

// Reject symlink ancestry before opening each directory and bind the opened
// handle to that observed inode. The private anchor and run must be0700;
// existing outer cache directories may be0755 but not group/other writable.
func openObservedDirectory(parent *os.Root, name string, private bool) (*os.Root, error) {
	before, err := parent.Lstat(name)
	if err != nil || !before.IsDir() || before.Mode()&os.ModeSymlink != 0 || before.Mode().Perm()&0022 != 0 || private && before.Mode().Perm() != 0700 {
		return nil, ErrStudy
	}
	child, err := parent.OpenRoot(name)
	if err != nil {
		return nil, ErrStudy
	}
	f, err := child.Open(".")
	if err != nil {
		child.Close()
		return nil, ErrStudy
	}
	opened, err := f.Stat()
	f.Close()
	after, afterErr := parent.Lstat(name)
	if err != nil || afterErr != nil || !os.SameFile(before, opened) || !os.SameFile(before, after) || after.Mode()&os.ModeSymlink != 0 {
		child.Close()
		return nil, ErrStudy
	}
	return child, nil
}

func syncDirectory(root *os.Root) error {
	f, err := root.Open(".")
	if err != nil {
		return ErrStudy
	}
	err = f.Sync()
	closed := f.Close()
	if err != nil || closed != nil {
		return ErrStudy
	}
	return nil
}

func privateOutput(root *os.Root, name string) (*os.Root, error) {
	const prefix = ".cache/statehint/v4/"
	if !strings.HasPrefix(name, prefix) || !opaque(strings.TrimPrefix(name, prefix)) {
		return nil, ErrStudy
	}
	cache, err := openObservedDirectory(root, ".cache", false)
	if err != nil {
		return nil, err
	}
	defer cache.Close()
	state, err := openObservedDirectory(cache, "statehint", false)
	if err != nil {
		return nil, err
	}
	defer state.Close()
	if err = state.Mkdir("v4", 0700); err != nil && !os.IsExist(err) {
		return nil, ErrStudy
	}
	if syncDirectory(state) != nil {
		return nil, ErrStudy
	}
	anchor, err := openObservedDirectory(state, "v4", true)
	if err != nil {
		return nil, err
	}
	defer anchor.Close()
	run := strings.TrimPrefix(name, prefix)
	if anchor.Mkdir(run, 0700) != nil {
		return nil, ErrStudy
	}
	if syncDirectory(anchor) != nil {
		return nil, ErrStudy
	}
	return openObservedDirectory(anchor, run, true)
}

func (j *journal) json(name string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return ErrStudy
	}
	return j.write(name, append(b, '\n'))
}

func modelBytes(model *statehint.Model) ([]byte, error) {
	var out bytes.Buffer
	if model.Save(&out) != nil {
		return nil, ErrStudy
	}
	return out.Bytes(), nil
}

func predict(model *statehint.Model, corpus statehintcorpus.Corpus) ([]statehint.Prediction, error) {
	rows := corpus.Rows()
	values := make([]statehint.Prediction, len(rows))
	var workspace statehint.Workspace
	for i, row := range rows {
		p, err := model.Predict(row.Text, &workspace)
		if err != nil {
			return nil, ErrStudy
		}
		values[i] = p
	}
	return values, nil
}

// Earlier index wins exact ties; all-abstain/ineligible candidates never win.
func choose(trials []Trial) int {
	best := -1
	for i, trial := range trials {
		if !trial.Validation.Eligible {
			continue
		}
		if best < 0 || trial.Validation.SeverityCost < trials[best].Validation.SeverityCost || trial.Validation.SeverityCost == trials[best].Validation.SeverityCost && trial.Validation.EightNLL < trials[best].Validation.EightNLL {
			best = i
		}
	}
	return best
}

// This is the single final-test boundary. Failure/nonqualification cannot call
// the test reader. The production callback saves, fsyncs and rereads the lock.
func finalStage(validation, calibration statehintfamily.Report, lock func() error, test func() error) (bool, error) {
	if !validation.Eligible || !calibration.Eligible {
		return false, nil
	}
	if lock == nil || test == nil {
		return false, ErrStudy
	}
	if err := lock(); err != nil {
		return false, err
	}
	return true, test()
}

func fitStudy(root *os.Root, j *journal, ready prepared) (StudyResult, error) {
	p := ready.plan
	result := StudyResult{Schema: "riido-statehint-v4-research-result-v1", Status: "unqualified_validation", PlanSHA: digest(ready.planBytes), SelectedArm: -1}
	if err := j.write("PLAN.before-fit.json", ready.planBytes); err != nil {
		return result, err
	}
	rows := ready.partitions[0].Rows()
	samples := make([]statehint.Sample, len(rows))
	for i, row := range rows {
		samples[i] = statehint.Sample{Text: row.Text, Label: row.Expected}
	}
	var models [4]*statehint.Model
	for i, arm := range p.Arms {
		model := statehint.NewModel()
		if arm.Warm {
			model = ready.parent.Clone()
		}
		fit, err := model.Fit(samples, statehint.FitOptions{Epochs: p.Epochs, BatchSize: p.Batch, LearningRate: arm.Rate, WeightDecay: p.Decay, Seed: p.Seed})
		wantSteps := uint64(2120)
		if arm.Warm {
			wantSteps += 1920
		}
		if err != nil || fit.Batches != 2120 || fit.TrainingSteps != wantSteps || model.Temperature() != 1 {
			return result, ErrStudy
		}
		values, err := predict(model, ready.partitions[1])
		if err != nil {
			return result, err
		}
		result.Counts.Validation += len(values)
		report, err := statehintfamily.Evaluate(ready.partitions[1], values)
		if err != nil {
			return result, err
		}
		artifact, err := modelBytes(model)
		if err != nil {
			return result, err
		}
		if err = j.write(arm.Name+".rsh", artifact); err != nil {
			return result, err
		}
		if err = j.json(arm.Name+".validation.json", values); err != nil {
			return result, err
		}
		result.Trials = append(result.Trials, Trial{arm, fit, report, digest(artifact)})
		models[i] = model
	}
	selectedIndex := choose(result.Trials)
	result.SelectedArm = selectedIndex
	if err := j.json("TRIALS.before-calibration.json", result); err != nil {
		return result, err
	}
	if selectedIndex < 0 {
		return result, j.json("RESULTS.first.json", result)
	}
	selected := models[selectedIndex]
	bestCalibration := -1
	var bestValues []statehint.Prediction
	for step := 5; step <= 50; step++ {
		temperature := float64(step) / 10
		if selected.SetTemperature(temperature) != nil {
			return result, ErrStudy
		}
		values, err := predict(selected, ready.partitions[2])
		if err != nil {
			return result, err
		}
		result.Counts.Calibration += len(values)
		report, err := statehintfamily.Evaluate(ready.partitions[2], values)
		if err != nil {
			return result, err
		}
		result.Calibration = append(result.Calibration, Calibration{temperature, report})
		if err := j.json(fmt.Sprintf("temperature-%03d.calibration.json", step), result.Calibration[len(result.Calibration)-1]); err != nil {
			return result, err
		}
		if report.Eligible && (bestCalibration < 0 || report.EightNLL < result.Calibration[bestCalibration].Report.EightNLL) {
			bestCalibration = len(result.Calibration) - 1
			bestValues = values
		}
	}
	if bestCalibration < 0 {
		result.Status = "unqualified_calibration"
		return result, j.json("RESULTS.first.json", result)
	}
	chosen := result.Calibration[bestCalibration]
	if selected.SetTemperature(chosen.Temperature) != nil {
		return result, ErrStudy
	}
	result.Temperature = chosen.Temperature
	artifact, err := modelBytes(selected)
	if err != nil {
		return result, err
	}
	result.ModelSHA = digest(artifact)
	if err = j.write("selected.rsh", artifact); err != nil {
		return result, err
	}
	saved, err := readPin(j.root, Pin{"selected.rsh", result.ModelSHA}, statehint.ArtifactBytes)
	if err != nil {
		return result, err
	}
	reloaded, err := statehint.Load(bytes.NewReader(saved))
	if err != nil {
		return result, ErrStudy
	}
	validationValues, err := predict(selected, ready.partitions[1])
	if err != nil {
		return result, err
	}
	result.Counts.ReloadParity += len(validationValues)
	for i, corpus := range []statehintcorpus.Corpus{ready.partitions[1], ready.partitions[2]} {
		want := validationValues
		if i == 1 {
			want = bestValues
		}
		got, err := predict(reloaded, corpus)
		if err != nil {
			return result, err
		}
		result.Counts.ReloadParity += len(got)
		for k := range want {
			if want[k] != got[k] {
				return result, ErrStudy
			}
		}
	}
	if err = j.json("selected.validation.json", validationValues); err != nil {
		return result, err
	}
	if err = j.json("selected.calibration.json", bestValues); err != nil {
		return result, err
	}
	lock := beforeTestLock{Schema: "riido-statehint-v4-selected-lock-before-test-v1", PlanSHA: result.PlanSHA, SourceCommit: p.SourceCommit, BinarySHA: p.BinarySHA, Sources: p.Sources, Partitions: p.Partitions, Manifest: p.Manifest, Audit: p.Audit, Rubric: p.Rubric, Definitions: p.Definitions, Parent: p.Parent, Arm: p.Arms[selectedIndex], ModelSHA: result.ModelSHA, Temperature: chosen.Temperature, TrainingSteps: selected.TrainingSteps(), Validation: result.Trials[selectedIndex].Validation, Calibration: chosen.Report, Gate: [2]float64{.9, .05}}
	result.TestOpened, err = finalStage(lock.Validation, lock.Calibration, func() error {
		if err := j.json("LOCK.before-test.json", lock); err != nil {
			return err
		}
		data, err := readBounded(j.root, "LOCK.before-test.json", 1<<20)
		var got beforeTestLock
		// Report has optional JSON fields; compare exact serialized lock bytes,
		// rather than imposing the plan decoder on a reporting structure.
		if err != nil || json.Unmarshal(data, &got) != nil {
			return ErrStudy
		}
		want, err := json.MarshalIndent(lock, "", "  ")
		if err != nil || !bytes.Equal(data, append(want, '\n')) {
			return ErrStudy
		}
		// Source/data/rubric/parent pins are stable through fitting. The sealed
		// test is deliberately excluded from this repeat preflight.
		checked, err := prepare(root, ready.planName, p.BinarySHA)
		if err != nil || digest(checked.planBytes) != result.PlanSHA {
			return ErrStudy
		}
		return nil
	}, func() error {
		corpus, err := loadPartition(root, p.Partitions[3], "test", ready.manifest, ready.metadata)
		if err != nil {
			return err
		}
		for i, model := range []*statehint.Model{reloaded, ready.parent} {
			values, err := predict(model, corpus)
			if err != nil {
				return err
			}
			report, err := statehintfamily.Evaluate(corpus, values)
			if err != nil {
				return err
			}
			name := "child.test.json"
			if i == 0 {
				result.ChildTest = report
				result.Counts.ChildTest = len(values)
			} else {
				result.ParentTest = report
				result.Counts.ParentTest = len(values)
				name = "parent.test.json"
			}
			if err = j.json(name, values); err != nil {
				return err
			}
		}
		rows := corpus.Rows()
		rules := make([]statehint.Prediction, len(rows))
		for i, row := range rows {
			value, err := statehint.RuleSpeechAct(row.Text)
			if err != nil {
				return ErrStudy
			}
			rules[i] = value
		}
		result.Counts.RuleCalls = len(rules)
		result.RulesTest, err = statehintfamily.Evaluate(corpus, rules)
		if err != nil {
			return err
		}
		return j.json("speechact.test.json", rules)
	})
	if err != nil {
		return result, err
	}
	result.Status = "completed_research_evaluation"
	return result, j.json("RESULTS.first.json", result)
}

// Run has no default fit. --check validates the three available partitions
// without predictions; --fit is an explicit local maintainer operation.
func Run(args []string, out, errOut io.Writer) error {
	flags := flag.NewFlagSet("riido-statehint-tune-v4", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	planName := flags.String("plan", "", "concrete ready plan")
	output := flags.String("out", "", "new ignored private output directory")
	check := flags.Bool("check", false, "preflight only, no fit or predictions")
	fit := flags.Bool("fit", false, "explicit fixed four-arm local study")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, err = fmt.Fprintln(errOut, "riido-statehint-tune-v4 --plan PLAN --check\nriido-statehint-tune-v4 --plan PLAN --fit --out .cache/statehint/v4/NEW-RUN\nNo network, publication, app integration or state changes. Test bytes stay sealed until qualification and a durable lock.")
			return err
		}
		return ErrStudy
	}
	if flags.NArg() != 0 || !localPath(*planName) || *check == *fit || *check && *output != "" || *fit && (!localPath(*output) || !strings.HasPrefix(*output, ".cache/statehint/v4/")) {
		return ErrStudy
	}
	root, err := os.OpenRoot(".")
	if err != nil {
		return ErrStudy
	}
	defer root.Close()
	binary, err := executableSHA()
	if err != nil {
		return err
	}
	ready, err := prepare(root, *planName, binary)
	if err != nil {
		return err
	}
	ready.planName = *planName
	if *check {
		_, err = fmt.Fprintln(out, "checked_original_manifest_and_three_partitions; no model forwards or fitting; test bytes unopened")
		return err
	}
	private, err := privateOutput(root, *output)
	if err != nil {
		return ErrStudy
	}
	defer private.Close()
	journal := &journal{root: private}
	result, err := fitStudy(root, journal, ready)
	if err != nil {
		result.Status = "failed_study; inspect preserved local stage artifacts"
		if saved := journal.json("FAILED.first.json", result); saved != nil {
			return saved
		}
		return err
	}
	return json.NewEncoder(out).Encode(struct {
		Status           string `json:"status"`
		ModelSHA         string `json:"model_sha256"`
		TestOpened       bool   `json:"test_opened"`
		MutationExecuted bool   `json:"mutation_executed"`
	}{Status: result.Status, ModelSHA: result.ModelSHA, TestOpened: result.TestOpened})
}
