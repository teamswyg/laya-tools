// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintcorpuscli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/statehintcorpus"
	"github.com/teamswyg/laya-tools/pkg/statehint"
)

func TestCountsOnlyAndFixedErrors(t *testing.T) {
	sha := strings.Repeat("a", 64)
	row := statehintcorpus.Row{Schema: "statehint-v4-original-train-seed-row-v1", Family: "original-family", Lineage: "original-lineage", Partition: "train", Wording: 1, Role: "prose", Applicable: true, Unit: "Fictional read cache change.", ClauseScopes: []string{"current_unit"}, AssertionForms: []string{"ongoing"}, Expected: statehint.Progress, AnnotationSource: "Original fictional fixture, not truth.", OntologyVersion: "statehint-intent-scope-v4-1200x2-v1", OntologyFreezeSHA: sha, License: "Apache-2.0"}
	var input bytes.Buffer
	for _, locale := range []string{"ko", "en"} {
		row.ID, row.Locale, row.Text = "fictional-"+locale, locale, "LOCAL-SOURCE-CANARY-"+locale
		if err := json.NewEncoder(&input).Encode(row); err != nil {
			t.Fatal(err)
		}
	}
	var out, errOut bytes.Buffer
	err := Run([]string{"--partition", "train", "--rubric-sha256", sha}, &input, &out, &errOut)
	if err != nil || errOut.Len() != 0 || strings.Contains(out.String(), "CANARY") || strings.Contains(out.String(), "original-family") || !strings.Contains(out.String(), `"families":1`) {
		t.Fatalf("counts-only output: %s, err=%v", out.String(), err)
	}
	for _, key := range []string{"semantic_verified", "training_allowed", "model_used", "mutation_executed"} {
		if !strings.Contains(out.String(), `"`+key+`":false`) {
			t.Fatal("authority flag missing or true")
		}
	}
	for _, args := range [][]string{{"--CANARY-private"}, {"--partition", "train", "--rubric-sha256", "CANARY-private"}, {"--partition", "train", "--rubric-sha256", sha, "CANARY-private"}} {
		out.Reset()
		errOut.Reset()
		err := Run(args, strings.NewReader("CANARY-private"), &out, &errOut)
		if err == nil || strings.Contains(err.Error()+out.String()+errOut.String(), "CANARY-private") {
			t.Fatal("invalid source or options leaked content")
		}
	}
}
