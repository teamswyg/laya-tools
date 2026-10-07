// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaimscli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
)

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
