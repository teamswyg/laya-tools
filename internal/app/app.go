package app

import (
	"encoding/json"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/inference"
	"github.com/teamswyg/laya-tools/internal/router"
	"github.com/teamswyg/laya-tools/internal/search"
	"time"
)

type App struct {
	Root        string
	Options     inference.Options
	RouteConfig router.Config
	engine      *inference.Engine
	loadErr     error
	attempted   bool
}

func (a *App) Engine() (*inference.Engine, error) {
	if !a.attempted {
		a.attempted = true
		a.engine, a.loadErr = inference.New(a.Options)
	}
	return a.engine, a.loadErr
}
func (a *App) Close() {
	if a.engine != nil {
		a.engine.Close()
	}
}

type Request struct {
	ID             json.RawMessage `json:"id,omitempty"`
	Op             string          `json:"op"`
	Query          string          `json:"query"`
	CandidateQuery string          `json:"candidate_query,omitempty"`
	Candidates     int             `json:"candidates,omitempty"`
	Limit          int             `json:"limit,omitempty"`
	Lexical        bool            `json:"lexical,omitempty"`
	Model          string          `json:"model,omitempty"`
}
type Response struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Result any             `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
	MS     float64         `json:"ms"`
}

func (a *App) Process(req Request) Response {
	start := time.Now()
	r := Response{ID: req.ID}
	var err error
	switch req.Op {
	case "search":
		var idx *search.Index
		idx, err = search.Load(a.Root)
		if err == nil {
			var scorer search.Scorer
			var warning string
			if !req.Lexical {
				var e *inference.Engine
				e, err = a.Engine()
				if err == nil {
					scorer = e
				} else {
					warning = "Laya unavailable: using lexical search. Run riidolaya setup or check model/runtime paths."
					err = nil
				}
			}
			k := req.Candidates
			if k == 0 {
				k = 8
			}
			limit := req.Limit
			if limit == 0 {
				limit = 3
			}
			var result search.Result
			result, err = idx.Search(req.Query, req.CandidateQuery, k, limit, scorer)
			if warning != "" {
				result.Warnings = append(result.Warnings, warning)
			}
			r.Result = result
		}
	case "route":
		var scorer router.Scorer
		if router.NeedsInference(req.Query, req.Model, a.RouteConfig) {
			if e, e2 := a.Engine(); e2 == nil {
				scorer = e
			}
		}
		r.Result, err = router.Route(req.Query, req.Model, a.RouteConfig, scorer)
	default:
		err = fmt.Errorf("unknown operation: %s", req.Op)
	}
	if err != nil {
		r.Result = nil
		r.Error = err.Error()
	}
	r.MS = float64(time.Since(start).Microseconds()) / 1000
	return r
}
