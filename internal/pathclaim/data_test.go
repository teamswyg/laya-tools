package pathclaim

import (
	"math"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/pairlearn"
)

// Every value here is authored test data. Repository strings exercise the
// pinned spelling contract and do not load any tasks, caches or model weights.
func smallRows() []Row {
	rows := make([]Row, 0, 8)
	for i := 0; i < 8; i++ {
		role, repo := "train", "Qiskit/qiskit"
		if i >= 4 {
			role, repo = "validation", "numpy/numpy"
		}
		row := Row{Role: role, Repository: repo, Group: i, BaselinePages: 2, HelperPages: 2}
		row.Features[0], row.Features[1] = 1, float64(i%4)/4
		switch i % 4 {
		case 0:
			row.BaselinePages, row.HelperPages = 6, 2
		case 1:
			row.BaselinePages, row.HelperPages = 2, 6
		case 3:
			row.BaselinePages, row.HelperPages = 4, 2
		}
		rows = append(rows, row)
	}
	return rows
}

func ownedReadyRows(t *testing.T) ([]Row, Readiness) {
	t.Helper()
	rows := make([]Row, 0, 4800)
	for i := 0; i < 2400; i++ {
		row := smallRows()[i%4]
		row.Group = i
		// Authored counts stay within these actual membership repository sizes.
		switch {
		case i < 1000:
			row.Repository = "Qiskit/qiskit"
		case i < 1400:
			row.Repository = "apache/airflow"
		case i < 1800:
			row.Repository = "conda/conda"
		case i < 2200:
			row.Repository = "pypa/pip"
		default:
			row.Repository = "huggingface/transformers"
		}
		rows = append(rows, row)
	}
	counts := [5]int{200, 500, 500, 1000, 200}
	for ri, repo := range ValidationRepositories() {
		for j := 0; j < counts[ri]; j++ {
			row := smallRows()[j%4]
			row.Role, row.Repository, row.Group = "validation", repo, len(rows)
			rows = append(rows, row)
		}
	}
	h, err := RowsSHA256(rows)
	if err != nil {
		t.Fatal(err)
	}
	r := Readiness{Schema: ReadinessSchema, PlanSHA256: PlanSHA256, MembershipSHA256: MembershipSHA256, NumericRowsSHA256: h, AvailabilityCheckpointSHA256: strings.Repeat("1", 64), SourceEvidenceSHA256: strings.Repeat("2", 64), InputSealSHA256: strings.Repeat("3", 64), RunnerRevision: strings.Repeat("4", 40), TrainingCoverage: 7335, ValidationCoverage: 5686, ProtectedFinal: 2402, TrainingEligible: 2400, ValidationEligible: 2400}
	for i, repo := range ValidationRepositories() {
		r.ValidationRepositories[i] = RepositoryCount{repo, counts[i]}
	}
	return rows, r
}

func TestPlanHashIsActualCommittedPlan(t *testing.T) {
	b, err := os.ReadFile("../../experiments/path-cost-claim/plan-46.json")
	if err != nil {
		t.Fatal(err)
	}
	// Independent digest of the definition file, not a training-data read.
	want := sha256Hex(b)
	if want != PlanSHA256 {
		t.Fatal("adapter does not match committed plan")
	}
}

func TestPreparationSignedTargetsZeroRowsAndOwnership(t *testing.T) {
	rows := smallRows()
	before := slices.Clone(rows)
	d, counts, err := Prepare(rows, "train", 0)
	if err != nil {
		t.Fatal(err)
	}
	if counts != (DatasetCounts{4, 1}) || len(d.Labels) != 4 || d.Labels[0] != 1 || d.Labels[1] != 0 || d.Labels[2] != 0 || d.Labels[3] != 1 {
		t.Fatal("signed targets or zero row changed", counts, d.Labels)
	}
	if d.SampleWeights[2] != 0 || d.SampleWeights[0] != 4/WeightScale || d.SampleWeights[1] != 4/WeightScale {
		t.Fatal("wrong global scale")
	}
	if !reflect.DeepEqual(rows, before) {
		t.Fatal("input rows mutated")
	}
	d.Values[0], d.Groups[0], d.Labels[0] = 9, 99, 9
	if !reflect.DeepEqual(rows, before) {
		t.Fatal("dataset aliases caller rows")
	}
	changed := slices.Clone(rows)
	for i := range changed {
		changed[i].Group += 100
		if changed[i].Role == "train" {
			changed[i].Repository = "apache/airflow"
		}
	}
	a, _, _ := Prepare(rows, "train", 1)
	c, _, err := Prepare(changed, "train", 1)
	if err != nil || !reflect.DeepEqual(a.Values, c.Values) || !reflect.DeepEqual(a.Indices, c.Indices) || !reflect.DeepEqual(a.Labels, c.Labels) || !reflect.DeepEqual(a.SampleWeights, c.SampleWeights) {
		t.Fatal("bookkeeping changed model inputs", err)
	}
}

func TestConstantScalingPreservesNLLAndFit(t *testing.T) {
	rows := smallRows()
	train, _, err := Prepare(rows, "train", 1)
	if err != nil {
		t.Fatal(err)
	}
	validation, _, err := Prepare(rows, "validation", 1)
	if err != nil {
		t.Fatal(err)
	}
	unscaledTrain, unscaledValidation := train, validation
	unscaledTrain.SampleWeights = slices.Clone(train.SampleWeights)
	unscaledValidation.SampleWeights = slices.Clone(validation.SampleWeights)
	for i := range unscaledTrain.SampleWeights {
		unscaledTrain.SampleWeights[i] *= WeightScale
	}
	for i := range unscaledValidation.SampleWeights {
		unscaledValidation.SampleWeights[i] *= WeightScale
	}
	w := make([]float64, 16)
	w[0], w[1] = .25, -.75
	if math.Abs(pairlearn.NLL(train, w)-pairlearn.NLL(unscaledTrain, w)) > 1e-14 {
		t.Fatal("constant scaling changed normalized loss")
	}
	config := frozenConfig(1729)
	fit, err := pairlearn.Fit(train, validation, config)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := pairlearn.Fit(unscaledTrain, unscaledValidation, config)
	if err != nil {
		t.Fatal(err)
	}
	if fit.Epoch != ref.Epoch || math.Abs(fit.ValidationNLL-ref.ValidationNLL) > 1e-12 {
		t.Fatal("scaling changed epoch or loss")
	}
	for i := range fit.Weights {
		if math.Abs(fit.Weights[i]-ref.Weights[i]) > 2e-7 {
			t.Fatal("scaling changed FP32 coefficients")
		}
	}
}

func TestPreparationBoundsAndRoles(t *testing.T) {
	for _, mode := range []string{"final", "unknown-role", "wrong-repo", "final-repo-as-train", "cross-role", "duplicate-group", "negative-group", "group-bound", "nan", "inf", "negative-feature", "feature-over-one", "zero-pages", "over-pages", "zero-total"} {
		t.Run(mode, func(t *testing.T) {
			rows := smallRows()
			switch mode {
			case "final":
				rows[7].Role = "final"
			case "unknown-role":
				rows[7].Role = "development"
			case "wrong-repo":
				rows[7].Repository = "numpy"
			case "final-repo-as-train":
				rows[0].Repository = "mesonbuild/meson"
			case "cross-role":
				rows[7].Repository = rows[0].Repository
			case "duplicate-group":
				rows[7].Group = rows[0].Group
			case "negative-group":
				rows[0].Group = -1
			case "group-bound":
				rows[0].Group = MaxRows
			case "nan":
				rows[0].Features[3] = math.NaN()
			case "inf":
				rows[0].Features[3] = math.Inf(1)
			case "negative-feature":
				rows[0].Features[3] = -.01
			case "feature-over-one":
				rows[0].Features[3] = 1.01
			case "zero-pages":
				rows[0].BaselinePages = 0
			case "over-pages":
				rows[0].HelperPages = MaxPages + 1
			case "zero-total":
				for i := range rows {
					rows[i].BaselinePages = rows[i].HelperPages
				}
			}
			if _, _, err := Prepare(rows, "train", 0); err == nil {
				t.Fatal("accepted invalid row contract")
			}
		})
	}
	rows := smallRows()
	rows[0].BaselinePages, rows[0].HelperPages = 1, MaxPages
	d, _, err := Prepare(rows, "train", 16)
	if err != nil || d.SampleWeights[0] != 5015/WeightScale || !(d.SampleWeights[0] < 1) {
		t.Fatal("maximum mathematical bound", err)
	}
	for _, lambda := range []float64{-.25, .5, 17, math.NaN(), math.Inf(1)} {
		if _, _, err := Prepare(rows, "train", lambda); err == nil {
			t.Fatal("unregistered penalty accepted")
		}
	}
	if _, _, err := Prepare(rows, "final", 0); err == nil {
		t.Fatal("final preparation accepted")
	}
	if _, _, err := Prepare(nil, "train", 0); err == nil {
		t.Fatal("empty input accepted")
	}
	if _, _, err := Prepare(make([]Row, MaxRows+1), "train", 0); err == nil {
		t.Fatal("row bound ignored")
	}
}

func TestReadinessRejectsWithoutFit(t *testing.T) {
	rows, ready := ownedReadyRows(t)
	if err := VerifyReadiness(rows, ready); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"schema", "plan", "membership", "rows-hash", "source", "checkpoint", "seal", "runner", "denominator", "final", "train-small", "val-small", "repo-small", "repo-wrong", "repo-count", "repo-over-cohort", "rows-count"} {
		t.Run(mode, func(t *testing.T) {
			r := ready
			input := rows
			switch mode {
			case "schema":
				r.Schema = "other"
			case "plan":
				r.PlanSHA256 = strings.Repeat("a", 64)
			case "membership":
				r.MembershipSHA256 = strings.Repeat("a", 64)
			case "rows-hash":
				r.NumericRowsSHA256 = strings.Repeat("a", 64)
			case "source":
				r.SourceEvidenceSHA256 = ""
			case "checkpoint":
				r.AvailabilityCheckpointSHA256 = ""
			case "seal":
				r.InputSealSHA256 = ""
			case "runner":
				r.RunnerRevision = strings.Repeat("A", 40)
			case "denominator":
				r.TrainingCoverage = 7334
			case "final":
				r.ProtectedFinal = 2401
			case "train-small":
				r.TrainingEligible = 2399
			case "val-small":
				r.ValidationEligible = 107
			case "repo-small":
				r.ValidationRepositories[0].Eligible = 99
			case "repo-wrong":
				r.ValidationRepositories[0].Repository = "lightning-ai/pytorch-lightning"
			case "repo-count":
				r.ValidationRepositories[0].Eligible++
			case "repo-over-cohort":
				r.ValidationRepositories[0].Eligible = 300
			case "rows-count":
				input = rows[:len(rows)-1]
				r.NumericRowsSHA256, _ = RowsSHA256(input)
			}
			result, err := Fit(input, r)
			if err == nil {
				t.Fatal("readiness bypass")
			}
			for _, trial := range result.Trials {
				if trial.Attempted || trial.Fitted {
					t.Fatal("fitting started before readiness")
				}
			}
		})
	}
}

func TestFitFrozenTenCandidatesOnOwnedSyntheticRows(t *testing.T) {
	rows, ready := ownedReadyRows(t)
	before := slices.Clone(rows)
	result, err := Fit(rows, ready)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rows, before) {
		t.Fatal("fit mutated numeric inputs")
	}
	i := 0
	for _, penalty := range Penalties() {
		for _, seed := range Seeds() {
			trial := result.Trials[i]
			if trial.Lambda != penalty || trial.Seed != seed || !trial.Attempted || !trial.Fitted || trial.Epoch < 1 || trial.Epoch > 100 || trial.Failure != "" {
				t.Fatal("registered candidate missing", i, trial)
			}
			if trial.Training.Rows != 2400 || trial.Validation.Rows != 2400 {
				t.Fatal("row count changed during fit")
			}
			for _, coefficient := range trial.Weights {
				if math.IsNaN(float64(coefficient)) || math.IsInf(float64(coefficient), 0) {
					t.Fatal("nonfinite fitted coefficient")
				}
			}
			i++
		}
	}
	if config := frozenConfig(1729); config.Dimension != 16 || config.Mode != "fp32" || config.LearningRate != .1 || config.L2 != .0001 || config.Epochs != 100 || config.Batch != 64 {
		t.Fatal("configuration drift")
	}
	threshold := Threshold{Disabled: true, BudgetPercent: 90, ValidationQuestions: 2400, ValidationPages: 6000}
	model, err := NewModel(result.Trials[0], ready, threshold)
	if err != nil || model.Readiness != ready || model.Coefficients != result.Trials[0].Weights {
		t.Fatal("successful synthetic trial/model link", err)
	}
	other := ready
	other.InputSealSHA256 = strings.Repeat("9", 64)
	if _, err := NewModel(result.Trials[0], other, threshold); err == nil {
		t.Fatal("trial moved to a different input seal")
	}
	wrongCounts := result.Trials[0]
	wrongCounts.Training.Rows--
	if _, err := NewModel(wrongCounts, ready, threshold); err == nil {
		t.Fatal("trial count mismatch ignored")
	}
}

func TestFitRetainsFailedPreparationCandidates(t *testing.T) {
	rows, ready := ownedReadyRows(t)
	for i := range rows {
		rows[i].BaselinePages = rows[i].HelperPages
	}
	ready.NumericRowsSHA256, _ = RowsSHA256(rows)
	result, err := Fit(rows, ready)
	if err == nil {
		t.Fatal("zero-weight candidates were silently omitted")
	}
	for i, trial := range result.Trials {
		if i < 2 {
			if trial.Lambda != 0 || trial.Attempted || trial.Fitted || trial.Failure == "" || trial.Training.ZeroWeightRows != 2400 {
				t.Fatal("failed zero-weight preparation missing", i, trial)
			}
		} else if !trial.Attempted || !trial.Fitted || trial.Failure != "" {
			t.Fatal("later fixed candidates not preserved", i, trial)
		}
	}
}
