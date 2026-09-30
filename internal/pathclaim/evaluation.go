package pathclaim

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"slices"

	"github.com/teamswyg/laya-tools/internal/fileeval"
)

// EvaluationRow holds measured paired outcomes, never model inference inputs.
// Work describes the original baseline+helper collection. Selecting a policy
// here does not execute search, prove skipped work or measure runtime savings.
type EvaluationRow struct {
	Row              Row
	Baseline, Helper fileeval.Score
	WorkKnown        bool
	Work             fileeval.WorkStats
	HelperFallback   bool
}

type Metrics struct {
	// AuxiliaryCalls counts replayed policy selections, including page ties
	// and fallbacks. It is not an operation or latency measured in this replay.
	Questions, Pages, AuxiliaryCalls int
	Wins, Losses, Ties               int
	Hit1, Hit10, Hit100, All10       int
	WorkKnownRows, SelectedFallbacks int
	// CounterfactualWork attributes recorded operation counts to this policy.
	// It is distinct from actual selected execution or measured time/memory.
	CounterfactualWork fileeval.WorkStats
}

type MetricGroup struct {
	Role, Repository string
	Metrics          Metrics
}

type Evaluation struct {
	InputSHA256             string
	Total                   Metrics
	Roles                   [2]MetricGroup
	Repositories            [25]MetricGroup
	ValidationMacroHit10    float64
	ValidationMacroComplete bool
}

func emptyEvaluation() Evaluation {
	var out Evaluation
	out.Roles[0].Role, out.Roles[1].Role = "train", "validation"
	for i, repo := range trainingRepositories() {
		out.Repositories[i] = MetricGroup{Role: "train", Repository: repo}
	}
	for i, repo := range ValidationRepositories() {
		out.Repositories[20+i] = MetricGroup{Role: "validation", Repository: repo}
	}
	return out
}

func repositorySlot(role, repository string) int {
	if role == "train" {
		repos := trainingRepositories()
		return slices.Index(repos[:], repository)
	}
	repos := ValidationRepositories()
	if i := slices.Index(repos[:], repository); i >= 0 {
		return 20 + i
	}
	return -1
}

func validateScore(score fileeval.Score, pages int) error {
	if score.Targets < 1 || score.Targets > 100000 || score.Mapped != score.Targets || score.FirstRank < 1 || score.FirstRank > 100000 || score.Targets+score.FirstRank-1 > 100000 || (score.FirstRank+19)/20 != pages {
		return fmt.Errorf("invalid fully mapped page outcome")
	}
	if score.Hit1 != (score.FirstRank == 1) || score.Hit10 != (score.FirstRank <= 10) || score.Hit100 != (score.FirstRank <= 100) || (score.All10 && (!score.Hit10 || score.FirstRank+score.Targets-1 > 10)) || (score.Targets == 1 && score.All10 != score.Hit10) {
		return fmt.Errorf("inconsistent quality flags")
	}
	if math.IsNaN(score.ReciprocalRank) || math.IsInf(score.ReciprocalRank, 0) || score.ReciprocalRank != 1/float64(score.FirstRank) {
		return fmt.Errorf("inconsistent reciprocal rank")
	}
	return nil
}

func validateEvaluationRows(rows []EvaluationRow) error {
	if len(rows) == 0 || len(rows) > MaxRows {
		return fmt.Errorf("evaluation row bound")
	}
	numeric := make([]Row, len(rows))
	for i, row := range rows {
		numeric[i] = row.Row
	}
	if err := validateRows(numeric); err != nil {
		return err
	}
	for _, row := range rows {
		if err := validateScore(row.Baseline, row.Row.BaselinePages); err != nil {
			return err
		}
		if err := validateScore(row.Helper, row.Row.HelperPages); err != nil {
			return err
		}
		if row.Baseline.Targets != row.Helper.Targets {
			return fmt.Errorf("paired target count mismatch")
		}
		w := row.Work
		if w.BaselineRankAttempts < 0 || w.AuxiliaryBuildAttempts < 0 || w.AuxiliaryRankAttempts < 0 || w.BaselineRankAttempts > 1 || w.AuxiliaryBuildAttempts > 1 || w.AuxiliaryRankAttempts > 1 {
			return fmt.Errorf("invalid per-task operation counts")
		}
		if (row.WorkKnown && w.BaselineRankAttempts != 1) || (!row.WorkKnown && w != (fileeval.WorkStats{})) {
			return fmt.Errorf("inconsistent operation availability")
		}
		if row.HelperFallback && (row.Baseline != row.Helper || row.Row.BaselinePages != row.Row.HelperPages) {
			return fmt.Errorf("fallback changed measured outcomes")
		}
	}
	return nil
}

func addMetrics(dst *Metrics, src Metrics) {
	dst.Questions += src.Questions
	dst.Pages += src.Pages
	dst.AuxiliaryCalls += src.AuxiliaryCalls
	dst.Wins += src.Wins
	dst.Losses += src.Losses
	dst.Ties += src.Ties
	dst.Hit1 += src.Hit1
	dst.Hit10 += src.Hit10
	dst.Hit100 += src.Hit100
	dst.All10 += src.All10
	dst.WorkKnownRows += src.WorkKnownRows
	dst.SelectedFallbacks += src.SelectedFallbacks
	dst.CounterfactualWork.BaselineRankAttempts += src.CounterfactualWork.BaselineRankAttempts
	dst.CounterfactualWork.AuxiliaryBuildAttempts += src.CounterfactualWork.AuxiliaryBuildAttempts
	dst.CounterfactualWork.AuxiliaryRankAttempts += src.CounterfactualWork.AuxiliaryRankAttempts
}

func rowMetrics(row EvaluationRow, selected bool) Metrics {
	score, pages := row.Baseline, row.Row.BaselinePages
	out := Metrics{Questions: 1}
	if row.WorkKnown {
		out.WorkKnownRows = 1
		out.CounterfactualWork.BaselineRankAttempts = row.Work.BaselineRankAttempts
	}
	if selected {
		out.AuxiliaryCalls = 1
		score, pages = row.Helper, row.Row.HelperPages
		if row.HelperFallback {
			out.SelectedFallbacks = 1
		}
		if row.WorkKnown {
			out.CounterfactualWork.AuxiliaryBuildAttempts = row.Work.AuxiliaryBuildAttempts
			out.CounterfactualWork.AuxiliaryRankAttempts = row.Work.AuxiliaryRankAttempts
		}
	}
	out.Pages = pages
	switch {
	case pages < row.Row.BaselinePages:
		out.Wins = 1
	case pages > row.Row.BaselinePages:
		out.Losses = 1
	default:
		out.Ties = 1
	}
	if score.Hit1 {
		out.Hit1 = 1
	}
	if score.Hit10 {
		out.Hit10 = 1
	}
	if score.Hit100 {
		out.Hit100 = 1
	}
	if score.All10 {
		out.All10 = 1
	}
	return out
}

func macroHit10(out Evaluation) (float64, bool) {
	var sum float64
	for i := 20; i < 25; i++ {
		m := out.Repositories[i].Metrics
		if m.Questions == 0 {
			return 0, false
		}
		sum += float64(m.Hit10) / float64(m.Questions)
	}
	return sum / 5, true
}

// Evaluate is a pure counterfactual comparison on supplied measured outcomes.
// It accepts bounded train/validation rows for diagnostics; no final row is
// permitted. Full unavailable/excluded coverage belongs to the sealed runner.
func Evaluate(rows []EvaluationRow, selected []bool) (Evaluation, error) {
	if len(rows) != len(selected) {
		return Evaluation{}, fmt.Errorf("selection row count mismatch")
	}
	if err := validateEvaluationRows(rows); err != nil {
		return Evaluation{}, err
	}
	out := emptyEvaluation()
	b, err := json.Marshal(rows)
	if err != nil {
		return Evaluation{}, err
	}
	h := sha256.Sum256(b)
	out.InputSHA256 = hex.EncodeToString(h[:])
	for i, row := range rows {
		m := rowMetrics(row, selected[i])
		role := 0
		if row.Row.Role == "validation" {
			role = 1
		}
		addMetrics(&out.Total, m)
		addMetrics(&out.Roles[role].Metrics, m)
		addMetrics(&out.Repositories[repositorySlot(row.Row.Role, row.Row.Repository)].Metrics, m)
	}
	out.ValidationMacroHit10, out.ValidationMacroComplete = macroHit10(out)
	return out, nil
}

func validateScores(rows []EvaluationRow, scores []float64) error {
	if len(rows) != len(scores) {
		return fmt.Errorf("score row count mismatch")
	}
	for _, score := range scores {
		if math.IsNaN(score) || math.IsInf(score, 0) {
			return fmt.Errorf("nonfinite linear selection score")
		}
	}
	return nil
}

func validateValidation(rows []EvaluationRow, ready Readiness) error {
	if err := ready.Validate(); err != nil {
		return err
	}
	if err := validateEvaluationRows(rows); err != nil {
		return err
	}
	if len(rows) != ready.ValidationEligible {
		return fmt.Errorf("actual validation count differs from readiness")
	}
	var counts [5]int
	for _, row := range rows {
		if row.Row.Role != "validation" {
			return fmt.Errorf("calibration requires validation only")
		}
		counts[repositorySlot("validation", row.Row.Repository)-20]++
	}
	for i, n := range counts {
		if n != ready.ValidationRepositories[i].Eligible {
			return fmt.Errorf("actual validation repository count differs from readiness")
		}
	}
	return nil
}

// calibrateCore is private so compact authored tests cannot become a public
// readiness override. Production selection goes through CalibrateValidation.
func calibrateCore(rows []EvaluationRow, scores []float64) (Threshold, error) {
	if err := validateEvaluationRows(rows); err != nil {
		return Threshold{}, err
	}
	for _, row := range rows {
		if row.Row.Role != "validation" {
			return Threshold{}, fmt.Errorf("calibration requires validation only")
		}
	}
	if err := validateScores(rows, scores); err != nil {
		return Threshold{}, err
	}
	best := Threshold{Disabled: true, BudgetPercent: 90, ValidationQuestions: len(rows)}
	type entry struct {
		score float64
		delta int
	}
	entries := make([]entry, len(rows))
	for i, row := range rows {
		best.ValidationPages += row.Row.BaselinePages
		entries[i] = entry{scores[i], row.Row.HelperPages - row.Row.BaselinePages}
	}
	slices.SortFunc(entries, func(a, b entry) int { return cmp.Compare(b.score, a.score) })
	pages, cap := best.ValidationPages, len(rows)*90/100
	for start := 0; start < len(entries); {
		end := start
		for end < len(entries) && entries[end].score == entries[start].score {
			pages += entries[end].delta
			end++
		}
		if end > cap {
			break
		}
		candidate := Threshold{MinimumScore: entries[start].score, BudgetPercent: 90, ValidationQuestions: len(rows), ValidationCalls: end, ValidationPages: pages}
		if candidate.ValidationPages < best.ValidationPages || (candidate.ValidationPages == best.ValidationPages && (candidate.ValidationCalls < best.ValidationCalls || (candidate.ValidationCalls == best.ValidationCalls && candidate.MinimumScore > best.MinimumScore))) {
			best = candidate
		}
		start = end
	}
	return best, nil
}

// CalibrateValidation chooses only a development threshold. The runner must
// first VerifyReadiness on the full numeric projection and verify source/seal
// evidence. This API independently enforces >=2400 validation rows, all five
// repository counts and the frozen source-provenance manifest shape.
func CalibrateValidation(validation []EvaluationRow, scores []float64, ready Readiness) (Threshold, error) {
	if err := validateValidation(validation, ready); err != nil {
		return Threshold{}, err
	}
	return calibrateCore(validation, scores)
}

func validateThreshold(t Threshold, compact bool) error {
	minimum := MinimumEligible
	if compact {
		minimum = 1
	}
	if t.BudgetPercent != 90 || t.ValidationQuestions < minimum || t.ValidationQuestions > 5686 || t.ValidationCalls < 0 || t.ValidationCalls > t.ValidationQuestions*90/100 || t.ValidationPages < t.ValidationQuestions || t.ValidationPages > t.ValidationQuestions*MaxPages || math.IsNaN(t.MinimumScore) || math.IsInf(t.MinimumScore, 0) || (t.Disabled && t.ValidationCalls != 0) {
		return fmt.Errorf("invalid fixed-budget threshold")
	}
	return nil
}

func evaluateThresholdCore(rows []EvaluationRow, scores []float64, t Threshold) (Evaluation, error) {
	if err := validateScores(rows, scores); err != nil {
		return Evaluation{}, err
	}
	selected := make([]bool, len(rows))
	for i, score := range scores {
		selected[i] = !t.Disabled && score >= t.MinimumScore
	}
	return Evaluate(rows, selected)
}

// EvaluateThreshold applies an already frozen development threshold to train
// and/or validation rows. Counterfactual calls may exceed 90% on training;
// calibration's cap is defined on its eligible validation population only.
func EvaluateThreshold(rows []EvaluationRow, scores []float64, threshold Threshold) (Evaluation, error) {
	if err := validateThreshold(threshold, false); err != nil {
		return Evaluation{}, err
	}
	return evaluateThresholdCore(rows, scores, threshold)
}

// BaselineGapScores reads only the fixed feature contract's index 9:
// top_two_gap_squashed. A smaller baseline gap receives a larger linear score.
// Target costs/labels are not inspected or used to orient this heuristic.
func BaselineGapScores(rows []Row) ([]float64, error) {
	if len(rows) == 0 || len(rows) > MaxRows {
		return nil, fmt.Errorf("gap score row bound")
	}
	out := make([]float64, len(rows))
	for i, row := range rows {
		if err := validateFeatures(row.Features); err != nil {
			return nil, err
		}
		out[i] = -row.Features[9]
	}
	return out, nil
}

type ScoredCandidate struct {
	Lambda float64
	Seed   uint64
	Epoch  int
	Scores []float64
}

type CandidateEvaluation struct {
	Lambda    float64
	Seed      uint64
	Epoch     int
	Threshold Threshold
	Metrics   Evaluation
}

type ValidationSelection struct {
	Candidates [5]CandidateEvaluation
	Selected   int
}

func betterCandidate(a, b CandidateEvaluation) bool {
	if a.Threshold.ValidationPages != b.Threshold.ValidationPages {
		return a.Threshold.ValidationPages < b.Threshold.ValidationPages
	}
	if a.Threshold.ValidationCalls != b.Threshold.ValidationCalls {
		return a.Threshold.ValidationCalls < b.Threshold.ValidationCalls
	}
	if a.Lambda != b.Lambda {
		return a.Lambda < b.Lambda
	}
	if a.Threshold.MinimumScore != b.Threshold.MinimumScore {
		return a.Threshold.MinimumScore > b.Threshold.MinimumScore
	}
	return a.Epoch < b.Epoch
}

// SelectValidation compares the five fixed penalties for one fixed seed. It
// offers no function for choosing a favorable seed. Inputs contain linear
// scores, never cost-weighted sigmoid probabilities or final labels.
func SelectValidation(validation []EvaluationRow, ready Readiness, candidates [5]ScoredCandidate) (ValidationSelection, error) {
	var out ValidationSelection
	if err := validateValidation(validation, ready); err != nil {
		return out, err
	}
	seed := candidates[0].Seed
	if seed != 1729 && seed != 2718 {
		return out, fmt.Errorf("unregistered selection seed")
	}
	for i, penalty := range Penalties() {
		candidate := candidates[i]
		if candidate.Lambda != penalty || candidate.Seed != seed || candidate.Epoch < 1 || candidate.Epoch > 100 {
			return ValidationSelection{}, fmt.Errorf("candidate order or metadata mismatch")
		}
		threshold, err := calibrateCore(validation, candidate.Scores)
		if err != nil {
			return ValidationSelection{}, err
		}
		metrics, err := evaluateThresholdCore(validation, candidate.Scores, threshold)
		if err != nil {
			return ValidationSelection{}, err
		}
		out.Candidates[i] = CandidateEvaluation{candidate.Lambda, candidate.Seed, candidate.Epoch, threshold, metrics}
		if i > 0 && betterCandidate(out.Candidates[i], out.Candidates[out.Selected]) {
			out.Selected = i
		}
	}
	return out, nil
}

type Controls struct {
	Baseline, AlwaysHelper Evaluation
	GapThreshold           Threshold
	Gap                    Evaluation
	ZeroPenalty            CandidateEvaluation
}

func ControlsForValidation(validation []EvaluationRow, ready Readiness, zero ScoredCandidate) (Controls, error) {
	var out Controls
	if err := validateValidation(validation, ready); err != nil {
		return out, err
	}
	if zero.Lambda != 0 || (zero.Seed != 1729 && zero.Seed != 2718) || zero.Epoch < 1 || zero.Epoch > 100 {
		return out, fmt.Errorf("invalid zero-penalty control")
	}
	selected := make([]bool, len(validation))
	var err error
	out.Baseline, err = Evaluate(validation, selected)
	if err != nil {
		return Controls{}, err
	}
	for i := range selected {
		selected[i] = true
	}
	out.AlwaysHelper, err = Evaluate(validation, selected)
	if err != nil {
		return Controls{}, err
	}
	rows := make([]Row, len(validation))
	for i, row := range validation {
		rows[i] = row.Row
	}
	scores, err := BaselineGapScores(rows)
	if err != nil {
		return Controls{}, err
	}
	out.GapThreshold, err = calibrateCore(validation, scores)
	if err != nil {
		return Controls{}, err
	}
	out.Gap, err = evaluateThresholdCore(validation, scores, out.GapThreshold)
	if err != nil {
		return Controls{}, err
	}
	t, err := calibrateCore(validation, zero.Scores)
	if err != nil {
		return Controls{}, err
	}
	m, err := evaluateThresholdCore(validation, zero.Scores, t)
	if err != nil {
		return Controls{}, err
	}
	out.ZeroPenalty = CandidateEvaluation{zero.Lambda, zero.Seed, zero.Epoch, t, m}
	return out, nil
}

func validMetrics(m Metrics) bool {
	if m.Questions < 0 || m.Questions > MaxRows || m.Pages < m.Questions || m.Pages > m.Questions*MaxPages || m.AuxiliaryCalls < 0 || m.AuxiliaryCalls > m.Questions || m.Wins < 0 || m.Wins > m.Questions || m.Losses < 0 || m.Losses > m.Questions || m.Ties < 0 || m.Ties > m.Questions || m.Wins+m.Losses+m.Ties != m.Questions || m.Wins+m.Losses > m.AuxiliaryCalls || m.Hit1 < 0 || m.Hit10 < m.Hit1 || m.Hit100 < m.Hit10 || m.Hit100 > m.Questions || m.All10 < 0 || m.All10 > m.Hit10 || m.WorkKnownRows < 0 || m.WorkKnownRows > m.Questions || m.SelectedFallbacks < 0 || m.SelectedFallbacks > min(m.AuxiliaryCalls, m.Ties) {
		return false
	}
	w := m.CounterfactualWork
	return w.BaselineRankAttempts == m.WorkKnownRows && w.AuxiliaryBuildAttempts >= 0 && w.AuxiliaryRankAttempts >= 0 && w.AuxiliaryBuildAttempts <= min(m.AuxiliaryCalls, m.WorkKnownRows) && w.AuxiliaryRankAttempts <= min(m.AuxiliaryCalls, m.WorkKnownRows)
}

func validateGateEvaluation(e Evaluation, ready Readiness) error {
	if !digest(e.InputSHA256, 64) || !validMetrics(e.Total) {
		return fmt.Errorf("invalid gate evaluation")
	}
	want := emptyEvaluation()
	var roles [2]Metrics
	for i, group := range e.Repositories {
		if group.Role != want.Repositories[i].Role || group.Repository != want.Repositories[i].Repository || !validMetrics(group.Metrics) {
			return fmt.Errorf("invalid repository metric group")
		}
		role := 0
		if i >= 20 {
			role = 1
		}
		addMetrics(&roles[role], group.Metrics)
		if i >= 20 && group.Metrics.Questions != ready.ValidationRepositories[i-20].Eligible {
			return fmt.Errorf("gate repository coverage mismatch")
		}
	}
	var total Metrics
	for i, role := range e.Roles {
		if role.Role != want.Roles[i].Role || role.Repository != "" || role.Metrics != roles[i] {
			return fmt.Errorf("gate role aggregation mismatch")
		}
		addMetrics(&total, role.Metrics)
	}
	if total != e.Total || roles[1].Questions != ready.ValidationEligible {
		return fmt.Errorf("gate aggregate coverage mismatch")
	}
	macro, complete := macroHit10(e)
	if !complete || !e.ValidationMacroComplete || e.ValidationMacroHit10 != macro {
		return fmt.Errorf("gate macro metric mismatch")
	}
	return nil
}

// exactMacroHit10 avoids rounding a true equal-repository tie differently. The
// product of the five frozen repository size ceilings is below 2e14; five
// weighted Hit10 sums remain well inside int64. Readiness enforces those bounds.
func exactMacroHit10(e Evaluation, ready Readiness) int64 {
	denominator := int64(1)
	for _, repo := range ready.ValidationRepositories {
		denominator *= int64(repo.Eligible)
	}
	var sum int64
	for i, repo := range ready.ValidationRepositories {
		sum += int64(e.Repositories[20+i].Metrics.Hit10) * (denominator / int64(repo.Eligible))
	}
	return sum
}

type SeedGate struct {
	Seed                                                                  uint64
	BaselinePages, AlwaysHelperPages, CallBudget, PooledHit10, MacroHit10 bool
	RepositoryPages                                                       [5]bool
	Pass                                                                  bool
}

type PilotGates struct {
	Seeds    [2]SeedGate
	BothPass bool
}

// PilotGate requires both precommitted seeds, never selects between them, and
// evaluates only validation metrics. Passing advances a private pilot; it does
// not approve final evaluation, actual savings, publication or production use.
func PilotGate(ready Readiness, baseline, always Evaluation, selected [2]CandidateEvaluation) (PilotGates, error) {
	var out PilotGates
	if err := ready.Validate(); err != nil {
		return out, err
	}
	for _, e := range []Evaluation{baseline, always, selected[0].Metrics, selected[1].Metrics} {
		if err := validateGateEvaluation(e, ready); err != nil {
			return out, err
		}
		if e.InputSHA256 != baseline.InputSHA256 {
			return out, fmt.Errorf("gate policies use different measured inputs")
		}
	}
	b, a := baseline.Roles[1].Metrics, always.Roles[1].Metrics
	if b.AuxiliaryCalls != 0 || b.Wins != 0 || b.Losses != 0 || a.AuxiliaryCalls != ready.ValidationEligible {
		return out, fmt.Errorf("invalid baseline or always-helper control")
	}
	out.BothPass = true
	for i, seed := range Seeds() {
		candidate := selected[i]
		if candidate.Seed != seed || !validPenalty(candidate.Lambda) || candidate.Epoch < 1 || candidate.Epoch > 100 {
			return PilotGates{}, fmt.Errorf("gate seed or candidate metadata mismatch")
		}
		if err := validateThreshold(candidate.Threshold, false); err != nil {
			return PilotGates{}, err
		}
		m := candidate.Metrics.Roles[1].Metrics
		if candidate.Threshold.ValidationQuestions != m.Questions || candidate.Threshold.ValidationCalls != m.AuxiliaryCalls || candidate.Threshold.ValidationPages != m.Pages {
			return PilotGates{}, fmt.Errorf("gate threshold outcome mismatch")
		}
		g := SeedGate{Seed: seed, BaselinePages: m.Pages*100 <= b.Pages*95, AlwaysHelperPages: m.Pages*100 <= a.Pages*99, CallBudget: m.AuxiliaryCalls <= m.Questions*90/100, PooledHit10: m.Hit10 >= b.Hit10, MacroHit10: exactMacroHit10(candidate.Metrics, ready) >= exactMacroHit10(baseline, ready)}
		g.Pass = g.BaselinePages && g.AlwaysHelperPages && g.CallBudget && g.PooledHit10 && g.MacroHit10
		for ri := 0; ri < 5; ri++ {
			g.RepositoryPages[ri] = candidate.Metrics.Repositories[20+ri].Metrics.Pages*100 <= baseline.Repositories[20+ri].Metrics.Pages*105
			g.Pass = g.Pass && g.RepositoryPages[ri]
		}
		out.Seeds[i] = g
		out.BothPass = out.BothPass && g.Pass
	}
	return out, nil
}
