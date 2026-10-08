package semanticframe

import (
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

// These original public software fixtures are not source-cohort examples,
// independently accepted data, reference labels or evidence of model quality.
func pinned(t *testing.T, name string, value any) PinnedBytes {
	t.Helper()
	b, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return rawPinned(name, b)
}
func rawPinned(name string, b []byte) PinnedBytes {
	return PinnedBytes{File: File{Path: name, SHA256: sourcecohort.Hash(b), Bytes: int64(len(b))}, Bytes: b}
}

func fixture(t *testing.T, amended bool) (Frame, AcceptedInputs) {
	t.Helper()
	source := rawPinned("source.txt", []byte("항목 A reports a check."))
	e := []Span{{Kind: "proposition", Start: 0, End: len(source.Bytes)}}
	g := InventoryGraph{Schema: InventorySchema, SourceID: "fixture-source", Source: source.File, Boundary: Span{Kind: "source_scope", Start: 0, End: len(source.Bytes)}, Components: []Component{}, Propositions: []Proposition{{"p1", "A reports a check", "positive", "current operation", []string{"r1"}, e}}, Oppositions: []Opposition{}, Temporal: []Temporal{}, Referents: []Referent{{"r1", "operation", nil, "operation A", e}}, Constraints: []Constraint{}, Dependencies: []string{}}
	in := AcceptedInputs{Source: source, Observation: rawPinned("observation.json", []byte(`{"software_fixture":true}`)), Inventory: pinned(t, "inventory.json", g), Predecessors: []PinnedBytes{}}
	selection, schema := "original", OriginalProducerSchema
	if amended {
		selection, schema = "amended", AmendedProducerSchema
		old := VersionBinding{Schema: VersionSchema, Selection: "original", Source: source.File, Observation: rawPinned("old-observation.json", []byte("old observation")).File, Inventory: rawPinned("old-inventory.json", []byte("old inventory")).File, Producer: rawPinned("old-producer.json", []byte("old producer")).File, SourceReview: rawPinned("old-review.json", []byte("old review")).File, InventoryAcceptance: []File{rawPinned("old-acceptance.json", []byte("old acceptance")).File}, Predecessors: []File{}}
		in.Predecessors = []PinnedBytes{pinned(t, "predecessor.json", old)}
	}
	producer := ProducerDeclaration{Schema: schema, ProducerID: "fixture-producer", SourceID: g.SourceID, Source: source.File, Observation: in.Observation.File, Inventory: in.Inventory.File, Predecessors: pinsOf(in.Predecessors)}
	in.Producer = pinned(t, "producer.json", producer)
	ref := func(name string) File { return rawPinned(name, []byte("declared reference only")).File }
	review := sourcecohort.Review{ID: g.SourceID, Source: source.File, SourceSchema: ref("source-schema.json"), ReviewerID: "fixture-reviewer", ReviewedUTC: "2026-01-01T00:00:00Z", Complete: true, Structural: true, SourceScope: "complete_source_situation", SemanticVerdict: "declared_pass", Holds: []sourcecohort.Hold{}, Evidence: []Span{g.Boundary}, Dependencies: []string{}, Observation: in.Observation.File, CheckBinding: ref("check-binding.json"), ReadStart: ref("read-start.json"), ReadResult: ref("read-start.json.result")}
	in.SourceReview = pinned(t, "review.json", review)
	acceptance := InventoryAcceptance{Schema: AcceptanceSchema, SourceID: g.SourceID, Selection: selection, Source: source.File, Observation: in.Observation.File, Inventory: in.Inventory.File, Producer: in.Producer.File, SourceReview: in.SourceReview.File, ReviewerID: "fixture-inventory-reviewer", Verdict: "declared_pass", Predecessors: pinsOf(in.Predecessors)}
	in.InventoryAcceptance = []PinnedBytes{pinned(t, "acceptance.json", acceptance)}
	v := VersionBinding{Schema: VersionSchema, Selection: selection, Source: source.File, Observation: in.Observation.File, Inventory: in.Inventory.File, Producer: in.Producer.File, SourceReview: in.SourceReview.File, InventoryAcceptance: pinsOf(in.InventoryAcceptance), Predecessors: pinsOf(in.Predecessors)}
	in.Version = pinned(t, "version.json", v)
	f := Frame{Schema: Schema, FamilyID: "fixture-family", SourceID: g.SourceID, Slot: 1, Source: source.File, Inventory: in.Inventory.File, Producer: in.Producer.File, VersionEvidence: in.Version.File, Observation: in.Observation.File, SourceReview: in.SourceReview.File, Definitions: rawPinned("definitions.json", []byte("source method")).File, InventorySchemaSource: rawPinned("inventory-method.go", []byte("inventory method")).File, Boundary: g.Boundary, Plan: Plan{Bindings: []Binding{{"/propositions/0", "utterance_content"}, {"/referent_support/0", "governing_constraint"}}, Evidence: e, Description: "Source-supported software fixture plan", WordingFreedom: "source_supported_rephrasing_only"}, Ambiguities: []Ambiguity{}, TargetMode: "one_complete_comment_text_only"}
	return f, in
}

func TestOriginalAndAmendedDeclarationsJoin(t *testing.T) {
	for _, amended := range []bool{false, true} {
		name := "original"
		if amended {
			name = "amended"
		}
		t.Run(name, func(t *testing.T) {
			f, in := fixture(t, amended)
			selection := "original"
			if amended {
				selection = "amended"
			}
			n, err := NormalizeAccepted(selection, in)
			if err != nil {
				t.Fatal(err)
			}
			s, err := JoinFrame(f, n, []string{f.SourceID})
			if err != nil {
				t.Fatal(err)
			}
			if s.State != "STRUCTURAL_DECLARATIONS_JOINED_QA_PENDING" || s.MeaningProven || s.TrainingEligible {
				t.Fatalf("unexpected authority: %+v", s)
			}
			// Caller-owned bytes cannot mutate the normalized retained source.
			in.Source.Bytes[0] = 'x'
			if _, err = JoinFrame(f, n, []string{f.SourceID}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSelectedVersionCannotAcceptSubstitutedEvidence(t *testing.T) {
	cases := []struct {
		name   string
		change func(*AcceptedInputs)
	}{
		{"inventory bytes", func(in *AcceptedInputs) { in.Inventory.Bytes = append(in.Inventory.Bytes, ' ') }},
		{"repinned inventory", func(in *AcceptedInputs) {
			in.Inventory = rawPinned(in.Inventory.File.Path, append(in.Inventory.Bytes, ' '))
		}},
		{"acceptance bytes", func(in *AcceptedInputs) {
			in.InventoryAcceptance[0].Bytes = append(in.InventoryAcceptance[0].Bytes, ' ')
		}},
		{"repinned acceptance", func(in *AcceptedInputs) {
			in.InventoryAcceptance[0] = rawPinned(in.InventoryAcceptance[0].File.Path, append(in.InventoryAcceptance[0].Bytes, ' '))
		}},
		{"missing independent acceptance", func(in *AcceptedInputs) { in.InventoryAcceptance = nil }},
		{"duplicate independent acceptance", func(in *AcceptedInputs) {
			in.InventoryAcceptance = append(in.InventoryAcceptance, in.InventoryAcceptance[0])
		}},
		{"wrong source", func(in *AcceptedInputs) { in.Source = rawPinned("source.txt", []byte("different source")) }},
		{"wrong review", func(in *AcceptedInputs) {
			var r sourcecohort.Review
			json.Unmarshal(in.SourceReview.Bytes, &r)
			r.Observation = rawPinned("other.json", []byte("other")).File
			in.SourceReview = pinned(t, "review.json", r)
		}},
		{"wrong wire branch", func(in *AcceptedInputs) {
			var p ProducerDeclaration
			json.Unmarshal(in.Producer.Bytes, &p)
			p.Schema = AmendedProducerSchema
			in.Producer = pinned(t, "producer.json", p)
		}},
		{"unknown inventory wire", func(in *AcceptedInputs) {
			in.Inventory = rawPinned("inventory.json", []byte(`{"schema":"historical-wire"}`))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, in := fixture(t, false)
			tc.change(&in)
			if _, err := NormalizeAccepted("original", in); err == nil {
				t.Fatal("substitution accepted")
			}
		})
	}
	_, in := fixture(t, true)
	if _, err := NormalizeAccepted("original", in); err == nil {
		t.Fatal("wrong explicit selection accepted")
	}
}

func TestFrameRejectsPointersRolesDependenciesAndSpans(t *testing.T) {
	cases := []struct {
		name   string
		change func(*Frame)
		code   string
	}{
		{"field rather than item", func(f *Frame) { f.Plan.Bindings[0].Pointer = "/propositions/0/polarity" }, "inventory_pointer"},
		{"noncanonical index", func(f *Frame) { f.Plan.Bindings[0].Pointer = "/propositions/00" }, "inventory_pointer"},
		{"escaped pointer", func(f *Frame) { f.Plan.Bindings[0].Pointer = "/propositions~1/0" }, "inventory_pointer"},
		{"index out of bounds", func(f *Frame) { f.Plan.Bindings[0].Pointer = "/propositions/1" }, "inventory_pointer"},
		{"conflicting role", func(f *Frame) {
			f.Plan.Bindings = append(f.Plan.Bindings, Binding{"/propositions/0", "unavailable_source_context"})
		}, "duplicate_or_conflicting_role"},
		{"invented claim state", func(f *Frame) { f.Plan.Bindings[0].Role = "completed" }, "disclosure_role"},
		{"missing subject", func(f *Frame) { f.Plan.Bindings = f.Plan.Bindings[:1] }, "plan_referent_dependency"},
		{"split UTF8", func(f *Frame) { f.Plan.Evidence = []Span{{Kind: "proposition", Start: 1, End: 3}} }, "plan_bounds"},
		{"partial source boundary", func(f *Frame) { f.Boundary.End-- }, "source_boundary"},
		{"wrong version pin", func(f *Frame) { f.VersionEvidence.SHA256 = strings.Repeat("a", 64) }, "frame_version_join"},
		{"unknown ambiguity item", func(f *Frame) {
			f.Ambiguities = []Ambiguity{{[]string{"/propositions/1"}, f.Plan.Evidence, "uncertain", "preserve_without_resolution"}}
		}, "ambiguity_pointer"},
		{"path alias", func(f *Frame) { f.Definitions.Path = "a/../definitions.json" }, "method_pin"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, in := fixture(t, false)
			n, err := NormalizeAccepted("original", in)
			if err != nil {
				t.Fatal(err)
			}
			tc.change(&f)
			_, err = JoinFrame(f, n, []string{f.SourceID})
			if err != Error(tc.code) {
				t.Fatalf("got %v, want %s", err, tc.code)
			}
		})
	}
}

func TestInventoryGraphRejectsDanglingEndpointsAndCycles(t *testing.T) {
	cases := []struct {
		name   string
		change func(*InventoryGraph)
	}{
		{"duplicate proposition", func(g *InventoryGraph) { g.Propositions = append(g.Propositions, g.Propositions[0]) }},
		{"dangling subject", func(g *InventoryGraph) { g.Propositions[0].Subjects = []string{"missing"} }},
		{"dangling opposition", func(g *InventoryGraph) {
			g.Oppositions = []Opposition{{"o1", "p1", "missing", "opposition", g.Propositions[0].Evidence}}
		}},
		{"dangling temporal", func(g *InventoryGraph) {
			g.Temporal = []Temporal{{"t1", "relative", []string{"missing"}, "anchor", g.Propositions[0].Evidence, g.Propositions[0].Evidence}}
		}},
		{"dangling constraint", func(g *InventoryGraph) {
			g.Constraints = []Constraint{{"quote", []string{"missing"}, "constraint", g.Propositions[0].Evidence}}
		}},
		{"referent self cycle", func(g *InventoryGraph) { p := "r1"; g.Referents[0].Parent = &p }},
		{"referent missing parent", func(g *InventoryGraph) { p := "missing"; g.Referents[0].Parent = &p }},
		{"source self dependency", func(g *InventoryGraph) { g.Dependencies = []string{g.SourceID} }},
		{"duplicate source dependency", func(g *InventoryGraph) { g.Dependencies = []string{"other", "other"} }},
		{"split codepoint", func(g *InventoryGraph) { g.Propositions[0].Evidence = []Span{{Kind: "proposition", Start: 0, End: 1}} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, in := fixture(t, false)
			var g InventoryGraph
			if err := json.Unmarshal(in.Inventory.Bytes, &g); err != nil {
				t.Fatal(err)
			}
			tc.change(&g)
			if err := validateInventory(g, in.Source.Bytes); err == nil {
				t.Fatal("bad graph accepted")
			}
		})
	}
	_, in := fixture(t, false)
	n, err := NormalizeAccepted("original", in)
	if err != nil {
		t.Fatal(err)
	}
	n.inventory.Dependencies = []string{"other"}
	f, _ := fixture(t, false)
	if _, err = JoinFrame(f, n, []string{f.SourceID}); err != Error("orphan_dependency") {
		t.Fatalf("orphan got %v", err)
	}
}

func TestClosedBoundedFrameDecode(t *testing.T) {
	f, _ := fixture(t, false)
	raw := pinned(t, "frame.json", f).Bytes
	decoded, err := DecodeFrame(raw)
	if err != nil || decoded.FamilyID != f.FamilyID {
		t.Fatalf("valid decode: %v", err)
	}
	bad := [][]byte{[]byte(`{}`), append(append([]byte{}, raw...), raw...), []byte(strings.Replace(string(raw), `"family_slot":1`, `"family_slot":1,"family_slot":2`, 1)), []byte(strings.Replace(string(raw), `"family_id"`, `"Family_ID"`, 1)), []byte(strings.Replace(string(raw), `"preserved_ambiguities":[]`, `"preserved_ambiguities":null`, 1)), []byte(strings.Replace(string(raw), `"fixture-family"`, `"\ud800"`, 1)), []byte(strings.Repeat(" ", MaxFrameBytes+1))}
	for i, b := range bad {
		if _, err := DecodeFrame(b); err == nil {
			t.Fatalf("bad JSON %d accepted", i)
		}
	}
}

// Repin the selected review together with its declared acceptance/version links,
// so failures below exercise metadata consistency rather than stale byte pins.
func replaceReview(t *testing.T, in *AcceptedInputs, r sourcecohort.Review) {
	t.Helper()
	in.SourceReview = pinned(t, "review.json", r)
	var a InventoryAcceptance
	if err := json.Unmarshal(in.InventoryAcceptance[0].Bytes, &a); err != nil {
		t.Fatal(err)
	}
	a.SourceReview = in.SourceReview.File
	in.InventoryAcceptance[0] = pinned(t, "acceptance.json", a)
	var v VersionBinding
	if err := json.Unmarshal(in.Version.Bytes, &v); err != nil {
		t.Fatal(err)
	}
	v.SourceReview = in.SourceReview.File
	v.InventoryAcceptance = pinsOf(in.InventoryAcceptance)
	in.Version = pinned(t, "version.json", v)
}

func TestRepinnedMalformedReviewMetadataIsRejected(t *testing.T) {
	cases := []struct {
		name   string
		change func(*sourcecohort.Review)
		code   string
	}{
		{"empty UTC", func(r *sourcecohort.Review) { r.ReviewedUTC = "" }, "review_utc"},
		{"non UTC", func(r *sourcecohort.Review) { r.ReviewedUTC = "2026-01-01T00:00:00+09:00" }, "review_utc"},
		{"empty read reference", func(r *sourcecohort.Review) { r.ReadStart = File{} }, "review_reference_metadata"},
		{"read-result mismatch", func(r *sourcecohort.Review) { r.ReadResult.Path = "another.result" }, "review_reference_metadata"},
		{"no complete Source span", func(r *sourcecohort.Review) { r.Evidence[0].Kind = "proposition" }, "review_boundary"},
		{"missing required dependency", func(r *sourcecohort.Review) { r.Dependencies = []string{"other"} }, "review_dependencies"},
		{"producer reviews own inventory", func(r *sourcecohort.Review) { r.ReviewerID = "fixture-producer" }, "source_review_join"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, in := fixture(t, false)
			var r sourcecohort.Review
			json.Unmarshal(in.SourceReview.Bytes, &r)
			tc.change(&r)
			replaceReview(t, &in, r)
			if _, err := NormalizeAccepted("original", in); err != Error(tc.code) {
				t.Fatalf("got %v want %s", err, tc.code)
			}
		})
	}
}

func TestRepinnedProducerAndVersionCannotSubstituteInventoryAcceptance(t *testing.T) {
	_, in := fixture(t, false)
	var g InventoryGraph
	json.Unmarshal(in.Inventory.Bytes, &g)
	g.Propositions[0].Description = "A different declared content"
	in.Inventory = pinned(t, "inventory.json", g)
	var p ProducerDeclaration
	json.Unmarshal(in.Producer.Bytes, &p)
	p.Inventory = in.Inventory.File
	in.Producer = pinned(t, "producer.json", p)
	var v VersionBinding
	json.Unmarshal(in.Version.Bytes, &v)
	v.Inventory = in.Inventory.File
	v.Producer = in.Producer.File
	in.Version = pinned(t, "version.json", v)
	// Source and observation remain correct, as does the standard Source-review.
	// The independently pinned old inventory acceptance cannot cover this hybrid.
	if _, err := NormalizeAccepted("original", in); err != Error("inventory_acceptance_join") {
		t.Fatalf("hybrid got %v", err)
	}
}

func TestNormalizationBounds(t *testing.T) {
	_, in := fixture(t, false)
	in.Source = rawPinned("source.txt", []byte(strings.Repeat("s", MaxSourceBytes+1)))
	if _, err := NormalizeAccepted("original", in); err != Error("byte_pin") {
		t.Fatalf("source bound: %v", err)
	}
	_, in = fixture(t, false)
	in.Inventory = rawPinned("inventory.json", []byte(strings.Repeat("i", MaxInventoryBytes+1)))
	if _, err := NormalizeAccepted("original", in); err != Error("byte_pin") {
		t.Fatalf("inventory bound: %v", err)
	}
	_, in = fixture(t, false)
	in.Predecessors = make([]PinnedBytes, MaxItems+1)
	if _, err := NormalizeAccepted("original", in); err != Error("evidence_count") {
		t.Fatalf("history bound: %v", err)
	}
}

func TestCrossRoleReferencePinsRemainConsistent(t *testing.T) {
	t.Run("review metadata conflicts with captured Source", func(t *testing.T) {
		_, in := fixture(t, false)
		var r sourcecohort.Review
		json.Unmarshal(in.SourceReview.Bytes, &r)
		r.SourceSchema.Path = in.Source.File.Path
		replaceReview(t, &in, r)
		if _, err := NormalizeAccepted("original", in); err != Error("conflicting_reference_pin") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("frame method conflicts with captured acceptance", func(t *testing.T) {
		f, in := fixture(t, false)
		n, err := NormalizeAccepted("original", in)
		if err != nil {
			t.Fatal(err)
		}
		f.Definitions.Path = in.InventoryAcceptance[0].File.Path
		if _, err = JoinFrame(f, n, []string{f.SourceID}); err != Error("conflicting_reference_pin") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("identical metadata method repetition is allowed", func(t *testing.T) {
		f, in := fixture(t, false)
		var r sourcecohort.Review
		json.Unmarshal(in.SourceReview.Bytes, &r)
		r.SourceSchema = f.Definitions
		replaceReview(t, &in, r)
		f.SourceReview = in.SourceReview.File
		f.VersionEvidence = in.Version.File
		n, err := NormalizeAccepted("original", in)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = JoinFrame(f, n, []string{f.SourceID}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("retained declaration conflicts with current acceptance", func(t *testing.T) {
		_, in := fixture(t, true)
		var old VersionBinding
		json.Unmarshal(in.Predecessors[0].Bytes, &old)
		old.InventoryAcceptance[0].Path = in.InventoryAcceptance[0].File.Path
		in.Predecessors[0] = pinned(t, "predecessor.json", old)
		var p ProducerDeclaration
		json.Unmarshal(in.Producer.Bytes, &p)
		p.Predecessors = pinsOf(in.Predecessors)
		in.Producer = pinned(t, "producer.json", p)
		var a InventoryAcceptance
		json.Unmarshal(in.InventoryAcceptance[0].Bytes, &a)
		a.Producer = in.Producer.File
		a.Predecessors = pinsOf(in.Predecessors)
		in.InventoryAcceptance[0] = pinned(t, "acceptance.json", a)
		var v VersionBinding
		json.Unmarshal(in.Version.Bytes, &v)
		v.Producer = in.Producer.File
		v.Predecessors = pinsOf(in.Predecessors)
		v.InventoryAcceptance = pinsOf(in.InventoryAcceptance)
		in.Version = pinned(t, "version.json", v)
		if _, err := NormalizeAccepted("amended", in); err != Error("conflicting_reference_pin") {
			t.Fatalf("got %v", err)
		}
	})
}

func TestTypedFrameBoundsBeforeParsingOrSorting(t *testing.T) {
	f, in := fixture(t, false)
	n, err := NormalizeAccepted("original", in)
	if err != nil {
		t.Fatal(err)
	}
	large := strings.Repeat("/", 1<<20)
	cases := []struct {
		name   string
		change func(*Frame)
	}{
		{"large binding pointer", func(f *Frame) { f.Plan.Bindings[0].Pointer = large }},
		{"large ambiguity pointer", func(f *Frame) {
			f.Ambiguities = []Ambiguity{{[]string{large}, f.Plan.Evidence, "uncertain", "preserve_without_resolution"}}
		}},
		{"large hash", func(f *Frame) { f.Definitions.SHA256 = large }},
		{"large path", func(f *Frame) { f.Definitions.Path = large }},
		{"large role", func(f *Frame) { f.Plan.Bindings[0].Role = large }},
		{"large collection", func(f *Frame) { f.Plan.Bindings = make([]Binding, MaxItems+1) }},
		{"aggregate encoded bytes", func(f *Frame) {
			a := Ambiguity{[]string{"/propositions/0"}, f.Plan.Evidence, strings.Repeat("x", 512), "preserve_without_resolution"}
			f.Ambiguities = make([]Ambiguity, MaxItems)
			for i := range f.Ambiguities {
				f.Ambiguities[i] = a
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, _ := fixture(t, false)
			tc.change(&f)
			if _, err := JoinFrame(f, n, []string{f.SourceID}); err != Error("typed_frame_bounds") {
				t.Fatalf("got %v", err)
			}
		})
	}
	// Input construction is excluded. This measures Go allocations for rejection,
	// not global RSS, native memory or GPU use. No tests in this package run parallel.
	f.Plan.Bindings[0].Pointer = large
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	_, err = JoinFrame(f, n, []string{f.SourceID})
	runtime.ReadMemStats(&after)
	if err != Error("typed_frame_bounds") {
		t.Fatal(err)
	}
	allocated := after.TotalAlloc - before.TotalAlloc
	t.Logf("Go allocated bytes for 1 MiB pointer rejection: %d", allocated)
	if allocated > 64<<10 {
		t.Fatalf("rejection allocated %d bytes", allocated)
	}
	badPin := in.Source.File
	badPin.SHA256 = large
	if allocations := testing.AllocsPerRun(10, func() {
		if validFile(badPin) {
			panic("bad hash accepted")
		}
	}); allocations != 0 {
		t.Fatalf("large hash validation allocations: %v", allocations)
	}
	badPin = in.Source.File
	badPin.Path = large
	if allocations := testing.AllocsPerRun(10, func() {
		if validFile(badPin) {
			panic("bad path accepted")
		}
	}); allocations != 0 {
		t.Fatalf("large path validation allocations: %v", allocations)
	}
}

func TestTypedFrameSizeMatchesCompactJSON(t *testing.T) {
	f, _ := fixture(t, false)
	f.Plan.Description = "quoted \" \\ control \x01 \n <>& 한글 \u2028\u2029"
	f.Ambiguities = []Ambiguity{{[]string{"/propositions/0"}, f.Plan.Evidence, "uncertain", "preserve_without_resolution"}}
	for _, nilArrays := range []bool{false, true} {
		if nilArrays {
			f.Ambiguities = nil
		}
		encoded, err := json.Marshal(f)
		if err != nil {
			t.Fatal(err)
		}
		size, ok := typedFrameSize(f)
		if !ok || size != len(encoded) {
			t.Fatalf("counter %d/%v actual %d", size, ok, len(encoded))
		}
	}
}
