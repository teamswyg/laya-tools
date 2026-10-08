package semanticframe

import (
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

var categories = [...]string{"inspected_components", "propositions", "opposition_relations", "temporal_support", "referent_support", "reply_quote_adoption_retraction_negation_constraints"}

func text(s string, max int) bool {
	return len(s) > 0 && len(s) <= max && utf8.ValidString(s) && !strings.ContainsRune(s, 0) && strings.TrimSpace(s) != ""
}
func spanOK(s Span, source []byte) bool {
	if s.Start < 0 || s.Start >= s.End || s.End > len(source) {
		return false
	}
	if s.Start < len(source) && !utf8.RuneStart(source[s.Start]) || s.End < len(source) && !utf8.RuneStart(source[s.End]) {
		return false
	}
	switch s.Kind {
	case "source_scope", "proposition", "opposition", "event_anchor", "communication_time", "absence", "not_applicable":
		return true
	}
	return false
}
func evidenceOK(es []Span, source []byte) bool {
	if len(es) == 0 || len(es) > MaxItems {
		return false
	}
	for _, e := range es {
		if !spanOK(e, source) {
			return false
		}
	}
	return true
}
func boundaryOK(s Span, source []byte) bool {
	return s.Kind == "source_scope" && s.Start == 0 && s.End == len(source) && len(source) > 0
}
func sortedIDs(ids []string) ([]string, error) {
	if len(ids) > MaxItems {
		return nil, Error("item_count")
	}
	out := append([]string{}, ids...)
	sort.Strings(out)
	for i, id := range out {
		if !identifier(id) || i > 0 && out[i-1] == id {
			return nil, Error("duplicate_id")
		}
	}
	return out, nil
}
func present(ids []string, id string) bool {
	i := sort.SearchStrings(ids, id)
	return i < len(ids) && ids[i] == id
}
func refsOK(ids, allowed []string, nonempty bool) bool {
	if nonempty && len(ids) == 0 {
		return false
	}
	sorted, err := sortedIDs(ids)
	if err != nil {
		return false
	}
	for _, id := range sorted {
		if !present(allowed, id) {
			return false
		}
	}
	return true
}

func validateInventory(g InventoryGraph, source []byte) error {
	if !boundaryOK(g.Boundary, source) || !utf8.Valid(source) {
		return Error("source_boundary")
	}
	if len(g.Components) > MaxItems || len(g.Propositions) > MaxItems || len(g.Oppositions) > MaxItems || len(g.Temporal) > MaxItems || len(g.Referents) > MaxItems || len(g.Constraints) > MaxItems {
		return Error("item_count")
	}
	componentIDs := make([]string, len(g.Components))
	propIDs := make([]string, len(g.Propositions))
	opIDs := make([]string, len(g.Oppositions))
	timeIDs := make([]string, len(g.Temporal))
	refIDs := make([]string, len(g.Referents))
	for i, x := range g.Components {
		componentIDs[i] = x.ID
		if !text(x.Description, 4096) || !evidenceOK(x.Evidence, source) {
			return Error("component")
		}
	}
	for i, x := range g.Propositions {
		propIDs[i] = x.ID
		if !text(x.Description, 4096) || !text(x.Scope, 4096) || !text(x.Polarity, 128) || !evidenceOK(x.Evidence, source) {
			return Error("proposition")
		}
	}
	for i, x := range g.Oppositions {
		opIDs[i] = x.ID
		if !text(x.Description, 4096) || !evidenceOK(x.Evidence, source) || x.Left == x.Right {
			return Error("opposition")
		}
	}
	for i, x := range g.Temporal {
		timeIDs[i] = x.ID
		if !text(x.Relation, 128) || !text(x.Anchor, 4096) || !evidenceOK(x.Evidence, source) || !evidenceOK(x.CommunicationEvidence, source) {
			return Error("temporal")
		}
	}
	for i, x := range g.Referents {
		refIDs[i] = x.ID
		if !text(x.Kind, 128) || !text(x.Description, 4096) || !evidenceOK(x.Evidence, source) {
			return Error("referent")
		}
	}
	var err error
	for _, ids := range [][]string{componentIDs, propIDs, opIDs, timeIDs, refIDs} {
		if _, err = sortedIDs(ids); err != nil {
			return err
		}
	}
	propIDs, _ = sortedIDs(propIDs)
	refIDs, _ = sortedIDs(refIDs)
	for _, x := range g.Propositions {
		if !refsOK(x.Subjects, refIDs, false) {
			return Error("proposition_referent")
		}
	}
	for _, x := range g.Oppositions {
		if !present(propIDs, x.Left) || !present(propIDs, x.Right) {
			return Error("opposition_endpoint")
		}
	}
	for _, x := range g.Temporal {
		if !refsOK(x.Propositions, propIDs, true) {
			return Error("temporal_endpoint")
		}
	}
	for _, x := range g.Constraints {
		if !text(x.Kind, 128) || !text(x.Description, 4096) || !evidenceOK(x.Evidence, source) || !refsOK(x.Propositions, propIDs, true) {
			return Error("constraint_endpoint")
		}
	}
	// Every parent must exist; following at most N links detects all cycles.
	refs := append([]Referent{}, g.Referents...)
	sort.Slice(refs, func(i, j int) bool { return refs[i].ID < refs[j].ID })
	for _, r := range refs {
		p := r.Parent
		for depth := 0; p != nil; depth++ {
			if depth >= len(refs) || !present(refIDs, *p) {
				return Error("referent_parent")
			}
			i := sort.Search(len(refs), func(i int) bool { return refs[i].ID >= *p })
			p = refs[i].Parent
		}
	}
	deps, err := sortedIDs(g.Dependencies)
	if err != nil {
		return err
	}
	if present(deps, g.SourceID) {
		return Error("self_dependency")
	}
	return nil
}

func pointer(p string, g InventoryGraph) (int, int, error) {
	if len(p) == 0 || len(p) > maxPointerBytes {
		return 0, 0, Error("inventory_pointer")
	}
	parts := strings.Split(p, "/")
	if len(parts) != 3 || parts[0] != "" || len(parts[2]) == 0 || len(parts[2]) > 3 || len(parts[2]) > 1 && parts[2][0] == '0' {
		return 0, 0, Error("inventory_pointer")
	}
	for _, c := range parts[2] {
		if c < '0' || c > '9' {
			return 0, 0, Error("inventory_pointer")
		}
	}
	i, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, Error("inventory_pointer")
	}
	lengths := [...]int{len(g.Components), len(g.Propositions), len(g.Oppositions), len(g.Temporal), len(g.Referents), len(g.Constraints)}
	for c, name := range categories {
		if parts[1] == name && i < lengths[c] {
			return c, i, nil
		}
	}
	return 0, 0, Error("inventory_pointer")
}

// DecodeFrame bounds and closes the V3 JSON shape. It does not open pins or
// interpret prose. JoinFrame additionally checks selected bytes and graph links.
func DecodeFrame(raw []byte) (Frame, error) {
	var f Frame
	if err := decode(raw, &f, MaxFrameBytes); err != nil {
		return f, err
	}
	return f, nil
}

// JoinFrame verifies structural plans on a byte-verified generic inventory.
// RequiredSourceIDs is the complete externally supplied registered cohort list,
// used only to reject orphan Source dependencies; it proves no whole-cohort QA.
// No component is forced to communicate a claim by being true in the Source.
func JoinFrame(f Frame, n NormalizedAccepted, requiredSourceIDs []string) (Summary, error) {
	s := Summary{State: "blocked", Bindings: len(f.Plan.Bindings), Ambiguities: len(f.Ambiguities), Limits: "Byte-verified generic declarations and structural joins only. Meaning, whole Source QA, fidelity, rights, authenticated independence, references and training remain unproved. Historical private-record adapters are not implemented."}
	fail := func(code string) (Summary, error) { return s, Error(code) }
	if _, ok := typedFrameSize(f); !ok {
		return fail("typed_frame_bounds")
	}
	if n.version.Schema != VersionSchema || len(n.source) == 0 {
		return fail("normalized_missing")
	}
	if f.Schema != Schema || !identifier(f.FamilyID) || f.SourceID != n.inventory.SourceID || f.Slot < 1 || f.Slot > 3 || f.DesiredLabels || f.TargetMode != "one_complete_comment_text_only" {
		return fail("frame_identity")
	}
	if f.Source != n.version.Source || f.Inventory != n.version.Inventory || f.Observation != n.version.Observation || f.Producer != n.version.Producer || f.SourceReview != n.version.SourceReview || f.VersionEvidence != n.versionPin {
		return fail("frame_version_join")
	}
	if !validFile(f.Definitions) || !validFile(f.InventorySchemaSource) || f.Definitions.Path == f.InventorySchemaSource.Path {
		return fail("method_pin")
	}
	allPins := []File{f.Source, f.Inventory, f.Observation, f.Producer, f.SourceReview, f.VersionEvidence, f.Definitions, f.InventorySchemaSource}
	if uniquePins(allPins) != nil {
		return fail("frame_pin")
	}
	if _, err := referencedPins(append(append([]File{}, n.references...), allPins...)); err != nil {
		return fail("conflicting_reference_pin")
	}
	if !boundaryOK(f.Boundary, n.source) || f.Boundary != n.inventory.Boundary {
		return fail("source_boundary")
	}
	// Cohort IDs are bounded independently from the per-inventory 256-item cap.
	if len(requiredSourceIDs) == 0 || len(requiredSourceIDs) > 400 {
		return fail("registered_sources")
	}
	for _, id := range requiredSourceIDs {
		if !identifier(id) {
			return fail("registered_sources")
		}
	}
	registered := append([]string{}, requiredSourceIDs...)
	sort.Strings(registered)
	for i, id := range registered {
		if !identifier(id) || i > 0 && registered[i-1] == id {
			return fail("registered_sources")
		}
	}
	if !present(registered, f.SourceID) {
		return fail("registered_sources")
	}
	for _, id := range n.inventory.Dependencies {
		if !present(registered, id) {
			return fail("orphan_dependency")
		}
	}
	if len(f.Plan.Bindings) == 0 || len(f.Plan.Bindings) > MaxItems || !evidenceOK(f.Plan.Evidence, n.source) || !text(f.Plan.Description, 1024) || f.Plan.WordingFreedom != "source_supported_rephrasing_only" || len(f.Ambiguities) > MaxItems {
		return fail("plan_bounds")
	}
	bindings := append([]Binding{}, f.Plan.Bindings...)
	sort.Slice(bindings, func(i, j int) bool { return bindings[i].Pointer < bindings[j].Pointer })
	var selected [6][MaxItems]bool
	for i, b := range bindings {
		if i > 0 && bindings[i-1].Pointer == b.Pointer {
			return fail("duplicate_or_conflicting_role")
		}
		if b.Role != "utterance_content" && b.Role != "governing_constraint" && b.Role != "unavailable_source_context" {
			return fail("disclosure_role")
		}
		c, j, err := pointer(b.Pointer, n.inventory)
		if err != nil {
			return fail("inventory_pointer")
		}
		selected[c][j] = true
	}
	// Declared graph edges from every selected item must retain their endpoints.
	// A different disclosure role may retain the endpoint; it need not be uttered.
	propSelectedIDs := make([]string, 0, len(n.inventory.Propositions))
	refSelectedIDs := make([]string, 0, len(n.inventory.Referents))
	for i, p := range n.inventory.Propositions {
		if selected[1][i] {
			propSelectedIDs = append(propSelectedIDs, p.ID)
		}
	}
	for i, r := range n.inventory.Referents {
		if selected[4][i] {
			refSelectedIDs = append(refSelectedIDs, r.ID)
		}
	}
	sort.Strings(propSelectedIDs)
	sort.Strings(refSelectedIDs)
	propSelected := func(id string) bool { return present(propSelectedIDs, id) }
	refSelected := func(id string) bool { return present(refSelectedIDs, id) }
	for i, p := range n.inventory.Propositions {
		if selected[1][i] {
			for _, id := range p.Subjects {
				if !refSelected(id) {
					return fail("plan_referent_dependency")
				}
			}
		}
	}
	for i, o := range n.inventory.Oppositions {
		if selected[2][i] && (!propSelected(o.Left) || !propSelected(o.Right)) {
			return fail("plan_opposition_dependency")
		}
	}
	for i, t := range n.inventory.Temporal {
		if selected[3][i] {
			for _, id := range t.Propositions {
				if !propSelected(id) {
					return fail("plan_temporal_dependency")
				}
			}
		}
	}
	for i, r := range n.inventory.Referents {
		if selected[4][i] && r.Parent != nil && !refSelected(*r.Parent) {
			return fail("plan_parent_dependency")
		}
	}
	for i, c := range n.inventory.Constraints {
		if selected[5][i] {
			for _, id := range c.Propositions {
				if !propSelected(id) {
					return fail("plan_constraint_dependency")
				}
			}
		}
	}
	for _, a := range f.Ambiguities {
		if len(a.Pointers) == 0 || len(a.Pointers) > MaxItems || !evidenceOK(a.Evidence, n.source) || !text(a.Description, 512) || a.Operation != "preserve_without_resolution" {
			return fail("ambiguity")
		}
		ps := append([]string{}, a.Pointers...)
		sort.Strings(ps)
		for i, p := range ps {
			c, j, err := pointer(p, n.inventory)
			if err != nil || i > 0 && ps[i-1] == p || !selected[c][j] {
				return fail("ambiguity_pointer")
			}
		}
	}
	s.State = "STRUCTURAL_DECLARATIONS_JOINED_QA_PENDING"
	return s, nil
}
