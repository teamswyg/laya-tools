package semanticframe

// ProjectedPlanSchema identifies a plan whose graph has a separately retained
// FULL acceptance anchor. JoinGraphPlan checks plan structure, not acceptance.
const ProjectedPlanSchema = "riido-projected-family-plan-v1"

// PlanAnchors are exact byte references already joined by a caller. This type
// contains no verdict and does not create a selected-version acceptance.
type PlanAnchors struct {
	Source, Graph, Observation, Producer, SourceReview, Version File
}

// JoinGraphPlan is the shared structural-plan endpoint for a FULL-aware bridge.
// It verifies graph/Source bytes and plan endpoints. Acceptance is deliberately
// absent: callers must independently join FULL-version acceptance before use.
func JoinGraphPlan(f Frame, source, graph PinnedBytes, anchors PlanAnchors, references []File, requiredSourceIDs []string) (Summary, error) {
	blocked := Summary{State: "blocked", Limits: "Plan structure only; no FULL acceptance, meaning or training permission is established."}
	if len(references) > 8192 {
		return blocked, Error("reference_count")
	}
	if err := verify(source, MaxSourceBytes); err != nil {
		return blocked, err
	}
	if err := verify(graph, MaxInventoryBytes); err != nil {
		return blocked, err
	}
	if anchors.Source != source.File || anchors.Graph != graph.File {
		return blocked, Error("graph_plan_anchor")
	}
	all := append([]File{anchors.Source, anchors.Graph, anchors.Observation, anchors.Producer, anchors.SourceReview, anchors.Version}, references...)
	all, err := referencedPins(all)
	if err != nil {
		return blocked, err
	}
	var g InventoryGraph
	if err = decode(graph.Bytes, &g, MaxInventoryBytes); err != nil {
		return blocked, err
	}
	if g.Schema != InventorySchema || g.Source != source.File {
		return blocked, Error("inventory_join")
	}
	if err = validateInventory(g, source.Bytes); err != nil {
		return blocked, err
	}
	out, err := joinPlan(f, planContext{source: source.Bytes, inventory: g, anchors: anchors, references: all, valid: true}, requiredSourceIDs, ProjectedPlanSchema)
	if err == nil {
		out.State = "GRAPH_PLAN_STRUCTURE_JOINED_QA_PENDING"
		out.Limits = "Byte-verified graph and structural plan only. FULL acceptance is a separate caller join; meaning, whole Source QA, rights and training remain unproved."
	}
	return out, err
}
