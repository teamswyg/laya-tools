package statehintpilot

import (
	"context"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintcatalog"
)

// Evaluate defaults to a conspicuously synthetic reader/catalog. Its fixture
// revision "1" is fabricated test data, never a substitute for a live revision.
func Evaluate(cases []Case, predictor statehintcatalog.PredictorFunc) (Report, error) {
	if predictor == nil {
		return Report{}, ErrInput
	}
	return EvaluateWithOptions(context.Background(), cases, predictor, Options{})
}

// EvaluateWithOptions validates all rows and deterministic budgets before any
// reader/predictor call, then processes chunks of at most64 references. Reader
// invocation counts do not bound a caller reader's internal network operations;
// callers must additionally supply time/cancellation budgets through context.
// Expected truth never enters the reader or predictor. Eight-class scores are
// preserved and a final display gate permits only the three pilot intents.
func EvaluateWithOptions(ctx context.Context, cases []Case, predictor statehintcatalog.Predictor, options Options) (Report, error) {
	if ctx == nil || predictor == nil {
		return Report{}, ErrInput
	}
	if f, ok := predictor.(statehintcatalog.PredictorFunc); ok && f == nil {
		return Report{}, ErrInput
	}
	if err := ctx.Err(); err != nil {
		return Report{}, err
	}
	options, bytesUsed, err := validate(cases, options)
	if err != nil {
		return Report{}, err
	}
	readerMode, fixtureCatalog, fixtureVersion := "caller_supplied_reader", "", ""
	if options.Reader == nil {
		options.Reader = syntheticFixtureReader{}
		options.Scope = statehintcatalog.Scope{WorkspaceRef: "synthetic-pilot-workspace", OwnerRef: "synthetic-pilot-owner"}
		readerMode, fixtureCatalog, fixtureVersion = "synthetic_fixture_reader", "synthetic_three_intent_catalog_v1", "1"
	}
	report := Report{Schema: "statehint-three-display-shadow-evaluation-v1", Mode: "shadow", ReaderMode: readerMode,
		FixtureCatalog: fixtureCatalog, SyntheticCanonicalVersion: fixtureVersion, CatalogConsistency: "unqualified",
		MappingDiagnosticScope: "three_intent_catalog_and_preserved_original_ref_conflicts",
		MinConfidence:          options.MinConfidence, MinMargin: options.MinMargin, AllowedIntents: AllowedIntents(), IntentOrder: statehint.Intents(),
		FamilyCounts: map[string]int{}, PerLanguage: map[string]Metrics{}, Overall: newMetrics(), Observations: make([]Observation, 0, len(cases)),
		Budgets: Budgets{CaseLimit: options.MaxCases, TextByteLimit: options.MaxTotalTextBytes, ReadCallLimit: options.MaxReadCalls,
			CasesUsed: len(cases), TextBytesUsed: bytesUsed, BatchSize: BatchSize}}
	tracked := &trackedReader{reader: options.Reader, limit: options.MaxReadCalls}
	predict := statehintcatalog.PredictorFunc(func(text string) (statehint.Prediction, error) {
		report.Budgets.PredictionCallsUsed++
		return predictor.Predict(text)
	})
	adapter := statehintcatalog.Adapter{MinConfidence: options.MinConfidence, MinMargin: options.MinMargin}
	for start := 0; start < len(cases); start += BatchSize {
		end := min(start+BatchSize, len(cases))
		requests := make([]statehintcatalog.Request, end-start)
		for i, c := range cases[start:end] {
			requests[i] = statehintcatalog.Request{WorkRef: c.ID, CommandRef: c.ID, Text: c.Text}
		}
		result, err := adapter.Propose(ctx, tracked, predict, options.Scope, requests)
		if err != nil {
			if err := ctx.Err(); err != nil {
				return Report{}, err
			}
			return Report{}, ErrEvaluation
		}
		if result.MutationExecuted || len(result.Proposals) != len(requests) || result.Mode != "shadow" || result.CatalogConsistency != "unqualified" {
			return Report{}, ErrEvaluation
		}
		mappingIssues := append(append([]statehintcatalog.MappingIssue(nil), result.MappingIssues...), tracked.filterIssues...)
		for i, proposal := range result.Proposals {
			c := cases[start+i]
			if proposal.WorkRef != c.ID || proposal.MutationExecuted || proposal.StateChange ||
				proposal.Plan != nil && (proposal.Plan.MutationExecuted || proposal.Plan.StateChange) {
				return Report{}, ErrEvaluation
			}
			observation := observe(c, proposal, tracked.snapshot, mappingIssues, options)
			report.Observations = append(report.Observations, observation)
			report.Overall.add(observation)
			language, exists := report.PerLanguage[c.Locale]
			if !exists {
				language = newMetrics()
			}
			language.add(observation)
			report.PerLanguage[c.Locale] = language
			report.FamilyCounts[c.Family]++
		}
	}
	report.Budgets.ReadCallsUsed = tracked.calls
	report.FamilyCount = len(report.FamilyCounts)
	for _, count := range report.FamilyCounts {
		if count > 1 {
			report.RepeatedFamilies++
		}
	}
	return report, nil
}

func opaque(value string) bool {
	if value == "" || len(value) > 256 {
		return false
	}
	for _, c := range value {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == '.' || c == ':') {
			return false
		}
	}
	return true
}

func validate(cases []Case, options Options) (Options, int, error) {
	if options.MaxCases == 0 {
		options.MaxCases = MaxCases
	}
	if options.MaxTotalTextBytes == 0 {
		options.MaxTotalTextBytes = MaxTotalTextBytes
	}
	if options.MinConfidence == 0 {
		options.MinConfidence = .9
	}
	if options.MinMargin == 0 {
		options.MinMargin = .05
	}
	if options.MaxCases < 1 || options.MaxCases > MaxCases || options.MaxTotalTextBytes < 1 || options.MaxTotalTextBytes > MaxTotalTextBytes ||
		math.IsNaN(options.MinConfidence) || math.IsInf(options.MinConfidence, 0) || options.MinConfidence < .9 || options.MinConfidence > 1 ||
		math.IsNaN(options.MinMargin) || math.IsInf(options.MinMargin, 0) || options.MinMargin < 0 || options.MinMargin > 1 {
		return Options{}, 0, ErrInput
	}
	if len(cases) == 0 {
		return Options{}, 0, ErrInput
	}
	if len(cases) > options.MaxCases {
		return Options{}, 0, ErrBudget
	}
	neededReads := (len(cases) + BatchSize - 1) / BatchSize
	if options.MaxReadCalls == 0 {
		options.MaxReadCalls = neededReads
	}
	if options.MaxReadCalls < 1 || options.MaxReadCalls > (MaxCases+BatchSize-1)/BatchSize {
		return Options{}, 0, ErrInput
	}
	if neededReads > options.MaxReadCalls {
		return Options{}, 0, ErrBudget
	}
	if options.Reader != nil {
		if !opaque(options.Scope.WorkspaceRef) || !opaque(options.Scope.OwnerRef) {
			return Options{}, 0, ErrInput
		}
	} else if options.Scope != (statehintcatalog.Scope{}) {
		return Options{}, 0, ErrInput
	}
	ids, locales := map[string]bool{}, map[string]bool{}
	bytesUsed := 0
	for _, c := range cases {
		if !opaque(c.ID) || ids[c.ID] || !opaque(c.Family) || !opaque(c.Locale) || len(c.Locale) > 32 ||
			len(c.Text) > statehint.MaxTextBytes || !utf8.ValidString(c.Text) || strings.ContainsRune(c.Text, 0) {
			return Options{}, 0, ErrInput
		}
		if _, ok := statehint.IntentIndex(c.Expected); !ok {
			return Options{}, 0, ErrInput
		}
		ids[c.ID], locales[c.Locale] = true, true
		if len(locales) > 32 {
			return Options{}, 0, ErrBudget
		}
		bytesUsed += len(c.Text)
		if bytesUsed > options.MaxTotalTextBytes {
			return Options{}, 0, ErrBudget
		}
	}
	return options, bytesUsed, nil
}

type trackedReader struct {
	reader       statehintcatalog.SnapshotReader
	limit, calls int
	snapshot     statehintcatalog.Snapshot
	filtered     statehintcatalog.Snapshot
	filterIssues []statehintcatalog.MappingIssue
}

func (r *trackedReader) Read(ctx context.Context, scope statehintcatalog.Scope, refs []string) (statehintcatalog.Snapshot, error) {
	if r.calls >= r.limit {
		return statehintcatalog.Snapshot{}, ErrBudget
	}
	r.calls++
	snapshot, err := r.reader.Read(ctx, scope, refs)
	if err != nil {
		return statehintcatalog.Snapshot{}, err
	}
	// Validate the complete original metadata before limiting the display
	// configuration. Invalid outside-scope entries must not disappear silently.
	if err := statehintcatalog.ValidateSnapshot(snapshot, scope, refs); err != nil {
		return statehintcatalog.Snapshot{}, err
	}
	r.snapshot = snapshot
	r.filtered, r.filterIssues = restrictCatalog(snapshot)
	return r.filtered, nil
}

func restrictCatalog(snapshot statehintcatalog.Snapshot) (statehintcatalog.Snapshot, []statehintcatalog.MappingIssue) {
	filtered := snapshot
	filtered.Works = append([]statehintcatalog.Work(nil), snapshot.Works...)
	for i := range filtered.Works {
		filtered.Works[i].StatePlanningAvailable = false
	}
	filtered.Catalog.Bindings = make([]statehintcatalog.Binding, 0, 3)
	filtered.Catalog.Emojis = make([]statehint.EmojiCandidate, 0, 3)
	refCounts := map[string]int{}
	for _, binding := range snapshot.Catalog.Bindings {
		for _, ref := range binding.LabelRefs {
			refCounts[ref]++
		}
	}
	var issues []statehintcatalog.MappingIssue
	for _, binding := range snapshot.Catalog.Bindings {
		if !allowed(binding.Intent) {
			continue
		}
		refs := make([]string, 0, len(binding.LabelRefs))
		for _, ref := range binding.LabelRefs {
			if refCounts[ref] != 1 {
				issues = append(issues, statehintcatalog.MappingIssue{Reason: "duplicate_label_ref", Ref: ref, Intent: binding.Intent})
				continue
			}
			refs = append(refs, ref)
		}
		filtered.Catalog.Bindings = append(filtered.Catalog.Bindings, statehintcatalog.Binding{Intent: binding.Intent, LabelRefs: refs})
	}
	for _, emoji := range snapshot.Catalog.Emojis {
		if allowed(emoji.Intent) {
			filtered.Catalog.Emojis = append(filtered.Catalog.Emojis, emoji)
		}
	}
	return filtered, issues
}

type syntheticFixtureReader struct{}

func (syntheticFixtureReader) Read(_ context.Context, scope statehintcatalog.Scope, refs []string) (statehintcatalog.Snapshot, error) {
	snapshot := statehintcatalog.Snapshot{Scope: scope, Works: make([]statehintcatalog.Work, len(refs)), Catalog: statehintcatalog.Catalog{}}
	for i, ref := range refs {
		snapshot.Works[i] = statehintcatalog.Work{Ref: ref, CanonicalVersion: "1", State: statehint.Todo, StatePlanningAvailable: false}
	}
	for _, intent := range AllowedIntents() {
		ref := "synthetic-pilot-label-" + string(intent)
		snapshot.Catalog.Labels = append(snapshot.Catalog.Labels, statehintcatalog.Label{Ref: ref, Active: true})
		snapshot.Catalog.Bindings = append(snapshot.Catalog.Bindings, statehintcatalog.Binding{Intent: intent, LabelRefs: []string{ref}})
	}
	snapshot.Catalog.Emojis = []statehint.EmojiCandidate{{Intent: statehint.Progress, Code: "1f6e0-fe0f"}, {Intent: statehint.CompletionReport, Code: "2705"}, {Intent: statehint.Question, Code: "2753"}}
	return snapshot, nil
}
