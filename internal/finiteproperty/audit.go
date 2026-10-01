package finiteproperty

import (
	"errors"

	"github.com/teamswyg/laya-tools/internal/typedbehavior"
)

type LiteralResult struct {
	Input   string      `json:"input"`
	Want    Observation `json:"want"`
	Got     Observation `json:"got"`
	Matches bool        `json:"matches"`
}

type Observation = typedbehavior.PropertyQuotedObservation

type SourceResult struct {
	State            string           `json:"state"`
	PropertyID       string           `json:"property_id"`
	SourceID         string           `json:"source_id"`
	LiteralCount     int              `json:"literal_count"`
	ObservedLiterals int              `json:"observed_literals"`
	Literals         [5]LiteralResult `json:"literals"`
	Matches          bool             `json:"matches_all_finite_literals"`
}

type CandidateResult struct {
	ID            string `json:"id"`
	SourceID      string `json:"source_id"`
	EvidenceIndex int    `json:"evidence_index"`
	Label         string `json:"finite_behavior_label"`
}

type RowResult struct {
	ID               string             `json:"id"`
	OriginalParentID string             `json:"original_parent_id"`
	Candidates       [3]CandidateResult `json:"candidates"`
	MatchingCount    int                `json:"matching_candidate_count"`
	Status           string             `json:"finite_status"`
}

type Results struct {
	Schema                      string           `json:"schema"`
	State                       string           `json:"state"`
	DatasetSHA256               string           `json:"dataset_sha256"`
	SourceResults               [12]SourceResult `json:"source_results"`
	Rows                        [12]RowResult    `json:"rows"`
	ActualCandidateCalls        int              `json:"actual_candidate_calls"`
	ObserverDispatchAttempts    int              `json:"observer_dispatch_attempts"`
	CompletedObservations       int              `json:"completed_observations"`
	ObserverFailures            int              `json:"observer_failures"`
	UniqueSourceInputs          int              `json:"unique_source_input_bindings"`
	CandidateLabels             int              `json:"projected_row_candidate_labels"`
	NewIndependentParents       int              `json:"new_independent_parents"`
	PreflightTextNormalizations int              `json:"preflight_text_bounds_normalizations"`
	RankingFeatureExtractions   int              `json:"ranking_feature_extractions"`
	RankingRuns                 int              `json:"ranking_runs"`
	Fits                        int              `json:"fits"`
	ModelCalls                  int              `json:"model_calls"`
	ProtectedFinalRead          bool             `json:"protected_final_read"`
	TrainingReady               bool             `json:"training_ready"`
}

// ExactMatch includes success/error identity, Count, the whole fixed array and
// no panic. Zero output with an omitted syntax error is still a mismatch.
func ExactMatch(got, want Observation) bool { return got == want }

// Audit runs only the closed compiled observations, once per source/literal.
// It does not extract text features, rank, use whole-function roles, or fit.
// Unsupported adapter bindings stop; no false label is created from a failure.
func Audit(d Dataset) (Results, error) {
	if err := Validate(d); err != nil {
		return Results{}, err
	}
	// Validate performs preparation/bounds checks before candidate observations.
	// Those text bounds are not part of the outcome loop or a ranking feature.
	raw, err := JSON(d)
	if err != nil {
		return Results{}, err
	}
	out := Results{Schema: "riido-finite-property-observation-56e-v2", State: "incomplete", DatasetSHA256: SHA256(raw), PreflightTextNormalizations: 192}
	for property, def := range d.Oracle {
		for source, id := range d.ConnectedSources {
			index := property*4 + source
			e := SourceResult{State: "incomplete", PropertyID: d.Parents[property*4].PropertyID, SourceID: id, LiteralCount: def.LiteralCount, Matches: true}
			for i := 0; i < def.LiteralCount; i++ {
				v := def.Literals[i]
				out.ObserverDispatchAttempts++
				out.UniqueSourceInputs++
				got, err := typedbehavior.ObserveQuotedProperty(id, v.Input)
				if err != nil {
					out.ObserverFailures++
					out.State = "observer_failed"
					e.State = "observer_failed"
					out.SourceResults[index] = e
					return out, errors.New("finite_observer_unsupported")
				}
				matches := ExactMatch(got, v.Want)
				e.Literals[i] = LiteralResult{Input: v.Input, Want: v.Want, Got: got, Matches: matches}
				e.Matches = e.Matches && matches
				out.CompletedObservations++
				out.ActualCandidateCalls++
				e.ObservedLiterals++
				out.SourceResults[index] = e
			}
			e.State = "complete"
			out.SourceResults[index] = e
		}
	}
	for i, parent := range d.Parents {
		row := RowResult{ID: parent.ID, OriginalParentID: parent.OriginalParentID}
		for j, candidate := range parent.Candidates {
			index := -1
			for k, evidence := range out.SourceResults {
				if evidence.SourceID == candidate.SourceID && evidence.PropertyID == parent.PropertyID {
					index = k
				}
			}
			if index < 0 {
				out.State = "evidence_binding_failed"
				return out, errors.New("finite_evidence_binding_missing")
			}
			label := "mismatches_finite_literals"
			if out.SourceResults[index].Matches {
				label = "matches_finite_literals"
				row.MatchingCount++
			}
			row.Candidates[j] = CandidateResult{candidate.ID, candidate.SourceID, index, label}
			out.CandidateLabels++
		}
		row.Status = "finite_answerable"
		if row.MatchingCount == 0 {
			row.Status = "finite_no_answer"
		}
		out.Rows[i] = row
	}
	out.State = "complete"
	return out, nil
}
