// routeeval compares fixed author-labeled development tasks with real Laya outputs.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/teamswyg/laya-tools/internal/assets"
	"github.com/teamswyg/laya-tools/internal/inference"
	"github.com/teamswyg/laya-tools/internal/router"
)

type task struct {
	ID        string `json:"id"`
	Stage     int    `json:"stage"`
	Language  string `json:"language"`
	Prompt    string `json:"prompt"`
	Expected  string `json:"expected"`
	Rationale string `json:"rationale"`
}
type row struct {
	Task       task          `json:"task"`
	Result     router.Result `json:"result"`
	RawCorrect bool          `json:"raw_correct"`
	MS         float64       `json:"ms"`
}
type summary struct {
	Cases           int       `json:"cases"`
	Judged          int       `json:"judged"`
	Correct         int       `json:"raw_correct"`
	Accepted        int       `json:"accepted"`
	AcceptedCorrect int       `json:"accepted_correct"`
	UnderRoutes     int       `json:"accepted_under_routes"`
	Confusion       [3][3]int `json:"confusion_expected_by_predicted"`
}

func tier(s string) int {
	switch s {
	case "fast":
		return 0
	case "standard":
		return 1
	case "strong":
		return 2
	}
	return -1
}
func load(path string) ([]task, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	var tasks []task
	if err = json.Unmarshal(b, &tasks); err != nil {
		return nil, "", err
	}
	seen := map[string]bool{}
	if len(tasks) == 0 || len(tasks) > 1000 {
		return nil, "", fmt.Errorf("require 1..1000 tasks")
	}
	for _, t := range tasks {
		if t.ID == "" || seen[t.ID] || t.Stage < 1 || t.Stage > 3 || tier(t.Expected) < 0 || t.Prompt == "" || t.Rationale == "" || (t.Language != "en" && t.Language != "ko") {
			return nil, "", fmt.Errorf("invalid golden task %q", t.ID)
		}
		seen[t.ID] = true
	}
	hash := sha256.Sum256(b)
	return tasks, hex.EncodeToString(hash[:]), nil
}
func run() error {
	path := flag.String("tasks", "benchmarks/routing-golden.json", "fixed public tasks")
	stage := flag.Int("stage", 3, "cumulative stage 1..3")
	flag.Parse()
	if *stage < 1 || *stage > 3 {
		return fmt.Errorf("stage must be 1..3")
	}
	tasks, hash, err := load(*path)
	if err != nil {
		return err
	}
	e, err := inference.New(inference.Options{ModelDir: assets.ModelDir("base"), Runtime: assets.Runtime(), Provider: "cpu", Threads: 4, MaxTokens: 512})
	if err != nil {
		return err
	}
	defer e.Close()
	c := router.Config{Fast: "example-fast", Standard: "example-standard", Strong: "example-strong", Threshold: .9}
	rows := []row{}
	s := summary{}
	for _, t := range tasks {
		if t.Stage > *stage {
			continue
		}
		start := time.Now()
		r, err := router.Route(t.Prompt, "", c, e)
		if err != nil {
			return err
		}
		expected, pred := tier(t.Expected), tier(r.SuggestedTier)
		// An English fixture silently falling back after native failure is not evidence.
		if t.Language == "en" && len(r.Probabilities) != 3 {
			return fmt.Errorf("%s has no native prediction", t.ID)
		}
		s.Cases++
		correct := pred == expected
		if pred >= 0 {
			s.Judged++
			s.Confusion[expected][pred]++
			if correct {
				s.Correct++
			}
		}
		if !r.Abstained {
			s.Accepted++
			if tier(r.Tier) == expected {
				s.AcceptedCorrect++
			}
			if tier(r.Tier) < expected {
				s.UnderRoutes++
			}
		}
		rows = append(rows, row{t, r, correct, float64(time.Since(start).Microseconds()) / 1000})
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"schema": 1, "fixture_sha256": hash, "stage": *stage, "os": runtime.GOOS, "arch": runtime.GOARCH, "go": runtime.Version(), "load_ms": e.LoadMS, "summary": s, "rows": rows, "note": "Author-labeled development golden set, fixed before inference. Not independent labels, coding success, or held-out calibration. Korean policy abstentions excluded from raw accuracy. Confusion order: fast, standard, strong. No downstream coding model executed."})
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
