package app

import (
	"github.com/teamswyg/laya-tools/internal/router"
	"testing"
)

func TestPolicyOnlyRouteDoesNotLoadWeights(t *testing.T) {
	for _, req := range []Request{{Op: "route", Query: "주석 오타 수정"}, {Op: "route", Query: "task", Model: "explicit"}, {Op: "route", Query: ""}} {
		a := &App{RouteConfig: router.Config{Threshold: .9, Strong: "strong"}}
		r := a.Process(req)
		if a.attempted {
			t.Fatal("policy-only request attempted native initialization")
		}
		if req.Query != "" && r.Error != "" {
			t.Fatal(r.Error)
		}
	}
}
