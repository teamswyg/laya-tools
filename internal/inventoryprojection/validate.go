package inventoryprojection

import (
	"sort"
	"unicode/utf8"
)

func span(p Span, size int, source *PinnedBytes) error {
	if p.Start < 0 || p.End <= p.Start || p.End > size {
		return Error("span_range")
	}
	switch p.Kind {
	case "source_scope", "proposition", "opposition", "event_anchor", "communication_time", "absence", "not_applicable":
	default:
		return Error("span_kind")
	}
	if source != nil {
		b := source.Bytes
		if p.Start < len(b) && !utf8.RuneStart(b[p.Start]) || p.End < len(b) && !utf8.RuneStart(b[p.End]) {
			return Error("utf8_span")
		}
	}
	return nil
}
func evidence(ps []Span, size int, source *PinnedBytes) error {
	if len(ps) > MaxItems {
		return Error("evidence_count")
	}
	for _, p := range ps {
		if err := span(p, size, source); err != nil {
			return err
		}
	}
	return nil
}

func validate(f Full, source *PinnedBytes) error {
	size := int(f.Source.Bytes)
	if f.Boundary.Kind != "source_scope" || f.Boundary.Start != 0 || f.Boundary.End != size {
		return Error("whole_boundary")
	}
	if err := span(f.Boundary, size, source); err != nil {
		return err
	}
	if !identifier(f.Coder.ID) || len(f.Coder.Agent) == 0 || len(f.Coder.Agent) > 128 {
		return Error("coder_metadata")
	}
	componentIDs := make([]string, len(f.Components))
	props := make([]string, len(f.Propositions))
	opps := make([]string, len(f.Oppositions))
	times := make([]string, len(f.Temporal))
	referents := make([]string, len(f.Referents))
	for i, p := range f.Components {
		componentIDs[i] = p.ID
		if err := evidence(p.Evidence, size, source); err != nil {
			return err
		}
	}
	for i, p := range f.Propositions {
		props[i] = p.ID
		if err := evidence(p.Evidence, size, source); err != nil {
			return err
		}
	}
	for i, p := range f.Oppositions {
		opps[i] = p.ID
		if err := evidence(p.Evidence, size, source); err != nil {
			return err
		}
	}
	for i, p := range f.Temporal {
		times[i] = p.ID
		if err := evidence(p.Evidence, size, source); err != nil {
			return err
		}
		if err := evidence(p.CommunicationEvidence, size, source); err != nil {
			return err
		}
	}
	for i, p := range f.Referents {
		referents[i] = p.ID
		if err := evidence(p.Evidence, size, source); err != nil {
			return err
		}
	}
	var err error
	if _, err = sortedIDs(componentIDs); err != nil {
		return err
	}
	if props, err = sortedIDs(props); err != nil {
		return err
	}
	if opps, err = sortedIDs(opps); err != nil {
		return err
	}
	if times, err = sortedIDs(times); err != nil {
		return err
	}
	if referents, err = sortedIDs(referents); err != nil {
		return err
	}
	for _, p := range f.Propositions {
		if err = refs(p.Subjects, referents); err != nil {
			return err
		}
	}
	for _, p := range f.Oppositions {
		if p.Left == p.Right || !present(props, p.Left) || !present(props, p.Right) {
			return Error("opposition_endpoint")
		}
	}
	for _, p := range f.Temporal {
		if err = refs(p.Propositions, props); err != nil {
			return err
		}
	}
	for _, p := range f.Constraints {
		if err = evidence(p.Evidence, size, source); err != nil {
			return err
		}
		if err = refs(p.Propositions, props); err != nil {
			return err
		}
	}
	if err = refs(f.NegativeIDs, props); err != nil {
		return err
	}
	parents := make([]struct {
		ID     string
		Parent *string
	}, 0, len(f.Referents))
	for _, p := range f.Referents {
		parents = append(parents, struct {
			ID     string
			Parent *string
		}{p.ID, p.Parent})
	}
	sort.Slice(parents, func(i, j int) bool { return parents[i].ID < parents[j].ID })
	for _, p := range parents {
		parent := p.Parent
		for depth := 0; parent != nil; depth++ {
			if depth >= len(parents) || !present(referents, *parent) {
				return Error("referent_parent")
			}
			j := sort.Search(len(parents), func(j int) bool { return parents[j].ID >= *parent })
			parent = parents[j].Parent
		}
	}
	for _, field := range f.OfficialFields {
		s := field.Support
		if err = refs(s.Propositions, props); err != nil {
			return err
		}
		if err = refs(s.Referents, referents); err != nil {
			return err
		}
		if err = refs(s.Oppositions, opps); err != nil {
			return err
		}
		if err = refs(s.Temporal, times); err != nil {
			return err
		}
		if err = evidence(s.Evidence, size, source); err != nil {
			return err
		}
	}
	dependencies, err := sortedIDs(f.Dependencies)
	if err != nil {
		return err
	}
	if present(dependencies, f.ID) {
		return Error("self_dependency")
	}
	return nil
}
