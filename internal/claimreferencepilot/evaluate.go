package claimreferencepilot

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// AcceptedSchemaSourceDigest returns the accepted blueprint-v2 source pin. It
// identifies the source contract, not a future runtime/adoption authorization.
func AcceptedSchemaSourceDigest() Digest {
	return Digest{0x68, 0xac, 0x44, 0x58, 0x12, 0xbf, 0xe4, 0x99, 0x6d, 0x7b, 0xd8, 0x5d, 0x1e, 0x1a, 0xe8, 0xb0, 0x55, 0xe9, 0x21, 0x2e, 0xb3, 0x47, 0x2b, 0xa5, 0x00, 0xe6, 0xaf, 0xaa, 0x7b, 0x39, 0xd4, 0x01}
}

// AcceptedSchemaReviewDigest identifies its independently issued source review.
func AcceptedSchemaReviewDigest() Digest {
	return Digest{0x77, 0xc0, 0x13, 0x88, 0xe7, 0x95, 0xbc, 0xcc, 0xdf, 0xf2, 0x7a, 0x4b, 0x6a, 0xba, 0xc0, 0xc6, 0x43, 0xf9, 0x97, 0xd9, 0x34, 0x5e, 0xa6, 0x13, 0x03, 0x7b, 0x5b, 0x1b, 0xf3, 0x5e, 0x6e, 0x35}
}

type joinedRecord struct {
	occurrences int
	usable      bool
	index       int
	states      [HeadCount]State
}
type recordGrid [PanelCount][LocaleCount][RaterCount][FamilyCount]joinedRecord

// Evaluate returns only aggregate metadata counts. All submitted attempts are
// counted, and a duplicate makes its entire joined slot unusable. No vote comes
// from a missing, invalid, blocked, unordered or overwritten reference.
func Evaluate(b Batch) Aggregate {
	a := Aggregate{}
	a.Completeness.RecordsRetained = len(b.Records)
	if !nonzero(b.SelectedInputsDigest) || b.PanelInputsDigest[0] != b.SelectedInputsDigest || b.PanelInputsDigest[1] != b.SelectedInputsDigest {
		a.Failures.Scopes++
	}
	checkBindings(&a, b, true)
	usage := ValidateUsage(b.Usage)
	a.Failures.Bindings += usage.Failures.Bindings
	for _, attempts := range b.Usage.Budget.RoleContentAttempts {
		if attempts == 0 {
			a.Failures.Bindings++
		}
	}
	checkLifecycle(&a, b)
	checkRoles(&a, b)
	checkFamilies(&a, b)
	var grid recordGrid
	var receipts [RecordCount]Digest
	var receiptIndices [RecordCount]int
	receiptCount := 0
	for i, record := range b.Records {
		a.Completeness.JudgmentsRetained += len(record.Judgments)
		repeatedReceipt := false
		if nonzero(record.Receipt) {
			for j, seen := range receipts[:receiptCount] {
				if seen == record.Receipt {
					a.Failures.Records++
					repeatedReceipt = true
					old := b.Records[receiptIndices[j]]
					if p, l, r, f, ok := recordIndices(old.Key); ok {
						grid[p][l][r][f].usable = false
					}
					break
				}
			}
			if receiptCount < RecordCount {
				receipts[receiptCount] = record.Receipt
				receiptIndices[receiptCount] = i
				receiptCount++
			}
		}
		p, l, r, f, ok := recordIndices(record.Key)
		if !ok {
			a.Completeness.InvalidRecords++
			a.Failures.Records++
			continue
		}
		slot := &grid[p][l][r][f]
		slot.occurrences++
		valid := checkRecord(&a, b, record, p, l, f) && !repeatedReceipt
		if record.Status == InvalidOrMissingInput || record.Status == RaterUnable {
			a.Completeness.BlockedRecords++
		}
		if !valid {
			a.Completeness.InvalidRecords++
			a.Failures.Records++
		}
		if slot.occurrences == 1 {
			slot.index = i
			slot.usable = valid && record.Status == Assessed
			if slot.usable {
				for h := range HeadCount {
					slot.states[h] = record.Judgments[h].FirstState
				}
			}
		} else {
			slot.usable = false
			a.Completeness.DuplicateAttempts++
			a.Failures.Records++
		}
	}
	for p := range PanelCount {
		for l := range LocaleCount {
			for r := range RaterCount {
				for f := range FamilyCount {
					slot := grid[p][l][r][f]
					if slot.occurrences == 0 {
						a.Completeness.MissingRecordSlots++
						continue
					}
					a.Completeness.UniqueRecordSlots++
					if !slot.usable {
						continue
					}
					a.Completeness.AssessedRecords++
					a.Completeness.UsableJudgments += HeadCount
					for h, state := range slot.states {
						a.StateSupport[p][l][r][h][int(state)-1]++
					}
				}
			}
			for f := range FamilyCount {
				left, right := grid[p][l][0][f], grid[p][l][1][f]
				if !left.usable || !right.usable {
					continue
				}
				for h := range HeadCount {
					x, y := left.states[h], right.states[h]
					a.Confusion[p][l][h][int(x)-1][int(y)-1]++
					if x == y {
						a.ExactAgreement[p][l][h]++
					}
					if (x == True) == (y == True) {
						a.PositiveAgreement[p][l][h]++
					}
					if p == 0 && x == True && y == True {
						a.FirstAgreedTrue[l][h]++
					}
				}
			}
		}
	}
	if a.Completeness.MissingRecordSlots != 0 || a.Completeness.BlockedRecords != 0 || len(b.Records) != RecordCount {
		a.Failures.Records++
	}
	checkAudit(&a, b, &grid)
	checkFloors(&a)
	a.Completeness.Passed = a.Completeness.Families == FamilyCount && a.Completeness.Inputs == InputCount &&
		a.Completeness.FidelityReviews == FamilyCount && a.Completeness.FrameReviews == FamilyCount &&
		a.Completeness.NaturalnessReviews == InputCount && a.Completeness.RecordsRetained == RecordCount &&
		a.Completeness.AssessedRecords == RecordCount && a.Completeness.UsableJudgments == JudgmentCount &&
		a.Completeness.InvalidRecords == 0 && a.Completeness.DuplicateAttempts == 0
	fail := a.Failures
	switch {
	case fail.Bindings+fail.Roles+fail.Inventory+fail.Scopes+fail.Screens+fail.Chronology+fail.Revisions+fail.Records+fail.Audit+fail.HardHolds > 0 || !a.Completeness.Passed:
		a.Verdict = Blocked
	case fail.SupportCells+fail.AgreedTrueCells+fail.ExactCells+fail.PositiveCells+fail.NaturalnessCells > 0:
		a.Verdict = Fail
	default:
		a.Verdict = Pass
	}
	// Qualification fields intentionally remain false for every possible result.
	return a
}

func checkBindings(a *Aggregate, b Batch, requireExecutionAdmission bool) {
	x := b.Bindings
	if x.SchemaSource != AcceptedSchemaSourceDigest() || x.SchemaAcceptance != AcceptedSchemaReviewDigest() ||
		!allDigests(x.CharterSource, x.CharterAcceptance, x.MethodScope, x.BrokerAdmission, x.AcquisitionCAP, x.ReviewCAP) || (requireExecutionAdmission && !nonzero(x.ExecutionAdmission)) || b.ProhibitedOperations != 0 {
		a.Failures.Bindings++
	}
	c := x.Controller
	if !allDigests(c.Receipt, c.ToolObservation, c.IsolationReceipt, c.ExposureReceipt) || c.SourceReceipt != x.CharterSource ||
		!allIDs(c.Provider, c.ExposedModelVersion, c.Broker, c.RunID) ||
		!honestModel(c.UnderlyingModelUnknown, c.UnderlyingModelIdentity, c.UnderlyingModelObservation) || c.UnderlyingModelIndependenceClaimed {
		a.Failures.Bindings++
	}
	if !before(c.Registered, x.Locked) {
		a.Failures.Chronology++
	}
}

func checkLifecycle(a *Aggregate, b Batch) {
	l := b.Lifecycle
	if !before(l.InitialDefinitionLocked, b.Bindings.Locked) || !before(b.Bindings.Locked, l.InputsLocked) ||
		!before(l.InputsLocked, l.FramesLocked) || !before(l.FramesLocked, l.PanelLocks[0]) ||
		!before(l.PanelLocks[0], l.Decision.Locked) || !before(l.Decision.Locked, l.PanelLocks[1]) ||
		!before(l.PanelLocks[1], l.FrameValuesReleased) || !before(l.FrameValuesReleased, l.FinalReviewLocked) {
		a.Failures.Chronology++
	}
	d := l.Decision
	valid := nonzero(d.Initial) && nonzero(d.Repeat) && d.Revisions <= 1 && d.Reviewer == Comparer &&
		d.FirstDissentPreserved && d.NoStructuralChanges && !d.HiddenFrameValuesUsed
	if d.Revisions == 0 {
		valid = valid && d.Repeat == d.Initial && !nonzero(d.DefectEvidence) && !nonzero(d.BilingualRepairEvidence)
	} else {
		valid = valid && d.Repeat != d.Initial && allDigests(d.DefectEvidence, d.BilingualRepairEvidence, d.AlternativeReadingsEvidence)
	}
	if !valid {
		a.Failures.Revisions++
	}
}

func checkRoles(a *Aggregate, b Batch) {
	for i, role := range b.Roles {
		expected := Role(i + 1)
		kind := DeclaredAIAuthorOrReviewer
		if expected >= FirstKORaterOne && expected <= RepeatENRaterTwo {
			kind = DeclaredAIReferenceNotHumanGold
		}
		valid := role.Present && role.Position == expected && role.Kind == kind && role.DeclaredAI && role.ForkTurnsNone && role.Available && !role.ForbiddenExposure &&
			allIDs(role.RoleID, role.AgentID, role.ContextID, role.ObservedAgentID, role.ObservedContextID, role.Provider, role.ExposedModelVersion) &&
			role.AgentID == role.ObservedAgentID && role.ContextID == role.ObservedContextID &&
			allDigests(role.ToolObservation, role.CapabilityReceipt, role.ExposureReceipt, role.InstructionReceipt, role.IsolationReceipt) &&
			role.SourceReceipt == b.Bindings.CharterSource && role.ControllerReceipt == b.Bindings.Controller.Receipt &&
			role.Provider == b.Bindings.Controller.Provider && role.ExposedModelVersion == b.Bindings.Controller.ExposedModelVersion &&
			honestModel(role.UnderlyingModelUnknown, role.UnderlyingModelIdentity, role.UnderlyingModelObservation) && !role.UnderlyingModelIndependenceClaimed &&
			validContexts(role.ContextIDs, role.ObservedContextID)
		if !valid {
			a.Failures.Roles++
		}
		if !atOrBefore(role.Registered, b.Bindings.Locked) || !before(b.Lifecycle.InitialDefinitionLocked, role.Registered) {
			a.Failures.Chronology++
		}
		for j := 0; j < i; j++ {
			other := b.Roles[j]
			if sameID(role.RoleID, other.RoleID) || sameID(role.AgentID, other.AgentID) || sameID(role.ObservedAgentID, other.ObservedAgentID) || contextsOverlap(role.ContextIDs, other.ContextIDs) {
				a.Failures.Roles++
			}
		}
	}
}

func checkFamilies(a *Aggregate, b Batch) {
	var screenReceipts [FamilyCount*2 + InputCount]Digest
	screenCount := 0
	for f, family := range b.Families {
		if family.Present {
			a.Completeness.Families++
		}
		if !family.Present || !validID(family.Key) || family.Stratum < 1 || family.Stratum > StratumCount {
			a.Failures.Inventory++
		} else {
			a.StratumFamilies[family.Stratum-1]++
		}
		for j := 0; j < f; j++ {
			if sameID(family.Key, b.Families[j].Key) || (nonzero(family.Source.Receipt) && family.Source.Receipt == b.Families[j].Source.Receipt) {
				a.Failures.Inventory++
			}
		}
		source := family.Source
		if !allDigests(source.Receipt, source.CreationReceipt) || source.Bytes == 0 || source.Bytes > 16384 || !versionSelection(source.EditorialRevisions, source.SelectedVersion, source.VersionsRetained) {
			a.Failures.Inventory++
		}
		if !before(b.Bindings.Locked, source.Created) || !before(source.Created, source.Saved) {
			a.Failures.Chronology++
		}
		var digests [LocaleCount]Digest
		for l, input := range family.Inputs {
			if input.Present {
				a.Completeness.Inputs++
			}
			digests[l] = input.CommentDigest
			if !validInputScope(input) {
				a.Failures.Scopes++
			}
			if !before(source.Saved, input.Locked) || !atOrBefore(input.Locked, b.Lifecycle.InputsLocked) {
				a.Failures.Chronology++
			}
			for oldF := 0; oldF <= f; oldF++ {
				for oldL := range LocaleCount {
					if oldF == f && oldL >= l {
						continue
					}
					if nonzero(input.Receipt) && input.Receipt == b.Families[oldF].Inputs[oldL].Receipt {
						a.Failures.Scopes++
					}
				}
			}
		}
		if family.Fidelity.Present {
			a.Completeness.FidelityReviews++
		}
		if family.Frame.Present {
			a.Completeness.FrameReviews++
		}
		for _, screen := range [4]Screen{family.Fidelity, family.Frame, family.Naturalness[0], family.Naturalness[1]} {
			if nonzero(screen.Receipt) {
				for _, seen := range screenReceipts[:screenCount] {
					if seen == screen.Receipt {
						a.Failures.Screens++
						break
					}
				}
				screenReceipts[screenCount] = screen.Receipt
				screenCount++
			}
		}
		lastInputLock := family.Inputs[0].Locked
		if before(lastInputLock, family.Inputs[1].Locked) {
			lastInputLock = family.Inputs[1].Locked
		}
		fidelity := checkScreen(a, family.Fidelity, SourceFidelityChecker, source.Receipt, digests, -1, b.Lifecycle.InputsLocked, lastInputLock)
		frame := checkScreen(a, family.Frame, BilingualFrameAuditor, source.Receipt, digests, -1, b.Lifecycle.FramesLocked, b.Lifecycle.InputsLocked)
		if !family.Fidelity.Accepted || !family.Frame.Accepted {
			a.Failures.HardHolds++
		}
		natural := fidelity && frame
		for l, screen := range family.Naturalness {
			if screen.Present {
				a.Completeness.NaturalnessReviews++
			}
			role := KONaturalnessReviewer + Role(l)
			valid := checkScreen(a, screen, role, source.Receipt, digests, l, b.Lifecycle.InputsLocked, family.Inputs[l].Locked)
			natural = natural && valid && screen.Accepted
		}
		if natural && family.Stratum >= 1 && family.Stratum <= StratumCount {
			a.Naturalness[family.Stratum-1]++
		}
	}
	for _, count := range a.StratumFamilies {
		if count != FamiliesPerStratum {
			a.Failures.Inventory++
		}
	}
}

func checkScreen(a *Aggregate, screen Screen, role Role, source Digest, inputs [LocaleCount]Digest, locale int, deadline, after Stamp) bool {
	valid := screen.Present && screen.Reviewer == role && allDigests(screen.Receipt, screen.SourceReceipt) && screen.SourceReceipt == source && screen.ReferencesHidden
	if locale < 0 {
		valid = valid && screen.InputDigests == inputs
	} else {
		valid = valid && screen.InputDigests[locale] == inputs[locale] && !nonzero(screen.InputDigests[1-locale])
	}
	if !valid {
		a.Failures.Screens++
	}
	if !before(after, screen.Locked) || !atOrBefore(screen.Locked, deadline) {
		a.Failures.Chronology++
		valid = false
	}
	if screen.HardHold || screen.MaterialLossOrInvention {
		a.Failures.HardHolds++
		valid = false
	}
	return valid
}

func checkRecord(a *Aggregate, b Batch, record ReferenceRecord, p, l, f int) bool {
	family := b.Families[f]
	input := family.Inputs[l]
	definition := b.Lifecycle.Decision.Initial
	priorLock := b.Lifecycle.FramesLocked
	if p == 1 {
		definition = b.Lifecycle.Decision.Repeat
		priorLock = b.Lifecycle.Decision.Locked
	}
	valid := family.Present && family.Key == record.FamilyKey && nonzero(record.Receipt) && record.RaterRole == RaterRole(record.Key.Panel, record.Key.Locale, record.Key.Rater) &&
		record.Kind == DeclaredAIReferenceNotHumanGold && record.Definition == definition && nonzero(record.Definition) &&
		record.PartnerLocaleHidden && record.ModelResultsHidden && record.DesiredStatesHidden && record.SourceAndStratumHidden && record.FrameValuesHidden && !record.HumanGoldOrOperationalAccuracyClaimed
	if !before(priorLock, record.Started) || !before(record.Started, record.Locked) || !atOrBefore(record.Locked, b.Lifecycle.PanelLocks[p]) {
		a.Failures.Chronology++
		valid = false
	}
	if record.Status == Assessed || record.Status == RaterUnable || nonzero(record.InputScopeReceipt) {
		valid = valid && validInputScope(input) && nonzero(record.InputScopeReceipt) && record.InputScopeReceipt == input.Receipt && record.InputDigest == input.CommentDigest && nonzero(record.InputDigest)
	} else if nonzero(record.InputDigest) {
		valid = false
	}
	switch record.Status {
	case Assessed:
		valid = valid && len(record.Judgments) == HeadCount && !nonzero(record.BlockedReasonReceipt)
		if len(record.Judgments) == HeadCount {
			for h, judgment := range record.Judgments {
				valid = valid && validJudgment(judgment, Head(h+1), int(input.Bytes))
			}
		}
	case InvalidOrMissingInput, RaterUnable:
		valid = valid && len(record.Judgments) == 0 && nonzero(record.BlockedReasonReceipt)
	default:
		valid = false
	}
	return valid
}

func validJudgment(j Judgment, expected Head, bytes int) bool {
	if j.Head != expected || !validState(j.FirstState) || !allDigests(j.ReasonReceipt, j.RuleReceipt) || j.Confidence < High || j.Confidence > NotRecordedWithReason ||
		!j.FirstImmutable || j.ConfidenceIsSemanticState || !validStates(j.PlausibleStates, j.FirstState) {
		return false
	}
	if (j.FirstState == True || j.FirstState == Unknown) && len(j.EvidenceSpans) == 0 {
		return false
	}
	for _, span := range j.EvidenceSpans {
		if span.StartByte < 0 || span.EndByte <= span.StartByte || span.EndByte > bytes {
			return false
		}
	}
	return true
}

func validInputScope(input InputScope) bool {
	return input.Present && allDigests(input.Receipt, input.CommentDigest) && input.CommentDigest == input.ModelEvidenceDigest && input.CommentDigest == input.ReferenceEvidenceDigest &&
		input.Bytes > 0 && input.Bytes <= 4096 && input.Codepoints > 0 && input.Codepoints <= input.Bytes && uint64(input.Bytes) <= 4*uint64(input.Codepoints) && input.ValidNonemptyUTF8 && input.NoTruncation && input.ContextUnits == 0 &&
		versionSelection(input.EditorialRevisions, input.SelectedVersion, input.VersionsRetained)
}

func checkAudit(a *Aggregate, b Batch, grid *recordGrid) {
	if !b.Audit.FirstReferencesPreserved || b.Audit.MajorityAutomaticRelabel || b.Audit.FirstReferenceOverwritten || b.Audit.UnresolvedCasesDropped || b.Audit.ConfidenceSubstituted {
		a.Failures.Audit++
	}
	for _, dissent := range b.Audit.Dissent {
		p, l, r, f, ok := recordIndices(dissent.Reference)
		if !ok || dissent.Head < ResponseRequested || dissent.Head > CompletionClaimed {
			a.Failures.Audit++
			continue
		}
		slot := grid[p][l][r][f]
		if !slot.usable || slot.occurrences != 1 {
			a.Failures.Audit++
			continue
		}
		record := b.Records[slot.index]
		state := slot.states[int(dissent.Head)-1]
		if dissent.ReferenceReceipt != record.Receipt || dissent.Definition != record.Definition || dissent.FirstState != state || !nonzero(dissent.ReasonReceipt) || !validStates(dissent.PlausibleStates, state) {
			a.Failures.Audit++
		}
	}
	var seen [PanelCount][LocaleCount][FamilyCount][HeadCount]bool
	for _, distribution := range b.Audit.Distributions {
		key := RecordKey{Panel: distribution.Panel, Locale: distribution.Locale, Rater: RaterOne, Family: distribution.Family}
		p, l, _, f, ok := recordIndices(key)
		if !ok || distribution.Head < ResponseRequested || distribution.Head > CompletionClaimed {
			a.Failures.Audit++
			continue
		}
		h := int(distribution.Head) - 1
		if seen[p][l][f][h] {
			a.Failures.Audit++
		}
		seen[p][l][f][h] = true
		var expected [StateCount]int
		var members [RaterCount]bool
		expectedMembers := 0
		for r := range RaterCount {
			if grid[p][l][r][f].usable {
				expected[int(grid[p][l][r][f].states[h])-1]++
				expectedMembers++
			}
		}
		valid := expectedMembers > 0 && len(distribution.Members) == expectedMembers && distribution.Counts == expected && nonzero(distribution.DependenceLimitReceipt)
		for _, member := range distribution.Members {
			mp, ml, mr, mf, ok := recordIndices(member)
			if !ok || mp != p || ml != l || mf != f {
				valid = false
				continue
			}
			if members[mr] || !grid[p][l][mr][f].usable {
				valid = false
			}
			members[mr] = true
		}
		definition := b.Lifecycle.Decision.Initial
		if p == 1 {
			definition = b.Lifecycle.Decision.Repeat
		}
		if distribution.Definition != definition || !valid {
			a.Failures.Audit++
		}
	}
}

func checkFloors(a *Aggregate) {
	for l := range LocaleCount {
		for h := range HeadCount {
			for r := range RaterCount {
				for s := range StateCount {
					if a.StateSupport[0][l][r][h][s] < SupportFloor {
						a.Failures.SupportCells++
					}
				}
			}
			if a.FirstAgreedTrue[l][h] < AgreedTrueFloor {
				a.Failures.AgreedTrueCells++
			}
			for p := range PanelCount {
				if a.ExactAgreement[p][l][h] < ExactFloor {
					a.Failures.ExactCells++
				}
				if a.PositiveAgreement[p][l][h] < PositiveFloor {
					a.Failures.PositiveCells++
				}
			}
		}
	}
	for _, count := range a.Naturalness {
		if count < NaturalnessFloor {
			a.Failures.NaturalnessCells++
		}
	}
}

func recordIndices(key RecordKey) (p, l, r, f int, ok bool) {
	if key.Panel < First || key.Panel > FreshRepeat || key.Locale < KO || key.Locale > EN || key.Rater < RaterOne || key.Rater > RaterTwo || key.Family >= FamilyCount {
		return 0, 0, 0, 0, false
	}
	return int(key.Panel) - 1, int(key.Locale) - 1, int(key.Rater) - 1, int(key.Family), true
}
func nonzero(d Digest) bool { return d != Digest{} }
func allDigests(ds ...Digest) bool {
	for _, d := range ds {
		if !nonzero(d) {
			return false
		}
	}
	return true
}
func validState(s State) bool { return s >= True && s <= Unknown }
func validStates(states []State, required State) bool {
	if len(states) < 1 || len(states) > StateCount {
		return false
	}
	var seen [StateCount]bool
	for _, state := range states {
		if !validState(state) || seen[int(state)-1] {
			return false
		}
		seen[int(state)-1] = true
	}
	return validState(required) && seen[int(required)-1]
}
func versionSelection(revisions, selected uint8, retained bool) bool {
	return revisions <= 1 && selected == revisions && retained
}
func validID(s string) bool {
	return len(s) > 0 && len(s) <= 256 && utf8.ValidString(s) && strings.IndexFunc(s, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) < 0
}
func allIDs(ids ...string) bool {
	for _, id := range ids {
		if !validID(id) {
			return false
		}
	}
	return true
}
func sameID(a, b string) bool { return a != "" && a == b }
func honestModel(unknown bool, identity string, observation Digest) bool {
	if unknown {
		return identity == "" && !nonzero(observation)
	}
	return validID(identity) && nonzero(observation)
}
func validContexts(contexts []string, observed string) bool {
	if len(contexts) == 0 || len(contexts) > RoleCount {
		return false
	}
	includesObserved := false
	for i, context := range contexts {
		if !validID(context) || strings.HasSuffix(context, "/") {
			return false
		}
		includesObserved = includesObserved || context == observed
		for j := 0; j < i; j++ {
			if contextOverlap(context, contexts[j]) {
				return false
			}
		}
	}
	return includesObserved
}
func contextOverlap(a, b string) bool {
	return a != "" && b != "" && (a == b || strings.HasPrefix(a, b+"/") || strings.HasPrefix(b, a+"/"))
}
func contextsOverlap(a, b []string) bool {
	for _, x := range a {
		for _, y := range b {
			if contextOverlap(x, y) {
				return true
			}
		}
	}
	return false
}
func validStamp(s Stamp) bool {
	_, offset := s.UTC.Zone()
	return !s.UTC.IsZero() && offset == 0 && s.Ordinal > 0
}
func before(a, b Stamp) bool {
	return validStamp(a) && validStamp(b) && (a.UTC.Before(b.UTC) || (a.UTC.Equal(b.UTC) && a.Ordinal < b.Ordinal))
}
func atOrBefore(a, b Stamp) bool {
	return validStamp(a) && validStamp(b) && ((a.UTC.Equal(b.UTC) && a.Ordinal == b.Ordinal) || before(a, b))
}
