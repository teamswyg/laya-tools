package acceptedfullbridge

import (
	"bytes"
	"path"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/inventoryprojection"
	"github.com/teamswyg/laya-tools/internal/semanticframe"
	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

func identifier(s string) bool {
	if len(s) == 0 || len(s) > 128 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

// The generic contract has two spellings for one declared actor label:
// label and /root/label. Other namespaces require an explicit future contract;
// do not infer identities by stripping arbitrary prefixes or nested paths.
func declaredActorLabel(s string) string {
	if len(s) > 128 {
		return ""
	}
	label := strings.TrimPrefix(s, "/root/")
	if !identifier(label) {
		return ""
	}
	return label
}
func validFile(f File) bool {
	if len(f.Path) == 0 || len(f.Path) > 4096 || !utf8.ValidString(f.Path) || strings.ContainsAny(f.Path, "\\:\x00") || path.IsAbs(f.Path) || path.Clean(f.Path) != f.Path || f.Path == "." || f.Path == ".." || strings.HasPrefix(f.Path, "../") || len(f.SHA256) != 64 || f.Bytes <= 0 || f.Bytes > MaxBytes {
		return false
	}
	for _, c := range f.SHA256 {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func verify(p PinnedBytes, max int) error {
	if len(p.Bytes) == 0 || len(p.Bytes) > max || !validFile(p.File) || p.File.Bytes != int64(len(p.Bytes)) || sourcecohort.Hash(p.Bytes) != p.File.SHA256 {
		return Error("byte_pin")
	}
	return nil
}
func checkPins(pins []File) error {
	for _, p := range pins {
		if !validFile(p) {
			return Error("reference_pin")
		}
	}
	ordered := append([]File{}, pins...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Path < ordered[j].Path })
	for i := 1; i < len(ordered); i++ {
		if ordered[i].Path == ordered[i-1].Path && ordered[i] != ordered[i-1] {
			return Error("conflicting_reference_pin")
		}
	}
	return nil
}
func files(ps []PinnedBytes) []File {
	out := make([]File, len(ps))
	for i, p := range ps {
		out[i] = p.File
	}
	return out
}
func sameFiles(a, b []File) bool {
	if len(a) != len(b) || len(a) > MaxItems {
		return false
	}
	for _, p := range append(append([]File{}, a...), b...) {
		if !validFile(p) {
			return false
		}
	}
	a = append([]File{}, a...)
	b = append([]File{}, b...)
	sort.Slice(a, func(i, j int) bool { return a[i].Path < a[j].Path })
	sort.Slice(b, func(i, j int) bool { return b[i].Path < b[j].Path })
	for i := range a {
		if a[i] != b[i] || i > 0 && (a[i].Path == a[i-1].Path || b[i].Path == b[i-1].Path) {
			return false
		}
	}
	return true
}
func sameKeys(a, b []string) bool {
	if len(a) != len(b) || len(a) > MaxItems {
		return false
	}
	for _, s := range append(append([]string{}, a...), b...) {
		if !identifier(s) {
			return false
		}
	}
	a = append([]string{}, a...)
	b = append([]string{}, b...)
	sort.Strings(a)
	sort.Strings(b)
	for i := range a {
		if a[i] != b[i] || i > 0 && (a[i] == a[i-1] || b[i] == b[i-1]) {
			return false
		}
	}
	return true
}

// Join executes the declared projector, verifies exact supplied graph/binding
// bytes, joins independent FULL-version declarations and verifies a graph plan.
// No input paths are opened, no declarations written, no histories promoted.
func Join(in Inputs) (Result, error) {
	var out Result
	if in.Projection.Source == nil || len(in.Acceptance) == 0 || len(in.Acceptance) > MaxItems || len(in.History) > MaxItems || len(in.RegisteredSourceIDs) == 0 || len(in.RegisteredSourceIDs) > 400 {
		return out, Error("input_count")
	}
	all := []PinnedBytes{in.Projection.Config, in.Projection.Full, in.Projection.Native, in.Projection.Mapping, in.Projection.Adapter, *in.Projection.Source, in.ProjectionBinding, in.Graph, in.Version, in.Observation, in.Producer, in.Review, in.CheckBinding, in.CheckReport, in.ReadStart, in.ReadResult, in.Frame}
	all = append(all, in.Acceptance...)
	all = append(all, in.History...)
	pins := make([]File, 0, len(all)+len(in.History)*7+12)
	total := 0
	for i, p := range all {
		max := MaxBytes
		if i == 5 {
			max = inventoryprojection.MaxSourceBytes
		}
		if i == 14 || i == 15 {
			max = 8192
		}
		if err := verify(p, max); err != nil {
			return out, err
		}
		total += len(p.Bytes)
		if total > MaxTotalBytes {
			return out, Error("total_bytes")
		}
		pins = append(pins, p.File)
		// All introduced JSON contracts are bounded before typed allocation.
		if i != 4 && i != 5 {
			if err := preflight(p.Bytes); err != nil {
				return out, err
			}
		}
	}
	if err := checkPins(pins); err != nil {
		return out, err
	}
	projection, err := inventoryprojection.Project(in.Projection)
	if err != nil {
		return out, Error("projection")
	}
	if projection.Binding.Graph != in.Graph.File || !bytes.Equal(projection.GraphBytes, in.Graph.Bytes) {
		return out, Error("derived_graph_join")
	}
	var binding inventoryprojection.Binding
	if err = closed(in.ProjectionBinding.Bytes, &binding); err != nil {
		return out, err
	}
	if binding != projection.Binding {
		return out, Error("projection_binding_join")
	}
	var v Version
	if err = closed(in.Version.Bytes, &v); err != nil {
		return out, err
	}
	if v.Schema != VersionSchema || v.SourceID != projection.Full.ID || (v.Selection != "original" && v.Selection != "amended") || v.Source != in.Projection.Source.File || v.Observation != in.Observation.File || v.Full != in.Projection.Full.File || v.Graph != in.Graph.File || v.Projection != in.ProjectionBinding.File || v.Producer != in.Producer.File || v.Review != in.Review.File || !sameFiles(v.Acceptance, files(in.Acceptance)) || !sameFiles(v.History, files(in.History)) {
		return out, Error("selected_version_join")
	}
	if v.ReviewLineage != "first_review" && v.ReviewLineage != "fresh_review" || v.Selection == "amended" && len(v.History) == 0 {
		return out, Error("selection_history")
	}
	var producer Producer
	if err = closed(in.Producer.Bytes, &producer); err != nil {
		return out, err
	}
	producerActor := declaredActorLabel(producer.ProducerActor)
	if producer.Schema != ProducerSchema || producer.SourceID != v.SourceID || producer.Selection != v.Selection || !identifier(producer.ProducerID) || producerActor == "" || producer.ProducerID != projection.Full.Coder.ID || producer.ProducerActor != projection.Full.Coder.Agent || producer.Source != v.Source || producer.Observation != v.Observation || producer.Full != v.Full {
		return out, Error("producer_join")
	}
	var review sourcecohort.Review
	if err = closed(in.Review.Bytes, &review); err != nil {
		return out, err
	}
	if review.ID != v.SourceID || review.Source != v.Source || review.Observation != v.Observation || !identifier(review.ReviewerID) || review.ReviewerID == producer.ProducerID || review.ReviewerID == producerActor || !review.Complete || !review.Structural || review.SourceScope != "complete_source_situation" || review.SemanticVerdict != "declared_pass" || len(review.Holds) != 0 || !sameKeys(review.Dependencies, projection.Full.Dependencies) {
		return out, Error("standard_review_join")
	}
	if err = reviewEvidence(review, in); err != nil {
		return out, err
	}
	pins = append(pins, review.SourceSchema, review.CheckBinding, review.ReadStart, review.ReadResult)
	var native inventoryprojection.NativeContract
	if err = closed(in.Projection.Native.Bytes, &native); err != nil {
		return out, err
	}
	for _, p := range in.Acceptance {
		var a Acceptance
		if err = closed(p.Bytes, &a); err != nil {
			return out, err
		}
		at, ok := utc(a.ReviewedUTC)
		rt, _ := utc(review.ReviewedUTC)
		if a.Schema != AcceptanceSchema || a.SourceID != v.SourceID || a.Selection != v.Selection || a.Source != v.Source || a.Observation != v.Observation || a.Full != v.Full || a.Producer != v.Producer || a.Review != v.Review || a.ReviewerID != review.ReviewerID || a.ReviewerID == producer.ProducerID || a.Verdict != "declared_pass" || a.Scope != AcceptanceScope || !sameKeys(a.OfficialFields, native.Fields) || !sameFiles(a.History, v.History) || !ok || at.Before(rt) {
			return out, Error("independent_full_acceptance_join")
		}
	}
	holds := 0
	inventoryHistory := make([]File, 0, len(in.History))
	reviewHistory := 0
	for _, p := range in.History {
		var h History
		if err = closed(p.Bytes, &h); err != nil {
			return out, err
		}
		if h.Schema != HistorySchema || h.SourceID != v.SourceID || h.Source != v.Source || h.Review == v.Review || h.Version == in.Version.File || (h.Disposition != "held" && h.Disposition != "superseded") {
			return out, Error("retained_history_join")
		}
		switch h.Relation {
		case "inventory_predecessor":
			if h.Full == v.Full {
				return out, Error("inventory_history_join")
			}
			inventoryHistory = append(inventoryHistory, p.File)
		case "review_predecessor":
			if h.Full != v.Full || h.Observation != v.Observation || h.Producer != v.Producer {
				return out, Error("review_history_join")
			}
			reviewHistory++
		default:
			return out, Error("history_relation")
		}
		if h.Disposition == "held" {
			holds++
		}
		pins = append(pins, h.Source, h.Observation, h.Full, h.Producer, h.Review, h.Version)
	}
	if v.Selection == "original" && len(inventoryHistory) != 0 || v.Selection == "amended" && len(inventoryHistory) == 0 || v.ReviewLineage == "first_review" && reviewHistory != 0 || v.ReviewLineage == "fresh_review" && reviewHistory == 0 || !sameFiles(producer.History, inventoryHistory) {
		return out, Error("selection_history")
	}
	// Include every native FULL provenance reference, including unopened records,
	// in the same conflict registry as historical and frame references.
	pins = append(pins, projection.Full.Source, projection.Full.Creation, projection.Full.ReadBinding)
	for _, r := range projection.Full.AdditionalReads {
		pins = append(pins, r.Source, r.Start, r.Result)
	}
	var f Frame
	if err = closed(in.Frame.Bytes, &f); err != nil {
		return out, err
	}
	if f.Schema != FrameSchema || f.Full != v.Full || f.Projection != v.Projection || f.Plan.InventorySchemaSource != binding.Native {
		return out, Error("full_frame_anchor")
	}
	pins = append(pins, f.Full, f.Projection, f.Plan.Source, f.Plan.Inventory, f.Plan.Producer, f.Plan.VersionEvidence, f.Plan.Observation, f.Plan.SourceReview, f.Plan.Definitions, f.Plan.InventorySchemaSource)
	if err = checkPins(pins); err != nil {
		return out, err
	}
	plan, err := semanticframe.JoinGraphPlan(f.Plan, *in.Projection.Source, in.Graph, semanticframe.PlanAnchors{Source: v.Source, Graph: v.Graph, Observation: v.Observation, Producer: v.Producer, SourceReview: v.Review, Version: in.Version.File}, pins, in.RegisteredSourceIDs)
	if err != nil {
		return out, Error("frame_plan")
	}
	out = Result{Projection: projection, Summary: Summary{State: "FULL_ACCEPTANCE_AND_GRAPH_PLAN_DECLARATIONS_JOINED_QA_PENDING", AcceptanceDeclarations: len(in.Acceptance), RetainedHolds: holds, Plan: plan, Limits: "Pinned generic FULL acceptance, actor/read/check lineage and derived graph plan links only. Historical native adapters, authenticated independence, semantic meaning, whole Source QA, rights and training remain unproved."}}
	return out, nil
}
