package main

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/teamswyg/laya-tools/internal/app"
	"github.com/teamswyg/laya-tools/internal/router"
	"github.com/teamswyg/laya-tools/pkg/planner"
)

func TestStrictPlanJSON(t *testing.T) {
	for _, input := range []string{`{"assessment":{},"typo":true}`, `{} {}`, `{"assessment":{"confidence":NaN}}`} {
		var r planner.Request
		if err := decodePlanJSON([]byte(input), &r); err == nil {
			t.Fatalf("accepted %s", input)
		}
	}
}

func TestPlanCLI(t *testing.T) {
	for _, tt := range []struct{ name, query, status, model string }{
		{"no native runtime", "", "recommend", "example-fast"},
		{"Korean abstention", "오타 수정", "hold", "example-strong"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			a := &app.App{}
			a.RouteConfig.Threshold = .9
			read, write, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			old := os.Stdout
			os.Stdout = write
			defer func() { os.Stdout = old; read.Close(); write.Close() }()
			err = runPlan(a, "../../examples/planner/config.json", "../../examples/planner/request.json", tt.query, true)
			write.Close()
			os.Stdout = old
			if err != nil {
				t.Fatal(err)
			}
			b, err := io.ReadAll(read)
			if err != nil {
				t.Fatal(err)
			}
			var result struct {
				Plan           planner.Plan   `json:"plan"`
				Classification *router.Result `json:"classification"`
			}
			if err = json.Unmarshal(b, &result); err != nil {
				t.Fatal(err)
			}
			if result.Plan.Status != tt.status || result.Plan.RecommendedModel != tt.model {
				t.Fatal(string(b))
			}
			if tt.query != "" && (result.Classification == nil || !result.Classification.Abstained || !result.Plan.Selection.Guarded) {
				t.Fatal("lost abstention", string(b))
			}
		})
	}
}
