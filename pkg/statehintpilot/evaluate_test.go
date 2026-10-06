package statehintpilot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintcatalog"
)

// Scores/readers in these policy tests are deliberately test doubles. No test
// result here is a native-reader observation or learned-model accuracy claim.
func fakePrediction(intent statehint.Intent, confidence float64) statehint.Prediction {
	var probabilities [statehint.IntentCount]float64
	for i := range probabilities {
		probabilities[i] = (1 - confidence) / 7
	}
	index, _ := statehint.IntentIndex(intent)
	probabilities[index] = confidence
	return statehint.Prediction{Intent: intent, Confidence: confidence, Margin: confidence - (1-confidence)/7, Probabilities: probabilities, Source: statehint.Learned, TrainingSteps: 1}
}
func oneCase() []Case {
	return []Case{{ID: "opaque-case", Family: "original-family", Locale: "ko", Text: "Original private source sentence is never report text", Expected: statehint.Question}}
}
func onePredictor() statehintcatalog.PredictorFunc {
	return func(string) (statehint.Prediction, error) { return fakePrediction(statehint.Question, .98), nil }
}

func TestEightClassConfusionThreeDisplayGateAndWrongProposals(t *testing.T) {
	cases := []Case{
		{ID: "case-1", Family: "family-1", Locale: "ko", Text: "original-one", Expected: statehint.Progress},
		{ID: "case-2", Family: "family-1", Locale: "ko", Text: "original-two", Expected: statehint.CompletionReport},
		{ID: "case-3", Family: "family-2", Locale: "en", Text: "original-three", Expected: statehint.CompletionReport},
		{ID: "case-4", Family: "family-3", Locale: "en", Text: "original-four", Expected: statehint.Planned},
		{ID: "case-5", Family: "family-4", Locale: "en", Text: "original-five", Expected: statehint.Question},
		{ID: "case-6", Family: "family-5", Locale: "ko", Text: "original-six", Expected: statehint.Question},
	}
	scores := []statehint.Prediction{fakePrediction(statehint.Progress, .98), fakePrediction(statehint.Question, .98), fakePrediction(statehint.CompletionReport, .8), fakePrediction(statehint.Planned, .98), fakePrediction(statehint.Question, .98), fakePrediction(statehint.Question, .98)}
	scores[4].GuardReason = "deliberate_test_guard"
	scores[5].Source = statehint.Untrained
	scores[5].TrainingSteps = 0
	calls := 0
	report, err := Evaluate(cases, func(text string) (statehint.Prediction, error) {
		if text != cases[calls].Text {
			t.Fatal("metadata entered predictor")
		}
		p := scores[calls]
		calls++
		return p, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	m := report.Overall
	if m.Rows != 6 || m.Classified != 6 || m.Correct != 5 || m.Eligible != 2 || m.Abstained != 4 || m.ProposalsCreated != 2 || m.WrongEligible != 1 || m.WrongProposals != 1 {
		t.Fatal(m)
	}
	if report.FamilyCount != 5 || report.RepeatedFamilies != 1 || report.FamilyCounts["family-1"] != 2 || report.IndependentProductSamples {
		t.Fatal(report.FamilyCounts)
	}
	if report.PerLanguage["ko"].Rows != 3 || report.PerLanguage["en"].Rows != 3 {
		t.Fatal(report.PerLanguage)
	}
	for index, intent := range AllowedIntents() {
		if m.Targets[index].Intent != intent {
			t.Fatal("fixed target order drifted", m.Targets)
		}
	}
	if m.Targets[0].Expected != 1 || m.Targets[0].Predicted != 1 || m.Targets[1].Expected != 2 || m.Targets[1].Predicted != 1 || m.Targets[2].Expected != 2 || m.Targets[2].Predicted != 3 {
		t.Fatal(m.Targets)
	}
	planned, _ := statehint.IntentIndex(statehint.Planned)
	if m.Confusion[planned][planned] != 1 || report.Observations[3].Eligible || report.Observations[3].ProposalCreated || report.Observations[3].Probability != scores[3].Probabilities {
		t.Fatal("outside intent forced or renormalized")
	}
	if report.Observations[4].Guard != "custom_predictor_guard" || report.Observations[4].ProposalCreated || report.Observations[5].Eligible {
		t.Fatal("guard or untrained model bypassed gate")
	}
	for _, o := range report.Observations {
		if o.MutationExecuted || o.StateChange {
			t.Fatal("unexpected mutation")
		}
	}
}

func TestManyRowsAreBatchedAt64WithDeterministicBudgets(t *testing.T) {
	cases := make([]Case, 65)
	for i := range cases {
		cases[i] = Case{ID: fmt.Sprintf("case-%03d", i), Family: "repeated-family", Locale: "en", Text: "original caller text", Expected: statehint.Question}
	}
	report, err := Evaluate(cases, onePredictor())
	if err != nil || report.Budgets.ReadCallsUsed != 2 || report.Budgets.ReadCallLimit != 2 || report.Budgets.PredictionCallsUsed != 65 || report.Budgets.BatchSize != 64 || report.FamilyCount != 1 || report.RepeatedFamilies != 1 {
		t.Fatal(report.Budgets, err)
	}
	if report.ReaderMode != "synthetic_fixture_reader" || report.FixtureCatalog != "synthetic_three_intent_catalog_v1" || report.SyntheticCanonicalVersion != "1" || report.PlacementVerified || report.MutationExecuted || report.StateChanges != 0 || report.CatalogConsistency != "unqualified" {
		t.Fatal("synthetic source hidden or qualified")
	}
}

type fakeReader struct {
	scope  statehintcatalog.Scope
	change func(*statehintcatalog.Snapshot)
	calls  int
}

func (r *fakeReader) Read(ctx context.Context, scope statehintcatalog.Scope, refs []string) (statehintcatalog.Snapshot, error) {
	r.calls++
	snapshot, err := (syntheticFixtureReader{}).Read(ctx, scope, refs)
	if r.change != nil {
		r.change(&snapshot)
	}
	return snapshot, err
}
func callerOptions(reader statehintcatalog.SnapshotReader) Options {
	return Options{Reader: reader, Scope: statehintcatalog.Scope{WorkspaceRef: "opaque-space", OwnerRef: "opaque-owner"}}
}

func TestEligibilityDistinguishesMissingCatalogFromExistingNoOp(t *testing.T) {
	reader := &fakeReader{change: func(s *statehintcatalog.Snapshot) {
		s.Catalog.Labels = nil
		s.Catalog.Bindings = nil
		s.Catalog.Emojis = nil
		s.Works[0].CanonicalVersion = ""
	}}
	report, err := EvaluateWithOptions(context.Background(), oneCase(), onePredictor(), callerOptions(reader))
	if err != nil {
		t.Fatal(err)
	}
	m := report.Overall
	o := report.Observations[0]
	if m.Eligible != 1 || m.ProposalsCreated != 0 || m.EligibleWithoutProposal != 1 || m.CurrentLabelMissing != 1 || m.CurrentEmojiMissing != 1 || m.ExistingAnnotationNoOp != 0 || o.StateReason != "disabled_in_display_pilot" || !o.RawReaderStatePlanningChecked {
		t.Fatal(m, o)
	}
	reader.change = func(s *statehintcatalog.Snapshot) {
		s.Works[0].ExistingOpaqueLabels = []string{"synthetic-pilot-label-question"}
		s.Works[0].CurrentEmoji = "2753"
	}
	report, err = EvaluateWithOptions(context.Background(), oneCase(), onePredictor(), callerOptions(reader))
	if err != nil || report.Overall.ExistingAnnotationNoOp != 1 || report.Overall.ProposalsCreated != 0 || report.Overall.CurrentLabelMissing != 0 || report.Overall.CurrentEmojiMissing != 0 {
		t.Fatal(report.Overall, err)
	}
	reader.change = func(s *statehintcatalog.Snapshot) { s.Catalog.Labels = nil; s.Catalog.Bindings = nil }
	report, err = EvaluateWithOptions(context.Background(), oneCase(), onePredictor(), callerOptions(reader))
	if err != nil || report.Overall.Eligible != 1 || report.Overall.ProposalsCreated != 1 || report.Overall.CurrentLabelMissing != 1 || !report.Observations[0].EmojiProposed || report.Observations[0].LabelProposalCount != 0 {
		t.Fatal(report.Overall, err)
	}
}

func TestCallerOutsideCatalogDoesNotExpandDisplayScope(t *testing.T) {
	reader := &fakeReader{change: func(s *statehintcatalog.Snapshot) {
		s.Catalog.Labels = append(s.Catalog.Labels, statehintcatalog.Label{Ref: "opaque-reference-label", Active: true})
		s.Catalog.Bindings = append(s.Catalog.Bindings, statehintcatalog.Binding{Intent: statehint.Reference, LabelRefs: []string{"opaque-reference-label"}})
		s.Catalog.Emojis = append(s.Catalog.Emojis, statehint.EmojiCandidate{Intent: statehint.Reference, Code: "1f4ce"})
	}}
	cases := oneCase()
	cases[0].Expected = statehint.Reference
	score := fakePrediction(statehint.Reference, .98)
	report, err := EvaluateWithOptions(context.Background(), cases, statehintcatalog.PredictorFunc(func(string) (statehint.Prediction, error) { return score, nil }), callerOptions(reader))
	if err != nil || report.Overall.Correct != 1 || report.Overall.Eligible != 0 || report.Overall.ProposalsCreated != 0 || report.Observations[0].LabelProposalCount != 0 || report.Observations[0].EmojiProposed || report.Observations[0].Probability != score.Probabilities {
		t.Fatal(report.Overall, err)
	}
}

func TestOutsideBlockerIsPreventedBeforeBridgeCreatesAnnotations(t *testing.T) {
	reader := &fakeReader{change: func(s *statehintcatalog.Snapshot) {
		s.Catalog.Labels = append(s.Catalog.Labels, statehintcatalog.Label{Ref: "opaque-blocker", Active: true})
		s.Catalog.Bindings = append(s.Catalog.Bindings, statehintcatalog.Binding{Intent: statehint.Blocker, LabelRefs: []string{"opaque-blocker"}})
		s.Catalog.Emojis = append(s.Catalog.Emojis, statehint.EmojiCandidate{Intent: statehint.Blocker, Code: "1f6a7"})
	}}
	tracked := &trackedReader{reader: reader, limit: 1}
	result, err := (statehintcatalog.Adapter{}).Propose(context.Background(), tracked, statehintcatalog.PredictorFunc(func(string) (statehint.Prediction, error) { return fakePrediction(statehint.Blocker, .98), nil }), callerOptions(reader).Scope, []statehintcatalog.Request{{WorkRef: "opaque-case", CommandRef: "opaque-case", Text: "original scope test"}})
	if err != nil {
		t.Fatal(err)
	}
	p := result.Proposals[0]
	if p.Prediction.Intent != statehint.Blocker || len(p.LabelRefs) != 0 || p.EmojiCode != "" || !p.NoOp {
		t.Fatal("outside proposal was created and merely hidden", p)
	}
	for _, binding := range tracked.filtered.Catalog.Bindings {
		if !allowed(binding.Intent) {
			t.Fatal("outside binding reached bridge")
		}
	}
	for _, emoji := range tracked.filtered.Catalog.Emojis {
		if !allowed(emoji.Intent) {
			t.Fatal("outside emoji reached bridge")
		}
	}
}

func TestOutsideMetadataErrorsAndCrossScopeConflictsSurviveFiltering(t *testing.T) {
	reader := &fakeReader{change: func(s *statehintcatalog.Snapshot) {
		s.Catalog.Emojis = append(s.Catalog.Emojis, statehint.EmojiCandidate{Intent: statehint.Blocker, Code: "not-a-canonical-code"})
	}}
	calls := 0
	predict := statehintcatalog.PredictorFunc(func(string) (statehint.Prediction, error) {
		calls++
		return fakePrediction(statehint.Question, .98), nil
	})
	if _, err := EvaluateWithOptions(context.Background(), oneCase(), predict, callerOptions(reader)); !errors.Is(err, ErrEvaluation) || calls != 0 {
		t.Fatal("invalid outside metadata hidden", err, calls)
	}
	reader.change = func(s *statehintcatalog.Snapshot) {
		s.Catalog.Bindings = append(s.Catalog.Bindings, statehintcatalog.Binding{Intent: statehint.Blocker, LabelRefs: []string{"synthetic-pilot-label-question"}})
	}
	report, err := EvaluateWithOptions(context.Background(), oneCase(), predict, callerOptions(reader))
	if err != nil || report.Overall.CurrentLabelMissing != 1 || report.Observations[0].LabelProposalCount != 0 || !report.Observations[0].EmojiProposed {
		t.Fatal("conflicting original binding became valid after filter", report.Overall, err)
	}
}

func TestFullValidationAndBudgetFailurePrecedeAnyCallbacks(t *testing.T) {
	for _, modify := range []func([]Case){
		func(c []Case) { c[1].ID = c[0].ID }, func(c []Case) { c[1].Family = "" }, func(c []Case) { c[1].Expected = "unknown" },
		func(c []Case) { c[1].Text = "\xff" }, func(c []Case) { c[1].Text = "text\x00text" }, func(c []Case) { c[1].Text = strings.Repeat("x", statehint.MaxTextBytes+1) },
	} {
		cases := append(oneCase(), Case{ID: "second-case", Family: "second-family", Locale: "en", Text: "second original text", Expected: statehint.Question})
		modify(cases)
		reader := &fakeReader{}
		calls := 0
		_, err := EvaluateWithOptions(context.Background(), cases, statehintcatalog.PredictorFunc(func(string) (statehint.Prediction, error) {
			calls++
			return fakePrediction(statehint.Question, .98), nil
		}), callerOptions(reader))
		if !errors.Is(err, ErrInput) || reader.calls != 0 || calls != 0 {
			t.Fatal("invalid later row partially evaluated", err, reader.calls, calls)
		}
	}
	reader := &fakeReader{}
	calls := 0
	predict := statehintcatalog.PredictorFunc(func(string) (statehint.Prediction, error) {
		calls++
		return fakePrediction(statehint.Question, .98), nil
	})
	options := callerOptions(reader)
	options.MaxTotalTextBytes = 1
	if _, err := EvaluateWithOptions(context.Background(), oneCase(), predict, options); !errors.Is(err, ErrBudget) || reader.calls != 0 || calls != 0 {
		t.Fatal("text budget evaluated", err)
	}
	cases := make([]Case, 65)
	for i := range cases {
		cases[i] = Case{ID: fmt.Sprintf("case-%d", i), Family: "group", Locale: "en", Text: "original", Expected: statehint.Question}
	}
	options = callerOptions(reader)
	options.MaxReadCalls = 1
	if _, err := EvaluateWithOptions(context.Background(), cases, predict, options); !errors.Is(err, ErrBudget) || reader.calls != 0 || calls != 0 {
		t.Fatal("read budget evaluated", err)
	}
}

func TestNoScopeMissingWorkAndCancellationDoNotPredict(t *testing.T) {
	reader := &fakeReader{}
	calls := 0
	predict := statehintcatalog.PredictorFunc(func(string) (statehint.Prediction, error) {
		calls++
		return fakePrediction(statehint.Question, .98), nil
	})
	if _, err := EvaluateWithOptions(context.Background(), oneCase(), predict, Options{Reader: reader}); !errors.Is(err, ErrInput) || reader.calls != 0 || calls != 0 {
		t.Fatal("caller reader received invented scope", err)
	}
	reader.change = func(s *statehintcatalog.Snapshot) { s.Scope.OwnerRef = "foreign-owner" }
	if _, err := EvaluateWithOptions(context.Background(), oneCase(), predict, callerOptions(reader)); !errors.Is(err, ErrEvaluation) || calls != 0 {
		t.Fatal("foreign scope classified", err)
	}
	reader.change = func(s *statehintcatalog.Snapshot) { s.Works = nil }
	report, err := EvaluateWithOptions(context.Background(), oneCase(), predict, callerOptions(reader))
	if err != nil || calls != 0 || report.Overall.UnavailableWork != 1 || report.Overall.Classified != 0 || report.Overall.Abstained != 1 {
		t.Fatal(report.Overall, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	before := reader.calls
	if _, err := EvaluateWithOptions(ctx, oneCase(), predict, callerOptions(reader)); !errors.Is(err, context.Canceled) || reader.calls != before || calls != 0 {
		t.Fatal("cancelled context evaluated", err)
	}
}

func TestReportDoesNotRetainTextOrCatalogReferences(t *testing.T) {
	report, err := Evaluate(oneCase(), onePredictor())
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(report)
	if err != nil || strings.Contains(string(encoded), oneCase()[0].Text) || strings.Contains(string(encoded), `"text"`) || strings.Contains(string(encoded), "synthetic-pilot-label-question") {
		t.Fatal("source text/catalog reference retained", err)
	}
	var parsed Case
	if err := json.Unmarshal([]byte(`{"id":"case","family":"family","locale":"en","text":"original","expected_intent":"question"}`), &parsed); err != nil || parsed.Expected != statehint.Question {
		t.Fatal("case JSON contract drift", err)
	}
	if !reflect.DeepEqual(report.AllowedIntents, AllowedIntents()) || len(report.Overall.Targets) != 3 {
		t.Fatal("target array drift")
	}
	score := fakePrediction(statehint.Question, .98)
	score.GuardReason = oneCase()[0].Text
	guarded, err := Evaluate(oneCase(), func(string) (statehint.Prediction, error) { return score, nil })
	if err != nil {
		t.Fatal(err)
	}
	encoded, err = json.Marshal(guarded)
	if err != nil || strings.Contains(string(encoded), oneCase()[0].Text) || guarded.Observations[0].Guard != "custom_predictor_guard" || guarded.Observations[0].Eligible {
		t.Fatal("caller diagnostic echoed source text", err)
	}
}

func TestPilotStateDisablementDoesNotBlameTheOriginalReader(t *testing.T) {
	reader := &fakeReader{change: func(s *statehintcatalog.Snapshot) {
		s.Works[0].StatePlanningAvailable = true
		s.Catalog.Statuses = []statehintcatalog.Status{{Ref: "opaque-active-one", StateType: statehint.Active, Active: true}, {Ref: "opaque-active-two", StateType: statehint.Active, Active: true}}
	}}
	report, err := EvaluateWithOptions(context.Background(), oneCase(), onePredictor(), callerOptions(reader))
	if err != nil {
		t.Fatal(err)
	}
	o := report.Observations[0]
	if !o.RawReaderStatePlanningChecked || !o.RawReaderStatePlanningAvailable || !o.PilotStatePlanningDisabled || o.StateCandidatesChecked || o.StateReason != "disabled_in_display_pilot" || o.StateChange || o.MutationExecuted {
		t.Fatal("pilot policy reported as reader/candidate failure", o)
	}
	reader.change = func(s *statehintcatalog.Snapshot) { s.Works[0].StatePlanningAvailable = false }
	report, err = EvaluateWithOptions(context.Background(), oneCase(), onePredictor(), callerOptions(reader))
	if err != nil || !report.Observations[0].RawReaderStatePlanningChecked || report.Observations[0].RawReaderStatePlanningAvailable || report.Observations[0].StateReason != "disabled_in_display_pilot" {
		t.Fatal(report.Observations, err)
	}
}

func TestLowConfidencePreservesMissingAndInactiveCatalogDiagnostics(t *testing.T) {
	for _, reason := range []string{"missing_label_ref", "inactive_label_ref"} {
		reader := &fakeReader{change: func(s *statehintcatalog.Snapshot) {
			s.Catalog.Bindings = []statehintcatalog.Binding{{Intent: statehint.Question, LabelRefs: []string{"synthetic-sensitive-label-ref"}}}
			if reason == "inactive_label_ref" {
				s.Catalog.Labels = []statehintcatalog.Label{{Ref: "synthetic-sensitive-label-ref", Active: false}}
			} else {
				s.Catalog.Labels = nil
			}
		}}
		predict := statehintcatalog.PredictorFunc(func(string) (statehint.Prediction, error) { return fakePrediction(statehint.Question, .8), nil })
		report, err := EvaluateWithOptions(context.Background(), oneCase(), predict, callerOptions(reader))
		if err != nil {
			t.Fatal(err)
		}
		o := report.Observations[0]
		if o.Eligible || o.ProposalCreated || !o.AnnotationChecked || !o.CurrentLabelMissing || o.CurrentEmojiMissing || o.AnnotationReason != "current_label_unavailable" || o.Reason != "below_confidence" || !o.MappingDiagnosticsChecked || report.Overall.AnnotationChecked != 1 || report.Overall.CurrentLabelMissing != 1 {
			t.Fatal("ineligible row erased independent mapping checks", o, report.Overall)
		}
		found := false
		for _, issue := range o.MappingIssues {
			if issue.Reason == reason && issue.Intent == statehint.Question {
				found = true
			}
		}
		if !found {
			t.Fatal("mapping issue not retained", o.MappingIssues)
		}
		encoded, err := json.Marshal(report)
		if err != nil || strings.Contains(string(encoded), "synthetic-sensitive-label-ref") || strings.Contains(string(encoded), `"ref"`) {
			t.Fatal("mapping diagnostic exposed reference", err)
		}
	}
}

func TestOutsideIntentHasUncheckedAnnotationsAndKeepsCatalogIssueCodes(t *testing.T) {
	reader := &fakeReader{change: func(s *statehintcatalog.Snapshot) {
		s.Catalog.Bindings = []statehintcatalog.Binding{{Intent: statehint.Question, LabelRefs: []string{"missing-current-question-label"}}}
		s.Catalog.Labels = nil
	}}
	cases := oneCase()
	cases[0].Expected = statehint.Blocker
	report, err := EvaluateWithOptions(context.Background(), cases, statehintcatalog.PredictorFunc(func(string) (statehint.Prediction, error) { return fakePrediction(statehint.Blocker, .98), nil }), callerOptions(reader))
	if err != nil {
		t.Fatal(err)
	}
	o := report.Observations[0]
	if o.AnnotationChecked || o.AnnotationReason != "annotation_not_applicable" || o.CurrentLabelMissing || o.CurrentEmojiMissing || !o.MappingDiagnosticsChecked || len(o.MappingIssues) != 1 || o.MappingIssues[0].Reason != "missing_label_ref" || report.Overall.AnnotationChecked != 0 {
		t.Fatal("unchecked scope read as known annotation availability", o)
	}
}
