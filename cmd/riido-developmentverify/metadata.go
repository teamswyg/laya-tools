// SPDX-License-Identifier: Apache-2.0
package main

type retainedUnknown struct {
	Request   string `json:"request_id"`
	Candidate int    `json:"source_candidate_position"`
	Fixture   int    `json:"fixture"`
	Field     string `json:"field"`
	Reason    string `json:"reason"`
}
type materialization struct {
	Schema         string            `json:"schema"`
	State          string            `json:"state"`
	Failure        string            `json:"failure_code"`
	Requests       int               `json:"requests"`
	Labels         int               `json:"candidate_labels"`
	Positive       int               `json:"positive_labels"`
	Negative       int               `json:"negative_labels"`
	Inputs         int               `json:"frozen_input_fixtures"`
	Selected       int               `json:"selected_candidate_observations"`
	Full           int               `json:"original_evidence_observations"`
	PositiveIndex1 int               `json:"positive_selected_index1_requests"`
	Prefix         bool              `json:"prior23_prefix_byte_exact"`
	RootCopied     bool              `json:"Root_adopted_supervision_copied"`
	GroupSHA       string            `json:"external_group_decision_sha256"`
	Unknowns       []retainedUnknown `json:"new_retained_unknown_predicates"`
	UnknownFalse   bool              `json:"unknown_converted_to_false"`
	DataSHA        string            `json:"data_sha256"`
	DataBytes      int               `json:"data_bytes"`
	Dispatches     int               `json:"LoadDevelopmentRow_dispatch_intents"`
	Returns        int               `json:"LoadDevelopmentRow_returns"`
	Matches        int               `json:"LoadDevelopmentRow_value_matches"`
	Reserved       bool              `json:"output_exclusively_reserved"`
	Written        int               `json:"output_written_bytes"`
	FileSynced     bool              `json:"output_file_sync_returned"`
	DirSynced      bool              `json:"output_directory_sync_returned"`
	LabelsAssigned bool              `json:"labels_assigned_by_preparer"`
	Qualified      bool              `json:"qualified_by_preparer"`
	Features       int               `json:"Features_calls"`
	Score          int               `json:"Score_calls"`
	Project        int               `json:"Project_calls"`
	Fit            int               `json:"Fit_calls"`
	Original       int               `json:"original_candidate_calls"`
	Model          int               `json:"model_calls"`
	Scope          string            `json:"scope"`
}

var expectedUnknowns = [8]retainedUnknown{
	{"next60-afero-sub-validation", 0, 0, "error_type", "observable_unavailable"},
	{"next60-afero-sub-validation", 2, 0, "error_type", "observable_unavailable"},
	{"next60-ini-quoted-comments", 2, 0, "comment", "observable_unavailable_preserved"},
	{"next60-ini-quoted-comments", 2, 0, "value", "observable_unavailable_preserved"},
	{"next60-ini-delete-index", 0, 0, "indexes_unchanged", "observable_unavailable_preserved"},
	{"next60-ini-delete-index", 0, 0, "sections", "observable_unavailable_preserved"},
	{"next60-ini-byte-budget", 0, 4, "error_kind_not", "unavailable_or_unobserved_saved_channel"},
	{"next60-afero-exclusive-write", 0, 1, "exclusive_open", "unavailable_or_unobserved_saved_channel"},
}

func verifyMetadata(raw []byte) error {
	if len(raw) != metadataBytes || digest(raw) != metadataSHA {
		return failure("metadata_pin")
	}
	var m materialization
	if !decode(raw, &m) {
		return failure("metadata_json")
	}
	if m.Schema != "riido-next60-27-or30-data-correspondence-v1" || m.State != "materialized_Root_adopted_data_only" || m.Failure != "" || m.Requests != 30 || m.Labels != 88 || m.Positive != 30 || m.Negative != 58 || m.Inputs != 146 || m.Selected != 430 || m.Full != 434 || m.PositiveIndex1 != 27 || !m.Prefix || !m.RootCopied || m.GroupSHA != "5c82aa14b01f34bc4af08e06c0d565eb5459cee27e55db82a2304e3c6d28f00d" || len(m.Unknowns) != len(expectedUnknowns) || m.UnknownFalse || m.DataSHA != dataSHA || m.DataBytes != dataBytes || m.Dispatches != 0 || m.Returns != 0 || m.Matches != 0 || !m.Reserved || m.Written != dataBytes || !m.FileSynced || !m.DirSynced || m.LabelsAssigned || m.Qualified || m.Features != 0 || m.Score != 0 || m.Project != 0 || m.Fit != 0 || m.Original != 0 || m.Model != 0 || m.Scope != "Copy pinned Root finite supervision and external source-group decision only; no truth, rights, unseen-source or training-readiness authority." {
		return failure("metadata_correspondence")
	}
	for i, want := range expectedUnknowns {
		if m.Unknowns[i] != want {
			return failure("metadata_unknown_preservation")
		}
	}
	return nil
}
