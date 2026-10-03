// Copyright 2026 teamswyg. SPDX-License-Identifier: Apache-2.0
// Inert source. Root supplies and hashes config; no model or original imports.
package main

import (
	"bytes"
	"errors"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
	"github.com/teamswyg/laya-tools/pkg/shortclaimdata"
	"io"
	"os"
	data35 "riido.local/development35prep"
	"runtime"
	"runtime/debug"
)

func load(raw []byte) (data35.Loaded, error) {
	return bridge(shortclaimdata.LoadDevelopmentRow(bytes.NewReader(raw)))
}

// bridge accepts an already returned result, so error accounting is tested without a Reader call.
func bridge(ex shortclaimdata.Example, e error) (data35.Loaded, error) {
	if e != nil {
		return data35.Loaded{ReaderReturned: true}, errors.New("reader_failed")
	}
	if !ex.Valid() {
		return data35.Loaded{ReaderReturned: true}, errors.New("reader_invalid")
	}
	m, p, s := ex.Metadata(), ex.Input().Prepared(), ex.Supervision()
	v := data35.Loaded{Valid: true, UnusedZero: true, ReaderReturned: true, Schema: p.Schema, Provenance: p.Provenance, ID: m.StableID, Request: p.Request, Role: m.Role, Family: m.SourceFamily, Revision: m.SourceRevision, TextRevision: m.TextRevision, Policy: m.FeaturePolicy, ScopeDescription: m.FiniteScope.Description, Group: m.WholeGroup, Inputs: m.FiniteScope.InputCount, Observations: m.FiniteScope.CandidateObservations, Count: p.Count, SupervisionCount: s.Count, All: m.FiniteScope.AllInputGuarantee, Unseen: m.FiniteScope.UnseenSourceGuarantee}
	if p.Count < 1 || p.Count > 8 || s.Count != p.Count {
		return v, errors.New("reader_count")
	}
	var candidates [8]shortclaim.Candidate
	for j, c := range p.Candidates {
		if j < p.Count {
			v.Candidates[j] = data35.LoadedCandidate{ID: c.ID, Text: c.Text, Label: s.Labels[j], Weight: s.Weights[j]}
			candidates[j] = shortclaim.Candidate{ID: c.ID, Text: c.Text}
		} else if c.ID != "" || c.Text != "" || c.NormalizedText != "" || s.Labels[j] || s.Weights[j] != 0 {
			v.UnusedZero = false
		}
	}
	// The extra 35 ValidateInput calls check all Prepared normalized strings/fields.
	// LoadDevelopmentRow already internally validates once; these are separate.
	v.ValidateAttempted = true
	expected, e := shortclaim.ValidateInput(shortclaim.Input{Schema: shortclaim.Schema, Request: p.Request, Candidates: candidates[:p.Count], Provenance: shortclaimdata.InputProvenance})
	if e != nil || p != expected.Prepared() {
		return v, errors.New("reader_prepared_mismatch")
	}
	return v, nil
}
func run(args []string) error {
	v, e := data35.Arguments(args, []string{"--config", "--config-sha256", "--out"})
	if e != nil {
		return e
	}
	c, e := data35.LoadConfig(v[0], v[1])
	if e != nil {
		return e
	}
	// Reserve empty exclusive data+receipt files and sync their directory BEFORE
	// reading adoption, accessing labels, checking source inputs or using Reader.
	reservation, e := data35.Reserve(v[2], c)
	if reservation != nil {
		defer reservation.Close()
	}
	if e != nil {
		return e
	}
	receipt := data35.NewReceipt(v[1])
	in, e := data35.Load(c)
	var data []byte
	if e == nil {
		var rows [35]data35.Row
		data, rows, e = data35.Compose(in, &receipt)
		if e == nil {
			e = data35.VerifyReader(data, rows, &receipt, load)
		}
	}
	return reservation.Finish(data, &receipt, e)
}
func main() {
	runtime.GOMAXPROCS(1)
	debug.SetMemoryLimit(256 << 20)
	if e := run(os.Args[1:]); e != nil {
		io.WriteString(os.Stderr, data35.FailureCode(e)+"\n")
		os.Exit(1)
	}
	io.WriteString(os.Stdout, "data35_and_35_project_Reader_matches_complete_no_model\n")
}
