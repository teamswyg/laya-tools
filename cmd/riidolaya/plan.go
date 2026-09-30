package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/teamswyg/laya-tools/internal/app"
	"github.com/teamswyg/laya-tools/internal/router"
	"github.com/teamswyg/laya-tools/pkg/planner"
)

func readPlanJSON(path string, value any) error {
	if path == "" {
		return fmt.Errorf("JSON file path is required")
	}
	var r io.Reader = os.Stdin
	if path != "-" {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	}
	b, err := io.ReadAll(io.LimitReader(r, (1<<20)+1))
	if err != nil {
		return err
	}
	if len(b) > 1<<20 {
		return fmt.Errorf("plan input exceeds 1 MiB")
	}
	return decodePlanJSON(b, value)
}

func runPlan(a *app.App, configPath, requestPath, query string, jsonOut bool) error {
	if configPath == "-" {
		return fmt.Errorf("--config must be a file; --request may use stdin")
	}
	var c planner.Config
	var r planner.Request
	if err := readPlanJSON(configPath, &c); err != nil {
		return err
	}
	if err := readPlanJSON(requestPath, &r); err != nil {
		return err
	}
	var assessment *router.Result
	if query != "" {
		if len(c.Catalog.QualityFloors) != 3 {
			return fmt.Errorf("Laya query requires three quality floors: fast, standard, strong")
		}
		// These are classification labels only; plan never launches these names.
		a.RouteConfig.Fast, a.RouteConfig.Standard, a.RouteConfig.Strong = "fast", "standard", "strong"
		response := a.Process(app.Request{Op: "route", Query: query})
		if response.Error != "" {
			return fmt.Errorf("classification: %s", response.Error)
		}
		result := response.Result.(router.Result)
		assessment = &result
		tiers := map[string]int{"fast": 0, "standard": 1, "strong": 2}
		r.Assessment.Tier = tiers[result.Tier]
		r.Assessment.Confidence = result.Confidence
		r.Assessment.Uncertain = result.Abstained || result.Truncated
	}
	p, err := planner.Build(c, r)
	if err != nil {
		return err
	}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(struct {
			Plan           planner.Plan   `json:"plan"`
			Classification *router.Result `json:"classification,omitempty"`
		}{p, assessment})
	}
	fmt.Printf("%s: %s\n", p.Status, p.Reason)
	if p.RecommendedModel != "" {
		fmt.Printf("recommended model: %s\n", p.RecommendedModel)
	}
	for _, v := range p.Selection.Candidates {
		fmt.Printf("%s eligible=%t reasons=%v\n", v.ID, v.Eligible, v.Reasons)
	}
	fmt.Println("Plan only. No downstream model was called or switched; estimates are not subscription usage.")
	return nil
}
