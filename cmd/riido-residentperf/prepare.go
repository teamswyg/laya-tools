package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"runtime"
	"slices"

	"github.com/teamswyg/laya-tools/internal/behaviorprobe"
	"github.com/teamswyg/laya-tools/internal/typedbehavior"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

const (
	featureEncoding = "tagged_count_uint64be_length_ordered_normalized_request_candidates_v1"
	wireEncoding    = "base64_exact_compact_json_without_lf_v1"
)

// Takes already read bounded bytes, verifies exactly those bytes, then decodes
// them. Never calls Evaluate, an observer, a candidate, or Load(path).
// The exact JSON hashes identify frozen public artifacts. Source provenance
// verification is a separate preparation gate; these are not truth labels.
func prepareCorpus(legacyRaw, typedRaw []byte) (Corpus, error) {
	if len(legacyRaw) > 64<<20 || len(typedRaw) > 64<<20 || hashBytes(legacyRaw) != legacySHA || hashBytes(typedRaw) != typedSHA {
		return Corpus{}, errors.New("original_fixture_hash_mismatch")
	}
	var legacy behaviorprobe.Dataset
	var typed typedbehavior.Dataset
	if strictJSON(legacyRaw, &legacy) != nil || strictJSON(typedRaw, &typed) != nil || legacy.Schema != behaviorprobe.DatasetSchema || legacy.Origin != behaviorprobe.DatasetOrigin || len(legacy.Parents) != 48 || typed.Schema != typedbehavior.Schema || typed.Origin != typedbehavior.Origin || len(typed.Parents) != 24 {
		return Corpus{}, errors.New("original_fixture_shape_mismatch")
	}
	legacyInputs, typedInputs := behaviorprobe.FeatureInputs(legacy), typedbehavior.FeatureInputs(typed)
	out := Corpus{Schema: corpusSchema, State: "prepared_protocol_only_no_labels", OS: runtime.GOOS, Arch: runtime.GOARCH, LegacyInputSHA256: legacySHA, TypedInputSHA256: typedSHA, FeatureEncoding: featureEncoding, WireEncoding: wireEncoding, OriginalParents: 72}
	for i, input := range legacyInputs {
		ids := make([]string, len(legacy.Parents[i].Candidates))
		for j, candidate := range legacy.Parents[i].Candidates {
			ids[j] = candidate.ID
		}
		if e := appendProjection(&out, "probes-56", legacy.Parents[i].ID, i, input.Request, input.Candidates, ids); e != nil {
			return Corpus{}, e
		}
	}
	for i, input := range typedInputs {
		ids := make([]string, len(typed.Parents[i].Candidates))
		for j, candidate := range typed.Parents[i].Candidates {
			ids[j] = candidate.ID
		}
		if e := appendProjection(&out, "probes-56b", typed.Parents[i].ID, i, input.Request, input.Candidates, ids); e != nil {
			return Corpus{}, e
		}
	}
	if e := validateCorpus(out); e != nil {
		return Corpus{}, e
	}
	return out, nil
}

// This helper has no alternate fixture acceptance path. The production caller
// has already checked both exact original fixture hashes. Authored unit inputs
// exercise deduplication without a new cohort/audit. Original IDs are metadata;
// only Request and ordered Text enter features.
func appendProjection(out *Corpus, fixture, id string, index int, request string, texts, originalIDs []string) error {
	if out == nil || len(out.Origins) >= 72 || index < 0 || !originID(id) || len(texts) < 1 || len(texts) > 8 || len(originalIDs) != len(texts) {
		return errors.New("projected_origin_invalid")
	}
	link := OriginLink{Fixture: fixture, ParentID: id, OriginalIndex: index, CandidateCount: len(texts), PayloadIndex: -1}
	for i, candidateID := range originalIDs {
		if !originID(candidateID) || slices.Contains(originalIDs[:i], candidateID) {
			return errors.New("projected_origin_invalid")
		}
		link.OriginalCandidateIDs[i] = candidateID
	}
	if len(texts) != 3 {
		link.ExcludedReason = "original_candidate_count_not_three"
		out.Origins = append(out.Origins, link)
		return nil
	}
	ids := candidateIDs()
	input := shortclaim.Input{Schema: textSchema, Request: request, Provenance: provenance, Candidates: make([]shortclaim.Candidate, 3)}
	for i, text := range texts {
		input.Candidates[i] = shortclaim.Candidate{ID: ids[i], Text: text}
	}
	p, e := shortclaim.Validate(input)
	if e != nil {
		return errors.New("projected_text_bounds_invalid")
	}
	raw, e := json.Marshal(input)
	if e != nil || len(raw)+1 > wireLimit {
		return errors.New("projected_wire_bounds_invalid")
	}
	link.FeatureSHA256 = featureSHA(p)
	// Stable O(n^2) bounded scan: at most 72 inputs, no map or cache. Keep the
	// first representative's exact raw wire, including its original raw text.
	// Lexical normalization drops punctuation: this is not a future cache key
	// for raw rule parsing, a semantic oracle, or any model.
	for i := range out.Payloads {
		if out.Payloads[i].FeatureSHA256 == link.FeatureSHA256 {
			link.PayloadIndex = i
			out.Origins = append(out.Origins, link)
			return nil
		}
	}
	payload := Payload{Wire: slices.Clone(raw), FeatureSHA256: link.FeatureSHA256, WireSHA256: hashBytes(append(slices.Clone(raw), '\n')), InputDigest: inputDigest(p)}
	payload.RawTextBytes[0], payload.NormalizedTextBytes[0], payload.NormalizedWords[0] = len(p.Request), len(p.NormalizedRequest), words(p.NormalizedRequest)
	for i := 0; i < 3; i++ {
		payload.RawTextBytes[i+1], payload.NormalizedTextBytes[i+1], payload.NormalizedWords[i+1] = len(p.Candidates[i].Text), len(p.Candidates[i].NormalizedText), words(p.Candidates[i].NormalizedText)
	}
	for i, kind := range baselineKinds() {
		_, payload.Expected[i], e = expectedResponse(p, kind)
		if e != nil {
			return e
		}
	}
	link.PayloadIndex = len(out.Payloads)
	out.Payloads, out.Origins = append(out.Payloads, payload), append(out.Origins, link)
	return nil
}

func originID(s string) bool {
	if len(s) < 1 || len(s) > shortclaim.MaxIDBytes {
		return false
	}
	for i, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || i > 0 && (c == '-' || c == '_' || c == '.')) {
			return false
		}
	}
	return true
}

// Readiness is protocol-only. Deterministic heuristic expectations are checked
// outside all official child/controller replay timing. No candidate source is
// run, and nothing here certifies semantic correctness or usefulness.
func validateCorpus(c Corpus) error {
	if c.Schema != corpusSchema || c.State != "prepared_protocol_only_no_labels" || c.OS != runtime.GOOS || c.Arch != runtime.GOARCH || c.LegacyInputSHA256 != legacySHA || c.TypedInputSHA256 != typedSHA || c.OriginalParents != 72 || len(c.Origins) != 72 || len(c.Payloads) < 2 || len(c.Payloads) > 72 || c.FeatureEncoding != featureEncoding || c.WireEncoding != wireEncoding {
		return errors.New("corpus_identity_invalid")
	}
	ids := candidateIDs()
	for i, payload := range c.Payloads {
		if len(payload.Wire)+1 > wireLimit {
			return errors.New("corpus_wire_bounds_invalid")
		}
		p, e := shortclaim.Load(bytes.NewReader(payload.Wire))
		if e != nil || p.Count != 3 || p.Provenance != provenance || featureSHA(p) != payload.FeatureSHA256 || inputDigest(p) != payload.InputDigest || hashBytes(append(slices.Clone(payload.Wire), '\n')) != payload.WireSHA256 {
			return errors.New("corpus_payload_binding_invalid")
		}
		input := shortclaim.Input{Schema: p.Schema, Request: p.Request, Provenance: p.Provenance, Candidates: make([]shortclaim.Candidate, 3)}
		for j := 0; j < 3; j++ {
			if p.Candidates[j].ID != ids[j] {
				return errors.New("corpus_metadata_invalid")
			}
			input.Candidates[j] = shortclaim.Candidate{ID: ids[j], Text: p.Candidates[j].Text}
		}
		canonical, e := json.Marshal(input)
		if e != nil || !bytes.Equal(canonical, payload.Wire) {
			return errors.New("corpus_wire_encoding_invalid")
		}
		for j := 0; j < i; j++ {
			if c.Payloads[j].FeatureSHA256 == payload.FeatureSHA256 {
				return errors.New("corpus_feature_duplicate")
			}
		}
		if payload.RawTextBytes[0] != len(p.Request) || payload.NormalizedTextBytes[0] != len(p.NormalizedRequest) || payload.NormalizedWords[0] != words(p.NormalizedRequest) {
			return errors.New("corpus_text_distribution_invalid")
		}
		for j := 0; j < 3; j++ {
			if payload.RawTextBytes[j+1] != len(p.Candidates[j].Text) || payload.NormalizedTextBytes[j+1] != len(p.Candidates[j].NormalizedText) || payload.NormalizedWords[j+1] != words(p.Candidates[j].NormalizedText) {
				return errors.New("corpus_text_distribution_invalid")
			}
		}
		for j, kind := range baselineKinds() {
			_, expected, e := expectedResponse(p, kind)
			if e != nil || expected != payload.Expected[j] {
				return errors.New("corpus_response_expectation_invalid")
			}
		}
	}
	// Bounded shape/binding checks cannot authenticate parent IDs. The caller
	// must also replay projection from both exact original byte streams and
	// compare the canonical corpus before use.
	var referenced [72]bool
	nextPayload := 0
	for i, link := range c.Origins {
		fixture, index := "probes-56", i
		if i >= 48 {
			fixture, index = "probes-56b", i-48
		}
		if link.Fixture != fixture || link.OriginalIndex != index || !originID(link.ParentID) || link.CandidateCount < 1 || link.CandidateCount > 8 {
			return errors.New("corpus_origin_shape_invalid")
		}
		for j := 0; j < i; j++ {
			if c.Origins[j].Fixture == fixture && c.Origins[j].ParentID == link.ParentID {
				return errors.New("corpus_origin_shape_invalid")
			}
		}
		for j, candidateID := range link.OriginalCandidateIDs {
			if j >= link.CandidateCount {
				if candidateID != "" {
					return errors.New("corpus_origin_metadata_invalid")
				}
				continue
			}
			if !originID(candidateID) || slices.Contains(link.OriginalCandidateIDs[:j], candidateID) {
				return errors.New("corpus_origin_metadata_invalid")
			}
		}
		if link.CandidateCount != 3 {
			if link.PayloadIndex != -1 || link.ExcludedReason != "original_candidate_count_not_three" || link.FeatureSHA256 != "" {
				return errors.New("corpus_exclusion_invalid")
			}
			continue
		}
		if link.PayloadIndex < 0 || link.PayloadIndex >= len(c.Payloads) || link.ExcludedReason != "" || link.FeatureSHA256 != c.Payloads[link.PayloadIndex].FeatureSHA256 {
			return errors.New("corpus_selection_invalid")
		}
		if !referenced[link.PayloadIndex] {
			if link.PayloadIndex != nextPayload {
				return errors.New("corpus_selection_order_invalid")
			}
			referenced[link.PayloadIndex] = true
			nextPayload++
		}
	}
	if nextPayload != len(c.Payloads) {
		return errors.New("corpus_selection_unreferenced")
	}
	return nil
}
