// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Maintainer-only fixed architecture study. AI references do not qualify a model.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"time"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
	"github.com/teamswyg/laya-tools/pkg/statehintclaimsmlp"
)

const (
	anchor           = ".cache/statehint-claims-mlp-study"
	inputBudget      = 8 << 20
	metadataBudget   = 1 << 20
	rowSchema        = "riido-three-claims-training-row-v1"
	strataSchema     = "riido-three-claims-mlp-strata-v1"
	receiptSchema    = "riido-three-claims-mlp-writer-receipt-v1"
	linearBytes      = 73988
	mlpBytes         = 131972
	nllFloor         = 1e-15
	bootstrapSamples = 10000
	newSteps         = 2120
)

var errStudy = errors.New("fixed MLP study pinned input or private output invalid")
var errEvaluationAlias = errors.New("pinned input aliases sealed evaluation")
var errInputAlias = errors.New("pinned inputs alias one another")

type manifest struct{ training, reference string }

func frozenManifest() manifest {
	return manifest{"c5b5a0adbc721d73213d4565c6a7c430094f31b8e7502fb5249e7ed40a3da837", "cfd35ee70a23f94a8d470b7dea596244a91ac7f1410475e8a42959693c99864b"}
}

type pins struct{ trainingName, trainingSHA, evaluationName, evaluationSHA, strataName, strataSHA, receiptName, receiptSHA, referenceName, referenceSHA string }
type inputRole struct {
	name, sha    string
	limit, exact int
}

func (p pins) roles() []inputRole {
	return []inputRole{{p.evaluationName, p.evaluationSHA, inputBudget, 0}, {p.trainingName, p.trainingSHA, inputBudget, 0}, {p.strataName, p.strataSHA, metadataBudget, 0}, {p.receiptName, p.receiptSHA, metadataBudget, 0}, {p.referenceName, p.referenceSHA, linearBytes, linearBytes}}
}

// Inspect every leaf before any body read. Identity guards cover hardlinks as
// well as paths; retained snapshots also stop a later aliased replacement.
func preflightInputs(root *os.Root, p pins) (map[string]os.FileInfo, error) {
	roles := p.roles()
	snapshots := make(map[string]os.FileInfo, len(roles))
	for i, role := range roles {
		for j := 0; j < i; j++ {
			if role.name == roles[j].name {
				if j == 0 {
					return nil, errEvaluationAlias
				}
				return nil, errInputAlias
			}
		}
		st, e := inspectInput(root, role.name, role.limit)
		if e != nil || role.exact > 0 && st.Size() != int64(role.exact) {
			return nil, errStudy
		}
		for j := 0; j < i; j++ {
			if os.SameFile(st, snapshots[roles[j].name]) {
				if j == 0 {
					return nil, errEvaluationAlias
				}
				return nil, errInputAlias
			}
		}
		snapshots[role.name] = st
	}
	return snapshots, nil
}
func readSnapshot(root *os.Root, role inputRole, before os.FileInfo) ([]byte, error) {
	if !validSHA(role.sha) {
		return nil, errStudy
	}
	f, e := openInput(root, role.name, role.limit, role.exact)
	if e != nil {
		return nil, errStudy
	}
	st, e := f.Stat()
	if e != nil || !sameSnapshot(before, st) {
		f.Close()
		return nil, errStudy
	}
	b, re := io.ReadAll(io.LimitReader(f, int64(role.limit)+1))
	after, se := f.Stat()
	ce := f.Close()
	if re != nil || se != nil || ce != nil || !sameSnapshot(before, after) || len(b) != int(before.Size()) || len(b) > role.limit || role.exact > 0 && len(b) != role.exact || digest(b) != role.sha {
		return nil, errStudy
	}
	return b, nil
}

func privateOutput(root *os.Root, name string) (*os.Root, error) {
	return privateOutputWithSync(root, name, syncDirectory)
}

// Directory entries need their own sync after file sync or rename. A failed
// durability guard prevents the next Fit invocation.
func syncDirectory(root *os.Root) error {
	directory, e := root.Open(".")
	if e != nil {
		return errStudy
	}
	syncErr, closeErr := directory.Sync(), directory.Close()
	if syncErr != nil || closeErr != nil {
		return errStudy
	}
	return nil
}
func privateOutputWithSync(root *os.Root, name string, sync func(*os.Root) error) (*os.Root, error) {
	if !local(name) || filepath.Dir(name) != anchor {
		return nil, errStudy
	}
	if e := root.Mkdir(".cache", 0700); e != nil && !errors.Is(e, os.ErrExist) {
		return nil, errStudy
	}
	if sync(root) != nil {
		return nil, errStudy
	}
	cache, e := openDirectory(root, ".cache")
	if e != nil {
		return nil, errStudy
	}
	defer cache.Close()
	base := filepath.Base(anchor)
	if e := cache.Mkdir(base, 0700); e != nil && !errors.Is(e, os.ErrExist) {
		return nil, errStudy
	}
	if sync(cache) != nil {
		return nil, errStudy
	}
	study, e := openDirectory(cache, base)
	if e != nil {
		return nil, errStudy
	}
	defer study.Close()
	st, e := study.Stat(".")
	if e != nil || st.Mode().Perm() != 0700 {
		return nil, errStudy
	}
	child := filepath.Base(name)
	if study.Mkdir(child, 0700) != nil {
		return nil, errStudy
	}
	if sync(study) != nil {
		return nil, errStudy
	}
	out, e := openDirectory(study, child)
	if e != nil {
		return nil, errStudy
	}
	st, e = out.Stat(".")
	if e != nil || st.Mode().Perm() != 0700 {
		out.Close()
		return nil, errStudy
	}
	if sync(out) != nil {
		out.Close()
		return nil, errStudy
	}
	return out, nil
}
func writeBytes(root *os.Root, name string, data []byte) error {
	if writeFileBytes(root, name, data) != nil {
		return errStudy
	}
	return syncDirectory(root)
}

// Only consumption's atomic temporary file uses this helper directly; its
// creation/removal and final name are persisted by syncing after the rename.
func writeFileBytes(root *os.Root, name string, data []byte) error {
	f, e := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return errStudy
	}
	n, e := f.Write(data)
	if e == nil && n != len(data) {
		e = io.ErrShortWrite
	}
	if e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e != nil || ce != nil {
		return errStudy
	}
	return nil
}

type recipe struct {
	Epochs      int        `json:"epochs"`
	Batch       int        `json:"batch_size"`
	Rate        float64    `json:"learning_rate"`
	Decay       float64    `json:"weight_decay"`
	BiasDecay   float64    `json:"bias_weight_decay"`
	Seed        int64      `json:"seed"`
	Temperature float64    `json:"temperature"`
	Gate        [2]float64 `json:"confidence_margin"`
	Optimizer   string     `json:"optimizer"`
	Shuffle     string     `json:"shuffle_contract"`
	SameOrder   bool       `json:"same_sample_and_shuffle_order_both_arms"`
	NewUpdates  int        `json:"new_updates_per_arm"`
}

func fixedRecipe() recipe {
	return recipe{40, 32, .001, .01, 0, 1729, 1, [2]float64{.9, .05}, "fresh_adamw_beta1_0.9_beta2_0.999_epsilon_1e-8", "PCG_seed1729_xor_0x9e3779b97f4a7c15_each_epoch_in_place_shuffle", true, newSteps}
}

type fitReport struct {
	Samples        int     `json:"samples"`
	Epochs         int     `json:"epochs"`
	Batches        int     `json:"batches"`
	TrainingSteps  uint64  `json:"training_steps"`
	MeanLoss       float64 `json:"mean_loss"`
	Initialization string  `json:"initialization"`
	Seed           int64   `json:"initialization_seed"`
	BatchSize      int     `json:"batch_size"`
	LearningRate   float64 `json:"learning_rate"`
	WeightDecay    float64 `json:"weight_decay"`
}
type modelReport struct {
	SHA                           string `json:"sha256"`
	ArtifactBytes                 int    `json:"artifact_bytes"`
	TrainingSteps                 uint64 `json:"training_steps"`
	InitializationSeed            int64  `json:"initialization_seed"`
	Architecture                  string `json:"architecture"`
	ParameterFloats               int    `json:"parameter_floats"`
	ReadOnlyExact                 bool   `json:"read_only_exact_save_parity"`
	ModelStructBytes              uint64 `json:"model_struct_bytes_not_total_heap"`
	InferenceWorkspaceStructBytes uint64 `json:"inference_workspace_struct_bytes_not_total_heap"`
	TrainingWorkspaceStructBytes  uint64 `json:"training_workspace_struct_bytes_excludes_shuffle_slice_buffer"`
}
type candidateReport struct {
	modelReport
	ReloadByteExact       bool `json:"reload_exact_byte_parity"`
	ReloadPredictExact    bool `json:"reload_exact_prediction_parity"`
	ReloadScoresExact     bool `json:"reload_exact_score_parity"`
	PredictionCalls       int  `json:"evaluation_prediction_calls"`
	ScoreCalls            int  `json:"evaluation_score_calls"`
	ReloadPredictionCalls int  `json:"reload_prediction_calls"`
	ReloadScoreCalls      int  `json:"reload_score_calls"`
}

// Only text enters model calls. Workspaces belong to this one scorer and are
// never shared across arms. These small adapters reconcile the two concrete APIs.
type modelAdapter struct {
	save     func(io.Writer) error
	predict  func(string) (statehintclaims.Prediction, error)
	scores   func(string) (statehintclaims.ScoreResult, error)
	metadata modelReport
}

func linearAdapter(m *statehintclaims.Model) modelAdapter {
	var w statehintclaims.Workspace
	meta := modelReport{TrainingSteps: m.TrainingSteps(), InitializationSeed: m.InitializationSeed(), Architecture: "2048_contextual_three_independent_linear_categorical_heads", ParameterFloats: statehintclaims.ParameterCount, ModelStructBytes: uint64(reflect.TypeOf(statehintclaims.Model{}).Size()), InferenceWorkspaceStructBytes: uint64(reflect.TypeOf(w).Size()), TrainingWorkspaceStructBytes: uint64(reflect.TypeOf(statehintclaims.TrainingWorkspace{}).Size())}
	return modelAdapter{m.Save, func(text string) (statehintclaims.Prediction, error) { return m.Predict(text, &w) }, func(text string) (statehintclaims.ScoreResult, error) { return m.Scores(text, &w) }, meta}
}
func mlpAdapter(m *statehintclaimsmlp.Model) modelAdapter {
	var w statehintclaimsmlp.Workspace
	meta := m.Metadata()
	summary := modelReport{TrainingSteps: meta.TrainingSteps, InitializationSeed: meta.InitializationSeed, Architecture: meta.Architecture, ParameterFloats: meta.ParameterFloats, ModelStructBytes: uint64(reflect.TypeOf(statehintclaimsmlp.Model{}).Size()), InferenceWorkspaceStructBytes: uint64(reflect.TypeOf(w).Size()), TrainingWorkspaceStructBytes: uint64(reflect.TypeOf(statehintclaimsmlp.TrainingWorkspace{}).Size())}
	return modelAdapter{m.Save, func(text string) (statehintclaims.Prediction, error) { return m.Predict(text, &w) }, func(text string) (statehintclaims.ScoreResult, error) { return m.Scores(text, &w) }, summary}
}

type studyOperations struct {
	fitLinear     func([]statehintclaims.Sample) (modelAdapter, fitReport, error)
	fitMLP        func([]statehintclaims.Sample) (modelAdapter, fitReport, error)
	loadLinear    func([]byte, uint64) (modelAdapter, error)
	loadMLP       func([]byte) (modelAdapter, error)
	syncDirectory func(*os.Root) error
}

func nativeOperations() studyOperations {
	return studyOperations{
		syncDirectory: syncDirectory,
		fitLinear: func(samples []statehintclaims.Sample) (modelAdapter, fitReport, error) {
			m := statehintclaims.NewModel()
			var w statehintclaims.TrainingWorkspace
			r, e := m.Fit(samples, statehintclaims.FitOptions{Epochs: 40, BatchSize: 32, LearningRate: .001, WeightDecay: .01, Seed: 1729}, &w)
			return linearAdapter(m), fitReport{r.Samples, r.Epochs, r.Batches, r.TrainingSteps, r.MeanLoss, r.Initialization, r.Seed, 32, .001, .01}, e
		},
		fitMLP: func(samples []statehintclaims.Sample) (modelAdapter, fitReport, error) {
			m := statehintclaimsmlp.NewModel()
			var w statehintclaimsmlp.TrainingWorkspace
			r, e := m.Fit(samples, &w)
			return mlpAdapter(m), fitReport{r.Samples, r.Epochs, r.Batches, r.TrainingSteps, r.MeanLoss, r.Initialization, r.Seed, r.BatchSize, r.LearningRate, r.WeightDecay}, e
		},
		loadLinear: func(data []byte, steps uint64) (modelAdapter, error) {
			m, e := statehintclaims.Load(bytes.NewReader(data))
			if e != nil || m.TrainingSteps() != steps || m.InitializationSeed() != 1729 || m.Temperature() != 1 {
				return modelAdapter{}, errStudy
			}
			return linearAdapter(m), nil
		},
		loadMLP: func(data []byte) (modelAdapter, error) {
			m, e := statehintclaimsmlp.Load(bytes.NewReader(data))
			if e != nil {
				return modelAdapter{}, errStudy
			}
			meta := m.Metadata()
			if meta.TrainingSteps != newSteps || meta.TrainingSamples != 1680 || meta.InitializationSeed != 1729 || meta.Temperature != 1 || meta.HiddenUnits != 16 || meta.Qualified || meta.StateAuthority {
				return modelAdapter{}, errStudy
			}
			return mlpAdapter(m), nil
		},
	}
}
func validFit(r fitReport, m modelAdapter, initialization string) bool {
	return r.Samples == 1680 && r.Epochs == 40 && r.Batches == newSteps && r.TrainingSteps == newSteps && r.Seed == 1729 && r.BatchSize == 32 && r.LearningRate == .001 && r.WeightDecay == .01 && r.Initialization == initialization && !math.IsNaN(r.MeanLoss) && !math.IsInf(r.MeanLoss, 0) && r.MeanLoss >= 0 && m.metadata.TrainingSteps == newSteps && m.metadata.InitializationSeed == 1729
}
func serialize(m modelAdapter, exact int) ([]byte, error) {
	var b bytes.Buffer
	if m.save == nil || m.save(&b) != nil || b.Len() != exact {
		return nil, errStudy
	}
	return b.Bytes(), nil
}
func observe(m modelAdapter, p prepared) ([]observation, error) {
	r := make([]observation, len(p.rows))
	for i, row := range p.rows {
		prediction, e := m.predict(row.Text)
		if e != nil || prediction.Source != statehintclaims.Learned || prediction.TrainingSteps != m.metadata.TrainingSteps {
			return nil, errStudy
		}
		score, e := m.scores(row.Text)
		if e != nil || score.WordCount == 0 {
			return nil, errStudy
		}
		r[i] = observation{prediction, score}
	}
	return r, nil
}
func exactReload(out *os.Root, name string, m modelAdapter, data []byte, p prepared, obs []observation, load func([]byte) (modelAdapter, error)) (candidateReport, error) {
	stored, e := readPinned(out, name, digest(data), len(data), len(data))
	if e != nil {
		return candidateReport{}, errStudy
	}
	loaded, e := load(stored)
	if e != nil {
		return candidateReport{}, errStudy
	}
	reencoded, e := serialize(loaded, len(data))
	if e != nil || !bytes.Equal(stored, reencoded) {
		return candidateReport{}, errStudy
	}
	round, e := observe(loaded, p)
	if e != nil || len(round) != len(obs) {
		return candidateReport{}, errStudy
	}
	for i := range round {
		if round[i] != obs[i] {
			return candidateReport{}, errStudy
		}
	}
	unchanged, e := serialize(m, len(data))
	if e != nil || !bytes.Equal(data, unchanged) {
		return candidateReport{}, errStudy
	}
	summary := m.metadata
	summary.SHA, summary.ArtifactBytes, summary.ReadOnlyExact = digest(data), len(data), true
	return candidateReport{summary, true, true, true, len(obs), len(obs), len(obs), len(obs)}, nil
}

type consumption struct {
	Schema         string  `json:"schema"`
	Status         string  `json:"status"`
	FitCalls       int     `json:"actual_new_fit_calls"`
	FitInvocations [2]int  `json:"fit_invocations_cold_linear_cold_mlp16"`
	CompletedFits  [2]bool `json:"completed_fits_cold_linear_cold_mlp16"`
	SavedModels    [2]bool `json:"saved_models_cold_linear_cold_mlp16"`
	Phase          string  `json:"phase"`
	TrainingSHA    string  `json:"training_sha256"`
	NoRetry        bool    `json:"no_implicit_retry"`
}

// Each invocation is consumed durably before calling Fit. Fresh output names
// cannot retry or overwrite an existing ledger/artifact on a later failure.
func writeConsumption(out *os.Root, r consumption) error {
	return writeConsumptionWithSync(out, r, syncDirectory)
}
func writeConsumptionWithSync(out *os.Root, r consumption, sync func(*os.Root) error) error {
	b, e := json.MarshalIndent(r, "", "  ")
	if e != nil {
		return errStudy
	}
	if writeFileBytes(out, ".fit-consumption.tmp", append(b, '\n')) != nil {
		return errStudy
	}
	if out.Rename(".fit-consumption.tmp", "fit-consumption.json") != nil {
		return errStudy
	}
	return sync(out)
}

type studyReport struct {
	Schema                    string                   `json:"schema"`
	Status                    string                   `json:"status"`
	TrainingSHA               string                   `json:"training_sha256"`
	EvaluationSHA             string                   `json:"evaluation_sha256"`
	StrataSHA                 string                   `json:"strata_sha256"`
	ReceiptSHA                string                   `json:"writer_receipt_sha256"`
	ReferenceStatus           string                   `json:"reference_status"`
	TrainingCounts            counts                   `json:"training_counts"`
	EvaluationCounts          counts                   `json:"evaluation_counts"`
	Heads                     [3]string                `json:"head_order"`
	States                    [3]statehintclaims.State `json:"state_order"`
	Recipe                    recipe                   `json:"fixed_recipe"`
	ColdLinear                candidateReport          `json:"cold_linear"`
	ColdMLP                   candidateReport          `json:"cold_mlp16"`
	PriorReference            modelReport              `json:"descriptive_prior_float_reference"`
	PriorRole                 string                   `json:"prior_reference_role"`
	PriorEqualBudget          bool                     `json:"prior_reference_equal_initialization_or_update_budget_control"`
	Fits                      [2]fitReport             `json:"one_fit_each_cold_linear_cold_mlp16"`
	FitLossMeaning            string                   `json:"mean_fit_loss_meaning"`
	FitCalls                  int                      `json:"actual_new_fit_calls"`
	ColdLinearEvaluation      evaluationReport         `json:"cold_linear_final_evaluation"`
	ColdMLPEvaluation         evaluationReport         `json:"cold_mlp16_final_evaluation"`
	PriorEvaluation           evaluationReport         `json:"descriptive_prior_reference_final_evaluation"`
	Primary                   pairedReport             `json:"primary_mlp16_minus_cold_linear_paired_comparison"`
	Progress                  progressReport           `json:"frozen_research_progress_rule"`
	Qualified                 bool                     `json:"semantic_quality_qualified"`
	OriginalQualified         bool                     `json:"original_qualified_model_exists"`
	QualificationChanged      bool                     `json:"existing_qualification_rules_changed"`
	Selection                 bool                     `json:"selection_performed"`
	Calibration               bool                     `json:"calibration_performed"`
	CalTestAccessed           bool                     `json:"original_or_existing_calibration_test_accessed"`
	StateWrites               int                      `json:"application_state_writes"`
	FitsSavedBeforeEvaluation bool                     `json:"both_fits_and_saves_before_evaluation_body_or_prior_load"`
	ElapsedNS                 int64                    `json:"fit_evaluation_save_reload_nanoseconds"`
	GoHeap                    uint64                   `json:"go_heap_alloc_bytes_not_os_rss"`
	OSRSSMeasured             bool                     `json:"os_rss_measured_in_process"`
	OSRSSMeasurement          string                   `json:"os_rss_measurement"`
	CPUProfile                bool                     `json:"private_cpu_profile_written"`
}

func fitStudy(root, out *os.Root, train prepared, s strata, receipt receipt, p pins, snapshots map[string]os.FileInfo, cpuProfile bool, ops studyOperations) (result studyReport, err error) {
	started := time.Now()
	sync := ops.syncDirectory
	if sync == nil {
		sync = syncDirectory
	}
	ledger := consumption{Schema: "riido-three-claims-mlp-fit-consumption-v1", Status: "prepared_no_fit_invoked", Phase: "before_fit", TrainingSHA: p.trainingSHA, NoRetry: true}
	ledgerAttempted := false
	defer func() {
		if err != nil && ledgerAttempted {
			ledger.Status = "failed_before_fit_invocation"
			if ledger.FitCalls > 0 {
				ledger.Status = "failed_after_fit_invocation_consumed"
			}
			if writeConsumptionWithSync(out, ledger, sync) != nil {
				err = errStudy
			}
		}
	}()
	consumeFit := func(arm int, phase string) error {
		ledgerAttempted = true
		ledger.Phase = phase + "_durability_guard"
		reserved := ledger
		reserved.Phase, reserved.Status = phase, "fit_invocation_consumed"
		reserved.FitCalls++
		reserved.FitInvocations[arm]++
		if writeConsumptionWithSync(out, reserved, sync) != nil {
			return errStudy
		}
		ledger = reserved
		return nil
	}
	var profile *os.File
	finishProfile := func() error {
		if profile == nil {
			return nil
		}
		pprof.StopCPUProfile()
		se, ce := profile.Sync(), profile.Close()
		profile = nil
		if se != nil || ce != nil {
			return errStudy
		}
		return nil
	}
	if cpuProfile {
		profile, err = out.OpenFile("cpu.pprof", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return result, errStudy
		}
		if sync(out) != nil {
			profile.Close()
			profile = nil
			return result, errStudy
		}
		if pprof.StartCPUProfile(profile) != nil {
			profile.Close()
			profile = nil
			return result, errStudy
		}
		defer func() {
			if finishProfile() != nil {
				ledger.Phase = "cpu_profile_finish"
				err = errStudy
			}
		}()
	}
	samples := make([]statehintclaims.Sample, len(train.rows))
	for i, r := range train.rows {
		samples[i] = statehintclaims.Sample{Text: r.Text, Targets: r.labels}
	}
	before, e := inspectInput(root, p.evaluationName, inputBudget)
	if e != nil || !sameSnapshot(snapshots[p.evaluationName], before) {
		return result, errStudy
	}
	if consumeFit(0, "cold_linear_fixed_fit") != nil {
		return result, errStudy
	}
	linear, linearFit, e := ops.fitLinear(samples)
	if e != nil || !validFit(linearFit, linear, "fresh_zero") {
		return result, errStudy
	}
	ledger.CompletedFits[0] = true
	ledger.Phase = "linear_save_before_evaluation"
	linearData, e := serialize(linear, linearBytes)
	if e != nil || writeBytes(out, "linear.rsc", linearData) != nil {
		return result, errStudy
	}
	ledger.SavedModels[0] = true
	if consumeFit(1, "cold_mlp16_fixed_fit") != nil {
		return result, errStudy
	}
	mlp, mlpFit, e := ops.fitMLP(samples)
	if e != nil || !validFit(mlpFit, mlp, statehintclaimsmlp.Initialization) {
		return result, errStudy
	}
	ledger.CompletedFits[1] = true
	ledger.Phase = "mlp_save_before_evaluation"
	mlpData, e := serialize(mlp, mlpBytes)
	if e != nil || writeBytes(out, "mlp.rcm", mlpData) != nil {
		return result, errStudy
	}
	ledger.SavedModels[1] = true
	ledger.Status = "both_fixed_fits_consumed_and_artifacts_saved"
	ledger.Phase = "evaluation_body_read_and_pin"
	if writeConsumptionWithSync(out, ledger, sync) != nil {
		return result, errStudy
	}
	data, e := readSnapshot(root, p.roles()[0], snapshots[p.evaluationName])
	if e != nil {
		return result, errStudy
	}
	ledger.Phase = "evaluation_schema_parse"
	evaluation, e := prepare(data, "dev")
	if e != nil {
		return result, errStudy
	}
	ledger.Phase = "evaluation_reference_support_audit"
	if validateEvaluation(evaluation, train, s, receipt) != nil {
		return result, errStudy
	}
	ledger.Phase = "prior_reference_body_read_and_pin"
	priorData, e := readSnapshot(root, p.roles()[4], snapshots[p.referenceName])
	if e != nil {
		return result, errStudy
	}
	ledger.Phase = "prior_reference_numeric_load"
	prior, e := ops.loadLinear(priorData, 4240)
	if e != nil {
		return result, errStudy
	}
	ledger.Phase = "evaluation_predictions_and_aggregate_diagnostics"
	linearObs, e := observe(linear, evaluation)
	if e != nil {
		return result, errStudy
	}
	mlpObs, e := observe(mlp, evaluation)
	if e != nil {
		return result, errStudy
	}
	priorObs, e := observe(prior, evaluation)
	if e != nil {
		return result, errStudy
	}
	linearEval, e := evaluate(evaluation, linearObs)
	if e != nil {
		return result, errStudy
	}
	mlpEval, e := evaluate(evaluation, mlpObs)
	if e != nil {
		return result, errStudy
	}
	priorEval, e := evaluate(evaluation, priorObs)
	if e != nil {
		return result, errStudy
	}
	delta, e := paired(evaluation, s, mlpObs, linearObs)
	if e != nil || delta.Lineages != [2]int{40, 20} {
		return result, errStudy
	}
	ledger.Phase = "both_artifacts_reload_prediction_score_byte_parity"
	linearSaved, e := exactReload(out, "linear.rsc", linear, linearData, evaluation, linearObs, func(data []byte) (modelAdapter, error) { return ops.loadLinear(data, newSteps) })
	if e != nil {
		return result, errStudy
	}
	mlpSaved, e := exactReload(out, "mlp.rcm", mlp, mlpData, evaluation, mlpObs, ops.loadMLP)
	if e != nil {
		return result, errStudy
	}
	unchanged, e := serialize(prior, linearBytes)
	if e != nil || !bytes.Equal(priorData, unchanged) {
		return result, errStudy
	}
	priorSummary := prior.metadata
	priorSummary.SHA, priorSummary.ArtifactBytes, priorSummary.ReadOnlyExact = digest(priorData), len(priorData), true
	ledger.Phase = "cpu_profile_finish"
	if finishProfile() != nil {
		return result, errStudy
	}
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	result = studyReport{Schema: "riido-three-claims-fixed-mlp16-study-report-v1", Status: "ai_reference_research_only_not_human_or_product_evidence", TrainingSHA: p.trainingSHA, EvaluationSHA: p.evaluationSHA, StrataSHA: p.strataSHA, ReceiptSHA: p.receiptSHA, ReferenceStatus: receipt.ReferenceStatus, TrainingCounts: train.counts, EvaluationCounts: s.counts, Heads: headOrder(), States: statehintclaims.States(), Recipe: fixedRecipe(), ColdLinear: linearSaved, ColdMLP: mlpSaved, PriorReference: priorSummary, PriorRole: "descriptive_fixed_prior_float_preview_4240_history_updates_cold_arms_2120_each_not_equal_budget_or_selector", Fits: [2]fitReport{linearFit, mlpFit}, FitLossMeaning: "mean_online_fit_trajectory_loss_across_all_epochs_not_final_evaluation_ce", FitCalls: 2, ColdLinearEvaluation: linearEval, ColdMLPEvaluation: mlpEval, PriorEvaluation: priorEval, Primary: delta, Progress: progress(delta, mlpEval, linearEval), FitsSavedBeforeEvaluation: true, ElapsedNS: time.Since(started).Nanoseconds(), GoHeap: memory.HeapAlloc, OSRSSMeasurement: "external_process_measurement_required_go_heap_is_not_os_rss", CPUProfile: cpuProfile}
	ledger.Phase = "aggregate_report_write"
	encoded, e := json.MarshalIndent(result, "", "  ")
	if e != nil || writeBytes(out, "report.json", append(encoded, '\n')) != nil {
		return studyReport{}, errStudy
	}
	ledger.Status, ledger.Phase = "completed_two_fixed_fits_no_selection", "complete"
	if writeConsumptionWithSync(out, ledger, sync) != nil {
		return studyReport{}, errStudy
	}
	return result, nil
}

func run(args []string, output, errorOutput io.Writer) error {
	return runWithOperations(args, output, errorOutput, frozenManifest(), nativeOperations())
}

// Tests supply owned fixtures and phase observers. CLI has no alternate pins,
// fit count, hyperparameter, model-selection or evaluator override.
func runWithOperations(args []string, output, errorOutput io.Writer, m manifest, ops studyOperations) error {
	f := flag.NewFlagSet("riido-statehint-claims-mlp-study", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	var p pins
	f.StringVar(&p.trainingName, "training", "", "unchanged pinned training JSONL")
	f.StringVar(&p.trainingSHA, "training-sha256", "", "training exact SHA")
	f.StringVar(&p.evaluationName, "evaluation", "", "fresh sealed evaluation JSONL")
	f.StringVar(&p.evaluationSHA, "evaluation-sha256", "", "evaluation exact SHA")
	f.StringVar(&p.strataName, "strata", "", "text-free evaluation metadata")
	f.StringVar(&p.strataSHA, "strata-sha256", "", "strata exact SHA")
	f.StringVar(&p.receiptName, "receipt", "", "text-free writer receipt")
	f.StringVar(&p.receiptSHA, "receipt-sha256", "", "receipt exact SHA")
	f.StringVar(&p.referenceName, "reference", "", "descriptive prior float reference")
	f.StringVar(&p.referenceSHA, "reference-sha256", "", "prior reference exact SHA")
	check := f.Bool("check", false, "validate admission without evaluation or prior body reads, numeric models, fits or files")
	name := f.String("out", "", "fresh private study directory")
	profile := f.Bool("cpu-profile", false, "private CPU profile")
	if e := f.Parse(args); e != nil {
		if errors.Is(e, flag.ErrHelp) {
			_, e = fmt.Fprintln(errorOutput, "riido-statehint-claims-mlp-study: required --training/--training-sha256 --evaluation/--evaluation-sha256 --strata/--strata-sha256 --receipt/--receipt-sha256 --reference/--reference-sha256; --check OR --out .cache/statehint-claims-mlp-study/NEW [--cpu-profile]. Check inspects evaluation/reference leaves only; no Load, Predict, Fit or output files. Actual mode performs exactly two fixed cold Fits and saves both models before evaluation-body access or prior-reference Load. AI-reference research only.")
			return e
		}
		return errStudy
	}
	if f.NArg() != 0 || p.trainingSHA != m.training || p.referenceSHA != m.reference || *check && (*name != "" || *profile) || !*check && (!local(*name) || filepath.Dir(*name) != anchor) {
		return errStudy
	}
	for _, role := range p.roles() {
		if !local(role.name) || !validSHA(role.sha) {
			return errStudy
		}
	}
	root, e := os.OpenRoot(".")
	if e != nil {
		return errStudy
	}
	defer root.Close()
	snapshots, e := preflightInputs(root, p)
	if e != nil {
		return e
	}
	trainingData, e := readSnapshot(root, p.roles()[1], snapshots[p.trainingName])
	if e != nil {
		return errStudy
	}
	train, e := prepare(trainingData, "fit")
	if e != nil || !trainingScope(train.counts) {
		return errStudy
	}
	strataData, e := readSnapshot(root, p.roles()[2], snapshots[p.strataName])
	if e != nil {
		return errStudy
	}
	s, e := prepareStrata(strataData, train)
	if e != nil {
		return errStudy
	}
	receiptData, e := readSnapshot(root, p.roles()[3], snapshots[p.receiptName])
	if e != nil {
		return errStudy
	}
	r, e := validateReceipt(receiptData, p.trainingSHA, p.evaluationSHA, p.strataSHA, s)
	if e != nil || snapshots[p.evaluationName].Size() != r.EvaluationBytes {
		return errStudy
	}
	if *check {
		return json.NewEncoder(output).Encode(struct {
			Status     string `json:"status"`
			Training   counts `json:"training_counts"`
			Evaluation counts `json:"evaluation_receipt_counts"`
		}{"checked pinned training metadata receipt and leaf identities; no evaluation or prior body read, Load, Predict, Fit or output files", train.counts, s.counts})
	}
	sync := ops.syncDirectory
	if sync == nil {
		sync = syncDirectory
	}
	private, e := privateOutputWithSync(root, *name, sync)
	if e != nil {
		return errStudy
	}
	defer private.Close()
	result, e := fitStudy(root, private, train, s, r, p, snapshots, *profile, ops)
	if e != nil {
		return errStudy
	}
	return json.NewEncoder(output).Encode(struct {
		Status    string `json:"status"`
		Qualified bool   `json:"semantic_quality_qualified"`
		Progress  bool   `json:"research_progress"`
	}{result.Status, false, result.Progress.ResearchProgress})
}
func main() {
	runtime.GOMAXPROCS(2)
	debug.SetMemoryLimit(512 << 20)
	if run(os.Args[1:], os.Stdout, os.Stderr) != nil {
		fmt.Fprintln(os.Stderr, "fixed MLP study failed; check pinned canonical inputs and fresh private output; consumed fit ledger and saved artifacts are preserved")
		os.Exit(1)
	}
}
