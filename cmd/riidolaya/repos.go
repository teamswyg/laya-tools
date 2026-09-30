package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/teamswyg/laya-tools/internal/app"
	"github.com/teamswyg/laya-tools/internal/repojudge"
	"github.com/teamswyg/laya-tools/pkg/reporouter"
)

func runRepos(a *app.App, path, query string, withLaya, serve, jsonOut bool, threshold float64) error {
	var repos []reporouter.Repository
	if path == "-" {
		return fmt.Errorf("--catalog must be a file")
	}
	if err := readPlanJSON(path, &repos); err != nil {
		return err
	}
	idx, err := reporouter.New(repos)
	if err != nil {
		return err
	}
	var judge reporouter.Judge
	if withLaya {
		judge = repojudge.Judge{Engine: a.Engine}
	}
	cfg := reporouter.DefaultConfig()
	cfg.Threshold = threshold
	if serve {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 4096), 16384)
		for scanner.Scan() {
			var request struct {
				ID    json.RawMessage `json:"id"`
				Query string          `json:"query"`
			}
			var result reporouter.Result
			err := decodePlanJSON(scanner.Bytes(), &request)
			if err == nil {
				result, err = idx.Preview(request.Query, cfg, judge)
			}
			response := struct {
				ID     json.RawMessage    `json:"id,omitempty"`
				Result *reporouter.Result `json:"result,omitempty"`
				Error  string             `json:"error,omitempty"`
			}{ID: request.ID}
			if err != nil {
				response.Error = err.Error()
			} else {
				response.Result = &result
			}
			if err = json.NewEncoder(os.Stdout).Encode(response); err != nil {
				return err
			}
		}
		return scanner.Err()
	}
	result, err := idx.Preview(query, cfg, judge)
	if err != nil {
		return err
	}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	fmt.Printf("PREVIEW %s: %s\n", result.Status, result.Reason)
	if result.Suggested != "" {
		fmt.Printf("candidate repository: %s\n", result.Suggested)
	}
	for _, v := range result.Candidates {
		fmt.Printf("%s lexical=%.3f\n", v.Name, v.LexicalScore)
	}
	fmt.Println("No repository was opened, cloned, modified, or authorized.")
	return nil
}
