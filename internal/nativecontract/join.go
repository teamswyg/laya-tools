package nativecontract

import (
	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

// Join verifies the exact native metadata wire and declared relationships.
// No referenced record is opened and no verdict or graph acceptance is created.
func Join(in Inputs) (Result, error) {
	var out Result
	if in.Kind != Original && in.Kind != Amended || !registered(in.Expected.ID, in.RegisteredSourceIDs) || !identifier(in.Expected.ReviewerID) || !identifier(in.Expected.ProducerID) || actorLabel(in.Expected.ProducerActor) == "" || in.Expected.ReviewerID == in.Expected.ProducerID || in.Expected.ReviewerID == actorLabel(in.Expected.ProducerActor) || in.Kind == Amended && !keysSame(in.Expected.OfficialKeys, in.Expected.OfficialKeys) {
		return out, Error("expected_identity")
	}
	all := []PinnedBytes{in.Review, in.Acceptance, in.Version, in.Disposition}
	if in.Witnesses != nil {
		w := in.Witnesses
		all = append(all, w.ReadStart, w.ReadResult, w.Check, w.Report, w.HeldReview)
	}
	total := 0
	pins := make([]File, 0, 128)
	for _, p := range all {
		if e := verify(p); e != nil {
			return out, e
		}
		total += len(p.Bytes)
		if total > MaxTotalBytes {
			return out, Error("total_bytes")
		}
		if e := preflight(p.Bytes); e != nil {
			return out, e
		}
		pins = append(pins, p.File)
	}
	ex := in.Expected
	pins = append(pins, ex.Source, ex.SourceSchema, ex.OriginalObservation, ex.OriginalFull, ex.OriginalProducer, ex.SelectedObservation, ex.SelectedFull, ex.SelectedProducer, ex.OwnInventory, ex.ReadStart, ex.ReadResult, ex.Start, ex.Check, ex.Report, ex.HeldReview)
	if ex.Source.Bytes > 16384 || ex.ReadResult.Path != ex.ReadStart.Path+".result" || ex.HeldReview == in.Review.File {
		return out, Error("expected_lineage")
	}
	if in.Kind == Original {
		pins = append(pins, ex.Method)
		if ex.SelectedObservation != ex.OriginalObservation || ex.SelectedFull != ex.OriginalFull || ex.SelectedProducer != ex.OriginalProducer {
			return out, Error("original_unchanged_anchors")
		}
	} else {
		pins = append(pins, ex.PairScope, ex.AcceptanceScope)
		if ex.SelectedFull == ex.OriginalFull || ex.SelectedProducer == ex.OriginalProducer {
			return out, Error("amended_distinct_anchors")
		}
	}
	var review sourcecohort.Review
	var d Disposition
	if e := closed(in.Review.Bytes, &review); e != nil {
		return out, e
	}
	if e := closed(in.Disposition.Bytes, &d); e != nil {
		return out, e
	}
	if review.ID != ex.ID || review.ReviewerID != ex.ReviewerID || review.Source != ex.Source || review.SourceSchema != ex.SourceSchema || review.Observation != ex.SelectedObservation || review.ReadStart != ex.ReadStart || review.ReadResult != ex.ReadResult || review.CheckBinding != ex.Check || len(review.Dependencies) > MaxItems || len(review.Evidence) > MaxItems || len(review.Holds) > MaxItems {
		return out, Error("standard_review_join")
	}
	if _, ok := UTC(review.ReviewedUTC); !ok {
		return out, Error("review_utc")
	}
	for _, id := range review.Dependencies {
		if !registered(id, in.RegisteredSourceIDs) {
			return out, Error("review_dependencies")
		}
	}
	if d.ID != ex.ID || d.Reviewer != ex.ReviewerID || d.Version != in.Version.File || d.Review != in.Review.File || d.HeldReview != ex.HeldReview || d.SemanticHolds < 0 || d.TechnicalHolds < 0 || d.SemanticHolds > MaxItems || d.TechnicalHolds > MaxItems || d.OldPass || d.WholeQA || d.Training {
		return out, Error("current_disposition_join")
	}
	if _, ok := UTC(d.UTC); !ok {
		return out, Error("disposition_utc")
	}
	selected := Selected{SourceID: ex.ID, ReviewerID: ex.ReviewerID, Kind: in.Kind, Source: ex.Source, Observation: ex.SelectedObservation, Full: ex.SelectedFull, Producer: ex.SelectedProducer, OriginalObservation: ex.OriginalObservation, OriginalFull: ex.OriginalFull, OriginalProducer: ex.OriginalProducer, Review: in.Review.File, Acceptance: in.Acceptance.File, Version: in.Version.File, Disposition: in.Disposition.File, HeldReview: ex.HeldReview, ReadStart: ex.ReadStart, ReadResult: ex.ReadResult, Check: ex.Check, Report: ex.Report, CurrentSemanticHolds: d.SemanticHolds, CurrentTechnicalHolds: d.TechnicalHolds}
	var copies []CopyRead
	if in.Kind == Original {
		var a OriginalAcceptance
		var v OriginalVersion
		if e := closed(in.Acceptance.Bytes, &a); e != nil {
			return out, e
		}
		if e := closed(in.Version.Bytes, &v); e != nil {
			return out, e
		}
		if a.Schema != OriginalAcceptanceSchema || v.Schema != OriginalVersionSchema || d.Schema != OriginalDispositionSchema || v.Selection != "reconsidered_original" || !v.Unchanged || v.OldPass || v.WholeQA || v.Training || a.WholeQA || a.Training {
			return out, Error("original_contract")
		}
		if a.ID != ex.ID || v.ID != ex.ID || a.Reviewer != ex.ReviewerID || v.Reviewer != ex.ReviewerID || a.Source != ex.Source || v.Source != ex.Source || a.Observation != ex.OriginalObservation || v.Observation != ex.OriginalObservation || a.Full != ex.OriginalFull || v.Full != ex.OriginalFull || a.Producer != ex.OriginalProducer || v.Producer != ex.OriginalProducer || a.Review != in.Review.File || v.Review != in.Review.File || v.Acceptance != in.Acceptance.File || a.HeldReview != ex.HeldReview || v.HeldReview != ex.HeldReview || a.OwnInventory != ex.OwnInventory || v.OwnInventory != ex.OwnInventory || a.ReadStart != ex.ReadStart || v.ReadStart != ex.ReadStart || a.ReadResult != ex.ReadResult || v.ReadResult != ex.ReadResult || a.Start != ex.Start || v.Start != ex.Start || a.Check != ex.Check || v.Check != ex.Check || a.Report != ex.Report || v.Report != ex.Report || a.Method != ex.Method || v.Method != ex.Method {
			return out, Error("original_native_join")
		}
		if _, ok := UTC(a.ReviewedUTC); !ok {
			return out, Error("acceptance_utc")
		}
		if _, ok := UTC(v.UTC); !ok {
			return out, Error("version_utc")
		}
		selected.Verdict = a.Verdict
		selected.Scope = a.Scope
		selected.Unresolved = a.Unresolved
		copies = a.Copies
		pins = append(pins, a.Rationale)
		out.OriginalAcceptance = &a
		out.OriginalVersion = &v
	} else {
		var a AmendedAcceptance
		var v AmendedVersion
		if e := closed(in.Acceptance.Bytes, &a); e != nil {
			return out, e
		}
		if e := closed(in.Version.Bytes, &v); e != nil {
			return out, e
		}
		if a.Schema != AmendedAcceptanceSchema || v.Schema != AmendedVersionSchema || d.Schema != AmendedDispositionSchema || v.Selection != "independently_reviewed_amended_pair" || !v.Unchanged || v.OldPass || v.WholeQA || v.Training || a.OldPass || a.WholeQA || a.Training || !keysSame(a.OfficialKeys, ex.OfficialKeys) {
			return out, Error("amended_contract")
		}
		if a.ID != ex.ID || v.ID != ex.ID || a.Reviewer != ex.ReviewerID || v.Reviewer != ex.ReviewerID || a.Source != ex.Source || v.Source != ex.Source || a.OriginalObservation != ex.OriginalObservation || v.OriginalObservation != ex.OriginalObservation || a.OriginalFull != ex.OriginalFull || v.OriginalFull != ex.OriginalFull || a.OriginalProducer != ex.OriginalProducer || v.OriginalProducer != ex.OriginalProducer || a.Observation != ex.SelectedObservation || v.Observation != ex.SelectedObservation || a.Full != ex.SelectedFull || v.Full != ex.SelectedFull || a.Pair != ex.SelectedProducer || v.Pair != ex.SelectedProducer || a.PairScope != ex.PairScope || v.PairScope != ex.PairScope || a.AcceptanceScope != ex.AcceptanceScope || v.AcceptanceScope != ex.AcceptanceScope || a.Review != in.Review.File || v.Review != in.Review.File || v.Acceptance != in.Acceptance.File || a.OwnInventory != ex.OwnInventory || v.OwnInventory != ex.OwnInventory || a.ReadStart != ex.ReadStart || v.ReadStart != ex.ReadStart || a.ReadResult != ex.ReadResult || v.ReadResult != ex.ReadResult || a.Start != ex.Start || v.Start != ex.Start || a.Check != ex.Check || v.Check != ex.Check || a.Report != ex.Report || v.Report != ex.Report || a.HeldReview != ex.HeldReview || v.HeldReview != ex.HeldReview {
			return out, Error("amended_native_join")
		}
		if _, ok := UTC(a.ReviewedUTC); !ok {
			return out, Error("acceptance_utc")
		}
		if _, ok := UTC(v.UTC); !ok {
			return out, Error("version_utc")
		}
		selected.Verdict = a.Verdict
		selected.Scope = a.Scope
		selected.Unresolved = a.Unresolved
		copies = a.Copies
		pins = append(pins, a.Rationale)
		out.AmendedAcceptance = &a
		out.AmendedVersion = &v
	}
	if selected.Scope != Scope || selected.Unresolved < 0 || selected.Unresolved > MaxItems || selected.Verdict != review.SemanticVerdict {
		return Result{}, Error("independent_verdict_join")
	}
	switch selected.Verdict {
	case "declared_pass":
		if !review.Complete || !review.Structural || review.SourceScope != "complete_source_situation" || len(review.Holds) != 0 || selected.Unresolved != 0 || d.SemanticHolds != 0 || d.TechnicalHolds != 0 {
			return Result{}, Error("pass_with_holds")
		}
	case "declared_hold":
		if len(review.Holds) == 0 && selected.Unresolved == 0 && d.SemanticHolds == 0 && d.TechnicalHolds == 0 {
			return Result{}, Error("hold_without_retained_reason")
		}
	default:
		return Result{}, Error("declared_verdict")
	}
	if len(copies) > MaxItems {
		return Result{}, Error("copy_count")
	}
	for _, c := range copies {
		if c.Result.Path != c.Start.Path+".result" {
			return Result{}, Error("copy_read_pair")
		}
		pins = append(pins, c.Input, c.Start, c.Result)
	}
	if e := checkPins(pins); e != nil {
		return Result{}, e
	}
	if in.Witnesses != nil {
		if e := witnessJoin(review, in, pins); e != nil {
			return Result{}, e
		}
		w := in.Witnesses
		out.RawWitnesses = &Witnesses{ReadStart: pinnedCopy(w.ReadStart), ReadResult: pinnedCopy(w.ReadResult), Check: pinnedCopy(w.Check), Report: pinnedCopy(w.Report), HeldReview: pinnedCopy(w.HeldReview)}
	}
	out.Selected = selected
	out.Review = review
	out.Disposition = d
	out.RawReview = pinnedCopy(in.Review)
	out.RawAcceptance = pinnedCopy(in.Acceptance)
	out.RawVersion = pinnedCopy(in.Version)
	out.RawDisposition = pinnedCopy(in.Disposition)
	out.Summary = Summary{State: "NATIVE_METADATA_DECLARATIONS_JOINED_QA_PENDING", DeclaredVerdict: selected.Verdict, HeldReviewReferenceRetained: true, OfficialKeysVerified: in.Kind == Amended, WitnessContentVerified: in.Witnesses != nil, Limits: "Only the exact new original-reconsideration/amended native contracts and pinned caller-supplied metadata are joined. Original acceptance has no explicit seven-key field; that completeness remains a scope declaration. Referenced Source/Obs/FULL/producer/pair/scope/rationale/read targets are not opened; producer/pair bodies and all other historical schemas are not decoded. Optional public read/check/old-Review witnesses are verified only when supplied. Exact original metadata bytes/pins and held-review references retained, no verdict minted or old hold promoted. No authentication, provider independence, meaning, rights, whole Source QA or training admission."}
	return out, nil
}
