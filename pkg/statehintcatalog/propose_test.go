package statehintcatalog

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/laya-tools/pkg/statehint"
)

// All readers and scores below are deliberately test doubles. They are not
// actual configured application reads, trained model inference or native tests.
type testReader struct {
	Snapshot Snapshot
	Err      error
	Calls    int
	Refs     []string
}

func (r *testReader) Read(_ context.Context, _ Scope, refs []string) (Snapshot, error) {
	r.Calls++
	r.Refs = append([]string(nil), refs...)
	return r.Snapshot, r.Err
}
func testPrediction(intent statehint.Intent, confidence float64) statehint.Prediction {
	var probabilities [statehint.IntentCount]float64
	for i := range probabilities {
		probabilities[i] = (1 - confidence) / 7
	}
	index, _ := statehint.IntentIndex(intent)
	probabilities[index] = confidence
	return statehint.Prediction{Intent: intent, Probabilities: probabilities, Confidence: confidence,
		Margin: confidence - (1-confidence)/7, Source: statehint.Learned, TrainingSteps: 1}
}
func testScope() Scope { return Scope{WorkspaceRef: "opaque-space", OwnerRef: "opaque-owner"} }
func testSnapshot() Snapshot {
	return Snapshot{Scope: testScope(), Works: []Work{{Ref: "opaque-work", State: statehint.Todo}}, Catalog: Catalog{
		Labels:   []Label{{Ref: "opaque-label", Active: true}},
		Bindings: []Binding{{Intent: statehint.Question, LabelRefs: []string{"opaque-label"}}},
		Emojis:   []statehint.EmojiCandidate{{Intent: statehint.Question, Code: "2753"}},
	}}
}
func testRequests() []Request {
	return []Request{{WorkRef: "opaque-work", CommandRef: "opaque-command", Text: "Original caller text never fetched by the reader"}}
}
func proposeTest(t *testing.T, snapshot Snapshot, p statehint.Prediction) (Result, *testReader) {
	t.Helper()
	reader := &testReader{Snapshot: snapshot}
	result, err := (Adapter{}).Propose(context.Background(), reader, PredictorFunc(func(string) (statehint.Prediction, error) { return p, nil }), testScope(), testRequests())
	if err != nil {
		t.Fatal(err)
	}
	return result, reader
}

func TestMissingVersionProducesAnnotationsWithoutInventedRevision(t *testing.T) {
	result, reader := proposeTest(t, testSnapshot(), testPrediction(statehint.Question, .98))
	if reader.Calls != 1 || !reflect.DeepEqual(reader.Refs, []string{"opaque-work"}) {
		t.Fatal("reader did not receive only requested references")
	}
	p := result.Proposals[0]
	if p.Plan != nil || p.CanonicalVersion != "" || p.StatePlanningAvailable || p.StateReason != "unavailable_no_canonical_version" ||
		!reflect.DeepEqual(p.LabelRefs, []string{"opaque-label"}) || p.EmojiCode != "2753" || p.StateChange || p.MutationExecuted || result.MutationExecuted || result.Mode != "shadow" || result.CatalogConsistency != "unqualified" {
		t.Fatal(p, result)
	}
	encoded, err := json.Marshal(result)
	if err != nil || strings.Contains(string(encoded), testRequests()[0].Text) || strings.Contains(string(encoded), "expected_version") || strings.Contains(string(encoded), `"mutation_executed":true`) {
		t.Fatal("response exposed text, invented revision, or mutation")
	}
}

func TestInvalidCanonicalVersionsRemainAnnotationOnly(t *testing.T) {
	for _, version := range []string{"0", "01", "+1", "-1", "1.0", " 1", "18446744073709551615", "not-a-version"} {
		snapshot := testSnapshot()
		snapshot.Works[0].CanonicalVersion = version
		snapshot.Works[0].StatePlanningAvailable = true
		result, _ := proposeTest(t, snapshot, testPrediction(statehint.Question, .98))
		p := result.Proposals[0]
		if p.Plan != nil || p.CanonicalVersion != "" || p.StatePlanningAvailable || p.StateReason != "unavailable_invalid_canonical_version" || len(p.LabelRefs) != 1 {
			t.Fatal(version, p)
		}
	}
}

func TestUnavailableWorkIsNotClassified(t *testing.T) {
	snapshot := testSnapshot()
	snapshot.Works = nil
	reader := &testReader{Snapshot: snapshot}
	calls := 0
	result, err := (Adapter{}).Propose(context.Background(), reader, PredictorFunc(func(string) (statehint.Prediction, error) {
		calls++
		return testPrediction(statehint.Question, .98), nil
	}), testScope(), testRequests())
	if err != nil || calls != 0 || len(result.Proposals) != 1 || result.Proposals[0].Prediction != nil || !result.Proposals[0].NoOp || result.Proposals[0].StateReason != "work_unavailable" {
		t.Fatal(result, err, calls)
	}
}

func TestExplicitBindingsSkipGroupsMissingInactiveAndDuplicateRefs(t *testing.T) {
	snapshot := testSnapshot()
	snapshot.Catalog.Labels = append(snapshot.Catalog.Labels, Label{Ref: "opaque-group", Active: true, GroupContainer: true}, Label{Ref: "opaque-inactive"}, Label{Ref: "opaque-duplicate", Active: true}, Label{Ref: "opaque-duplicate", Active: true})
	snapshot.Catalog.Bindings[0].LabelRefs = append(snapshot.Catalog.Bindings[0].LabelRefs, "opaque-group", "opaque-inactive", "opaque-missing", "opaque-duplicate")
	result, _ := proposeTest(t, snapshot, testPrediction(statehint.Question, .98))
	if !reflect.DeepEqual(result.Proposals[0].LabelRefs, []string{"opaque-label"}) || len(result.MappingIssues) != 4 {
		t.Fatal(result)
	}
	snapshot.Catalog.Bindings = append(snapshot.Catalog.Bindings, Binding{Intent: statehint.Reference, LabelRefs: []string{"opaque-label"}})
	result, _ = proposeTest(t, snapshot, testPrediction(statehint.Question, .98))
	if len(result.Proposals[0].LabelRefs) != 0 {
		t.Fatal("one opaque label assigned multiple intents")
	}
	// No name-based fallback or automatic label creation exists.
	snapshot.Catalog.Bindings = nil
	result, _ = proposeTest(t, snapshot, testPrediction(statehint.Question, .98))
	if len(result.Proposals[0].LabelRefs) != 0 {
		t.Fatal("unconfigured label inferred")
	}
}

func TestDuplicateIntentAndEmojiConfigurationAbstains(t *testing.T) {
	snapshot := testSnapshot()
	snapshot.Catalog.Bindings = append(snapshot.Catalog.Bindings, Binding{Intent: statehint.Question, LabelRefs: []string{"opaque-other"}})
	snapshot.Catalog.Emojis = append(snapshot.Catalog.Emojis, statehint.EmojiCandidate{Intent: statehint.Question, Code: "1f440"})
	result, _ := proposeTest(t, snapshot, testPrediction(statehint.Question, .98))
	if len(result.Proposals[0].LabelRefs) != 0 || result.Proposals[0].EmojiCode != "" || !result.Proposals[0].NoOp {
		t.Fatal(result)
	}
}

func TestStatusAmbiguityRequiresCurrentPreferredReference(t *testing.T) {
	snapshot := testSnapshot()
	snapshot.Works[0].CanonicalVersion = "7"
	snapshot.Works[0].StatePlanningAvailable = true
	snapshot.Catalog.Statuses = []Status{{Ref: "opaque-active-one", StateType: statehint.Active, Active: true}, {Ref: "opaque-active-two", StateType: statehint.Active, Active: true}}
	result, _ := proposeTest(t, snapshot, testPrediction(statehint.Progress, .98))
	p := result.Proposals[0]
	if p.StatePlanningAvailable || p.Plan != nil || p.CandidateStatusRef != "" || p.StateReason != "ambiguous_current_status" {
		t.Fatal(p)
	}
	snapshot.Catalog.PreferredStatusRefs = map[statehint.State]string{statehint.Active: "opaque-active-two"}
	result, _ = proposeTest(t, snapshot, testPrediction(statehint.Progress, .98))
	p = result.Proposals[0]
	if !p.StatePlanningAvailable || p.Plan == nil || p.Plan.ExpectedVersion != 7 || p.CandidateStatusRef != "opaque-active-two" || p.NextState != statehint.Todo || p.StateChange || p.MutationExecuted || p.Plan.StateChange || p.Plan.MutationExecuted || p.StateReason != "no_matching_trusted_event" {
		t.Fatal(p)
	}
	snapshot.Catalog.Statuses[1].Active = false
	result, _ = proposeTest(t, snapshot, testPrediction(statehint.Progress, .98))
	p = result.Proposals[0]
	if p.StatePlanningAvailable || p.CandidateStatusRef != "" || p.StateReason != "preferred_status_unavailable" {
		t.Fatal("inactive preferred status selected", p)
	}
}

func TestRepeatedAnnotationAndGuardedPredictionAreNoOps(t *testing.T) {
	snapshot := testSnapshot()
	snapshot.Works[0].ExistingOpaqueLabels = []string{"opaque-label"}
	snapshot.Works[0].CurrentEmoji = "2753"
	result, _ := proposeTest(t, snapshot, testPrediction(statehint.Question, .98))
	if !result.Proposals[0].NoOp || result.Proposals[0].EmojiCode != "" || len(result.Proposals[0].LabelRefs) != 0 {
		t.Fatal(result)
	}
	snapshot = testSnapshot()
	snapshot.Works[0].CanonicalVersion = "4"
	snapshot.Works[0].StatePlanningAvailable = true
	p := testPrediction(statehint.Question, .98)
	p.GuardReason = "deliberately_guarded_test_score"
	result, _ = proposeTest(t, snapshot, p)
	if !result.Proposals[0].NoOp || result.Proposals[0].Plan != nil || result.Proposals[0].StateReason != "unavailable_prediction_guard" {
		t.Fatal(result)
	}
	p = testPrediction(statehint.Question, .85)
	result, _ = proposeTest(t, snapshot, p)
	if !result.Proposals[0].NoOp || len(result.Proposals[0].LabelRefs) != 0 || result.Proposals[0].EmojiCode != "" {
		t.Fatal(result)
	}
}

func TestBoundsScopeAndOpaqueReferencesFailBeforePrediction(t *testing.T) {
	for _, modify := range []func(*Snapshot){
		func(s *Snapshot) { s.Scope.OwnerRef = "other-owner" },
		func(s *Snapshot) { s.Works[0].Ref = "unrequested-work" },
		func(s *Snapshot) { s.Works = append(s.Works, s.Works[0]) },
		func(s *Snapshot) { s.Catalog.Labels = make([]Label, MaxLabels+1) },
		func(s *Snapshot) { s.Catalog.Statuses = make([]Status, MaxStatuses+1) },
		func(s *Snapshot) { s.Works = make([]Work, MaxWorks+1) },
		func(s *Snapshot) { s.Catalog.Emojis[0].Code = "❓" },
	} {
		snapshot := testSnapshot()
		modify(&snapshot)
		reader := &testReader{Snapshot: snapshot}
		calls := 0
		_, err := (Adapter{}).Propose(context.Background(), reader, PredictorFunc(func(string) (statehint.Prediction, error) {
			calls++
			return testPrediction(statehint.Question, .98), nil
		}), testScope(), testRequests())
		if !errors.Is(err, ErrSnapshot) || calls != 0 {
			t.Fatal("invalid snapshot classified", err, calls)
		}
	}
	reader := &testReader{Snapshot: testSnapshot()}
	tooMany := make([]Request, MaxWorks+1)
	if _, err := (Adapter{}).Propose(context.Background(), reader, PredictorFunc(func(string) (statehint.Prediction, error) { return testPrediction(statehint.Question, .98), nil }), testScope(), tooMany); !errors.Is(err, ErrInput) || reader.Calls != 0 {
		t.Fatal("oversized request read", err)
	}
	if _, err := (Adapter{MinConfidence: .89}).Propose(context.Background(), reader, PredictorFunc(func(string) (statehint.Prediction, error) { return testPrediction(statehint.Question, .98), nil }), testScope(), testRequests()); !errors.Is(err, ErrInput) || reader.Calls != 0 {
		t.Fatal("confidence floor lowered", err)
	}
}

func TestReadWindowsAreReportedWithoutAtomicityOrAuthority(t *testing.T) {
	snapshot := testSnapshot()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	snapshot.ReadTimes = ReadTimes{Works: ReadWindow{Started: start, Completed: start.Add(time.Second)}, Labels: ReadWindow{Started: start.Add(2 * time.Second), Completed: start.Add(3 * time.Second)}}
	result, _ := proposeTest(t, snapshot, testPrediction(statehint.Question, .98))
	if result.ReadTimes != snapshot.ReadTimes || result.CatalogConsistency != "unqualified" || result.MutationExecuted {
		t.Fatal(result)
	}
	snapshot.ReadTimes.Statuses = ReadWindow{Started: start, Completed: start.Add(-time.Second)}
	reader := &testReader{Snapshot: snapshot}
	if _, err := (Adapter{}).Propose(context.Background(), reader, PredictorFunc(func(string) (statehint.Prediction, error) { return testPrediction(statehint.Question, .98), nil }), testScope(), testRequests()); !errors.Is(err, ErrSnapshot) {
		t.Fatal("backwards read time accepted", err)
	}
}

func TestCancellationAndReaderErrorsDoNotExposePrivateDetails(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	reader := &testReader{Snapshot: testSnapshot()}
	predict := PredictorFunc(func(string) (statehint.Prediction, error) { return testPrediction(statehint.Question, .98), nil })
	if _, err := (Adapter{}).Propose(ctx, reader, predict, testScope(), testRequests()); !errors.Is(err, context.Canceled) || reader.Calls != 0 {
		t.Fatal(err, reader.Calls)
	}
	reader.Err = errors.New("sensitive-reader-detail")
	if _, err := (Adapter{}).Propose(context.Background(), reader, predict, testScope(), testRequests()); !errors.Is(err, ErrSnapshot) || strings.Contains(err.Error(), "sensitive-reader-detail") {
		t.Fatal("reader details exposed", err)
	}
}
