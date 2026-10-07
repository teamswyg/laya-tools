package claimreferencepilot

const (
	AIInvocationCeiling         = 49
	MetadataInvocationCeiling   = 17
	ContentInvocationCeiling    = 25
	ControllerInvocationCeiling = 7
	AcquisitionByteCeiling      = 8 << 20
	ReviewByteCeiling           = 16 << 20
)

// RoleContentInvocationCeilings returns the declared finite role ceilings in
// fixed Role order. These ceilings are metadata, not permission to invoke roles.
func RoleContentInvocationCeilings() [RoleCount]uint32 {
	return [RoleCount]uint32{2, 2, 2, 4, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 1, 2}
}

type PlannedInventory struct {
	Sources, Families, Inputs, Panels, RatersPerLocale, Records, Judgments uint32
	StratumFamilies                                                        [StratumCount]uint32
	Heads                                                                  [HeadCount]Head
	States                                                                 [StateCount]State
}

type PlannedScope struct {
	MaximumUTF8Bytes                                                                             uint32
	ContextUnits                                                                                 uint8
	CompleteComment, TextOnly, ValidNonemptyUTF8, NoTruncation, SameEvidenceForAllHeadsAndStates bool
	HiddenSourceWorld, HiddenPartnerLocale                                                       bool
}

// InvocationBudget includes every failed/blocked attempt in its counters.
// Recorder/runtime-helper verification remains outside the actual pilot ledger;
// it cannot be used to invent role execution receipts or to refill this budget.
type InvocationBudget struct {
	TotalLimit, MetadataLimit, ContentLimit, ControllerLimit uint32
	RoleContentLimits                                        [RoleCount]uint32
	MetadataAttempts, ControllerAttempts                     uint32
	RoleContentAttempts                                      [RoleCount]uint32
	AllAttemptsRetained                                      bool
}

// Usage carries the inclusive final ledger. Limits remain the admitted finite
// ceilings; originals, failed attempts and all metadata bytes count toward use.
type Usage struct {
	Budget                                  InvocationBudget
	AcquisitionBytesLimit, ReviewBytesLimit uint64
	AcquisitionBytesUsed, ReviewBytesUsed   uint64
}

// ValidateUsage checks all role-specific and total invocation ceilings and the
// inclusive byte caps. It can validate either pre-word or later usage without
// claiming that a recorded attempt actually occurred.
func ValidateUsage(u Usage) Readiness {
	r := Readiness{}
	b := u.Budget
	if b.TotalLimit != AIInvocationCeiling || b.MetadataLimit != MetadataInvocationCeiling || b.ContentLimit != ContentInvocationCeiling || b.ControllerLimit != ControllerInvocationCeiling || b.RoleContentLimits != RoleContentInvocationCeilings() || !b.AllAttemptsRetained {
		r.Failures.Bindings++
	}
	var content uint64
	for i, attempts := range b.RoleContentAttempts {
		content += uint64(attempts)
		if attempts > b.RoleContentLimits[i] {
			r.Failures.Bindings++
		}
	}
	if b.MetadataAttempts > b.MetadataLimit || b.ControllerAttempts > b.ControllerLimit || content > uint64(b.ContentLimit) || uint64(b.MetadataAttempts)+uint64(b.ControllerAttempts)+content > uint64(b.TotalLimit) {
		r.Failures.Bindings++
	}
	if u.AcquisitionBytesLimit != AcquisitionByteCeiling || u.ReviewBytesLimit != ReviewByteCeiling || u.AcquisitionBytesUsed > u.AcquisitionBytesLimit || u.ReviewBytesUsed > u.ReviewBytesLimit {
		r.Failures.Bindings++
	}
	r.Ready = r.Failures.Bindings == 0
	return r
}

// Plan is usable before any protected pilot IDs, input digests, words or labels
// exist. This pre-word check deliberately does not call completed-pilot Evaluate.
type Plan struct {
	Bindings                                Bindings
	Roles                                   [RoleCount]RoleBinding
	InitialDefinition                       Digest
	InitialDefinitionLocked                 Stamp
	Inventory                               PlannedInventory
	Scope                                   PlannedScope
	Budget                                  InvocationBudget
	AcquisitionBytesLimit, ReviewBytesLimit uint64
	AcquisitionBytesUsed, ReviewBytesUsed   uint64
	WordsAlreadyCreated                     bool
	ProhibitedOperations                    uint64
}

// Readiness is an aggregate structural preflight, never pilot PASS or authority.
type Readiness struct {
	Ready                        bool
	Failures                     Failures
	QualifiedModel               bool `json:"qualified_model"`
	HumanGold                    bool `json:"human_gold"`
	IndependentModelsEstablished bool `json:"independent_models_established"`
}

// ValidateBindings checks existing receipt/role identity declarations without
// requiring future source documents, inputs or references. Caller-supplied tool
// receipts remain assertions; only the actual broker can authenticate them.
// ExecutionAdmission is not required here: the distinct later S2 admission
// follows preflight and is required by completed-pilot Evaluate.
func ValidateBindings(bindings Bindings, roles [RoleCount]RoleBinding, initialDefinitionLocked Stamp) Readiness {
	b := Batch{Bindings: bindings, Roles: roles, Lifecycle: Lifecycle{InitialDefinitionLocked: initialDefinitionLocked}}
	a := Aggregate{}
	checkBindings(&a, b, false)
	checkRoles(&a, b)
	if !before(initialDefinitionLocked, bindings.Locked) {
		a.Failures.Chronology++
	}
	return Readiness{Ready: a.Failures.Bindings+a.Failures.Roles+a.Failures.Chronology == 0, Failures: a.Failures}
}

// ValidatePlan checks the fixed, admitted pre-word plan. It cannot turn an empty
// pilot into completed PASS, issue a CAP, or authorize content creation.
func ValidatePlan(p Plan) Readiness {
	r := ValidateBindings(p.Bindings, p.Roles, p.InitialDefinitionLocked)
	x := p.Inventory
	if x.Sources != FamilyCount || x.Families != FamilyCount || x.Inputs != InputCount || x.Panels != PanelCount || x.RatersPerLocale != RaterCount || x.Records != RecordCount || x.Judgments != JudgmentCount {
		r.Failures.Inventory++
	}
	for _, count := range x.StratumFamilies {
		if count != FamiliesPerStratum {
			r.Failures.Inventory++
		}
	}
	for h, head := range x.Heads {
		if head != Head(h+1) {
			r.Failures.Inventory++
		}
	}
	for s, state := range x.States {
		if state != State(s+1) {
			r.Failures.Inventory++
		}
	}
	scope := p.Scope
	if scope.MaximumUTF8Bytes != 4096 || scope.ContextUnits != 0 || !scope.CompleteComment || !scope.TextOnly || !scope.ValidNonemptyUTF8 || !scope.NoTruncation || !scope.SameEvidenceForAllHeadsAndStates || !scope.HiddenSourceWorld || !scope.HiddenPartnerLocale {
		r.Failures.Scopes++
	}
	usage := ValidateUsage(Usage{Budget: p.Budget, AcquisitionBytesLimit: p.AcquisitionBytesLimit, ReviewBytesLimit: p.ReviewBytesLimit, AcquisitionBytesUsed: p.AcquisitionBytesUsed, ReviewBytesUsed: p.ReviewBytesUsed})
	r.Failures.Bindings += usage.Failures.Bindings
	for _, count := range p.Budget.RoleContentAttempts {
		if count != 0 {
			r.Failures.Bindings++
		}
	}
	if !nonzero(p.InitialDefinition) || p.WordsAlreadyCreated || p.ProhibitedOperations != 0 {
		r.Failures.Bindings++
	}
	r.Ready = r.Ready && r.Failures.Inventory+r.Failures.Scopes+r.Failures.Bindings == 0
	return r
}
