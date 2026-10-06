package statehintplacement

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintcatalog"
)

// These are explicit synthetic owner-reader/prediction doubles. No native
// content, endpoint, credential, live admission or trained-model claim is used.
const heldFixture = "Original already-held content for a development-work question"

func anchorFixture(t *testing.T) Anchor {
	t.Helper()
	digest, err := TextDigest(heldFixture)
	if err != nil {
		t.Fatal(err)
	}
	return Anchor{Scope: Scope{WorkspaceRef: "opaque-space", OwnerRef: "opaque-owner"}, WorkRef: "opaque-development-work", OwnerWorkRevision: "p12-w9", ContentRef: "opaque-message", TextSHA256: digest}
}
func TestOpaqueRevisionAndDerivedDigestStaySeparate(t *testing.T) {
	a := anchorFixture(t)
	result := ValidatePlacement(a, a)
	if result.Status != Matched || result.ActualVerified || result.MutationExecuted || result.OwnerContentRevisionAvailable {
		t.Fatal(result)
	}
	current := a
	current.OwnerWorkRevision = "p13-w9"
	if result := ValidatePlacement(a, current); result.Status != Stale || result.Reason != "owner_work_revision_changed" {
		t.Fatal("profile revision was discarded", result)
	}
	a.OwnerWorkRevision = ""
	if result := ValidatePlacement(a, a); result.Status != Unavailable {
		t.Fatal("derived hash supplied a missing owner revision", result)
	}
	if a.OwnerContentRevision != "" {
		t.Fatal("content revision was minted from digest")
	}
}
func TestScopeWorkContentAndOwnerRevisionMustMatchExactly(t *testing.T) {
	a := anchorFixture(t)
	for _, change := range []struct {
		reason string
		modify func(*Anchor)
	}{
		{"scope_changed", func(x *Anchor) { x.Scope.WorkspaceRef = "other-space" }},
		{"scope_changed", func(x *Anchor) { x.Scope.OwnerRef = "other-owner" }},
		{"work_ref_changed", func(x *Anchor) { x.WorkRef = "opaque-game-session" }},
		{"content_ref_changed", func(x *Anchor) { x.ContentRef = "other-message" }},
		{"owner_work_revision_changed", func(x *Anchor) { x.OwnerWorkRevision = "p12-w10" }},
		{"owner_content_revision_changed", func(x *Anchor) { x.OwnerContentRevision = "owner-issued-edit-2" }},
		{"text_changed", func(x *Anchor) { x.TextSHA256 = strings.Repeat("a", 64) }},
	} {
		current := a
		change.modify(&current)
		if got := ValidatePlacement(a, current); got.Status != Stale || got.Reason != change.reason || got.ActualVerified {
			t.Fatal(change.reason, got)
		}
	}
}
func TestMissingOrInvalidHeaderMetadataIsUnavailable(t *testing.T) {
	a := anchorFixture(t)
	for _, modify := range []func(*Anchor){
		func(x *Anchor) { x.Scope.OwnerRef = "" }, func(x *Anchor) { x.WorkRef = "" }, func(x *Anchor) { x.ContentRef = "" }, func(x *Anchor) { x.OwnerWorkRevision = "" },
		func(x *Anchor) { x.TextSHA256 = "" }, func(x *Anchor) { x.TextSHA256 = strings.ToUpper(x.TextSHA256) },
		func(x *Anchor) { x.OwnerWorkRevision = "\xff" }, func(x *Anchor) { x.OwnerWorkRevision = "p12-w9\x00" }, func(x *Anchor) { x.OwnerWorkRevision = strings.Repeat("x", 257) },
	} {
		current := a
		modify(&current)
		if got := ValidatePlacement(a, current); got.Status != Unavailable {
			t.Fatal(got)
		}
	}
	for _, text := range []string{"\xff", "source\x00text", strings.Repeat("x", statehint.MaxTextBytes+1)} {
		if _, err := TextDigest(text); !errors.Is(err, ErrInput) {
			t.Fatal("invalid held text accepted", err)
		}
	}
	lower, _ := TextDigest("original")
	upper, _ := TextDigest("Original")
	if lower == upper {
		t.Fatal("exact bytes were normalized")
	}
}

type revisionDouble struct {
	current     Anchor
	currentBody string
	calls       int
}

func (r *revisionDouble) ReadAnchor(_ context.Context, _ Scope, _ string, _ string) (Anchor, error) {
	r.calls++
	a := r.current
	// The synthetic owner boundary hashes its own currently held body; it
	// never receives the request body, expected hash or expected revision.
	if r.currentBody != "" {
		a.TextSHA256, _ = TextDigest(r.currentBody)
	} else {
		a.TextSHA256 = ""
	}
	return a, nil
}

type catalogDouble struct {
	calls  int
	change func(*statehintcatalog.Snapshot)
}

type catalogFunc func(context.Context, Scope, []string) (statehintcatalog.Snapshot, error)

func (f catalogFunc) Read(ctx context.Context, scope Scope, refs []string) (statehintcatalog.Snapshot, error) {
	return f(ctx, scope, refs)
}

func (c *catalogDouble) Read(_ context.Context, scope Scope, refs []string) (statehintcatalog.Snapshot, error) {
	c.calls++
	s := statehintcatalog.Snapshot{Scope: scope, Works: []statehintcatalog.Work{{Ref: refs[0], State: statehint.Todo, StatePlanningAvailable: true}}, Catalog: statehintcatalog.Catalog{
		Labels: []statehintcatalog.Label{{Ref: "opaque-question-label", Active: true}}, Bindings: []statehintcatalog.Binding{{Intent: statehint.Question, LabelRefs: []string{"opaque-question-label"}}}, Emojis: []statehint.EmojiCandidate{{Intent: statehint.Question, Code: "2753"}},
	}}
	if c.change != nil {
		c.change(&s)
	}
	return s, nil
}
func fakeScore(intent statehint.Intent, confidence float64) statehint.Prediction {
	var probabilities [statehint.IntentCount]float64
	for i := range probabilities {
		probabilities[i] = (1 - confidence) / 7
	}
	index, _ := statehint.IntentIndex(intent)
	probabilities[index] = confidence
	return statehint.Prediction{Intent: intent, Probabilities: probabilities, Confidence: confidence, Margin: confidence - (1-confidence)/7, Source: statehint.Learned, TrainingSteps: 1}
}
func requestFixture(t *testing.T) Request {
	return Request{Expected: anchorFixture(t), HeldText: heldFixture, CommandRef: "opaque-display-command"}
}
func TestMatchedProposalRechecksOwnerAndNeverClaimsVerification(t *testing.T) {
	request := requestFixture(t)
	revisions := &revisionDouble{current: request.Expected, currentBody: heldFixture}
	catalog := &catalogDouble{}
	calls := 0
	adapter := Adapter{Revisions: revisions, Catalog: catalog, Predictor: statehintcatalog.PredictorFunc(func(text string) (statehint.Prediction, error) {
		calls++
		if text != heldFixture {
			t.Fatal("model received different content")
		}
		return fakeScore(statehint.Question, .98), nil
	})}
	result, err := adapter.Propose(context.Background(), request)
	if err != nil || result.Placement.Status != Matched || !result.Eligible || len(result.LabelRefs) != 1 || result.EmojiCode != "2753" || result.ActualVerified || result.Placement.ActualVerified || result.MutationExecuted || result.StateChange || result.CatalogConsistency != "unqualified" || revisions.calls != 2 || catalog.calls != 1 || calls != 1 {
		t.Fatal(result, err, revisions.calls, catalog.calls, calls)
	}
	if request.Expected.OwnerContentRevision != "" || revisions.current.OwnerContentRevision != "" {
		t.Fatal("missing owner revision was minted")
	}
	encoded, err := json.Marshal(result)
	if err != nil || strings.Contains(string(encoded), heldFixture) || strings.Contains(string(encoded), request.Expected.OwnerWorkRevision) || strings.Contains(string(encoded), request.Expected.TextSHA256) {
		t.Fatal("source/revision/digest echoed into report", err)
	}
}
func TestChangedOwnerContentAfterScoringDiscardsAllAnnotations(t *testing.T) {
	request := requestFixture(t)
	revisions := &revisionDouble{current: request.Expected, currentBody: heldFixture}
	catalog := &catalogDouble{}
	adapter := Adapter{Revisions: revisions, Catalog: catalog, Predictor: statehintcatalog.PredictorFunc(func(string) (statehint.Prediction, error) {
		revisions.currentBody = "Original owner-held content edited during classification"
		return fakeScore(statehint.Question, .98), nil
	})}
	result, err := adapter.Propose(context.Background(), request)
	if err != nil || result.Placement.Status != Stale || result.Reason != "text_changed" || len(result.LabelRefs) != 0 || result.EmojiCode != "" || result.Eligible || !result.NoOp {
		t.Fatal(result, err)
	}
}
func TestUnknownCurrentContentCannotReflectModelHashAsEvidence(t *testing.T) {
	request := requestFixture(t)
	revisions := &revisionDouble{current: request.Expected}
	catalog := &catalogDouble{}
	calls := 0
	adapter := Adapter{Revisions: revisions, Catalog: catalog, Predictor: statehintcatalog.PredictorFunc(func(string) (statehint.Prediction, error) { calls++; return fakeScore(statehint.Question, .98), nil })}
	result, err := adapter.Propose(context.Background(), request)
	if err != nil || result.Placement.Status != Unavailable || catalog.calls != 0 || calls != 0 {
		t.Fatal("request digest replaced unknown owner content", result, err)
	}
}
func TestCatalogMustReferToSameScopeAndWorkBeforeModelCall(t *testing.T) {
	for _, change := range []func(*statehintcatalog.Snapshot){func(s *statehintcatalog.Snapshot) { s.Scope.OwnerRef = "different-owner" }, func(s *statehintcatalog.Snapshot) { s.Works[0].Ref = "opaque-game-session" }, func(s *statehintcatalog.Snapshot) { s.Works = nil }} {
		request := requestFixture(t)
		calls := 0
		catalog := &catalogDouble{change: change}
		adapter := Adapter{Revisions: &revisionDouble{current: request.Expected, currentBody: heldFixture}, Catalog: catalog, Predictor: statehintcatalog.PredictorFunc(func(string) (statehint.Prediction, error) { calls++; return fakeScore(statehint.Question, .98), nil })}
		result, err := adapter.Propose(context.Background(), request)
		if err != nil || result.Placement.Status != Unavailable || len(result.LabelRefs) != 0 || calls != 0 {
			t.Fatal(result, err, calls)
		}
	}
}
func TestDisplayScopeFilteringAndReplayUseNoSharedCache(t *testing.T) {
	request := requestFixture(t)
	revisions := &revisionDouble{current: request.Expected, currentBody: heldFixture}
	catalog := &catalogDouble{change: func(s *statehintcatalog.Snapshot) {
		s.Catalog.Labels = append(s.Catalog.Labels, statehintcatalog.Label{Ref: "opaque-blocker-label", Active: true})
		s.Catalog.Bindings = append(s.Catalog.Bindings, statehintcatalog.Binding{Intent: statehint.Blocker, LabelRefs: []string{"opaque-blocker-label"}})
		s.Catalog.Emojis = append(s.Catalog.Emojis, statehint.EmojiCandidate{Intent: statehint.Blocker, Code: "1f6a7"})
	}}
	adapter := Adapter{Revisions: revisions, Catalog: catalog, Predictor: statehintcatalog.PredictorFunc(func(string) (statehint.Prediction, error) { return fakeScore(statehint.Blocker, .98), nil })}
	result, err := adapter.Propose(context.Background(), request)
	if err != nil || result.Eligible || len(result.LabelRefs) != 0 || result.EmojiCode != "" || result.Prediction.Intent != statehint.Blocker {
		t.Fatal("outside intent forced or proposed", result, err)
	}
	adapter.Predictor = statehintcatalog.PredictorFunc(func(string) (statehint.Prediction, error) { return fakeScore(statehint.Question, .98), nil })
	first, err := adapter.Propose(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := adapter.Propose(context.Background(), request)
	if err != nil || len(first.LabelRefs) != 1 || len(second.LabelRefs) != 1 || revisions.calls != 6 {
		t.Fatal("process cache changed replay", second, err, revisions.calls)
	}
	catalog.change = func(s *statehintcatalog.Snapshot) {
		s.Works[0].ExistingOpaqueLabels = []string{"opaque-question-label"}
		s.Works[0].CurrentEmoji = "2753"
	}
	third, err := adapter.Propose(context.Background(), request)
	if err != nil || !third.NoOp || len(third.LabelRefs) != 0 || third.EmojiCode != "" {
		t.Fatal("current annotations duplicated", third, err)
	}
}
func TestLowConfidenceAndHeldTextMismatchCannotCreateMarkers(t *testing.T) {
	request := requestFixture(t)
	revisions := &revisionDouble{current: request.Expected, currentBody: heldFixture}
	catalog := &catalogDouble{}
	adapter := Adapter{Revisions: revisions, Catalog: catalog, Predictor: statehintcatalog.PredictorFunc(func(string) (statehint.Prediction, error) { return fakeScore(statehint.Question, .85), nil })}
	result, err := adapter.Propose(context.Background(), request)
	if err != nil || result.Eligible || len(result.LabelRefs) != 0 || result.Reason != "below_confidence" {
		t.Fatal(result, err)
	}
	request.HeldText = "Different already-held text"
	before := revisions.calls
	result, err = adapter.Propose(context.Background(), request)
	if err != nil || result.Placement.Status != Stale || revisions.calls != before {
		t.Fatal("held text mismatch reached owner reader", result, err)
	}
}

func TestCancellationAfterEachSuccessfulCallbackStopsThePipeline(t *testing.T) {
	for _, stage := range []string{"first_anchor", "catalog", "predictor", "final_anchor"} {
		ctx, cancel := context.WithCancel(context.Background())
		request := requestFixture(t)
		revisionCalls, catalogCalls, predictionCalls := 0, 0, 0
		adapter := Adapter{
			Revisions: RevisionReaderFunc(func(context.Context, Scope, string, string) (Anchor, error) {
				revisionCalls++
				if stage == "first_anchor" && revisionCalls == 1 || stage == "final_anchor" && revisionCalls == 2 {
					cancel()
				}
				return request.Expected, nil
			}),
			Catalog: catalogFunc(func(ctx context.Context, scope Scope, refs []string) (statehintcatalog.Snapshot, error) {
				catalogCalls++
				snapshot, err := (&catalogDouble{}).Read(ctx, scope, refs)
				if stage == "catalog" {
					cancel()
				}
				return snapshot, err
			}),
			Predictor: statehintcatalog.PredictorFunc(func(string) (statehint.Prediction, error) {
				predictionCalls++
				if stage == "predictor" {
					cancel()
				}
				return fakeScore(statehint.Question, .98), nil
			}),
		}
		result, err := adapter.Propose(ctx, request)
		cancel()
		if !errors.Is(err, context.Canceled) || result.Placement.Status != Unavailable || result.Reason != "context_cancelled" || len(result.LabelRefs) != 0 || result.Prediction != nil || result.EmojiCode != "" {
			t.Fatal(stage, result, err)
		}
		switch stage {
		case "first_anchor":
			if revisionCalls != 1 || catalogCalls != 0 || predictionCalls != 0 {
				t.Fatal(stage, revisionCalls, catalogCalls, predictionCalls)
			}
		case "catalog":
			if revisionCalls != 1 || catalogCalls != 1 || predictionCalls != 0 {
				t.Fatal(stage, revisionCalls, catalogCalls, predictionCalls)
			}
		case "predictor":
			if revisionCalls != 1 || catalogCalls != 1 || predictionCalls != 1 {
				t.Fatal(stage, revisionCalls, catalogCalls, predictionCalls)
			}
		case "final_anchor":
			if revisionCalls != 2 || catalogCalls != 1 || predictionCalls != 1 {
				t.Fatal(stage, revisionCalls, catalogCalls, predictionCalls)
			}
		}
	}
}
