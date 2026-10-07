package claimreferencepilot

import (
	"testing"
	"time"
)

func TestIdentityRequiresObservedDistinctContextAndHonestAI(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Batch)
	}{
		{"missing source binding", func(b *Batch) { b.Bindings.CharterSource = Digest{} }},
		{"wrong schema", func(b *Batch) { b.Bindings.SchemaSource = digest(1) }},
		{"missing execution admission", func(b *Batch) { b.Bindings.ExecutionAdmission = Digest{} }},
		{"missing controller", func(b *Batch) { b.Bindings.Controller.Receipt = Digest{} }},
		{"duplicate role id", func(b *Batch) { b.Roles[1].RoleID = b.Roles[0].RoleID }},
		{"duplicate canonical agent", func(b *Batch) {
			b.Roles[1].AgentID = b.Roles[0].AgentID
			b.Roles[1].ObservedAgentID = b.Roles[0].ObservedAgentID
		}},
		{"duplicate context", func(b *Batch) {
			x := b.Roles[0].ContextID
			b.Roles[1].ContextID = x
			b.Roles[1].ObservedContextID = x
			b.Roles[1].ContextIDs = []string{x}
		}},
		{"overlapping context", func(b *Batch) {
			x := b.Roles[0].ContextID + "/nested"
			b.Roles[1].ContextID = x
			b.Roles[1].ObservedContextID = x
			b.Roles[1].ContextIDs = []string{x}
		}},
		{"copied context exposure", func(b *Batch) { b.Roles[1].ContextIDs = append(b.Roles[1].ContextIDs, b.Roles[0].ContextID) }},
		{"canonical observation mismatch", func(b *Batch) { b.Roles[0].AgentID = "unobserved-agent" }},
		{"missing observation receipt", func(b *Batch) { b.Roles[0].ToolObservation = Digest{} }},
		{"inherited fork", func(b *Batch) { b.Roles[0].ForkTurnsNone = false }},
		{"undeclared AI", func(b *Batch) { b.Roles[0].DeclaredAI = false }},
		{"wrong position", func(b *Batch) { b.Roles[0].Position = Comparer }},
		{"wrong rater kind", func(b *Batch) { b.Roles[4].Kind = DeclaredAIAuthorOrReviewer }},
		{"unavailable role", func(b *Batch) { b.Roles[0].Available = false }},
		{"forbidden exposure", func(b *Batch) { b.Roles[0].ForbiddenExposure = true }},
		{"hidden identity not explicitly unknown", func(b *Batch) { b.Roles[0].UnderlyingModelUnknown = false }},
		{"fabricated hidden identity", func(b *Batch) { b.Roles[0].UnderlyingModelIdentity = "unobserved-hidden-model" }},
		{"independent-model assertion", func(b *Batch) { b.Roles[0].UnderlyingModelIndependenceClaimed = true }},
		{"unknown controller not declared", func(b *Batch) { b.Bindings.Controller.UnderlyingModelUnknown = false }},
		{"controller independence assertion", func(b *Batch) { b.Bindings.Controller.UnderlyingModelIndependenceClaimed = true }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) { b := abstractBatch(); test.mutate(&b); wantVerdict(t, b, Blocked) })
	}
}

func TestFixedInventoryScopeAndScreenJoins(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Batch)
	}{
		{"reused family", func(b *Batch) { b.Families[1].Key = b.Families[0].Key }},
		{"unbalanced strata", func(b *Batch) { b.Families[0].Stratum = 2 }},
		{"zero stratum", func(b *Batch) { b.Families[0].Stratum = 0 }},
		{"missing family", func(b *Batch) { b.Families[0].Present = false }},
		{"reused source receipt", func(b *Batch) { b.Families[1].Source.Receipt = b.Families[0].Source.Receipt }},
		{"missing creation", func(b *Batch) { b.Families[0].Source.CreationReceipt = Digest{} }},
		{"oversize source", func(b *Batch) { b.Families[0].Source.Bytes = 16385 }},
		{"missing selected input", func(b *Batch) { b.Families[0].Inputs[0].Present = false }},
		{"duplicate scope receipt", func(b *Batch) { b.Families[0].Inputs[1].Receipt = b.Families[0].Inputs[0].Receipt }},
		{"empty bytes", func(b *Batch) { b.Families[0].Inputs[0].Bytes = 0 }},
		{"oversize bytes", func(b *Batch) { b.Families[0].Inputs[0].Bytes = 4097 }},
		{"impossible utf8 ratio", func(b *Batch) { b.Families[0].Inputs[0].Codepoints = 1 }},
		{"invalid UTF8 attestation", func(b *Batch) { b.Families[0].Inputs[0].ValidNonemptyUTF8 = false }},
		{"truncation", func(b *Batch) { b.Families[0].Inputs[0].NoTruncation = false }},
		{"context added", func(b *Batch) { b.Families[0].Inputs[0].ContextUnits = 1 }},
		{"model/reference evidence differs", func(b *Batch) { b.Families[0].Inputs[0].ModelEvidenceDigest = digest(2) }},
		{"panel input inventory changed", func(b *Batch) { b.PanelInputsDigest[1] = digest(2) }},
		{"wrong selected revision", func(b *Batch) { b.Families[0].Inputs[0].EditorialRevisions = 1 }},
		{"two editorial revisions", func(b *Batch) {
			b.Families[0].Inputs[0].EditorialRevisions = 2
			b.Families[0].Inputs[0].SelectedVersion = 2
		}},
		{"original variants dropped", func(b *Batch) { b.Families[0].Source.VersionsRetained = false }},
		{"wrong fidelity source join", func(b *Batch) { b.Families[0].Fidelity.SourceReceipt = b.Families[1].Source.Receipt }},
		{"wrong frame input join", func(b *Batch) { b.Families[0].Frame.InputDigests[0] = b.Families[1].Inputs[0].CommentDigest }},
		{"wrong naturalness actor", func(b *Batch) { b.Families[0].Naturalness[0].Reviewer = ENNaturalnessReviewer }},
		{"naturalness sees partner", func(b *Batch) { b.Families[0].Naturalness[0].InputDigests[1] = b.Families[0].Inputs[1].CommentDigest }},
		{"frame reference values visible", func(b *Batch) { b.Families[0].Frame.ReferencesHidden = false }},
		{"screen receipt reused", func(b *Batch) { b.Families[1].Frame.Receipt = b.Families[0].Frame.Receipt }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) { b := abstractBatch(); test.mutate(&b); wantVerdict(t, b, Blocked) })
	}
}

func TestPhaseLocksAndUTCOrdinalTies(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Batch)
	}{
		{"missing phase lock", func(b *Batch) { b.Lifecycle.PanelLocks[0] = Stamp{} }},
		{"source before binding", func(b *Batch) { b.Families[0].Source.Created = stamp(29) }},
		{"save before creation", func(b *Batch) { b.Families[0].Source.Saved = stamp(39) }},
		{"selected input after freeze", func(b *Batch) { b.Families[0].Inputs[0].Locked = stamp(81) }},
		{"fidelity before input", func(b *Batch) { b.Families[0].Fidelity.Locked = stamp(59) }},
		{"naturalness before selected locale", func(b *Batch) { b.Families[0].Naturalness[1].Locked = stamp(59) }},
		{"frame before wording freeze", func(b *Batch) { b.Families[0].Frame.Locked = stamp(79) }},
		{"reference before frame lock", func(b *Batch) { b.Records[0].Started = stamp(99) }},
		{"reference beyond panel lock", func(b *Batch) { b.Records[0].Locked = stamp(131) }},
		{"decision before all first locks", func(b *Batch) { b.Lifecycle.Decision.Locked = stamp(129) }},
		{"repeat before decision", func(b *Batch) { row(b, 1, 0, 0, 0).Started = stamp(139) }},
		{"frame release before repeat lock", func(b *Batch) { b.Lifecycle.FrameValuesReleased = stamp(169) }},
		{"review before frame release", func(b *Batch) { b.Lifecycle.FinalReviewLocked = stamp(179) }},
		{"identical first start and lock", func(b *Batch) { b.Records[0].Started = b.Records[0].Locked }},
		{"zero ordinal", func(b *Batch) { b.Records[0].Started.Ordinal = 0 }},
		{"non UTC stamp", func(b *Batch) { b.Records[0].Started.UTC = b.Records[0].Started.UTC.In(time.FixedZone("offset", 3600)) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) { b := abstractBatch(); test.mutate(&b); wantVerdict(t, b, Blocked) })
	}
	base := stamp(1).UTC
	if !before(Stamp{UTC: base, Ordinal: 1}, Stamp{UTC: base, Ordinal: 2}) || before(Stamp{UTC: base, Ordinal: 2}, Stamp{UTC: base, Ordinal: 1}) || before(Stamp{UTC: base, Ordinal: 1}, Stamp{UTC: base, Ordinal: 1}) {
		t.Fatal("timestamp ties not resolved strictly by ordinal")
	}
	if !before(Stamp{UTC: base, Ordinal: 99}, Stamp{UTC: base.Add(time.Second), Ordinal: 1}) {
		t.Fatal("ordinal incorrectly superseded distinct timestamps")
	}
	if !atOrBefore(Stamp{UTC: base, Ordinal: 1}, Stamp{UTC: base.In(time.FixedZone("zero", 0)), Ordinal: 1}) {
		t.Fatal("same UTC instant depends on timezone object identity")
	}
}

func TestSoleDefinitionRevisionCannotChangeWordsOrFirstVotes(t *testing.T) {
	b := abstractBatch()
	d := &b.Lifecycle.Decision
	d.Revisions = 1
	d.Repeat = digest(9800)
	d.DefectEvidence = digest(9801)
	d.BilingualRepairEvidence = digest(9802)
	d.AlternativeReadingsEvidence = digest(9803)
	for i := range b.Records {
		if b.Records[i].Key.Panel == FreshRepeat {
			b.Records[i].Definition = d.Repeat
		}
	}
	wantVerdict(t, b, Pass)
	tests := []struct {
		name   string
		mutate func(*Batch)
	}{
		{"two revisions", func(b *Batch) { b.Lifecycle.Decision.Revisions = 2 }},
		{"revision has no clause evidence", func(b *Batch) { b.Lifecycle.Decision.DefectEvidence = Digest{} }},
		{"revision uses hidden frames", func(b *Batch) { b.Lifecycle.Decision.HiddenFrameValuesUsed = true }},
		{"scope or ontology changed", func(b *Batch) { b.Lifecycle.Decision.NoStructuralChanges = false }},
		{"first dissent discarded", func(b *Batch) { b.Lifecycle.Decision.FirstDissentPreserved = false }},
		{"wrong repeat definition", func(b *Batch) { row(b, 1, 0, 0, 0).Definition = digest(12) }},
		{"unchanged decision changes definition", func(b *Batch) { b.Lifecycle.Decision.Revisions = 0 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testBatch := b
			testBatch.Records = append([]ReferenceRecord(nil), b.Records...)
			test.mutate(&testBatch)
			wantVerdict(t, testBatch, Blocked)
		})
	}
}

func TestDissentAndDistributionResolveImmutableHeadLocalePanel(t *testing.T) {
	b := abstractBatch()
	record := b.Records[0]
	b.Audit.Dissent = []Dissent{{Reference: record.Key, ReferenceReceipt: record.Receipt, Definition: record.Definition, ReasonReceipt: digest(9700), Head: ResponseRequested, FirstState: True, PlausibleStates: []State{True, Unknown}}}
	b.Audit.Distributions = []Distribution{{Panel: First, Locale: KO, Family: 0, Head: ResponseRequested, Definition: record.Definition, Members: []RecordKey{record.Key, row(&b, 0, 0, 1, 0).Key}, Counts: [3]int{2, 0, 0}, DependenceLimitReceipt: digest(9701)}}
	a := wantVerdict(t, b, Pass)
	if a.FirstAgreedTrue[0][0] != 16 {
		t.Fatal("dissent changed semantic votes")
	}
	tests := []struct {
		name   string
		mutate func(*Batch)
	}{
		{"dissent first state changed", func(b *Batch) { b.Audit.Dissent[0].FirstState = Unknown }},
		{"dissent locale mismatch", func(b *Batch) { b.Audit.Dissent[0].Reference.Locale = EN }},
		{"dissent panel mismatch", func(b *Batch) { b.Audit.Dissent[0].Reference.Panel = FreshRepeat }},
		{"dissent version mismatch", func(b *Batch) { b.Audit.Dissent[0].Definition = digest(1) }},
		{"missing dissent head", func(b *Batch) { b.Audit.Dissent[0].Head = 0 }},
		{"wrong distribution total", func(b *Batch) { b.Audit.Distributions[0].Counts = [3]int{1, 1, 1} }},
		{"duplicate member", func(b *Batch) { b.Audit.Distributions[0].Members = []RecordKey{record.Key, record.Key} }},
		{"missing member", func(b *Batch) { b.Audit.Distributions[0].Members = []RecordKey{record.Key} }},
		{"cross locale member", func(b *Batch) { b.Audit.Distributions[0].Members[0].Locale = EN }},
		{"duplicate distribution cell", func(b *Batch) { b.Audit.Distributions = append(b.Audit.Distributions, b.Audit.Distributions[0]) }},
		{"majority relabel", func(b *Batch) { b.Audit.MajorityAutomaticRelabel = true }},
		{"confidence replaced first state", func(b *Batch) { b.Audit.ConfidenceSubstituted = true }},
		{"unresolved dropped", func(b *Batch) { b.Audit.UnresolvedCasesDropped = true }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			x := abstractBatch()
			x.Audit.Dissent = append([]Dissent(nil), b.Audit.Dissent...)
			x.Audit.Distributions = append([]Distribution(nil), b.Audit.Distributions...)
			x.Audit.Distributions[0].Members = append([]RecordKey(nil), b.Audit.Distributions[0].Members...)
			test.mutate(&x)
			wantVerdict(t, x, Blocked)
		})
	}
}

func abstractPlan() Plan {
	b := abstractBatch()
	p := Plan{Bindings: b.Bindings, Roles: b.Roles, InitialDefinition: b.Lifecycle.Decision.Initial, InitialDefinitionLocked: b.Lifecycle.InitialDefinitionLocked,
		Inventory:             PlannedInventory{Sources: 48, Families: 48, Inputs: 96, Panels: 2, RatersPerLocale: 2, Records: 384, Judgments: 1152, Heads: [3]Head{ResponseRequested, CurrentActivityClaimed, CompletionClaimed}, States: [3]State{True, False, Unknown}},
		Scope:                 PlannedScope{MaximumUTF8Bytes: 4096, CompleteComment: true, TextOnly: true, ValidNonemptyUTF8: true, NoTruncation: true, SameEvidenceForAllHeadsAndStates: true, HiddenSourceWorld: true, HiddenPartnerLocale: true},
		Budget:                InvocationBudget{TotalLimit: 49, MetadataLimit: 17, ContentLimit: 25, ControllerLimit: 7, RoleContentLimits: RoleContentInvocationCeilings(), MetadataAttempts: 17, AllAttemptsRetained: true},
		AcquisitionBytesLimit: 8 << 20, ReviewBytesLimit: 16 << 20}
	p.Bindings.ExecutionAdmission = Digest{} // S2 is issued only after this preflight.
	for k := range StratumCount {
		p.Inventory.StratumFamilies[k] = 6
	}
	return p
}

func TestPreWordBindingsAndPlanAreSeparateFromCompletedPilot(t *testing.T) {
	p := abstractPlan()
	readiness := ValidatePlan(p)
	if !readiness.Ready || readiness.QualifiedModel || readiness.HumanGold || readiness.IndependentModelsEstablished {
		t.Fatalf("valid abstract plan blocked or qualified: %+v", readiness)
	}
	bindings := ValidateBindings(p.Bindings, p.Roles, p.InitialDefinitionLocked)
	if !bindings.Ready {
		t.Fatal("preword bindings require nonexistent words")
	}
	completed := abstractBatch()
	completed.Bindings.ExecutionAdmission = Digest{}
	wantVerdict(t, completed, Blocked)
	empty := Batch{Bindings: p.Bindings, Roles: p.Roles, Lifecycle: Lifecycle{InitialDefinitionLocked: p.InitialDefinitionLocked}}
	a := wantVerdict(t, empty, Blocked)
	if a.Completeness.AssessedRecords != 0 || a.Completeness.MissingRecordSlots != 384 {
		t.Fatal("preword readiness became empty-pilot PASS")
	}
	tests := []struct {
		name   string
		mutate func(*Plan)
	}{
		{"wrong inventory", func(p *Plan) { p.Inventory.Families = 47 }},
		{"wrong stratum size", func(p *Plan) { p.Inventory.StratumFamilies[0] = 5 }},
		{"state zero", func(p *Plan) { p.Inventory.States[0] = 0 }},
		{"swapped heads", func(p *Plan) { p.Inventory.Heads[0], p.Inventory.Heads[1] = p.Inventory.Heads[1], p.Inventory.Heads[0] }},
		{"missing broker admission", func(p *Plan) { p.Bindings.BrokerAdmission = Digest{} }},
		{"words already created", func(p *Plan) { p.WordsAlreadyCreated = true }},
		{"prohibited operation", func(p *Plan) { p.ProhibitedOperations = 1 }},
		{"hidden content attempt", func(p *Plan) { p.Budget.RoleContentAttempts[0] = 1 }},
		{"metadata overflow", func(p *Plan) { p.Budget.MetadataAttempts = 18 }},
		{"controller overflow", func(p *Plan) { p.Budget.ControllerAttempts = 8 }},
		{"extra invocation allowance", func(p *Plan) { p.Budget.TotalLimit = 50 }},
		{"role retry allowance", func(p *Plan) { p.Budget.RoleContentLimits[4] = 2 }},
		{"attempts discarded", func(p *Plan) { p.Budget.AllAttemptsRetained = false }},
		{"cap expanded", func(p *Plan) { p.AcquisitionBytesLimit++ }},
		{"cap overflow", func(p *Plan) { p.ReviewBytesUsed = p.ReviewBytesLimit + 1 }},
		{"context added", func(p *Plan) { p.Scope.ContextUnits = 1 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := abstractPlan()
			test.mutate(&p)
			if ValidatePlan(p).Ready {
				t.Fatal("invalid plan became ready")
			}
		})
	}
}

func TestActualInclusiveUsageBounds(t *testing.T) {
	b := abstractBatch()
	if !ValidateUsage(b.Usage).Ready {
		t.Fatal("legal role maxima at49 attempts and exact byte caps rejected")
	}
	wantVerdict(t, b, Pass)
	tests := []struct {
		name   string
		mutate func(*Usage)
	}{
		{"50 total attempts", func(u *Usage) { u.Budget.MetadataAttempts = 18 }},
		{"rater second invocation under total ceiling", func(u *Usage) { u.Budget.MetadataAttempts = 15; u.Budget.RoleContentAttempts[4] = 2 }},
		{"source third invocation", func(u *Usage) { u.Budget.RoleContentAttempts[0] = 3 }},
		{"fidelity fifth invocation", func(u *Usage) { u.Budget.RoleContentAttempts[3] = 5 }},
		{"acquisition cap plus one", func(u *Usage) { u.AcquisitionBytesUsed++ }},
		{"review cap plus one", func(u *Usage) { u.ReviewBytesUsed++ }},
		{"failed attempts dropped", func(u *Usage) { u.Budget.AllAttemptsRetained = false }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := abstractBatch()
			test.mutate(&b.Usage)
			if ValidateUsage(b.Usage).Ready {
				t.Fatal("invalid actual usage passed")
			}
			wantVerdict(t, b, Blocked)
		})
	}
	b = abstractBatch()
	b.Usage.Budget.RoleContentAttempts[4] = 0
	wantVerdict(t, b, Blocked)
}
