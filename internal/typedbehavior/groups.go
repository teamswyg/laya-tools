package typedbehavior

import (
	"errors"
	"slices"

	"github.com/teamswyg/laya-tools/internal/behaviorprobe"
	"github.com/teamswyg/laya-tools/internal/lexicalhint"
)

type RelationSource struct {
	ID                   string
	Core                 string
	NormalizedCodeSHA256 string
	Components           []ComponentPin
}
type RelationParent struct {
	ID, Prototype, Core, Request string
	CandidateTexts               []string
	Sources                      []RelationSource
}
type GroupReport struct {
	Schema                   string                    `json:"schema"`
	Parents                  int                       `json:"parents"`
	PrototypeCount           int                       `json:"prototype_count"`
	ConnectedGroups          int                       `json:"connected_groups"`
	MinimumGroups            int                       `json:"minimum_operational_groups"`
	MeetsGroupMinimum        bool                      `json:"meets_group_minimum"`
	LabeledConnectedGroups   int                       `json:"labeled_connected_groups"`
	MinimumLabeledGroups     int                       `json:"minimum_labeled_connected_groups"`
	MeetsLabeledGroupMinimum bool                      `json:"meets_labeled_group_minimum"`
	TrainingReady            bool                      `json:"training_ready"`
	Edges                    []behaviorprobe.GroupEdge `json:"group_edges"`
	Groups                   []behaviorprobe.Group     `json:"groups"`
	Policy                   string                    `json:"policy"`
}

// CombinedGroups preserves the v1 audit and applies a separately declared
// conservative relation policy to the combined development cohorts.
func CombinedGroups(legacy behaviorprobe.Dataset, lr behaviorprobe.Report, d Dataset, r Report) (GroupReport, error) {
	if len(legacy.Parents) != lr.Parents || len(d.Parents) != r.Parents {
		return GroupReport{}, errors.New("group_cohort_count_mismatch")
	}
	var parents []RelationParent
	for _, p := range legacy.Parents {
		contract := slices.IndexFunc(lr.Contracts, func(c behaviorprobe.ContractEvidence) bool { return c.Prototype == p.Prototype })
		if contract < 0 {
			return GroupReport{}, errors.New("legacy_group_contract_missing")
		}
		item := RelationParent{ID: p.ID, Prototype: p.Prototype, Core: lr.Contracts[contract].CoreTemplate, Request: p.Request}
		for _, c := range p.Candidates {
			n := slices.IndexFunc(lr.Sources, func(s behaviorprobe.SourcePin) bool { return s.ID == c.SourceID })
			if n < 0 {
				return GroupReport{}, errors.New("legacy_group_source_missing")
			}
			s := lr.Sources[n]
			item.CandidateTexts = append(item.CandidateTexts, c.Text)
			item.Sources = append(item.Sources, RelationSource{"v1:" + s.ID, s.CoreTemplate, s.NormalizedCodeSHA256, []ComponentPin{{ID: "v1:" + s.ID, Kind: "legacy_candidate_function", SHA256: s.CodeSHA256, NormalizedBehaviorSHA256: s.NormalizedCodeSHA256}}})
		}
		parents = append(parents, item)
	}
	for _, p := range d.Parents {
		contract := slices.IndexFunc(r.Contracts, func(c ContractSpec) bool { return c.Prototype == p.Prototype })
		if contract < 0 {
			return GroupReport{}, errors.New("typed_group_contract_missing")
		}
		item := RelationParent{ID: p.ID, Prototype: p.Prototype, Core: r.Contracts[contract].Core, Request: p.Request}
		for _, c := range p.Candidates {
			n := slices.IndexFunc(r.Sources, func(s SourcePin) bool { return s.ID == c.SourceID })
			if n < 0 {
				return GroupReport{}, errors.New("typed_group_source_missing")
			}
			s := r.Sources[n]
			item.CandidateTexts = append(item.CandidateTexts, c.Text)
			item.Sources = append(item.Sources, RelationSource{"v2:" + s.ID, s.Core, s.NormalizedCodeSHA256, s.Components})
		}
		parents = append(parents, item)
	}
	groupReport, err := auditRelations(parents)
	if err != nil {
		return GroupReport{}, err
	}
	if len(lr.Outcomes) != len(legacy.Parents) || len(r.Outcomes) != len(d.Parents) {
		return GroupReport{}, errors.New("group_outcome_count_mismatch")
	}
	var labeled []string
	for i, outcome := range lr.Outcomes {
		if outcome.ParentID != legacy.Parents[i].ID || !slices.Contains([]string{"known", "no_answer", "unknown"}, outcome.State) {
			return GroupReport{}, errors.New("legacy_group_outcome_identity_invalid")
		}
		if outcome.State != "unknown" {
			labeled = append(labeled, outcome.ParentID)
		}
	}
	for i, outcome := range r.Outcomes {
		if outcome.ParentID != d.Parents[i].ID || !slices.Contains([]string{"known", "no_answer", "unknown"}, outcome.State) {
			return GroupReport{}, errors.New("typed_group_outcome_identity_invalid")
		}
		if outcome.State != "unknown" {
			labeled = append(labeled, outcome.ParentID)
		}
	}
	applyLabelEvidence(&groupReport, labeled)
	return groupReport, nil
}

func applyLabelEvidence(r *GroupReport, labeledParents []string) {
	r.LabeledConnectedGroups = 0
	for _, group := range r.Groups {
		if slices.ContainsFunc(group.Parents, func(id string) bool { return slices.Contains(labeledParents, id) }) {
			r.LabeledConnectedGroups++
		}
	}
	r.MeetsLabeledGroupMinimum = r.LabeledConnectedGroups >= r.MinimumLabeledGroups
}

func unique(out []string, s string) []string {
	if !slices.Contains(out, s) {
		return append(out, s)
	}
	return out
}

func auditRelations(parents []RelationParent) (GroupReport, error) {
	if len(parents) < 1 || len(parents) > 240 {
		return GroupReport{}, errors.New("group_parent_count_invalid")
	}
	owners := make([]int, len(parents))
	for i := range owners {
		owners[i] = i
	}
	find := func(i int) int {
		for owners[i] != i {
			i = owners[i]
		}
		return i
	}
	r := GroupReport{Schema: "riido-combined-behavior-groups-v2", Parents: len(parents), MinimumGroups: 15, MinimumLabeledGroups: 15,
		Policy: "Transitive union of prototypes, declared behavioral cores, normalized requests/candidate captions, candidate sources including negatives, normalized candidate code, shared authored semantic component identities/exact declarations and normalized behavioral function/method copies. Pure type shapes or scalar enum layouts alone do not demonstrate copied behavior; exact types/globals still participate and all are pinned. Standard imports and observer infrastructure are excluded only from grouping, never provenance. The original v1 result is immutable; combined v2 relations may be stricter. Unknown parents retain all relation edges; only groups containing known/no-answer parents count toward the separate labeled minimum15. A minimum of15 operational or labeled groups does not establish statistical independence, role readiness, generalization or production eligibility."}
	var prototypes []string
	for i, a := range parents {
		if !validID(a.ID) || a.Prototype == "" || a.Core == "" {
			return GroupReport{}, errors.New("group_parent_identity_invalid")
		}
		prototypes = unique(prototypes, a.Prototype)
		for j := i + 1; j < len(parents); j++ {
			b := parents[j]
			if a.ID == b.ID {
				return GroupReport{}, errors.New("group_parent_duplicate")
			}
			var reasons []string
			if a.Prototype == b.Prototype {
				reasons = unique(reasons, "shared_prototype_paraphrase_or_counterexample")
			}
			if a.Core == b.Core {
				reasons = unique(reasons, "shared_core_template")
			}
			if lexicalhint.NormalizeText(a.Request) == lexicalhint.NormalizeText(b.Request) {
				reasons = unique(reasons, "normalized_request_duplicate")
			}
			for _, x := range a.CandidateTexts {
				for _, y := range b.CandidateTexts {
					if lexicalhint.NormalizeText(x) == lexicalhint.NormalizeText(y) {
						reasons = unique(reasons, "normalized_candidate_caption_duplicate")
					}
				}
			}
			for _, x := range a.Sources {
				for _, y := range b.Sources {
					if x.ID == y.ID {
						reasons = unique(reasons, "shared_candidate_source_including_negative")
					}
					if x.Core != "" && x.Core == y.Core {
						reasons = unique(reasons, "shared_candidate_core_template")
					}
					if x.NormalizedCodeSHA256 != "" && x.NormalizedCodeSHA256 == y.NormalizedCodeSHA256 {
						reasons = unique(reasons, "normalized_candidate_code_duplicate")
					}
					for _, cx := range x.Components {
						for _, cy := range y.Components {
							if cx.ID == cy.ID {
								reasons = unique(reasons, "shared_authored_semantic_component")
							}
							if cx.Kind == cy.Kind && cx.SHA256 != "" && cx.SHA256 == cy.SHA256 {
								reasons = unique(reasons, "exact_authored_declaration_copy")
							}
							if cx.NormalizedBehaviorSHA256 != "" && cx.NormalizedBehaviorSHA256 == cy.NormalizedBehaviorSHA256 {
								reasons = unique(reasons, "normalized_behavioral_component_copy")
							}
						}
					}
				}
			}
			if len(reasons) > 0 {
				r.Edges = append(r.Edges, behaviorprobe.GroupEdge{Left: a.ID, Right: b.ID, Reasons: reasons})
				owners[find(j)] = find(i)
			}
		}
	}
	for i, p := range parents {
		root := find(i)
		index := slices.IndexFunc(r.Groups, func(g behaviorprobe.Group) bool { return g.ID == root })
		if index < 0 {
			r.Groups = append(r.Groups, behaviorprobe.Group{ID: root})
			index = len(r.Groups) - 1
		}
		g := &r.Groups[index]
		g.Parents = append(g.Parents, p.ID)
		g.Prototypes = unique(g.Prototypes, p.Prototype)
		g.CoreTemplates = unique(g.CoreTemplates, p.Core)
		for _, s := range p.Sources {
			g.Sources = unique(g.Sources, s.ID)
		}
	}
	r.PrototypeCount = len(prototypes)
	r.ConnectedGroups = len(r.Groups)
	r.MeetsGroupMinimum = r.ConnectedGroups >= r.MinimumGroups
	return r, nil
}
