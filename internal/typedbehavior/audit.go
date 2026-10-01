package typedbehavior

import (
	"encoding/json"
	"errors"
	"slices"

	"github.com/teamswyg/laya-tools/internal/behaviorprobe"
)

type SourceArtifact struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
}
type ControlTruth struct {
	SourceID         string `json:"source_id"`
	ExpectedControl  string `json:"expected_control"`
	VectorsChecked   int    `json:"vectors_checked"`
	FailedVectors    int    `json:"failed_vectors"`
	TruthTableSHA256 string `json:"truth_table_sha256"`
}
type Report struct {
	Schema                        string                  `json:"schema"`
	DatasetSHA256                 string                  `json:"typed_dataset_sha256"`
	SourceArtifacts               []SourceArtifact        `json:"source_artifacts"`
	Parents                       int                     `json:"parents"`
	Candidates                    int                     `json:"candidates"`
	KnownParents                  int                     `json:"known_parents"`
	UnknownParents                int                     `json:"unknown_parents"`
	NoAnswerParents               int                     `json:"no_answer_parents"`
	PrototypeCount                int                     `json:"prototype_count"`
	NaturalLanguageReviewRequired bool                    `json:"natural_language_review_required"`
	TrainingReady                 bool                    `json:"training_ready"`
	Partitioned                   bool                    `json:"partitioned"`
	FinalEligible                 bool                    `json:"final_eligible"`
	Outcomes                      []behaviorprobe.Outcome `json:"outcomes"`
	Sources                       []SourcePin             `json:"sources"`
	Controls                      []ControlTruth          `json:"controls"`
	Contracts                     []ContractSpec          `json:"contracts"`
	InfrastructurePolicy          string                  `json:"infrastructure_policy"`
	Limitations                   []string                `json:"limitations"`
}

func Contracts() []ContractSpec {
	return append(stateContractEvidence(), flowContractEvidence()...)
}

func checkSource(id string) controlCheck {
	if slices.ContainsFunc(sourcesState(), func(s sourceSpec) bool { return s.id == id }) {
		return checkState(id)
	}
	return checkFlow(id)
}

func Evaluate(d Dataset) (Report, error) {
	pins, e := SourcePins()
	if e != nil {
		return Report{}, e
	}
	if e = validateDataset(d, pins); e != nil {
		return Report{}, e
	}
	raw, e := json.Marshal(d)
	if e != nil {
		return Report{}, errors.New("typed_dataset_encoding_failed")
	}
	r := Report{
		Schema: "riido-typed-behavior-truth-v2", DatasetSHA256: sum(raw), Parents: len(d.Parents),
		NaturalLanguageReviewRequired: true, Sources: pins, Contracts: Contracts(),
		SourceArtifacts:      SourceArtifacts(),
		InfrastructurePolicy: "Allowlisted Go standard imports and the observer infrastructure do not connect task groups. They remain immutable provenance. Every referenced authored global, type, method and helper is a semantic component; unresolved authored closure stops the audit. Candidate code normalization remains conservative. Different implementations in one coordinated authoring pipeline do not constitute independent data populations.",
		Limitations: []string{
			"Finite authored literal vectors demonstrate scoped observations, not global correctness or held-out semantic generalization.",
			"Source implementation fidelity and natural-language request/candidate fidelity require separate independent reviews.",
			"The coordinated synthetic authoring pipeline and English captions are not diverse production provenance.",
			"Safely observed no-panic contract violations are mismatches; unreliable observers are unknown and cannot provide boolean labels.",
			"No roles, splits, fits, model calls, weights, protected-final access or activation are created by this audit.",
		},
	}
	checks := make([]controlCheck, len(pins))
	for i, s := range sourceSpecs() {
		contract := slices.IndexFunc(r.Contracts, func(c ContractSpec) bool { return c.Prototype == s.prototype })
		if contract < 0 || pins[i].ID != s.id {
			return Report{}, errors.New("typed_registry_contract_mismatch")
		}
		c := checkSource(s.id)
		if c.Unknown || c.Checked != r.Contracts[contract].Vectors || c.Failed < 0 || c.Failed > c.Checked || s.correct && c.Failed != 0 || !s.correct && c.Failed == 0 {
			return Report{}, errors.New("typed_literal_control_not_verified")
		}
		checks[i] = c
		role := "wrong"
		if s.correct {
			role = "correct"
		}
		r.Controls = append(r.Controls, ControlTruth{s.id, role, c.Checked, c.Failed, r.Contracts[contract].TruthTableSHA256})
	}
	var prototypes []string
	for _, p := range d.Parents {
		if !slices.Contains(prototypes, p.Prototype) {
			prototypes = append(prototypes, p.Prototype)
		}
		r.Candidates += len(p.Candidates)
		o := behaviorprobe.Outcome{ParentID: p.ID, State: "known", Acceptable: []int{}, Candidates: []behaviorprobe.CandidateTruth{}}
		if p.ContractID == "unknown-ambiguous" || p.ContractID == "unknown-incomplete-caption" {
			o.State, o.Reason = "unknown", "request_policy_ambiguous_no_literal_oracle"
			if p.ContractID == "unknown-incomplete-caption" {
				o.Reason = "bounded_caption_does_not_express_complete_contract"
			}
			for _, c := range p.Candidates {
				o.Candidates = append(o.Candidates, behaviorprobe.CandidateTruth{CandidateID: c.ID, State: "unknown"})
			}
			r.UnknownParents++
		} else {
			for index, c := range p.Candidates {
				s := slices.IndexFunc(pins, func(s SourcePin) bool { return s.ID == c.SourceID })
				observed := checks[s]
				state := "rejected"
				if observed.Unknown {
					state, o.State, o.Reason = "unknown", "unknown", "candidate_observer_unreliable"
				} else if observed.Failed == 0 {
					state = "acceptable"
					o.Acceptable = append(o.Acceptable, index)
				}
				o.Candidates = append(o.Candidates, behaviorprobe.CandidateTruth{CandidateID: c.ID, State: state, VectorsChecked: observed.Checked, FailedVectors: observed.Failed})
			}
			if o.State == "unknown" {
				o.Acceptable = []int{}
				r.UnknownParents++
			} else {
				r.KnownParents++
				if len(o.Acceptable) == 0 {
					o.State = "no_answer"
					r.NoAnswerParents++
				}
			}
		}
		r.Outcomes = append(r.Outcomes, o)
	}
	r.PrototypeCount = len(prototypes)
	return r, nil
}
