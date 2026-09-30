package pathclaim

import (
	"math"
	"reflect"
	"slices"
	"testing"

	"github.com/teamswyg/laya-tools/internal/fileeval"
	"github.com/teamswyg/laya-tools/internal/searchclaim"
)

func evaluationScore(rank int) fileeval.Score {
	return fileeval.Score{Targets: 1, Mapped: 1, FirstRank: rank, Hit1: rank == 1, Hit10: rank <= 10, Hit100: rank <= 100, All10: rank <= 10, ReciprocalRank: 1 / float64(rank)}
}

// Compact fixtures are authored arithmetic examples. They never cross the
// public >=2400 readiness gate or invoke training/search/source downloads.
func compactEvaluation(base, helper []int) []EvaluationRow {
	rows := make([]EvaluationRow, len(base))
	for i := range rows {
		b, h := (base[i]-1)*20+2, (helper[i]-1)*20+2
		rows[i] = EvaluationRow{Row: Row{Role: "validation", Repository: "numpy/numpy", Group: i, BaselinePages: base[i], HelperPages: helper[i]}, Baseline: evaluationScore(b), Helper: evaluationScore(h), WorkKnown: true, Work: fileeval.WorkStats{BaselineRankAttempts: 1, AuxiliaryBuildAttempts: 1, AuxiliaryRankAttempts: 1}}
		rows[i].Row.Features[0] = 1
	}
	return rows
}

func readyEvaluation(t *testing.T) ([]EvaluationRow, Readiness) {
	t.Helper()
	_, ready := ownedReadyRows(t)
	var rows []EvaluationRow
	for ri, repo := range ValidationRepositories() {
		for j := 0; j < ready.ValidationRepositories[ri].Eligible; j++ {
			row := compactEvaluation([]int{2}, []int{3})[0]
			row.Row.Repository, row.Row.Group = repo, len(rows)
			rows = append(rows, row)
		}
	}
	return rows, ready
}

func TestCalibrationTieGroupsCapDisabledAndInputOrder(t *testing.T) {
	rows := compactEvaluation([]int{3, 3, 3, 3, 3, 3}, []int{2, 4, 1, 2, 5, 1})
	scores := []float64{5, 4, 4, 3, 2, 1}
	before, scoreBefore := slices.Clone(rows), slices.Clone(scores)
	threshold, err := calibrateCore(rows, scores)
	if err != nil {
		t.Fatal(err)
	}
	if threshold.Disabled || threshold.MinimumScore != 3 || threshold.ValidationCalls != 4 || threshold.ValidationPages != 15 {
		t.Fatal("wrong whole-tie threshold", threshold)
	}
	if !reflect.DeepEqual(rows, before) || !reflect.DeepEqual(scores, scoreBefore) {
		t.Fatal("calibration mutated input")
	}
	rows[1], rows[2], scores[1], scores[2] = rows[2], rows[1], scores[2], scores[1]
	reordered, err := calibrateCore(rows, scores)
	if err != nil || threshold != reordered {
		t.Fatal("score tie order changed selection", err)
	}
	evaluation, err := evaluateThresholdCore(rows, scores, threshold)
	if err != nil || evaluation.Total.AuxiliaryCalls != 4 || evaluation.Total.Pages != 15 {
		t.Fatal("threshold replay differs", err)
	}
	for _, n := range []int{1, 10} {
		base, helper, equal := make([]int, n), make([]int, n), make([]float64, n)
		for i := range base {
			base[i], helper[i], equal[i] = 2, 1, 1
		}
		got, err := calibrateCore(compactEvaluation(base, helper), equal)
		if err != nil || !got.Disabled || got.ValidationCalls != 0 {
			t.Fatal("indivisible group split to fill floor cap", n, got, err)
		}
	}
	tied := compactEvaluation([]int{2, 2, 2}, []int{2, 2, 2})
	got, err := calibrateCore(tied, []float64{3, 2, 1})
	if err != nil || !got.Disabled || got.ValidationCalls != 0 {
		t.Fatal("disabled baseline lost equal-page tie", err)
	}
	negative := compactEvaluation([]int{2, 2}, []int{1, 3})
	got, err = calibrateCore(negative, []float64{-.5, -1})
	if err != nil || got.Disabled || got.MinimumScore != -.5 || got.ValidationCalls != 1 {
		t.Fatal("negative linear margins were clamped", got, err)
	}
}

func TestEvaluationPairedQualityCallsAndCounterfactualWork(t *testing.T) {
	rows := compactEvaluation([]int{3, 1, 2, 1}, []int{1, 3, 2, 1})
	rows[3].HelperFallback = true
	rows[3].Work.AuxiliaryRankAttempts = 0
	selection := []bool{true, true, true, true}
	got, err := Evaluate(rows, selection)
	if err != nil {
		t.Fatal(err)
	}
	if got.Total.Pages != 7 || got.Total.AuxiliaryCalls != 4 || got.Total.Wins != 1 || got.Total.Losses != 1 || got.Total.Ties != 2 || got.Total.Hit10 != 2 || got.Total.All10 != 2 {
		t.Fatal("paired outcome accounting", got.Total)
	}
	if got.Total.CounterfactualWork != (fileeval.WorkStats{BaselineRankAttempts: 4, AuxiliaryBuildAttempts: 4, AuxiliaryRankAttempts: 3}) || got.Total.SelectedFallbacks != 1 {
		t.Fatal("operation attribution or fallback lost", got.Total)
	}
	if got.ValidationMacroComplete {
		t.Fatal("single-repository sample described as five-repository macro")
	}
	baseline, err := Evaluate(rows, make([]bool, len(rows)))
	if err != nil || baseline.Total.Pages != 7 || baseline.Total.AuxiliaryCalls != 0 || baseline.Total.Wins != 0 || baseline.Total.Losses != 0 || baseline.Total.Ties != 4 || baseline.Total.CounterfactualWork.AuxiliaryBuildAttempts != 0 || baseline.Total.CounterfactualWork.BaselineRankAttempts != 4 {
		t.Fatal("baseline or counterfactual work", baseline.Total, err)
	}
	if baseline.InputSHA256 != got.InputSHA256 {
		t.Fatal("policy choice changed input evidence")
	}
	unknown := slices.Clone(rows)
	for i := range unknown {
		unknown[i].WorkKnown = false
		unknown[i].Work = fileeval.WorkStats{}
	}
	u, err := Evaluate(unknown, selection)
	if err != nil || u.Total.WorkKnownRows != 0 || u.Total.CounterfactualWork != (fileeval.WorkStats{}) || u.Total.AuxiliaryCalls != 4 {
		t.Fatal("unknown work fabricated", err)
	}
	train := rows[0]
	train.Row.Role, train.Row.Repository, train.Row.Group = "train", "Qiskit/qiskit", 99
	rows = append(rows, train)
	selection = append(selection, false)
	mixed, err := Evaluate(rows, selection)
	if err != nil || mixed.Roles[0].Metrics.Questions != 1 || mixed.Roles[1].Metrics.Questions != 4 || mixed.Total.Questions != 5 {
		t.Fatal("role aggregation", err)
	}
}

func TestEvaluationCorruptRowsAndCallDataRejected(t *testing.T) {
	for _, mode := range []string{"final", "duplicate-group", "negative-page", "over-page", "wrong-rank-page", "unmapped", "zero-target", "paired-target", "impossible-target-bound", "hit-flag", "all10", "impossible-all10-bound", "rank-nan", "work-negative", "work-over-one", "work-unknown", "work-missing-baseline", "fallback-change", "feature-nan", "score-nan", "score-inf", "selection-count", "score-count"} {
		t.Run(mode, func(t *testing.T) {
			rows := compactEvaluation([]int{2, 2}, []int{1, 3})
			scores := []float64{1, 0}
			selected := []bool{true, false}
			switch mode {
			case "final":
				rows[0].Row.Role = "final"
			case "duplicate-group":
				rows[1].Row.Group = rows[0].Row.Group
			case "negative-page":
				rows[0].Row.BaselinePages = -1
			case "over-page":
				rows[0].Row.HelperPages = 5001
			case "wrong-rank-page":
				rows[0].Row.BaselinePages = 3
			case "unmapped":
				rows[0].Baseline.Mapped = 0
			case "zero-target":
				rows[0].Baseline.Targets = 0
			case "paired-target":
				rows[0].Baseline.Targets = 2
				rows[0].Baseline.Mapped = 2
			case "impossible-target-bound":
				rows[0].Row.BaselinePages = 5000
				rows[0].Baseline = evaluationScore(100000)
				rows[0].Baseline.Targets, rows[0].Baseline.Mapped = 2, 2
				rows[0].Helper.Targets, rows[0].Helper.Mapped = 2, 2
			case "hit-flag":
				rows[0].Helper.Hit10 = false
			case "all10":
				rows[0].Helper.All10 = false
			case "impossible-all10-bound":
				rows[0].Helper = evaluationScore(10)
				rows[0].Baseline.Targets, rows[0].Baseline.Mapped = 2, 2
				rows[0].Helper.Targets, rows[0].Helper.Mapped = 2, 2
			case "rank-nan":
				rows[0].Baseline.ReciprocalRank = math.NaN()
			case "work-negative":
				rows[0].Work.AuxiliaryRankAttempts = -1
			case "work-over-one":
				rows[0].Work.AuxiliaryBuildAttempts = 2
			case "work-unknown":
				rows[0].WorkKnown = false
			case "work-missing-baseline":
				rows[0].Work.BaselineRankAttempts = 0
			case "fallback-change":
				rows[0].HelperFallback = true
			case "feature-nan":
				rows[0].Row.Features[3] = math.NaN()
			case "score-nan":
				scores[0] = math.NaN()
			case "score-inf":
				scores[0] = math.Inf(1)
			case "selection-count":
				selected = selected[:1]
			case "score-count":
				scores = scores[:1]
			}
			if mode == "selection-count" {
				if _, err := Evaluate(rows, selected); err == nil {
					t.Fatal("selection count ignored")
				}
			} else if _, err := calibrateCore(rows, scores); err == nil {
				t.Fatal("corrupt calibration input accepted")
			}
		})
	}
}

func TestEvaluationMetricBoundsRejectImpossibleSelections(t *testing.T) {
	base := Metrics{Questions: 3, Pages: 6, AuxiliaryCalls: 1, Wins: 1, Ties: 2, WorkKnownRows: 3, CounterfactualWork: fileeval.WorkStats{BaselineRankAttempts: 3, AuxiliaryBuildAttempts: 1, AuxiliaryRankAttempts: 1}}
	if !validMetrics(base) {
		t.Fatal("valid compact metrics rejected")
	}
	for _, mode := range []string{"negative-calls", "too-many-calls", "unselected-gain", "overflow-wins", "fallback-win", "unknown-work-attribution"} {
		bad := base
		switch mode {
		case "negative-calls":
			bad.AuxiliaryCalls = -1
		case "too-many-calls":
			bad.AuxiliaryCalls = 4
		case "unselected-gain":
			bad.Wins, bad.Ties = 2, 1
		case "overflow-wins":
			bad.Questions, bad.Pages, bad.AuxiliaryCalls, bad.WorkKnownRows = 0, 0, 0, 0
			bad.Wins, bad.Losses, bad.Ties = math.MaxInt, math.MaxInt, 2
			bad.CounterfactualWork = fileeval.WorkStats{}
		case "fallback-win":
			bad.Wins, bad.Ties, bad.AuxiliaryCalls, bad.SelectedFallbacks = 3, 0, 3, 1
		case "unknown-work-attribution":
			bad.WorkKnownRows, bad.CounterfactualWork.BaselineRankAttempts = 0, 0
		}
		if validMetrics(bad) {
			t.Fatal("impossible policy metrics accepted", mode)
		}
	}
}

func TestCalibrationPublicReadinessGateAndThresholdApplication(t *testing.T) {
	rows, ready := readyEvaluation(t)
	scores := make([]float64, len(rows))
	for i := range rows {
		if i%2 == 0 {
			rows[i].Row.HelperPages = 1
			rows[i].Helper = evaluationScore(2)
			scores[i] = 1
		}
	}
	threshold, err := CalibrateValidation(rows, scores, ready)
	if err != nil || threshold.ValidationCalls != 1200 || threshold.ValidationPages != 3600 {
		t.Fatal("ready calibration", threshold, err)
	}
	selected, err := EvaluateThreshold(rows, scores, threshold)
	if err != nil || selected.Total.Pages != 3600 || selected.Total.AuxiliaryCalls != 1200 || !selected.ValidationMacroComplete {
		t.Fatal("ready threshold replay", err)
	}
	if _, err := CalibrateValidation(rows[:len(rows)-1], scores[:len(scores)-1], ready); err == nil {
		t.Fatal("manifest/actual count mismatch accepted")
	}
	wrongRepo := slices.Clone(rows)
	wrongRepo[0].Row.Repository = "numpy/numpy"
	if _, err := CalibrateValidation(wrongRepo, scores, ready); err == nil {
		t.Fatal("manifest repository count mismatch accepted")
	}
	small := compactEvaluation([]int{2}, []int{1})
	if _, err := CalibrateValidation(small, []float64{1}, ready); err == nil {
		t.Fatal("compact sample bypassed public readiness")
	}
	badReady := ready
	badReady.ValidationEligible = 2399
	if _, err := CalibrateValidation(rows, scores, badReady); err == nil {
		t.Fatal("minimum validation count bypassed")
	}
	badReady = ready
	badReady.SourceEvidenceSHA256 = ""
	if _, err := CalibrateValidation(rows, scores, badReady); err == nil {
		t.Fatal("source evidence requirement bypassed")
	}
	train := slices.Clone(rows)
	for i := range train {
		train[i].Row.Role = "train"
		train[i].Row.Repository = "Qiskit/qiskit"
	}
	if _, err := CalibrateValidation(train, scores, ready); err == nil {
		t.Fatal("training labels used for threshold")
	}
	if _, err := EvaluateThreshold(train, scores, threshold); err != nil {
		t.Fatal("frozen threshold cannot evaluate training", err)
	}
	bad := threshold
	bad.ValidationCalls = 2161
	if _, err := EvaluateThreshold(rows, scores, bad); err == nil {
		t.Fatal("corrupt call cap accepted")
	}
}

func TestGapScoresTargetIndependenceAndFeatureContract(t *testing.T) {
	names := searchclaim.PathFeatureNames()
	if names[9] != "top_two_gap_squashed" {
		t.Fatal("baseline gap schema index changed")
	}
	rows := []Row{{Features: [16]float64{9: .2}}, {Features: [16]float64{9: .8}}}
	a, err := BaselineGapScores(rows)
	if err != nil || !reflect.DeepEqual(a, []float64{-.2, -.8}) {
		t.Fatal("gap orientation", a, err)
	}
	rows[0].Role = "final"
	rows[0].Repository = "unread/raw"
	rows[0].BaselinePages = -1
	rows[0].HelperPages = 99999
	b, err := BaselineGapScores(rows)
	if err != nil || !reflect.DeepEqual(a, b) {
		t.Fatal("labels/bookkeeping entered gap scores", err)
	}
	rows[0].Features[9] = math.NaN()
	if _, err := BaselineGapScores(rows); err == nil {
		t.Fatal("invalid gap feature accepted")
	}
}

func TestSelectionFixedPenaltyOrderSeedIsolationAndTieBreaks(t *testing.T) {
	rows, ready := readyEvaluation(t)
	var candidates [5]ScoredCandidate
	for i, lambda := range Penalties() {
		candidates[i] = ScoredCandidate{Lambda: lambda, Seed: 1729, Epoch: 4, Scores: make([]float64, len(rows))}
	}
	// Constant scores have one whole group exceeding the cap, so disabled wins.
	selection, err := SelectValidation(rows, ready, candidates)
	if err != nil || selection.Selected != 0 {
		t.Fatal("smaller lambda disabled tie", err)
	}
	for _, candidate := range selection.Candidates {
		if !candidate.Threshold.Disabled || candidate.Threshold.ValidationCalls != 0 {
			t.Fatal("constant-score cap failed")
		}
	}
	for i := range rows {
		if i%2 == 0 {
			rows[i].Row.HelperPages = 1
			rows[i].Helper = evaluationScore(2)
			candidates[2].Scores[i] = 1
		}
	}
	selection, err = SelectValidation(rows, ready, candidates)
	if err != nil || selection.Selected != 2 || selection.Candidates[2].Threshold.ValidationPages != 3600 {
		t.Fatal("page winner not selected", err)
	}
	for _, mode := range []string{"lambda-order", "mixed-seed", "bad-epoch", "bad-score", "bad-count"} {
		bad := candidates
		switch mode {
		case "lambda-order":
			bad[0].Lambda = .25
		case "mixed-seed":
			bad[4].Seed = 2718
		case "bad-epoch":
			bad[4].Epoch = 101
		case "bad-score":
			bad[4].Scores = slices.Clone(bad[4].Scores)
			bad[4].Scores[0] = math.Inf(1)
		case "bad-count":
			bad[4].Scores = bad[4].Scores[:1]
		}
		if _, err := SelectValidation(rows, ready, bad); err == nil {
			t.Fatal("invalid candidate family accepted", mode)
		}
	}
	a := CandidateEvaluation{Lambda: 1, Epoch: 4, Threshold: Threshold{ValidationPages: 10, ValidationCalls: 3, MinimumScore: .5}}
	for _, mode := range []string{"pages", "calls", "lambda", "threshold", "epoch"} {
		b := a
		switch mode {
		case "pages":
			b.Threshold.ValidationPages--
		case "calls":
			b.Threshold.ValidationCalls--
		case "lambda":
			b.Lambda = .25
		case "threshold":
			b.Threshold.MinimumScore = .75
		case "epoch":
			b.Epoch = 3
		}
		if !betterCandidate(b, a) || betterCandidate(a, b) {
			t.Fatal("deterministic tie precedence", mode)
		}
	}
}

func pilotFixtures(t *testing.T) ([]EvaluationRow, Readiness, Evaluation, Evaluation, [2]CandidateEvaluation) {
	t.Helper()
	rows, ready := readyEvaluation(t)
	selected := make([]bool, len(rows))
	for i := range rows {
		if i%10 == 0 {
			rows[i].Row.HelperPages = 1
			rows[i].Helper = evaluationScore(2)
			selected[i] = true
		}
	}
	baseline, err := Evaluate(rows, make([]bool, len(rows)))
	if err != nil {
		t.Fatal(err)
	}
	all := make([]bool, len(rows))
	for i := range all {
		all[i] = true
	}
	always, err := Evaluate(rows, all)
	if err != nil {
		t.Fatal(err)
	}
	m, err := Evaluate(rows, selected)
	if err != nil {
		t.Fatal(err)
	}
	var candidates [2]CandidateEvaluation
	for i, seed := range Seeds() {
		candidates[i] = CandidateEvaluation{Lambda: 1, Seed: seed, Epoch: 4, Threshold: Threshold{MinimumScore: 1, BudgetPercent: 90, ValidationQuestions: len(rows), ValidationCalls: m.Total.AuxiliaryCalls, ValidationPages: m.Total.Pages}, Metrics: m}
	}
	return rows, ready, baseline, always, candidates
}

func TestPilotGateExactBoundaryBothSeedsAndCorruptMetrics(t *testing.T) {
	_, ready, baseline, always, selected := pilotFixtures(t)
	gate, err := PilotGate(ready, baseline, always, selected)
	if err != nil || !gate.BothPass {
		t.Fatal("exact five-percent boundary", gate, err)
	}
	selected[1].Metrics = baseline
	selected[1].Threshold = Threshold{Disabled: true, BudgetPercent: 90, ValidationQuestions: 2400, ValidationPages: baseline.Total.Pages}
	gate, err = PilotGate(ready, baseline, always, selected)
	if err != nil || gate.BothPass || !gate.Seeds[0].Pass || gate.Seeds[1].BaselinePages {
		t.Fatal("favorable seed hides failed replication", gate, err)
	}
	_, _, _, _, selected = pilotFixtures(t)
	for _, mode := range []string{"seed-order", "input", "aggregate", "macro", "cap", "threshold-outcome"} {
		bad := selected
		switch mode {
		case "seed-order":
			bad[1].Seed = 1729
		case "input":
			bad[1].Metrics.InputSHA256 = baseline.InputSHA256[:63] + "0"
			if bad[1].Metrics.InputSHA256 == baseline.InputSHA256 {
				bad[1].Metrics.InputSHA256 = baseline.InputSHA256[:63] + "1"
			}
		case "aggregate":
			bad[1].Metrics.Total.Pages++
		case "macro":
			bad[1].Metrics.ValidationMacroHit10++
		case "cap":
			bad[1].Threshold.ValidationCalls = 2161
		case "threshold-outcome":
			bad[1].Threshold.ValidationPages++
		}
		if _, err := PilotGate(ready, baseline, always, bad); err == nil {
			t.Fatal("corrupt gate metadata accepted", mode)
		}
	}
}

func TestPilotPooledGainCannotHideMacroQualityLoss(t *testing.T) {
	rows, ready := readyEvaluation(t)
	selected := make([]bool, len(rows))
	var pandas, pyca int
	for i := range rows {
		switch rows[i].Row.Repository {
		case "pandas-dev/pandas":
			if pandas < 500 {
				rows[i].Row.HelperPages = 1
				rows[i].Helper = evaluationScore(2)
				selected[i] = true
			}
			pandas++
		case "pyca/cryptography":
			if pyca < 120 {
				rows[i].Row.BaselinePages = 1
				rows[i].Baseline = evaluationScore(2)
				rows[i].Row.HelperPages = 1
				rows[i].Helper = evaluationScore(11)
				selected[i] = true
			}
			pyca++
		}
	}
	base, err := Evaluate(rows, make([]bool, len(rows)))
	if err != nil {
		t.Fatal(err)
	}
	all := make([]bool, len(rows))
	for i := range all {
		all[i] = true
	}
	always, err := Evaluate(rows, all)
	if err != nil {
		t.Fatal(err)
	}
	m, err := Evaluate(rows, selected)
	if err != nil {
		t.Fatal(err)
	}
	var candidates [2]CandidateEvaluation
	for i, seed := range Seeds() {
		candidates[i] = CandidateEvaluation{Lambda: 1, Seed: seed, Epoch: 4, Threshold: Threshold{MinimumScore: 1, BudgetPercent: 90, ValidationQuestions: len(rows), ValidationCalls: m.Total.AuxiliaryCalls, ValidationPages: m.Total.Pages}, Metrics: m}
	}
	gate, err := PilotGate(ready, base, always, candidates)
	if err != nil || gate.BothPass || !gate.Seeds[0].PooledHit10 || gate.Seeds[0].MacroHit10 || !gate.Seeds[0].BaselinePages {
		t.Fatal("pooled count hides small-repository quality loss", gate, err)
	}
}

func TestPilotAlwaysHelperOnePercentAndRepositoryFivePercentBoundaries(t *testing.T) {
	for _, extraHarm := range []bool{false, true} {
		t.Run("always-helper-extra-harm="+boolName(extraHarm), func(t *testing.T) {
			rows, ready := readyEvaluation(t)
			selection := make([]bool, len(rows))
			for i := range rows {
				rows[i].Row.HelperPages = 2
				rows[i].Helper = evaluationScore(22)
				if i < 345 {
					rows[i].Row.HelperPages = 1
					rows[i].Helper = evaluationScore(2)
					selection[i] = true
				} else if i < 390 {
					rows[i].Row.HelperPages = 3
					rows[i].Helper = evaluationScore(42)
				}
			}
			selection[345] = extraHarm
			base, always, candidates := gatePolicies(t, rows, selection)
			got, err := PilotGate(ready, base, always, candidates)
			if err != nil || got.BothPass == extraHarm || got.Seeds[0].AlwaysHelperPages == extraHarm || !got.Seeds[0].BaselinePages {
				t.Fatal("one-percent boundary was rounded", got, err)
			}
			if always.Total.Pages != 4500 || candidates[0].Metrics.Total.Pages != 4455+intBool(extraHarm) {
				t.Fatal("fixture is not the exact one-percent boundary")
			}
		})
	}
	for _, harms := range []int{20, 21} {
		rows, ready := readyEvaluation(t)
		selection := make([]bool, len(rows))
		var lightning, pandas int
		for i := range rows {
			switch rows[i].Row.Repository {
			case "Lightning-AI/lightning":
				selection[i] = lightning < harms
				lightning++
			case "pandas-dev/pandas":
				if pandas < 500 {
					rows[i].Row.HelperPages = 1
					rows[i].Helper = evaluationScore(2)
					selection[i] = true
				}
				pandas++
			}
		}
		base, always, candidates := gatePolicies(t, rows, selection)
		got, err := PilotGate(ready, base, always, candidates)
		want := harms == 20
		if err != nil || got.BothPass != want || got.Seeds[0].RepositoryPages[0] != want || !got.Seeds[0].BaselinePages || !got.Seeds[0].AlwaysHelperPages {
			t.Fatal("pooled gains hid repository boundary", harms, got, err)
		}
	}
}

func intBool(value bool) int {
	if value {
		return 1
	}
	return 0
}

func boolName(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func gatePolicies(t *testing.T, rows []EvaluationRow, selection []bool) (Evaluation, Evaluation, [2]CandidateEvaluation) {
	t.Helper()
	base, err := Evaluate(rows, make([]bool, len(rows)))
	if err != nil {
		t.Fatal(err)
	}
	all := make([]bool, len(rows))
	for i := range all {
		all[i] = true
	}
	always, err := Evaluate(rows, all)
	if err != nil {
		t.Fatal(err)
	}
	m, err := Evaluate(rows, selection)
	if err != nil {
		t.Fatal(err)
	}
	var candidates [2]CandidateEvaluation
	for i, seed := range Seeds() {
		candidates[i] = CandidateEvaluation{Lambda: 1, Seed: seed, Epoch: 4, Threshold: Threshold{MinimumScore: 1, BudgetPercent: 90, ValidationQuestions: len(rows), ValidationCalls: m.Total.AuxiliaryCalls, ValidationPages: m.Total.Pages}, Metrics: m}
	}
	return base, always, candidates
}

func TestControlsUseSameRowsAndNinetyPercentCalibration(t *testing.T) {
	rows, ready := readyEvaluation(t)
	scores := make([]float64, len(rows))
	for i := range rows {
		rows[i].Row.Features[9] = .8
		if i%2 == 0 {
			rows[i].Row.HelperPages = 1
			rows[i].Helper = evaluationScore(2)
			rows[i].Row.Features[9] = .2
			scores[i] = 1
		}
	}
	controls, err := ControlsForValidation(rows, ready, ScoredCandidate{Lambda: 0, Seed: 1729, Epoch: 4, Scores: scores})
	if err != nil {
		t.Fatal(err)
	}
	if controls.Baseline.Total.AuxiliaryCalls != 0 || controls.AlwaysHelper.Total.AuxiliaryCalls != len(rows) || controls.Gap.Total.AuxiliaryCalls != 1200 || controls.GapThreshold.MinimumScore != -.2 || controls.ZeroPenalty.Metrics.Total.Pages != controls.Gap.Total.Pages {
		t.Fatal("control policy mismatch", controls)
	}
	for _, control := range []Evaluation{controls.AlwaysHelper, controls.Gap, controls.ZeroPenalty.Metrics} {
		if control.InputSHA256 != controls.Baseline.InputSHA256 {
			t.Fatal("control cohorts differ")
		}
	}
	if _, err := ControlsForValidation(rows, ready, ScoredCandidate{Lambda: 1, Seed: 1729, Epoch: 4, Scores: scores}); err == nil {
		t.Fatal("nonzero penalty substituted for zero control")
	}
}
