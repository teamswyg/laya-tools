package nativecontract

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

// Entirely authored synthetic software metadata; no actual cohort artifacts.
func raw(name string, b []byte) PinnedBytes {
	return PinnedBytes{File: File{Path: name, SHA256: sourcecohort.Hash(b), Bytes: int64(len(b))}, Bytes: b}
}
func encoded(t testing.TB, name string, v any) PinnedBytes {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return raw(name, b)
}
func ref(name string) File { return raw(name, []byte("synthetic opaque target: "+name)).File }
func fixture(t testing.TB, kind, verdict string, witnesses bool) Inputs {
	t.Helper()
	e := Anchors{ID: "fixture-source", ReviewerID: "fixture-reviewer", ProducerID: "fixture-coder", ProducerActor: "/root/fixture-coder", Source: ref("source.txt"), SourceSchema: ref("schema.json"), OriginalObservation: ref("original-observation.json"), OriginalFull: ref("original-full.json"), OriginalProducer: ref("original-producer.json"), OwnInventory: ref("own-inventory.json"), Start: ref("fresh-start.json"), Method: ref("method-limitation.json"), PairScope: ref("pair-scope.json"), AcceptanceScope: ref("acceptance-scope.json"), OfficialKeys: []string{"field_a", "field_b", "field_c", "field_d", "field_e", "field_f", "field_g"}}
	e.SelectedObservation = e.OriginalObservation
	e.SelectedFull = e.OriginalFull
	e.SelectedProducer = e.OriginalProducer
	if kind == Amended {
		e.SelectedObservation = ref("candidate-observation.json")
		e.SelectedFull = ref("candidate-full.json")
		e.SelectedProducer = ref("candidate-pair.json")
	}
	start := encoded(t, "original-read.start", readStart{Version: "riido-reviewpacket/v1", State: "started", UTC: "2026-01-01T00:00:00Z", Expected: e.Source.SHA256, Max: 16384, Policy: "regular-file; no final symlink; nonblocking open; identity checked; bounded read; no retries", Limit: 1, Actor: e.ReviewerID})
	result := encoded(t, start.File.Path+".result", readResult{Version: "riido-reviewpacket/v1", State: "verified", StartSHA: start.File.SHA256, Expected: e.Source.SHA256, Actual: e.Source.SHA256, Bytes: e.Source.Bytes, Attempts: 1, ReadUTC: "2026-01-01T00:01:00Z", UTC: "2026-01-01T00:02:00Z", Errors: []issue{}})
	report := encoded(t, "check-report.json", sourcecohort.CheckReport{StructuralValid: true, MandatoryComplete: true, SourcePhaseScope: true, Scope: "synthetic metadata controls"})
	check := encoded(t, "check-binding.json", sourcecohort.CheckBinding{Schema: "riido-sourcecohort-check-binding-v1", Source: e.Source, Observation: e.SelectedObservation, SourceSchema: e.SourceSchema, Report: report.File, ExecutedUTC: "2026-01-01T00:03:00Z"})
	e.ReadStart = start.File
	e.ReadResult = result.File
	e.Check = check.File
	e.Report = report.File
	old := sourcecohort.Review{ID: e.ID, Source: e.Source, SourceSchema: e.SourceSchema, ReviewerID: e.ReviewerID, ReviewedUTC: "2026-01-01T00:03:00Z", Complete: true, Structural: true, SourceScope: "complete_source_situation", SemanticVerdict: "declared_hold", Holds: []sourcecohort.Hold{{Category: "synthetic", Reason: "retained old hold"}}, Evidence: []sourcecohort.Span{{Kind: "source_scope", Start: 0, End: int(e.Source.Bytes)}}, Dependencies: []string{}, Observation: e.OriginalObservation, CheckBinding: e.Check, ReadStart: e.ReadStart, ReadResult: e.ReadResult}
	oldBytes := encoded(t, "old-held-review.json", old)
	e.HeldReview = oldBytes.File
	r := old
	r.SemanticVerdict = verdict
	r.Observation = e.SelectedObservation
	r.ReviewedUTC = "2026-01-01T00:04:00Z"
	r.Holds = []sourcecohort.Hold{}
	unresolved, held := 0, 0
	if verdict == "declared_hold" {
		r.Holds = []sourcecohort.Hold{{Category: "synthetic", Reason: "new independent hold retained"}}
		unresolved = 1
		held = 1
	}
	rb := encoded(t, "selected-review.json", r)
	copies := []CopyRead{{Input: e.SelectedFull, Start: ref("copy.start"), Result: ref("copy.start.result")}}
	var ab, vb PinnedBytes
	ds := OriginalDispositionSchema
	if kind == Original {
		a := OriginalAcceptance{Schema: OriginalAcceptanceSchema, ID: e.ID, Reviewer: e.ReviewerID, ReviewedUTC: "2026-01-01T00:05:00Z", Source: e.Source, Observation: e.OriginalObservation, Full: e.OriginalFull, Producer: e.OriginalProducer, Review: rb.File, OwnInventory: e.OwnInventory, ReadStart: e.ReadStart, ReadResult: e.ReadResult, Start: e.Start, Copies: copies, Check: e.Check, Report: e.Report, Scope: Scope, Verdict: verdict, Unresolved: unresolved, HeldReview: e.HeldReview, Method: e.Method, Rationale: ref("rationale.json")}
		ab = encoded(t, "acceptance.json", a)
		v := OriginalVersion{Schema: OriginalVersionSchema, ID: e.ID, Reviewer: e.ReviewerID, UTC: "2026-01-01T00:06:00Z", Selection: "reconsidered_original", Source: e.Source, Observation: e.OriginalObservation, Full: e.OriginalFull, Producer: e.OriginalProducer, Review: rb.File, Acceptance: ab.File, HeldReview: e.HeldReview, Start: e.Start, ReadStart: e.ReadStart, ReadResult: e.ReadResult, Check: e.Check, Report: e.Report, OwnInventory: e.OwnInventory, Method: e.Method, Unchanged: true}
		vb = encoded(t, "version.json", v)
	} else {
		a := AmendedAcceptance{Schema: AmendedAcceptanceSchema, ID: e.ID, Reviewer: e.ReviewerID, ReviewedUTC: "2026-01-01T00:05:00Z", Source: e.Source, OriginalObservation: e.OriginalObservation, OriginalFull: e.OriginalFull, OriginalProducer: e.OriginalProducer, Observation: e.SelectedObservation, Full: e.SelectedFull, Pair: e.SelectedProducer, PairScope: e.PairScope, AcceptanceScope: e.AcceptanceScope, Review: rb.File, OwnInventory: e.OwnInventory, ReadStart: e.ReadStart, ReadResult: e.ReadResult, Start: e.Start, Check: e.Check, Report: e.Report, HeldReview: e.HeldReview, Rationale: ref("rationale.json"), Copies: copies, OfficialKeys: append([]string{}, e.OfficialKeys...), Scope: Scope, Verdict: verdict, Unresolved: unresolved}
		ab = encoded(t, "acceptance.json", a)
		v := AmendedVersion{Schema: AmendedVersionSchema, ID: e.ID, Reviewer: e.ReviewerID, UTC: "2026-01-01T00:06:00Z", Selection: "independently_reviewed_amended_pair", Source: e.Source, OriginalObservation: e.OriginalObservation, OriginalFull: e.OriginalFull, OriginalProducer: e.OriginalProducer, Observation: e.SelectedObservation, Full: e.SelectedFull, Pair: e.SelectedProducer, Review: rb.File, Acceptance: ab.File, HeldReview: e.HeldReview, PairScope: e.PairScope, AcceptanceScope: e.AcceptanceScope, Start: e.Start, ReadStart: e.ReadStart, ReadResult: e.ReadResult, Check: e.Check, Report: e.Report, OwnInventory: e.OwnInventory, Unchanged: true}
		vb = encoded(t, "version.json", v)
		ds = AmendedDispositionSchema
	}
	d := Disposition{Schema: ds, ID: e.ID, Reviewer: e.ReviewerID, UTC: "2026-01-01T00:07:00Z", Version: vb.File, Review: rb.File, SemanticHolds: held, HeldReview: e.HeldReview}
	in := Inputs{Kind: kind, Review: rb, Acceptance: ab, Version: vb, Disposition: encoded(t, "disposition.json", d), Expected: e, RegisteredSourceIDs: []string{e.ID}}
	if witnesses {
		in.Witnesses = &Witnesses{ReadStart: start, ReadResult: result, Check: check, Report: report, HeldReview: oldBytes}
	}
	return in
}
func mutateDoc(t testing.TB, p *PinnedBytes, key string, value any) {
	t.Helper()
	var m map[string]json.RawMessage
	if e := json.Unmarshal(p.Bytes, &m); e != nil {
		t.Fatal(e)
	}
	m[key], _ = json.Marshal(value)
	*p = encoded(t, p.File.Path, m)
}
func syncNativePins(t testing.TB, in *Inputs) {
	t.Helper()
	mutateDoc(t, &in.Version, "standard_review", in.Review.File)
	mutateDoc(t, &in.Version, "full_inventory_acceptance", in.Acceptance.File)
	mutateDoc(t, &in.Disposition, "selected_review", in.Review.File)
	mutateDoc(t, &in.Disposition, "selected_version", in.Version.File)
}

func mutateIfDifferent(t testing.TB, p *PinnedBytes, key string, value any) {
	t.Helper()
	var m map[string]json.RawMessage
	_ = json.Unmarshal(p.Bytes, &m)
	next, _ := json.Marshal(value)
	if bytes.Equal(m[key], next) {
		return
	}
	mutateDoc(t, p, key, value)
}
func syncWitnessPins(t testing.TB, in *Inputs) {
	t.Helper()
	w := in.Witnesses
	mutateIfDifferent(t, &w.ReadResult, "start_receipt_sha256", w.ReadStart.File.SHA256)
	mutateIfDifferent(t, &w.Check, "report", w.Report.File)
	e := &in.Expected
	e.ReadStart = w.ReadStart.File
	e.ReadResult = w.ReadResult.File
	e.Check = w.Check.File
	e.Report = w.Report.File
	e.HeldReview = w.HeldReview.File
	mutateDoc(t, &in.Review, "read_start", e.ReadStart)
	mutateDoc(t, &in.Review, "read_result", e.ReadResult)
	mutateDoc(t, &in.Review, "check_binding", e.Check)
	for _, p := range []*PinnedBytes{&in.Acceptance, &in.Version} {
		mutateDoc(t, p, "original_read_start", e.ReadStart)
		mutateDoc(t, p, "original_read_result", e.ReadResult)
		mutateDoc(t, p, "check_binding", e.Check)
		mutateDoc(t, p, "check_report", e.Report)
		mutateDoc(t, p, "retained_held_review", e.HeldReview)
	}
	mutateDoc(t, &in.Acceptance, "standard_review", in.Review.File)
	mutateDoc(t, &in.Disposition, "retained_held_review", e.HeldReview)
	syncNativePins(t, in)
}

func TestOriginalAmendedPassHoldAndExactRawRetention(t *testing.T) {
	for _, kind := range []string{Original, Amended} {
		for _, verdict := range []string{"declared_pass", "declared_hold"} {
			for _, w := range []bool{false, true} {
				in := fixture(t, kind, verdict, w)
				out, e := Join(in)
				if e != nil {
					t.Fatalf("%s/%s/%t:%v", kind, verdict, w, e)
				}
				if out.Selected.Verdict != verdict || out.Summary.DeclaredVerdict != verdict || out.Summary.MeaningProven || out.Summary.TrainingEligible || out.Summary.OfficialKeysVerified != (kind == Amended) || out.Summary.WitnessContentVerified != w || out.Selected.Full != in.Expected.SelectedFull || out.Selected.HeldReview != in.Expected.HeldReview {
					t.Fatal("declaration altered")
				}
				if !bytes.Equal(out.RawAcceptance.Bytes, in.Acceptance.Bytes) || out.RawAcceptance.File != in.Acceptance.File || !bytes.Equal(out.RawVersion.Bytes, in.Version.Bytes) || !bytes.Equal(out.RawReview.Bytes, in.Review.Bytes) || !bytes.Equal(out.RawDisposition.Bytes, in.Disposition.Bytes) {
					t.Fatal("original raw lost")
				}
				in.Acceptance.Bytes[0] = '!'
				if out.RawAcceptance.Bytes[0] == '!' {
					t.Fatal("output aliases caller buffer")
				}
			}
		}
	}
}
func TestNativeLinkAndScopeRefusals(t *testing.T) {
	cases := []struct {
		name   string
		kind   string
		mutate func(*Inputs)
	}{
		{"acceptance-source", Amended, func(in *Inputs) {
			mutateDoc(t, &in.Acceptance, "source", ref("different-source.txt"))
			syncNativePins(t, in)
		}},
		{"candidate-full-graph-substitution", Amended, func(in *Inputs) {
			mutateDoc(t, &in.Acceptance, "candidate_full_inventory", ref("graph.json"))
			syncNativePins(t, in)
		}},
		{"original-full-in-version", Amended, func(in *Inputs) {
			mutateDoc(t, &in.Version, "original_full_inventory", ref("changed-original-full.json"))
			mutateDoc(t, &in.Disposition, "selected_version", in.Version.File)
		}},
		{"wrong-pair", Amended, func(in *Inputs) {
			mutateDoc(t, &in.Acceptance, "candidate_pair_binding", ref("wrong-pair.json"))
			syncNativePins(t, in)
		}},
		{"wrong-reviewer", Original, func(in *Inputs) { mutateDoc(t, &in.Acceptance, "reviewer_id", "other-reviewer"); syncNativePins(t, in) }},
		{"old-hold-pass", Original, func(in *Inputs) {
			mutateDoc(t, &in.Version, "old_held_record_pass", true)
			mutateDoc(t, &in.Disposition, "selected_version", in.Version.File)
		}},
		{"whole-QA", Amended, func(in *Inputs) { mutateDoc(t, &in.Acceptance, "whole_source_qa_pass", true); syncNativePins(t, in) }},
		{"training", Original, func(in *Inputs) { mutateDoc(t, &in.Disposition, "training_eligible", true) }},
		{"scope", Original, func(in *Inputs) { mutateDoc(t, &in.Acceptance, "scope", "official_only"); syncNativePins(t, in) }},
		{"incomplete-official-key-set", Amended, func(in *Inputs) {
			mutateDoc(t, &in.Acceptance, "complete_official_field_keys", []string{"field_a"})
			syncNativePins(t, in)
		}},
		{"duplicate-official-key", Amended, func(in *Inputs) {
			keys := append([]string{}, in.Expected.OfficialKeys...)
			keys[1] = keys[0]
			mutateDoc(t, &in.Acceptance, "complete_official_field_keys", keys)
			syncNativePins(t, in)
		}},
		{"producer-self-review", Amended, func(in *Inputs) { in.Expected.ProducerID = in.Expected.ReviewerID }},
		{"producer-actor-self-review", Amended, func(in *Inputs) { in.Expected.ProducerActor = "/root/" + in.Expected.ReviewerID }},
		{"arbitrary-actor-prefix", Original, func(in *Inputs) { in.Expected.ProducerActor = "/other/fixture-coder" }},
		{"unregistered-source", Original, func(in *Inputs) { in.RegisteredSourceIDs = []string{"different-source"} }},
		{"duplicate-source-roster", Original, func(in *Inputs) { in.RegisteredSourceIDs = append(in.RegisteredSourceIDs, in.Expected.ID) }},
		{"old-review-selected-again", Original, func(in *Inputs) { in.Expected.HeldReview = in.Review.File }},
		{"copy-result-pair", Original, func(in *Inputs) {
			mutateDoc(t, &in.Acceptance, "copy_read_receipts", []CopyRead{{Input: in.Expected.SelectedFull, Start: ref("copy.start"), Result: ref("different.result")}})
			syncNativePins(t, in)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := fixture(t, tc.kind, "declared_pass", false)
			tc.mutate(&in)
			if out, e := Join(in); e == nil || out.Summary.MeaningProven || out.Summary.TrainingEligible {
				t.Fatal("bad native join accepted", e)
			}
		})
	}
}
func TestDeclaredVerdictCannotBeMinted(t *testing.T) {
	in := fixture(t, Amended, "declared_hold", false)
	mutateDoc(t, &in.Acceptance, "verdict", "declared_pass")
	syncNativePins(t, &in)
	if _, e := Join(in); e == nil {
		t.Fatal("hold promoted")
	}
	in = fixture(t, Original, "declared_pass", false)
	mutateDoc(t, &in.Disposition, "current_semantic_holds", 1)
	if _, e := Join(in); e == nil {
		t.Fatal("pass with hold accepted")
	}
	in = fixture(t, Original, "declared_hold", false)
	mutateDoc(t, &in.Acceptance, "verdict", "unknown")
	syncNativePins(t, &in)
	if _, e := Join(in); e == nil {
		t.Fatal("new verdict invented")
	}
}
func TestClosedKeysRequiredFlagsUnicodeAndBounds(t *testing.T) {
	good := fixture(t, Original, "declared_pass", false).Acceptance
	for _, kind := range []string{"missing-flag", "duplicate-key", "extra-field", "wrong-case", "null-array", "bad-unicode", "trailing", "depth", "collection", "oversize"} {
		t.Run(kind, func(t *testing.T) {
			in := fixture(t, Original, "declared_pass", false)
			b := append([]byte{}, good.Bytes...)
			switch kind {
			case "missing-flag":
				b = bytes.Replace(b, []byte(`,"training_eligible":false`), nil, 1)
			case "duplicate-key":
				b = bytes.Replace(b, []byte(`"schema":`), []byte(`"schema":"duplicate","schema":`), 1)
			case "extra-field":
				b = bytes.Replace(b, []byte(`"schema":`), []byte(`"extra":0,"schema":`), 1)
			case "wrong-case":
				b = bytes.Replace(b, []byte(`"schema":`), []byte(`"Schema":`), 1)
			case "null-array":
				b = bytes.Replace(b, []byte(`"copy_read_receipts":[`), []byte(`"copy_read_receipts":null,"other":[`), 1)
			case "bad-unicode":
				b = bytes.Replace(b, []byte(`"reviewer_id":"fixture-reviewer"`), []byte(`"reviewer_id":"\ud800"`), 1)
			case "trailing":
				b = append(b, []byte(`{}`)...)
			case "depth":
				b = []byte(strings.Repeat("[", 34) + "0" + strings.Repeat("]", 34))
			case "collection":
				b = []byte("[" + strings.Repeat("0,", 256) + "0]")
			case "oversize":
				b = append(b, []byte(strings.Repeat(" ", MaxBytes))...)
			}
			in.Acceptance = raw(in.Acceptance.File.Path, b)
			syncNativePins(t, &in)
			if _, e := Join(in); e == nil {
				t.Fatal("bad raw accepted")
			}
		})
	}
	in := fixture(t, Amended, "declared_pass", false)
	in.Acceptance.Bytes[0] = '!'
	if _, e := Join(in); e == nil {
		t.Fatal("unpinned byte mutation")
	}
}
func TestConflictingOpaquePinsAndWitnesses(t *testing.T) {
	in := fixture(t, Original, "declared_pass", false)
	bad := in.Expected.SelectedFull
	bad.SHA256 = strings.Repeat("b", 64)
	mutateDoc(t, &in.Acceptance, "semantic_rationale", bad)
	syncNativePins(t, &in)
	if _, e := Join(in); e == nil {
		t.Fatal("opaque conflicting path pin")
	}
	cases := []struct {
		name   string
		change func(*Inputs)
	}{
		{"readactor", func(in *Inputs) { mutateDoc(t, &in.Witnesses.ReadStart, "actor", "other-reviewer") }},
		{"check-source", func(in *Inputs) { mutateDoc(t, &in.Witnesses.Check, "source", ref("other-source")) }},
		{"old-hold-removed", func(in *Inputs) { mutateDoc(t, &in.Witnesses.HeldReview, "holds", []sourcecohort.Hold{}) }},
		{"unsupported-report", func(in *Inputs) { mutateDoc(t, &in.Witnesses.Report, "source_phase_scope", false) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := fixture(t, Amended, "declared_pass", true)
			tc.change(&in)
			syncWitnessPins(t, &in)
			if _, e := Join(in); e == nil {
				t.Fatal("bad witness accepted")
			}
		})
	}
}
func TestOriginalHasNoMintedSevenKeyField(t *testing.T) {
	for _, keys := range [][]string{nil, {"unused-caller-key"}} {
		in := fixture(t, Original, "declared_pass", false)
		in.Expected.OfficialKeys = keys
		out, e := Join(in)
		if e != nil || out.Summary.OfficialKeysVerified {
			t.Fatal("original contract wrongly requires/invents seven-key field", e)
		}
	}
	in := fixture(t, Original, "declared_pass", false)
	mutateDoc(t, &in.Acceptance, "complete_official_field_keys", in.Expected.OfficialKeys)
	syncNativePins(t, &in)
	if _, e := Join(in); e == nil {
		t.Fatal("unrecognized original field silently admitted")
	}
	in = fixture(t, Amended, "declared_pass", false)
	a := in.Expected.SelectedObservation
	in.Expected.SelectedObservation = in.Expected.OriginalObservation
	mutateDoc(t, &in.Acceptance, "candidate_observation", in.Expected.SelectedObservation)
	mutateDoc(t, &in.Version, "candidate_observation", in.Expected.SelectedObservation)
	mutateDoc(t, &in.Review, "observation", in.Expected.SelectedObservation)
	mutateDoc(t, &in.Acceptance, "standard_review", in.Review.File)
	syncNativePins(t, &in)
	if _, e := Join(in); e != nil {
		t.Fatal("support-only amended Obs exact-original pin should be valid", a, e)
	}
}

func TestRetainedWitnessVerdictAndAllPinRegistry(t *testing.T) {
	for _, name := range []string{"unsupported-verdict", "conflicting-old-check", "invalid-old-read-path"} {
		t.Run(name, func(t *testing.T) {
			in := fixture(t, Original, "declared_pass", true)
			switch name {
			case "unsupported-verdict":
				mutateDoc(t, &in.Witnesses.HeldReview, "independent_semantic_verdict", "invented_pass")
			case "conflicting-old-check":
				bad := in.Expected.Check
				bad.SHA256 = strings.Repeat("b", 64)
				mutateDoc(t, &in.Witnesses.HeldReview, "check_binding", bad)
			case "invalid-old-read-path":
				bad := in.Expected.ReadStart
				bad.Path = "../invalid.start"
				mutateDoc(t, &in.Witnesses.HeldReview, "read_start", bad)
			}
			syncWitnessPins(t, &in)
			if _, e := Join(in); e == nil {
				t.Fatal("bad fully repinned old witness accepted")
			}
		})
	}
	in := fixture(t, Original, "declared_pass", true)
	mutateDoc(t, &in.Witnesses.HeldReview, "independent_semantic_verdict", "declared_pass")
	mutateDoc(t, &in.Witnesses.HeldReview, "holds", []sourcecohort.Hold{{Category: "technical", Reason: "synthetic retained failed read"}})
	syncWitnessPins(t, &in)
	out, e := Join(in)
	if e != nil || !bytes.Equal(out.RawWitnesses.HeldReview.Bytes, in.Witnesses.HeldReview.Bytes) || out.Disposition.OldPass {
		t.Fatal("technical-held semantic declaration altered/promoted", e)
	}
}
