// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaimscli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
	"github.com/teamswyg/laya-tools/pkg/statehintclaimsmlp"
	"github.com/teamswyg/laya-tools/pkg/statehintclaimtrit"
)

func TestExplicitUntrainedSharedHeadModelStaysResearchOnly(t *testing.T) {
	var artifact bytes.Buffer
	if err := statehintclaimsmlp.NewModel().Save(&artifact); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "original-untrained.rcm")
	if err := os.WriteFile(path, artifact.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	input := strings.NewReader("{\"text\":\"original synthetic report\"}\n{\"text\":\"a\",\"text\":\"b\"}\n{\"text\":\"second synthetic report\"}\n")
	var output, errors bytes.Buffer
	if err := Run([]string{"--model", path, "--jsonl"}, input, &output, &errors); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 3 {
		t.Fatal("warm responses", len(lines))
	}
	for i, line := range lines {
		var r response
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			t.Fatal(err)
		}
		if r.Qualified || r.Writes || r.Mode != "research_preview" || r.Quantization != nil || r.Architecture == nil || r.Architecture.HiddenUnits != 16 || r.Architecture.Qualified || r.Architecture.StateAuthority {
			t.Fatal("unsafe architecture response")
		}
		if i == 1 {
			if r.Prediction != nil || r.Reason != "invalid_request" {
				t.Fatal("duplicate request admitted")
			}
			continue
		}
		if r.Prediction == nil || r.Prediction.Source != statehintclaims.Untrained || r.Prediction.TrainingSteps != 0 {
			t.Fatal("fabricated learned model")
		}
		for _, h := range r.Prediction.Heads {
			if h.State != statehintclaims.Unknown || h.UnknownReason != "untrained" {
				t.Fatal("untrained claim")
			}
		}
	}
	corrupt := append([]byte(nil), artifact.Bytes()...)
	corrupt[len(corrupt)-1] ^= 1
	if err := os.WriteFile(path, corrupt, 0600); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if Run([]string{"--model", path, "--text", "original text"}, nil, &output, &errors) == nil || output.Len() != 0 {
		t.Fatal("corrupt RCM admitted")
	}
}

func TestWarmStreamRejectsAmbiguousRequestsAndNeverClaimsQualification(t *testing.T) {
	var artifact bytes.Buffer
	if e := statehintclaims.NewModel().Save(&artifact); e != nil {
		t.Fatal(e)
	}
	p := filepath.Join(t.TempDir(), "original-untrained.rsc")
	if e := os.WriteFile(p, artifact.Bytes(), 0600); e != nil {
		t.Fatal(e)
	}
	input := strings.NewReader("{\"text\":\"original sample\"}\n{\"text\":\"a\",\"text\":\"b\"}\n{\"Text\":\"c\"}\n{\"text\":\"d\",\"context\":{}}\n{\"text\":null}\n{\"text\":\"e\"} {}\n{\"text\":\"another original sample\"}\n")
	var out, errs bytes.Buffer
	if e := Run([]string{"--model", p, "--jsonl"}, input, &out, &errs); e != nil {
		t.Fatal(e)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 7 {
		t.Fatalf("responses %d", len(lines))
	}
	for i, line := range lines {
		var r response
		if e := json.Unmarshal([]byte(line), &r); e != nil {
			t.Fatal(e)
		}
		if r.Writes || r.Qualified || r.Mode != "research_preview" {
			t.Fatal("unsafe response")
		}
		if i == 0 || i == 6 {
			if r.Prediction == nil || r.Prediction.Source != statehintclaims.Untrained || r.Prediction.TrainingSteps != 0 {
				t.Fatal("fabricated learned prediction")
			}
			for _, h := range r.Prediction.Heads {
				if h.State != statehintclaims.Unknown || h.UnknownReason != "untrained" {
					t.Fatal("untrained hint")
				}
			}
		} else if r.Reason != "invalid_request" || r.Prediction != nil {
			t.Fatal("ambiguous request accepted")
		}
	}
}

func TestExplicitTernaryPreviewKeepsUntrainedAndNoAuthority(t *testing.T) {
	parent := statehintclaims.NewModel()
	var original bytes.Buffer
	if e := parent.Save(&original); e != nil {
		t.Fatal(e)
	}
	hash := sha256.Sum256(original.Bytes())
	model, e := statehintclaimtrit.FromFloat(parent, hex.EncodeToString(hash[:]))
	if e != nil {
		t.Fatal(e)
	}
	var artifact bytes.Buffer
	if e := model.Save(&artifact); e != nil {
		t.Fatal(e)
	}
	p := filepath.Join(t.TempDir(), "original-untrained.rqt")
	if e := os.WriteFile(p, artifact.Bytes(), 0600); e != nil {
		t.Fatal(e)
	}
	var out, errs bytes.Buffer
	if e := Run([]string{"--model", p, "--text", "original fictional comment"}, nil, &out, &errs); e != nil {
		t.Fatal(e)
	}
	var result response
	if e := json.Unmarshal(out.Bytes(), &result); e != nil {
		t.Fatal(e)
	}
	if result.Writes || result.Qualified || result.Prediction == nil || result.Prediction.Source != statehintclaims.Untrained || result.Quantization == nil || result.Quantization.Mode != "ptq" || result.Quantization.NewOptimizerSteps != 0 || result.Quantization.StateAuthority || result.Quantization.QualityQualified {
		t.Fatal("projection fabricated training or authority")
	}
	for _, h := range result.Prediction.Heads {
		if h.State != statehintclaims.Unknown || h.UnknownReason != "untrained" {
			t.Fatal("untrained ternary hint")
		}
	}
}

func TestModelRequiredAndForeignArtifactRejected(t *testing.T) {
	var out bytes.Buffer
	if Run([]string{"--text", "original text"}, nil, &out, &out) == nil {
		t.Fatal("implicit model")
	}
	p := filepath.Join(t.TempDir(), "foreign.model")
	if e := os.WriteFile(p, []byte("RSH-old-model"), 0600); e != nil {
		t.Fatal(e)
	}
	if Run([]string{"--model", p, "--text", "original text"}, nil, &out, &out) == nil {
		t.Fatal("foreign model accepted")
	}
}
