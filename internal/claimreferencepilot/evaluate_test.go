package claimreferencepilot

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
)

// All fixtures are abstract metadata. They contain no pilot words, documents,
// real family identities, provider executions, model calls, or semantic claims.
func digest(n int) Digest { return Digest{byte(n), byte(n >> 8), byte(n >> 16), 0x7f} }
func stamp(n uint64) Stamp {
	return Stamp{UTC: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(n) * time.Second), Ordinal: n}
}

func abstractBatch() Batch {
	b := Batch{SelectedInputsDigest: digest(900), PanelInputsDigest: [2]Digest{digest(900), digest(900)}, Audit: Audit{FirstReferencesPreserved: true}}
	b.Usage = Usage{Budget: InvocationBudget{TotalLimit: 49, MetadataLimit: 17, ContentLimit: 25, ControllerLimit: 7, RoleContentLimits: RoleContentInvocationCeilings(), MetadataAttempts: 17, ControllerAttempts: 7, RoleContentAttempts: RoleContentInvocationCeilings(), AllAttemptsRetained: true}, AcquisitionBytesLimit: 8 << 20, ReviewBytesLimit: 16 << 20, AcquisitionBytesUsed: 8 << 20, ReviewBytesUsed: 16 << 20}
	b.Bindings = Bindings{
		SchemaSource: AcceptedSchemaSourceDigest(), SchemaAcceptance: AcceptedSchemaReviewDigest(),
		CharterSource: digest(1), CharterAcceptance: digest(2), MethodScope: digest(3), BrokerAdmission: digest(4), ExecutionAdmission: digest(5), AcquisitionCAP: digest(6), ReviewCAP: digest(7), Locked: stamp(30),
		Controller: ControllerBinding{Receipt: digest(8), SourceReceipt: digest(1), ToolObservation: digest(9), IsolationReceipt: digest(10), ExposureReceipt: digest(11), Provider: "abstract-provider", ExposedModelVersion: "abstract-version", Broker: "abstract-broker", RunID: "abstract-controller", UnderlyingModelUnknown: true, Registered: stamp(10)},
	}
	b.Lifecycle = Lifecycle{InitialDefinitionLocked: stamp(5), InputsLocked: stamp(80), FramesLocked: stamp(100), PanelLocks: [2]Stamp{stamp(130), stamp(170)}, FrameValuesReleased: stamp(180), FinalReviewLocked: stamp(190),
		Decision: DefinitionDecision{Initial: digest(12), Repeat: digest(12), Reviewer: Comparer, FirstDissentPreserved: true, NoStructuralChanges: true, Locked: stamp(140)},
	}
	for i := range RoleCount {
		kind := DeclaredAIAuthorOrReviewer
		if i >= 4 && i < 12 {
			kind = DeclaredAIReferenceNotHumanGold
		}
		id, context := fmt.Sprintf("abstract-agent-%d", i), fmt.Sprintf("abstract-context-%d", i)
		b.Roles[i] = RoleBinding{Present: true, Position: Role(i + 1), RoleID: fmt.Sprintf("abstract-role-%d", i), AgentID: id, ContextID: context, ObservedAgentID: id, ObservedContextID: context, ContextIDs: []string{context}, ToolObservation: digest(100 + i), SourceReceipt: digest(1), ControllerReceipt: digest(8), CapabilityReceipt: digest(200 + i), ExposureReceipt: digest(300 + i), InstructionReceipt: digest(400 + i), IsolationReceipt: digest(500 + i), Kind: kind, DeclaredAI: true, ForkTurnsNone: true, Available: true, Provider: "abstract-provider", ExposedModelVersion: "abstract-version", UnderlyingModelUnknown: true, Registered: stamp(20)}
	}
	for f := range FamilyCount {
		family := Family{Present: true, Key: fmt.Sprintf("abstract-family-%d", f), Stratum: uint8(f/6 + 1), Source: Source{Receipt: digest(1000 + f), CreationReceipt: digest(2000 + f), Bytes: 8, Created: stamp(40), Saved: stamp(50), VersionsRetained: true}}
		var inputs [2]Digest
		for l := range LocaleCount {
			inputs[l] = digest(4000 + f*2 + l)
			family.Inputs[l] = InputScope{Present: true, Receipt: digest(3000 + f*2 + l), CommentDigest: inputs[l], ModelEvidenceDigest: inputs[l], ReferenceEvidenceDigest: inputs[l], Bytes: 8, Codepoints: 8, ValidNonemptyUTF8: true, NoTruncation: true, VersionsRetained: true, Locked: stamp(60)}
		}
		family.Fidelity = Screen{Present: true, Reviewer: SourceFidelityChecker, SourceReceipt: family.Source.Receipt, InputDigests: inputs, Receipt: digest(5000 + f), Accepted: true, ReferencesHidden: true, Locked: stamp(70)}
		family.Frame = Screen{Present: true, Reviewer: BilingualFrameAuditor, SourceReceipt: family.Source.Receipt, InputDigests: inputs, Receipt: digest(6000 + f), Accepted: true, ReferencesHidden: true, Locked: stamp(90)}
		for l := range LocaleCount {
			family.Naturalness[l] = Screen{Present: true, Reviewer: KONaturalnessReviewer + Role(l), SourceReceipt: family.Source.Receipt, Receipt: digest(7000 + f*2 + l), Accepted: true, ReferencesHidden: true, Locked: stamp(70)}
			family.Naturalness[l].InputDigests[l] = inputs[l]
		}
		b.Families[f] = family
	}
	for p := range PanelCount {
		for l := range LocaleCount {
			for r := range RaterCount {
				for f := range FamilyCount {
					state := State(f/16 + 1)
					record := ReferenceRecord{Key: RecordKey{Panel: Panel(p + 1), Locale: Locale(l + 1), Rater: Rater(r + 1), Family: uint8(f)}, FamilyKey: b.Families[f].Key, RaterRole: RaterRole(Panel(p+1), Locale(l+1), Rater(r+1)), Kind: DeclaredAIReferenceNotHumanGold, Receipt: digest(8000 + len(b.Records)), Definition: digest(12), InputScopeReceipt: b.Families[f].Inputs[l].Receipt, InputDigest: b.Families[f].Inputs[l].CommentDigest, Status: Assessed, Started: stamp(uint64(110 + p*40)), Locked: stamp(uint64(120 + p*40)), PartnerLocaleHidden: true, ModelResultsHidden: true, DesiredStatesHidden: true, SourceAndStratumHidden: true, FrameValuesHidden: true}
					for h := range HeadCount {
						record.Judgments = append(record.Judgments, Judgment{Head: Head(h + 1), FirstState: state, EvidenceSpans: []Span{{0, 1}}, ReasonReceipt: digest(9000), RuleReceipt: digest(9001), Confidence: Low, PlausibleStates: []State{state}, FirstImmutable: true})
					}
					b.Records = append(b.Records, record)
				}
			}
		}
	}
	return b
}

func row(b *Batch, p, l, r, f int) *ReferenceRecord { return &b.Records[((p*2+l)*2+r)*48+f] }
func setState(b *Batch, p, l, r, f, h int, state State) {
	j := &row(b, p, l, r, f).Judgments[h]
	j.FirstState = state
	j.PlausibleStates = []State{state}
}
func firstCell(b *Batch, f int, state State) {
	for r := range RaterCount {
		setState(b, 0, 0, r, f, 0, state)
	}
}
func wantVerdict(t *testing.T, b Batch, verdict Verdict) Aggregate {
	t.Helper()
	got := Evaluate(b)
	if got.Verdict != verdict {
		t.Fatalf("verdict %s, want %s; failures=%+v completeness=%+v", got.Verdict, verdict, got.Failures, got.Completeness)
	}
	if got.QualifiedModel || got.HumanGold || got.IndependentModelsEstablished {
		t.Fatal("metadata guard asserted qualification")
	}
	return got
}

func TestExactSourcePins(t *testing.T) {
	for _, test := range []struct {
		got  Digest
		want string
	}{
		{AcceptedSchemaSourceDigest(), "68ac445812bfe4996d7bd85d1e1ae8b055e9212eb3472ba500e6afaa7b39d401"},
		{AcceptedSchemaReviewDigest(), "77c01388e795bcccdff27a4b6abac0c643f997d9345ea613037b5b1bf35e6e35"},
	} {
		if hex.EncodeToString(test.got[:]) != test.want {
			t.Errorf("wrong accepted source pin")
		}
	}
}

func TestCompleteAbstractMetadata(t *testing.T) {
	b := abstractBatch()
	before, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	a := wantVerdict(t, b, Pass)
	if !a.Completeness.Passed || a.Completeness.AssessedRecords != 384 || a.Completeness.UsableJudgments != 1152 {
		t.Fatalf("wrong fixed completeness %+v", a.Completeness)
	}
	for p := range PanelCount {
		for l := range LocaleCount {
			for h := range HeadCount {
				if a.ExactAgreement[p][l][h] != 48 || a.PositiveAgreement[p][l][h] != 48 {
					t.Fatal("wrong fixed agreement count")
				}
				for r := range RaterCount {
					if a.StateSupport[p][l][r][h] != [3]int{16, 16, 16} {
						t.Fatal("wrong per-rater marginal")
					}
				}
			}
		}
	}
	after, _ := json.Marshal(b)
	if !bytes.Equal(before, after) || !reflect.DeepEqual(a, Evaluate(b)) {
		t.Fatal("evaluation mutated input or was nondeterministic")
	}
	output, _ := json.Marshal(a)
	if bytes.Contains(output, []byte("abstract-")) || bytes.Contains(output, []byte("Receipt")) || bytes.Contains(output, []byte("Plausible")) {
		t.Fatal("case metadata escaped aggregate output")
	}
}

func TestSupportRequiresEachFirstRaterAndEachDistinctFamily(t *testing.T) {
	b := abstractBatch()
	for f := range FamilyCount {
		s := True
		if f >= 44 {
			s = False
		}
		if f >= 46 {
			s = Unknown
		}
		firstCell(&b, f, s)
	}
	a := wantVerdict(t, b, Pass)
	if a.StateSupport[0][0][0][0] != [3]int{44, 2, 2} {
		t.Fatal("support boundary not preserved")
	}
	setState(&b, 0, 0, 0, 46, 0, False)
	a = wantVerdict(t, b, Fail)
	if a.Failures.SupportCells != 1 || a.StateSupport[0][0][0][0][2] != 1 || a.StateSupport[0][0][1][0][2] != 2 {
		t.Fatal("pooled rater/repeat support hid a deficient first rater")
	}
}

func TestAgreedTrueBoundary(t *testing.T) {
	b := abstractBatch()
	for f := 8; f < 16; f++ {
		firstCell(&b, f, False)
	}
	a := wantVerdict(t, b, Pass)
	if a.FirstAgreedTrue[0][0] != 8 {
		t.Fatal("wrong agreed true count")
	}
	firstCell(&b, 7, False)
	a = wantVerdict(t, b, Fail)
	if a.FirstAgreedTrue[0][0] != 7 || a.Failures.AgreedTrueCells != 1 {
		t.Fatal("agreed true below boundary passed")
	}
}

func agreementBoundary() Batch {
	b := abstractBatch()
	for f := 0; f < 4; f++ {
		setState(&b, 0, 0, 1, f, 0, False)
	}
	for f := 16; f < 21; f++ {
		setState(&b, 0, 0, 1, f, 0, Unknown)
	}
	return b
}
func TestSeparateExactAndPositiveBoundaries(t *testing.T) {
	b := agreementBoundary()
	a := wantVerdict(t, b, Pass)
	if a.ExactAgreement[0][0][0] != 39 || a.PositiveAgreement[0][0][0] != 44 || a.Confusion[0][0][0][1][2] != 5 {
		t.Fatal("39/44 boundary or false/unknown confusion lost")
	}
	t.Run("38 exact with 44 positive", func(t *testing.T) {
		x := agreementBoundary()
		setState(&x, 0, 0, 1, 21, 0, Unknown)
		a := wantVerdict(t, x, Fail)
		if a.ExactAgreement[0][0][0] != 38 || a.PositiveAgreement[0][0][0] != 44 || a.Failures.ExactCells != 1 || a.Failures.PositiveCells != 0 {
			t.Fatal("nonpositive agreement replaced exact states")
		}
	})
	t.Run("39 exact with 43 positive", func(t *testing.T) {
		x := agreementBoundary()
		setState(&x, 0, 0, 1, 20, 0, False)
		setState(&x, 0, 0, 1, 4, 0, False)
		a := wantVerdict(t, x, Fail)
		if a.ExactAgreement[0][0][0] != 39 || a.PositiveAgreement[0][0][0] != 43 || a.Failures.ExactCells != 0 || a.Failures.PositiveCells != 1 {
			t.Fatal("positive boundary rounded or replaced")
		}
	})
	t.Run("repeat never erases first failure", func(t *testing.T) {
		x := agreementBoundary()
		setState(&x, 0, 0, 1, 21, 0, Unknown)
		a := wantVerdict(t, x, Fail)
		if a.ExactAgreement[1][0][0] != 48 || a.Failures.ExactCells != 1 {
			t.Fatal("repeat erased first panel")
		}
	})
}

func TestNaturalnessUsesBothLocalesAndAll48Fidelity(t *testing.T) {
	b := abstractBatch()
	for k := range StratumCount {
		b.Families[k*6+5].Naturalness[0].Accepted = false
	}
	a := wantVerdict(t, b, Pass)
	for _, n := range a.Naturalness {
		if n != 5 {
			t.Fatal("not five of six fixed pairs")
		}
	}
	b.Families[4].Naturalness[1].Accepted = false
	a = wantVerdict(t, b, Fail)
	if a.Naturalness[0] != 4 || a.Failures.NaturalnessCells != 1 || a.Completeness.NaturalnessReviews != 96 {
		t.Fatal("locale averaging or denominator shrink")
	}
	t.Run("one fidelity hard hold", func(t *testing.T) {
		x := abstractBatch()
		x.Families[0].Fidelity.HardHold = true
		a := wantVerdict(t, x, Blocked)
		if a.Failures.HardHolds != 1 {
			t.Fatal("all48 fidelity hold waived")
		}
	})
	t.Run("meaning invention", func(t *testing.T) {
		x := abstractBatch()
		x.Families[0].Naturalness[1].MaterialLossOrInvention = true
		wantVerdict(t, x, Blocked)
	})
	t.Run("44 frames are insufficient", func(t *testing.T) {
		x := abstractBatch()
		for f := 44; f < 48; f++ {
			x.Families[f].Frame.Present = false
		}
		a := wantVerdict(t, x, Blocked)
		if a.Completeness.FrameReviews != 44 {
			t.Fatal("missing frame denominator changed")
		}
	})
	t.Run("missing locale screen", func(t *testing.T) {
		x := abstractBatch()
		x.Families[0].Naturalness[0].Present = false
		a := wantVerdict(t, x, Blocked)
		if a.Completeness.NaturalnessReviews != 95 {
			t.Fatal("missing screen not retained")
		}
	})
}

func TestRetainedMissingBlockedDuplicateAndInvalidAttempts(t *testing.T) {
	t.Run("missing fixed slot", func(t *testing.T) {
		b := abstractBatch()
		b.Records = b.Records[1:]
		a := wantVerdict(t, b, Blocked)
		if a.Completeness.MissingRecordSlots != 1 || a.ExactAgreement[0][0][0] != 47 || a.Completeness.UsableJudgments != 1149 {
			t.Fatal("missing slot invented or denominator shrank")
		}
	})
	t.Run("duplicate destroys slot vote", func(t *testing.T) {
		b := abstractBatch()
		b.Records = append(b.Records, b.Records[0])
		a := wantVerdict(t, b, Blocked)
		if a.Completeness.RecordsRetained != 385 || a.Completeness.DuplicateAttempts != 1 || a.Completeness.UsableJudgments != 1149 || a.StateSupport[0][0][0][0][0] != 15 {
			t.Fatal("duplicate attempt counted as another family or discarded")
		}
	})
	t.Run("invalid input keeps null scope and no labels", func(t *testing.T) {
		b := abstractBatch()
		r := &b.Records[0]
		r.Status = InvalidOrMissingInput
		r.InputScopeReceipt = Digest{}
		r.InputDigest = Digest{}
		r.Judgments = nil
		r.BlockedReasonReceipt = digest(9900)
		a := wantVerdict(t, b, Blocked)
		if a.Completeness.BlockedRecords != 1 || a.Completeness.InvalidRecords != 0 || a.Completeness.JudgmentsRetained != 1149 {
			t.Fatal("valid blocked metadata fabricated semantic defaults")
		}
	})
	t.Run("unable requires valid scope", func(t *testing.T) {
		b := abstractBatch()
		r := &b.Records[0]
		r.Status = RaterUnable
		r.InputScopeReceipt = Digest{}
		r.InputDigest = Digest{}
		r.Judgments = nil
		r.BlockedReasonReceipt = digest(9900)
		a := wantVerdict(t, b, Blocked)
		if a.Completeness.InvalidRecords != 1 {
			t.Fatal("unable accepted null scope")
		}
	})
	t.Run("blocked cannot carry judgments", func(t *testing.T) {
		b := abstractBatch()
		b.Records[0].Status = RaterUnable
		b.Records[0].BlockedReasonReceipt = digest(9900)
		a := wantVerdict(t, b, Blocked)
		if a.Completeness.InvalidRecords != 1 {
			t.Fatal("blocked semantic votes accepted")
		}
	})
	t.Run("extra out of range retained", func(t *testing.T) {
		b := abstractBatch()
		r := b.Records[0]
		r.Key.Family = 48
		b.Records = append(b.Records, r)
		a := wantVerdict(t, b, Blocked)
		if a.Completeness.RecordsRetained != 385 || a.Completeness.InvalidRecords != 1 {
			t.Fatal("extra record sliced away")
		}
	})
}

func TestAllNonpositiveAllTrueAndUnableCannotPass(t *testing.T) {
	for _, test := range []struct {
		name            string
		state           State
		support, agreed int
	}{{"all true", True, 24, 0}, {"all false", False, 24, 6}, {"all unknown", Unknown, 24, 6}} {
		t.Run(test.name, func(t *testing.T) {
			b := abstractBatch()
			for p := range PanelCount {
				for l := range LocaleCount {
					for r := range RaterCount {
						for f := range FamilyCount {
							for h := range HeadCount {
								setState(&b, p, l, r, f, h, test.state)
							}
						}
					}
				}
			}
			a := wantVerdict(t, b, Fail)
			if a.Failures.SupportCells != test.support || a.Failures.AgreedTrueCells != test.agreed {
				t.Fatal("all-one-state shortcut")
			}
		})
	}
	b := abstractBatch()
	for i := range b.Records {
		r := &b.Records[i]
		r.Status = RaterUnable
		r.Judgments = nil
		r.BlockedReasonReceipt = digest(9900 + i)
	}
	a := wantVerdict(t, b, Blocked)
	if a.Completeness.BlockedRecords != 384 || a.Completeness.UsableJudgments != 0 || a.ExactAgreement[0][0][0] != 0 || a.PositiveAgreement[0][0][0] != 0 {
		t.Fatal("both unable became agreement")
	}
}

func TestRecordContractsAndZeroEnumsFailClosed(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Batch)
	}{
		{"zero state", func(b *Batch) { b.Records[0].Judgments[0].FirstState = 0 }},
		{"zero review status", func(b *Batch) { b.Records[0].Status = 0 }},
		{"zero head", func(b *Batch) { b.Records[0].Judgments[0].Head = 0 }},
		{"unordered heads", func(b *Batch) {
			b.Records[0].Judgments[0], b.Records[0].Judgments[1] = b.Records[0].Judgments[1], b.Records[0].Judgments[0]
		}},
		{"only two judgments", func(b *Batch) { b.Records[0].Judgments = b.Records[0].Judgments[:2] }},
		{"confidence substitution", func(b *Batch) { b.Records[0].Judgments[0].ConfidenceIsSemanticState = true }},
		{"overwritten first state", func(b *Batch) { b.Records[0].Judgments[0].FirstImmutable = false }},
		{"outside evidence span", func(b *Batch) { b.Records[0].Judgments[0].EvidenceSpans = []Span{{0, 9}} }},
		{"unknown requires input evidence", func(b *Batch) { setState(b, 0, 0, 0, 0, 0, Unknown); b.Records[0].Judgments[0].EvidenceSpans = nil }},
		{"duplicate plausible state", func(b *Batch) { b.Records[0].Judgments[0].PlausibleStates = []State{True, True} }},
		{"wrong family join", func(b *Batch) { b.Records[0].FamilyKey = b.Families[1].Key }},
		{"wrong locale scope join", func(b *Batch) { b.Records[0].InputScopeReceipt = b.Families[0].Inputs[1].Receipt }},
		{"wrong actor position", func(b *Batch) { b.Records[0].RaterRole = FirstENRaterOne }},
		{"unhidden partner", func(b *Batch) { b.Records[0].PartnerLocaleHidden = false }},
		{"human claim", func(b *Batch) { b.Records[0].HumanGoldOrOperationalAccuracyClaimed = true }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) { b := abstractBatch(); test.mutate(&b); wantVerdict(t, b, Blocked) })
	}
	t.Run("reused receipt removes both evidence slots", func(t *testing.T) {
		b := abstractBatch()
		b.Records[1].Receipt = b.Records[0].Receipt
		a := wantVerdict(t, b, Blocked)
		if a.Completeness.UsableJudgments != 1146 || a.StateSupport[0][0][0][0][0] != 14 {
			t.Fatal("reused evidence supplied distinct votes")
		}
	})
	t.Run("invalid scope supplies no semantic votes", func(t *testing.T) {
		b := abstractBatch()
		b.Families[0].Inputs[0].ValidNonemptyUTF8 = false
		a := wantVerdict(t, b, Blocked)
		if a.Completeness.UsableJudgments != 1140 || a.Completeness.InvalidRecords != 4 || a.ExactAgreement[0][0][0] != 47 {
			t.Fatal("invalid input remained assessed")
		}
	})
	var empty Batch
	if err := json.Unmarshal([]byte("{}"), &empty); err != nil {
		t.Fatal(err)
	}
	a := wantVerdict(t, empty, Blocked)
	if a.Completeness.UsableJudgments != 0 {
		t.Fatal("zero JSON/default enums fabricated votes")
	}
}
