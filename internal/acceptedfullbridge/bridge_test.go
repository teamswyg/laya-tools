package acceptedfullbridge

import (
	"bytes"
	"encoding/json"
	"github.com/teamswyg/laya-tools/internal/inventoryprojection"
	"github.com/teamswyg/laya-tools/internal/semanticframe"
	"github.com/teamswyg/laya-tools/internal/sourcecohort"
	"strings"
	"testing"
)

type Span = sourcecohort.Span

// Original public software fixtures. No admitted Source or private acceptance.
func raw(name string, b []byte) PinnedBytes {
	return PinnedBytes{File: File{Path: name, SHA256: sourcecohort.Hash(b), Bytes: int64(len(b))}, Bytes: b}
}
func encode(t testing.TB, name string, v any) PinnedBytes {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return raw(name, b)
}
func nativeBytes(t testing.TB, f inventoryprojection.Full) []byte {
	t.Helper()
	b, e := json.Marshal(f)
	if e != nil {
		t.Fatal(e)
	}
	fields, e := json.Marshal(f.OfficialFields)
	if e != nil {
		t.Fatal(e)
	}
	var obj bytes.Buffer
	obj.WriteByte('{')
	for i, x := range f.OfficialFields {
		if i > 0 {
			obj.WriteByte(',')
		}
		key, _ := json.Marshal(x.Name)
		value, _ := json.Marshal(x.Support)
		obj.Write(key)
		obj.WriteByte(':')
		obj.Write(value)
	}
	obj.WriteByte('}')
	return bytes.Replace(b, append([]byte(`"official_field_support":`), fields...), append([]byte(`"official_field_support":`), obj.Bytes()...), 1)
}
func projectionFixture(t testing.TB) (inventoryprojection.Inputs, inventoryprojection.Full) {
	t.Helper()
	source := raw("source.txt", []byte("항목 A; check."))
	boundary := Span{Kind: "source_scope", Start: 0, End: len(source.Bytes)}
	es := []Span{boundary}
	ref := raw("receipt.json", []byte("opaque reference"))
	support := inventoryprojection.Support{Status: "open-status", Value: nil, Reason: nil, Inspected: false, Description: "full-only support explanation", Propositions: []string{"pA"}, Referents: []string{}, Oppositions: []string{}, Temporal: []string{}, Evidence: es}
	f := inventoryprojection.Full{Schema: "public-software-fixture-full-v1", ID: "fixture-source", Source: source.File, Creation: ref.File, ReadBinding: ref.File, AdditionalReads: []inventoryprojection.AdditionalRead{}, Coder: inventoryprojection.Role{ID: "fixture-coder", Agent: "fixture-agent"}, Phase: "open-phase", Scope: "open-scope", Boundary: boundary, Complete: false, Components: []semanticframe.Component{{ID: "c1", Description: "component description", Evidence: es}}, Propositions: []semanticframe.Proposition{{ID: "pZ", Description: "first description", Polarity: "open polarity", Scope: "operation scope", Subjects: []string{"r1"}, Evidence: es}, {ID: "pA", Description: "second description", Polarity: "another open polarity", Scope: "scope", Subjects: []string{}, Evidence: es}}, NegativeIDs: []string{"pA"}, NegationDescription: "retained combined scope", Oppositions: []semanticframe.Opposition{}, Temporal: []semanticframe.Temporal{}, Referents: []semanticframe.Referent{{ID: "r1", Kind: "open referent kind", Parent: nil, Description: "referent description", Evidence: es}}, Constraints: []semanticframe.Constraint{}, OfficialFields: []inventoryprojection.OfficialField{{Name: "z_field", Support: support}, {Name: "a_field", Support: support}}, Dependencies: []string{}, Unresolved: []string{"retained unresolved question"}}
	in := inventoryprojection.Inputs{Full: raw("full.json", nativeBytes(t, f)), Native: encode(t, "native.json", inventoryprojection.NativeContract{Schema: f.Schema, Fields: []string{"a_field", "z_field"}}), Mapping: encode(t, "mapping.json", inventoryprojection.SupportedMapping()), Adapter: raw("adapter.go", []byte("public declared adapter identity"))}
	c := inventoryprojection.Config{Full: in.Full.File, Adapter: in.Adapter.File, Version: inventoryprojection.Version, Mapping: in.Mapping.File, Native: in.Native.File, GraphPath: "graph.json"}
	in.Config = encode(t, "config.json", c)
	return in, f
}
func repin(t testing.TB, in *inventoryprojection.Inputs, f inventoryprojection.Full) {
	t.Helper()
	in.Full = raw(in.Full.File.Path, nativeBytes(t, f))
	var c inventoryprojection.Config
	json.Unmarshal(in.Config.Bytes, &c)
	c.Full = in.Full.File
	in.Config = encode(t, in.Config.File.Path, c)
}

func fixture(t testing.TB) Inputs {
	t.Helper()
	projection, f := projectionFixture(t)
	source := raw("source.txt", []byte("항목 A; check."))
	projection.Source = &source
	r, err := inventoryprojection.Project(projection)
	if err != nil {
		t.Fatal(err)
	}
	obs := raw("observation.json", []byte(`{"public_fixture":"opaque observation"}`))
	method := raw("method.json", []byte(`{"public_fixture":"opaque method"}`))
	start := encode(t, "reads/source.start", readStart{Version: "riido-reviewpacket/v1", State: "started", UTC: "2026-01-01T00:00:00Z", Expected: source.File.SHA256, Max: 16384, Policy: "regular-file; no final symlink; nonblocking open; identity checked; bounded read; no retries", Limit: 1, Actor: "fixture-reviewer"})
	result := encode(t, start.File.Path+".result", readResult{Version: "riido-reviewpacket/v1", State: "verified", StartSHA: start.File.SHA256, Expected: source.File.SHA256, Actual: source.File.SHA256, Bytes: source.File.Bytes, Attempts: 1, ReadUTC: "2026-01-01T00:00:01Z", UTC: "2026-01-01T00:00:02Z"})
	report := encode(t, "check-report.json", sourcecohort.CheckReport{StructuralValid: true, MandatoryComplete: true, SourcePhaseScope: true, Scope: "public software fixture check"})
	check := encode(t, "check-binding.json", sourcecohort.CheckBinding{Schema: "riido-sourcecohort-check-binding-v1", Source: source.File, Observation: obs.File, SourceSchema: method.File, Report: report.File, ExecutedUTC: "2026-01-01T00:00:03Z"})
	review := encode(t, "review.json", sourcecohort.Review{ID: f.ID, Source: source.File, SourceSchema: method.File, ReviewerID: "fixture-reviewer", ReviewedUTC: "2026-01-01T00:00:04Z", Complete: true, Structural: true, SourceScope: "complete_source_situation", SemanticVerdict: "declared_pass", Holds: []sourcecohort.Hold{}, Evidence: []Span{f.Boundary}, Dependencies: []string{}, Observation: obs.File, CheckBinding: check.File, ReadStart: start.File, ReadResult: result.File})
	producer := encode(t, "producer.json", Producer{Schema: ProducerSchema, SourceID: f.ID, ProducerID: f.Coder.ID, ProducerActor: f.Coder.Agent, Selection: "original", Source: source.File, Observation: obs.File, Full: projection.Full.File, History: []File{}})
	acceptance := encode(t, "acceptance.json", Acceptance{Schema: AcceptanceSchema, SourceID: f.ID, Selection: "original", Source: source.File, Observation: obs.File, Full: projection.Full.File, Producer: producer.File, Review: review.File, ReviewerID: "fixture-reviewer", ReviewedUTC: "2026-01-01T00:00:05Z", Verdict: "declared_pass", Scope: AcceptanceScope, OfficialFields: []string{"a_field", "z_field"}, History: []File{}})
	projectionBinding := encode(t, "projection-binding.json", r.Binding)
	graph := raw(r.Binding.Graph.Path, r.GraphBytes)
	v := Version{Schema: VersionSchema, SourceID: f.ID, Selection: "original", ReviewLineage: "first_review", Source: source.File, Observation: obs.File, Full: projection.Full.File, Graph: graph.File, Projection: projectionBinding.File, Producer: producer.File, Review: review.File, Acceptance: []File{acceptance.File}, History: []File{}}
	version := encode(t, "version.json", v)
	frame := Frame{Schema: FrameSchema, Full: v.Full, Projection: v.Projection, Plan: semanticframe.Frame{Schema: semanticframe.ProjectedPlanSchema, FamilyID: "fixture-family", SourceID: f.ID, Slot: 1, Source: v.Source, Inventory: v.Graph, Producer: v.Producer, VersionEvidence: version.File, Observation: v.Observation, SourceReview: v.Review, Definitions: raw("definitions.json", []byte("opaque definitions")).File, InventorySchemaSource: projection.Native.File, Boundary: f.Boundary, Plan: semanticframe.Plan{Bindings: []semanticframe.Binding{{Pointer: "/propositions/1", Role: "utterance_content"}}, Evidence: []Span{f.Boundary}, Description: "public software communication plan", WordingFreedom: "source_supported_rephrasing_only"}, Ambiguities: []semanticframe.Ambiguity{}, TargetMode: "one_complete_comment_text_only"}}
	return Inputs{Projection: projection, ProjectionBinding: projectionBinding, Graph: graph, Version: version, Observation: obs, Producer: producer, Review: review, CheckBinding: check, CheckReport: report, ReadStart: start, ReadResult: result, Acceptance: []PinnedBytes{acceptance}, History: []PinnedBytes{}, Frame: encode(t, "frame.json", frame), RegisteredSourceIDs: []string{f.ID}}
}
func decoded[T any](t testing.TB, p PinnedBytes) T {
	t.Helper()
	var v T
	if json.Unmarshal(p.Bytes, &v) != nil {
		t.Fatal("fixture decode")
	}
	return v
}
func updateVersion(t testing.TB, in *Inputs, v Version) {
	t.Helper()
	in.Version = encode(t, in.Version.File.Path, v)
	f := decoded[Frame](t, in.Frame)
	f.Plan.VersionEvidence = in.Version.File
	in.Frame = encode(t, in.Frame.File.Path, f)
}
func amended(t testing.TB, in *Inputs, disposition string) {
	t.Helper()
	old := raw("old-version.json", []byte("opaque retained historical version"))
	h := History{Schema: HistorySchema, SourceID: "fixture-source", Source: in.Projection.Source.File, Observation: raw("old-observation.json", []byte("old observation")).File, Full: raw("old-full.json", []byte("old full")).File, Producer: raw("old-producer.json", []byte("old producer")).File, Review: raw("old-review.json", []byte("old review")).File, Version: old.File, Disposition: disposition, Relation: "inventory_predecessor"}
	in.History = []PinnedBytes{encode(t, "history.json", h)}
	p := decoded[Producer](t, in.Producer)
	p.Selection = "amended"
	p.History = files(in.History)
	in.Producer = encode(t, in.Producer.File.Path, p)
	a := decoded[Acceptance](t, in.Acceptance[0])
	a.Selection = "amended"
	a.Producer = in.Producer.File
	a.History = files(in.History)
	in.Acceptance[0] = encode(t, in.Acceptance[0].File.Path, a)
	v := decoded[Version](t, in.Version)
	v.Selection = "amended"
	v.Producer = in.Producer.File
	v.History = files(in.History)
	v.Acceptance = files(in.Acceptance)
	f := decoded[Frame](t, in.Frame)
	f.Plan.Producer = in.Producer.File
	in.Frame = encode(t, in.Frame.File.Path, f)
	updateVersion(t, in, v)
}
func TestFullAnchorAndDerivedGraphJoinWithoutQualification(t *testing.T) {
	in := fixture(t)
	r, err := Join(in)
	if err != nil {
		t.Fatal(err)
	}
	if r.Summary.State != "FULL_ACCEPTANCE_AND_GRAPH_PLAN_DECLARATIONS_JOINED_QA_PENDING" || r.Summary.MeaningProven || r.Summary.TrainingEligible || r.Projection.MeaningProven || !r.Projection.UTF8SpansVerified || !bytes.Equal(r.Projection.RetainedFullBytes, in.Projection.Full.Bytes) || r.Projection.Binding.Full == r.Projection.Binding.Graph {
		t.Fatal("wrong anchors or unearned qualification")
	}
	if _, err = semanticframe.NormalizeAccepted("original", semanticframe.AcceptedInputs{Version: in.Version}); err == nil {
		t.Fatal("FULL version must not become V3 acceptance")
	}
}
func TestAmendedVersionRetainsHeldHistory(t *testing.T) {
	in := fixture(t)
	amended(t, &in, "held")
	r, err := Join(in)
	if err != nil {
		t.Fatal(err)
	}
	if r.Summary.RetainedHolds != 1 {
		t.Fatal("history hold was lost")
	}
	h := decoded[History](t, in.History[0])
	h.Disposition = "declared_pass"
	in.History[0] = encode(t, in.History[0].File.Path, h)
	v := decoded[Version](t, in.Version)
	v.History = files(in.History)
	p := decoded[Producer](t, in.Producer)
	p.History = v.History
	in.Producer = encode(t, in.Producer.File.Path, p)
	v.Producer = in.Producer.File
	a := decoded[Acceptance](t, in.Acceptance[0])
	a.History = v.History
	a.Producer = v.Producer
	in.Acceptance[0] = encode(t, in.Acceptance[0].File.Path, a)
	v.Acceptance = files(in.Acceptance)
	updateVersion(t, &in, v)
	if _, err = Join(in); err != Error("retained_history_join") {
		t.Fatalf("promoted old hold: %v", err)
	}
}

func TestFreshReviewOfUnchangedOriginalRetainsProceduralHold(t *testing.T) {
	in := fixture(t)
	v := decoded[Version](t, in.Version)
	h := History{Schema: HistorySchema, SourceID: v.SourceID, Source: v.Source, Observation: v.Observation, Full: v.Full, Producer: v.Producer, Review: raw("old-held-review.json", []byte("opaque old procedural hold record")).File, Version: raw("old-version.json", []byte("opaque old selected version")).File, Disposition: "held", Relation: "review_predecessor"}
	in.History = []PinnedBytes{encode(t, "old-held-review-history.json", h)}
	a := decoded[Acceptance](t, in.Acceptance[0])
	a.History = files(in.History)
	in.Acceptance[0] = encode(t, in.Acceptance[0].File.Path, a)
	v.ReviewLineage = "fresh_review"
	v.History = files(in.History)
	v.Acceptance = files(in.Acceptance)
	updateVersion(t, &in, v)
	r, err := Join(in)
	if err != nil {
		t.Fatal(err)
	}
	if r.Summary.RetainedHolds != 1 || r.Projection.Binding.Full != h.Full || r.Projection.Full.Coder.ID != "fixture-coder" {
		t.Fatal("fresh review lost original inventory or held history")
	}
	// The old held Review cannot be selected as the current passed Review.
	h.Review = v.Review
	in.History[0] = encode(t, in.History[0].File.Path, h)
	a.History = files(in.History)
	in.Acceptance[0] = encode(t, in.Acceptance[0].File.Path, a)
	v.History = files(in.History)
	v.Acceptance = files(in.Acceptance)
	updateVersion(t, &in, v)
	if _, err = Join(in); err != Error("retained_history_join") {
		t.Fatalf("selected old held Review: %v", err)
	}
}
func TestSupportOnlyFullChangeDoesNotReusePriorAcceptance(t *testing.T) {
	in := fixture(t)
	before, err := Join(in)
	if err != nil {
		t.Fatal(err)
	}
	_, f := projectionFixture(t)
	f.OfficialFields[0].Support.Description = "a changed full-only explanation"
	repin(t, &in.Projection, f)
	projected, err := inventoryprojection.Project(in.Projection)
	if err != nil {
		t.Fatal(err)
	}
	if projected.Binding.Graph != before.Projection.Binding.Graph || projected.Binding.Full == before.Projection.Binding.Full {
		t.Fatal("fixture is not a support-only change")
	}
	in.ProjectionBinding = encode(t, in.ProjectionBinding.File.Path, projected.Binding)
	v := decoded[Version](t, in.Version)
	v.Full = in.Projection.Full.File
	v.Projection = in.ProjectionBinding.File
	p := decoded[Producer](t, in.Producer)
	p.Full = v.Full
	in.Producer = encode(t, in.Producer.File.Path, p)
	v.Producer = in.Producer.File
	fm := decoded[Frame](t, in.Frame)
	fm.Full = v.Full
	fm.Projection = v.Projection
	fm.Plan.Producer = v.Producer
	in.Frame = encode(t, in.Frame.File.Path, fm)
	updateVersion(t, &in, v)
	if _, err = Join(in); err != Error("independent_full_acceptance_join") {
		t.Fatalf("old FULL acceptance reused: %v", err)
	}
	a := decoded[Acceptance](t, in.Acceptance[0])
	a.Full = v.Full
	a.Producer = v.Producer
	in.Acceptance[0] = encode(t, in.Acceptance[0].File.Path, a)
	v.Acceptance = files(in.Acceptance)
	updateVersion(t, &in, v)
	if _, err = Join(in); err != nil {
		t.Fatal(err)
	}
}

func TestWrongSelectionActorScopeAndPinsReject(t *testing.T) {
	cases := []struct {
		name   string
		change func(*Inputs)
	}{
		{"missing independent acceptance", func(in *Inputs) { in.Acceptance = nil }},
		{"missing source bytes", func(in *Inputs) { in.Projection.Source = nil }},
		{"wrong selected version", func(in *Inputs) {
			v := decoded[Version](t, in.Version)
			v.Selection = "amended"
			updateVersion(t, in, v)
		}},
		{"graph substituted as full", func(in *Inputs) {
			v := decoded[Version](t, in.Version)
			v.Full = in.Graph.File
			updateVersion(t, in, v)
		}},
		{"wrong projection binding", func(in *Inputs) {
			b := decoded[inventoryprojection.Binding](t, in.ProjectionBinding)
			b.Full = in.Graph.File
			in.ProjectionBinding = encode(t, in.ProjectionBinding.File.Path, b)
		}},
		{"repinned nonderived graph", func(in *Inputs) { in.Graph = raw(in.Graph.File.Path, append(in.Graph.Bytes, ' ')) }},
		{"read actor impersonation", func(in *Inputs) {
			s := decoded[readStart](t, in.ReadStart)
			s.Actor = "fixture-coder"
			in.ReadStart = encode(t, in.ReadStart.File.Path, s)
			r := decoded[sourcecohort.Review](t, in.Review)
			r.ReadStart = in.ReadStart.File
			in.Review = encode(t, in.Review.File.Path, r)
			v := decoded[Version](t, in.Version)
			v.Review = in.Review.File
			updateVersion(t, in, v)
		}},
		{"held current standard review", func(in *Inputs) {
			r := decoded[sourcecohort.Review](t, in.Review)
			r.SemanticVerdict = "declared_hold"
			in.Review = encode(t, in.Review.File.Path, r)
			v := decoded[Version](t, in.Version)
			v.Review = in.Review.File
			updateVersion(t, in, v)
		}},
		{"partial official fields", func(in *Inputs) {
			a := decoded[Acceptance](t, in.Acceptance[0])
			a.OfficialFields = a.OfficialFields[:1]
			in.Acceptance[0] = encode(t, in.Acceptance[0].File.Path, a)
			v := decoded[Version](t, in.Version)
			v.Acceptance = files(in.Acceptance)
			updateVersion(t, in, v)
		}},
		{"support omitted in scope", func(in *Inputs) {
			a := decoded[Acceptance](t, in.Acceptance[0])
			a.Scope = "official_only"
			in.Acceptance[0] = encode(t, in.Acceptance[0].File.Path, a)
			v := decoded[Version](t, in.Version)
			v.Acceptance = files(in.Acceptance)
			updateVersion(t, in, v)
		}},
		{"producer self acceptance", func(in *Inputs) {
			a := decoded[Acceptance](t, in.Acceptance[0])
			a.ReviewerID = "fixture-coder"
			in.Acceptance[0] = encode(t, in.Acceptance[0].File.Path, a)
			v := decoded[Version](t, in.Version)
			v.Acceptance = files(in.Acceptance)
			updateVersion(t, in, v)
		}},
		{"acceptance before review", func(in *Inputs) {
			a := decoded[Acceptance](t, in.Acceptance[0])
			a.ReviewedUTC = "2026-01-01T00:00:01Z"
			in.Acceptance[0] = encode(t, in.Acceptance[0].File.Path, a)
			v := decoded[Version](t, in.Version)
			v.Acceptance = files(in.Acceptance)
			updateVersion(t, in, v)
		}},
		{"frame loses full anchor", func(in *Inputs) {
			f := decoded[Frame](t, in.Frame)
			f.Full = in.Graph.File
			in.Frame = encode(t, in.Frame.File.Path, f)
		}},
		{"history conflict with unopened native reference", func(in *Inputs) {
			amended(t, in, "held")
			h := decoded[History](t, in.History[0])
			h.Version = raw("receipt.json", []byte("conflicting contents")).File
			in.History[0] = encode(t, in.History[0].File.Path, h)
			v := decoded[Version](t, in.Version)
			v.History = files(in.History)
			p := decoded[Producer](t, in.Producer)
			p.History = v.History
			in.Producer = encode(t, in.Producer.File.Path, p)
			v.Producer = in.Producer.File
			a := decoded[Acceptance](t, in.Acceptance[0])
			a.History = v.History
			a.Producer = v.Producer
			in.Acceptance[0] = encode(t, in.Acceptance[0].File.Path, a)
			v.Acceptance = files(in.Acceptance)
			updateVersion(t, in, v)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := fixture(t)
			tc.change(&in)
			if _, err := Join(in); err == nil {
				t.Fatal("accepted invalid declaration")
			}
		})
	}
}
func TestNewContractBoundsBeforeTypedDecode(t *testing.T) {
	tests := []struct {
		name string
		raw  []byte
		code Error
	}{
		{"too large", []byte(strings.Repeat(" ", MaxBytes+1)), "byte_pin"},
		{"too many collection entries", []byte(`{"retained_history":[` + strings.Repeat(`{},`, MaxItems) + `{}]}`), "collection_bounds"},
		{"too deep", []byte(strings.Repeat("[", 34) + "0" + strings.Repeat("]", 34)), "json_depth"},
		{"duplicate key", []byte(`{"schema":"x","schema":"y"}`), "closed_json"},
		{"null field", []byte(`null`), "closed_json"},
		{"case alias", []byte(`{"Schema":"x"}`), "closed_json"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := fixture(t)
			in.Version = raw(in.Version.File.Path, tc.raw)
			if _, err := Join(in); err != tc.code {
				t.Fatalf("want %v, got %v", tc.code, err)
			}
		})
	}
}
func TestAggregateLimitAndCountRejectWholeInput(t *testing.T) {
	in := fixture(t)
	in.Acceptance = make([]PinnedBytes, MaxItems+1)
	if _, err := Join(in); err != Error("input_count") {
		t.Fatalf("count: %v", err)
	}
	in = fixture(t)
	p := raw("a.json", []byte(`"`+strings.Repeat("x", 65536)+`"`))
	in.Acceptance = make([]PinnedBytes, 129)
	for i := range in.Acceptance {
		in.Acceptance[i] = p
	}
	if _, err := Join(in); err != Error("total_bytes") {
		t.Fatalf("aggregate: %v", err)
	}
}

func replaceProducerActor(t testing.TB, in *Inputs, actor string) {
	t.Helper()
	_, full := projectionFixture(t)
	full.Coder.Agent = actor
	repin(t, &in.Projection, full)
	projected, err := inventoryprojection.Project(in.Projection)
	if err != nil {
		t.Fatal(err)
	}
	in.ProjectionBinding = encode(t, in.ProjectionBinding.File.Path, projected.Binding)
	in.Graph = raw(in.Graph.File.Path, projected.GraphBytes)
	p := decoded[Producer](t, in.Producer)
	p.ProducerActor = actor
	p.Full = in.Projection.Full.File
	in.Producer = encode(t, in.Producer.File.Path, p)
	a := decoded[Acceptance](t, in.Acceptance[0])
	a.Full = in.Projection.Full.File
	a.Producer = in.Producer.File
	in.Acceptance[0] = encode(t, in.Acceptance[0].File.Path, a)
	v := decoded[Version](t, in.Version)
	v.Full = in.Projection.Full.File
	v.Graph = in.Graph.File
	v.Projection = in.ProjectionBinding.File
	v.Producer = in.Producer.File
	v.Acceptance = files(in.Acceptance)
	f := decoded[Frame](t, in.Frame)
	f.Full = v.Full
	f.Projection = v.Projection
	f.Plan.Inventory = v.Graph
	f.Plan.Producer = v.Producer
	in.Frame = encode(t, in.Frame.File.Path, f)
	updateVersion(t, in, v)
}

// Independent peer found the bare-label case: different role IDs concealed the
// exact same declared producer/reviewer/read actor. Include its canonical alias.
func TestRejectSameDeclaredProducerReviewerActorAndCanonicalAlias(t *testing.T) {
	for _, actor := range []string{"fixture-reviewer", "/root/fixture-reviewer"} {
		t.Run(actor, func(t *testing.T) {
			in := fixture(t)
			replaceProducerActor(t, &in, actor)
			if _, err := Join(in); err != Error("standard_review_join") {
				t.Fatalf("same declared actor passed: %v", err)
			}
		})
	}
}

func TestDeclaredActorNamespaceIsExplicitAndNotAuthenticated(t *testing.T) {
	in := fixture(t)
	replaceProducerActor(t, &in, "/root/fixture-agent")
	if r, err := Join(in); err != nil || r.Summary.MeaningProven || r.Summary.TrainingEligible {
		t.Fatalf("distinct canonical actor: %v", err)
	}
	for _, actor := range []string{"/other/fixture-agent", "/root/nested/fixture-agent", "root:fixture-agent", "/root/"} {
		in := fixture(t)
		replaceProducerActor(t, &in, actor)
		if _, err := Join(in); err != Error("producer_join") {
			t.Fatalf("unsupported actor namespace passed: %v", err)
		}
	}
}
