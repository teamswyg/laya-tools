// Copyright 2026 teamswyg. Licensed under the Apache License, Version 2.0.
package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	core "riido.local/next60gjsontwooutside/twoliteral"
)

type CorePin struct {
	ID     string `json:"id"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}
type PlanRef struct {
	Path string  `json:"path"`
	Pin  CorePin `json:"pin"`
}

// Field order and tags match the sealed workerio.Plan exactly.
type CountPlan struct {
	Schema               string  `json:"schema"`
	Frozen               bool    `json:"frozen"`
	Literal              PlanRef `json:"literal"`
	Wants                PlanRef `json:"wants"`
	Captions             PlanRef `json:"captions"`
	RootWantFreeze       PlanRef `json:"Root_want_freeze"`
	SourceCorrection     PlanRef `json:"source_correction"`
	RootFamilyDecision   PlanRef `json:"Root_family_decision"`
	SourceManifest       PlanRef `json:"source_manifest"`
	RootCompilerReview   PlanRef `json:"Root_compiler_review"`
	RootSourceReview     PlanRef `json:"Root_source_review"`
	RootResourceReview   PlanRef `json:"Root_resource_review"`
	RootExecutionBinding PlanRef `json:"Root_execution_binding"`
	Worker               CorePin `json:"worker"`
	Controller           CorePin `json:"controller"`
	RequestOrder         []int   `json:"request_order"`
	Mapping              [][]int `json:"display_to_internal"`
	Fixtures             []int   `json:"fixture_counts"`
	MethodMaximum        int     `json:"method_maximum"`
	CallbackMaximum      int     `json:"callback_maximum"`
	FrameMaximum         int     `json:"frame_maximum"`
	OutputReserve        int     `json:"output_reserve_bytes"`
}

func refPin(r PlanRef) Pin {
	return Pin{ID: r.Pin.ID, Path: r.Path, Bytes: r.Pin.Bytes, SHA256: r.Pin.SHA256}
}
func sameFilePin(a, b Pin) bool {
	return a.Path == b.Path && a.Bytes == b.Bytes && a.SHA256 == b.SHA256
}
func planRefs(p CountPlan) []PlanRef {
	return []PlanRef{p.Literal, p.Wants, p.Captions, p.RootWantFreeze, p.SourceCorrection, p.RootFamilyDecision, p.SourceManifest, p.RootCompilerReview, p.RootSourceReview, p.RootResourceReview, p.RootExecutionBinding}
}
func loadCountPlan(c Config) (p CountPlan, e error) {
	b, e := checkPin(c.Plan, metadataCap)
	if e != nil {
		return
	}
	if e = strictCanonical(b, &p); e != nil {
		return
	}
	if p.Schema != "riido-gjson-two-native-plan-v1" || !p.Frozen || len(p.RequestOrder) != 2 || p.RequestOrder[0] != 21 || p.RequestOrder[1] != 22 || len(p.Mapping) != 2 || len(p.Mapping[0]) != 3 || len(p.Mapping[1]) != 3 || len(p.Fixtures) != 2 || p.Fixtures[0] != 5 || p.Fixtures[1] != 6 || p.MethodMaximum != 61 || p.CallbackMaximum != 33 || p.FrameMaximum != 384 || p.OutputReserve != 2097152 {
		return p, errCode("plan_count_contract")
	}
	for q, m := range [2][3]int{{0, 2, 1}, {1, 2, 0}} {
		for j, x := range m {
			if p.Mapping[q][j] != x {
				return p, errCode("plan_mapping")
			}
		}
	}
	exact := []struct {
		r PlanRef
		b int64
		h string
	}{
		{p.Literal, 6142, "bbaf3ad1856a66809ec51818a347b3bb46734d68a7e0bee2a36591879e7ef560"},
		{p.Wants, 5714, "6cef54345a729233d4de7591b513227b0bcc8fe463d471e3ffd0e12215652236"},
		{p.Captions, 2935, "a0cc56b4058d3b3d24325abbf863bd0ac3c26a578305cacf0b9fa962c24d8c5f"},
		{p.RootWantFreeze, 6889, "87285c2c5efe2f5de7539c96934e36629a35cdcb432f45b11361462c09dd40c2"},
		{p.SourceCorrection, 8510, "b5144e78100a9c4b235812c25bb13dc880fcb8ed4f4bbc2ccf85942ffbd8cf43"},
		{p.RootFamilyDecision, 3089, "1677d33aeb32c18c809e9ea717f375d5086d5b371e68f09f48fd24ae2f254b7f"},
	}
	for _, x := range exact {
		if x.r.Pin.Bytes != x.b || x.r.Pin.SHA256 != x.h {
			return p, errCode("plan_fixed_pin")
		}
	}
	if !sameFilePin(refPin(p.Captions), c.Captions) || !sameFilePin(refPin(p.Wants), c.Wants) || p.Worker.ID == "" || p.Controller.ID == "" || p.Worker.Bytes != c.Worker.Pin.Bytes || p.Worker.SHA256 != c.Worker.Pin.SHA256 || p.Controller.Bytes != c.Controller.Pin.Bytes || p.Controller.SHA256 != c.Controller.Pin.SHA256 {
		return p, errCode("plan_binary_config_binding")
	}
	for i, v := range []PlanRef{p.SourceManifest, p.RootCompilerReview, p.RootSourceReview, p.RootResourceReview} {
		if !sameFilePin(refPin(v), c.RootReceipts[i]) {
			return p, errCode("plan_review_binding")
		}
	}
	refs := planRefs(p)
	for i, v := range refs {
		for _, old := range refs[:i] {
			if old.Path == v.Path || equalASCII(old.Path, v.Path) {
				return p, errCode("plan_ref_alias")
			}
		}
		for _, other := range []Pin{c.Plan, c.Worker.Pin, c.Controller.Pin, c.Time} {
			if v.Path == other.Path || equalASCII(v.Path, other.Path) {
				return p, errCode("plan_binary_alias")
			}
		}
		if v.Path == c.Output || inside(v.Path, c.Output) || inside(c.Output, v.Path) {
			return p, errCode("plan_output_overlap")
		}
		if _, e = checkPin(refPin(v), evidenceCap); e != nil {
			return
		}
	}
	return p, nil
}

type childReservation struct {
	Schema                 string      `json:"schema"`
	PlanSHA256             string      `json:"plan_sha256"`
	Literal                CorePin     `json:"literal"`
	Wants                  CorePin     `json:"wants"`
	Captions               CorePin     `json:"captions"`
	RootFreeze             CorePin     `json:"Root_freeze"`
	Rows                   int         `json:"rows"`
	Methods                int         `json:"methods"`
	Callbacks              int         `json:"callbacks"`
	Frames                 int         `json:"frames"`
	OutputReserve          int         `json:"output_reserve_bytes"`
	InitialCounts          core.Counts `json:"initial_counts"`
	OriginalInitialization *int        `json:"original_initialization"`
	OriginalNested         *int        `json:"original_nested"`
}

func checkChildReservation(attempt string, p CountPlan, sha string) error {
	b, e := readOwnedOutput(filepath.Join(attempt, "reservation.json"), 16384)
	if e != nil {
		return e
	}
	var v childReservation
	if e = strictCompact(b, &v); e != nil {
		return e
	}
	if v.Schema != "riido-gjson-two-native-reservation-v1" || v.PlanSHA256 != sha || v.Literal != p.Literal.Pin || v.Wants != p.Wants.Pin || v.Captions != p.Captions.Pin || v.RootFreeze != p.RootWantFreeze.Pin || v.Rows != 33 || v.Methods != 61 || v.Callbacks != 33 || v.Frames != 384 || v.OutputReserve != 2097152 || v.InitialCounts != (core.Counts{}) || v.OriginalInitialization != nil || v.OriginalNested != nil {
		return errCode("child_reservation_binding")
	}
	return nil
}

type Carrier struct {
	Fixtures [11]core.Fixture
	Wants    [11]core.WantRecord
}

func loadCarrier(p CountPlan) (out Carrier, e error) {
	l, e := checkPin(refPin(p.Literal), evidenceCap)
	if e != nil {
		return
	}
	w, e := checkPin(refPin(p.Wants), evidenceCap)
	if e != nil {
		return
	}
	var a, b map[string]json.RawMessage
	if json.Unmarshal(l, &a) != nil || json.Unmarshal(w, &b) != nil {
		return out, errCode("carrier_json")
	}
	var fs []core.Fixture
	var ws []core.WantRecord
	if json.Unmarshal(a["fixtures"], &fs) != nil || json.Unmarshal(b["records"], &ws) != nil || len(fs) != 11 || len(ws) != 11 {
		return out, errCode("carrier_arity")
	}
	for i, f := range fs {
		q, fi := 0, i
		if i >= 5 {
			q = 1
			fi = i - 5
		}
		if f.Ordinal != i || f.Request != q || f.Index != fi || ws[i].Ordinal != i || f.RequestID != ws[i].RequestID || f.ID != ws[i].FixtureID {
			return out, errCode("carrier_order")
		}
		out.Fixtures[i] = f
		out.Wants[i] = ws[i]
	}
	return out, nil
}
func equalRaw(a, b []byte) bool {
	var x, y bytes.Buffer
	return json.Compact(&x, a) == nil && json.Compact(&y, b) == nil && bytes.Equal(x.Bytes(), y.Bytes())
}
func bindRow(row core.Row, d int, carrier Carrier) bool {
	q, f, x, j, ok := schedule(d)
	if !ok {
		return false
	}
	i := f
	if q == 1 {
		i += 5
	}
	v := carrier.Fixtures[i]
	w := carrier.Wants[i]
	return row.Ordinal == d && row.Request == q && row.Fixture == f && row.Display == x && row.Internal == j && row.RequestID == v.RequestID && row.FixtureID == v.ID && row.InputPointer == v.InputReference.Pointer && equalRaw(row.Input, v.Input) && equalRaw(row.Want, w.Want) && row.Truth == nil && row.Role == nil && row.Weight == nil
}
