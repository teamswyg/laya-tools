// perfsuite replays fixed public scenarios, without downstream model execution.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"
)

type scenario struct {
	name, bin string
	args      []string
	native    bool
}
type row struct {
	Name         string          `json:"name"`
	Repeat       int             `json:"repeat"`
	OK           bool            `json:"ok"`
	WallMS       float64         `json:"wall_ms"`
	CPUSeconds   float64         `json:"cpu_seconds"`
	PeakRSSBytes int64           `json:"peak_rss_bytes"`
	Output       json.RawMessage `json:"output,omitempty"`
	Error        string          `json:"error,omitempty"`
}

func scenarios() []scenario {
	s := []scenario{
		{"native-warm", "riidolaya", []string{"bench", "--iterations", "5", "--threads", "4"}, true},
		{"search-lexical", "riidolaya", []string{"search", "--root", ".", "--lexical", "--json", "routing confidence"}, false},
		{"search-laya", "riidolaya", []string{"search", "--root", ".", "--candidates", "3", "--limit", "3", "--threads", "4", "--json", "routing confidence"}, true},
		{"plan-downgrade", "riidolaya", []string{"plan", "--config", "examples/planner/config.json", "--request", "examples/planner/request.json", "--json"}, false},
		{"plan-upgrade", "riidolaya", []string{"plan", "--config", "examples/planner/config.json", "--request", "examples/planner/upgrade.json", "--json"}, false},
		{"repositories-lexical", "repoeval", nil, false},
		{"repositories-laya", "repoeval", []string{"--laya"}, true},
	}
	for _, q := range []struct{ name, text string }{
		{"route-easy", "Fix a spelling mistake in a comment."},
		{"route-standard", "Add input validation and tests to a bounded form handler."},
		{"route-hard", "Diagnose a concurrent transaction race across services and design a safe schema migration."},
		{"route-korean-guard", "동시성 오류를 분석하고 수정해줘"},
	} {
		s = append(s, scenario{q.name, "riidolaya", []string{"route", "--fast-model", "example-fast", "--standard-model", "example-standard", "--strong-model", "example-strong", "--threads", "4", "--json", q.text}, q.name != "route-korean-guard"})
	}
	return s
}

// Native scenarios must not silently succeed through an inference-error fallback.
func validate(s scenario, b []byte) error {
	if !json.Valid(b) {
		return fmt.Errorf("invalid JSON result")
	}
	var v map[string]json.RawMessage
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	if result, ok := v["result"]; ok {
		if err := json.Unmarshal(result, &v); err != nil {
			return err
		}
	}
	if !s.native {
		return nil
	}
	if s.name == "native-warm" {
		if len(v["probabilities"]) == 0 {
			return fmt.Errorf("missing native probabilities")
		}
		return nil
	}
	if s.name == "repositories-laya" {
		var n int
		_ = json.Unmarshal(v["model_judgments"], &n)
		if n == 0 {
			return fmt.Errorf("no native repository judgments")
		}
		return nil
	}
	if s.name == "search-laya" {
		var warnings []string
		_ = json.Unmarshal(v["warnings"], &warnings)
		if len(warnings) != 0 {
			return fmt.Errorf("search reported fallback warnings")
		}
		var results []struct {
			Relevance *float64 `json:"relevance"`
		}
		_ = json.Unmarshal(v["results"], &results)
		if len(results) == 0 {
			return fmt.Errorf("no search candidates")
		}
		for _, r := range results {
			if r.Relevance == nil {
				return fmt.Errorf("missing native relevance")
			}
		}
		return nil
	}
	var probs []float64
	_ = json.Unmarshal(v["probabilities"], &probs)
	if len(probs) != 3 {
		return fmt.Errorf("missing native route probabilities")
	}
	return nil
}
func measure(s scenario, repeat int) row {
	r := row{Name: s.name, Repeat: repeat}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "./bin/"+s.bin, s.args...)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	start := time.Now()
	err := cmd.Run()
	r.WallMS = float64(time.Since(start).Microseconds()) / 1000
	if cmd.ProcessState != nil {
		r.CPUSeconds = (cmd.ProcessState.UserTime() + cmd.ProcessState.SystemTime()).Seconds()
		if u, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage); ok {
			r.PeakRSSBytes = u.Maxrss
			if runtime.GOOS == "linux" {
				r.PeakRSSBytes *= 1024
			}
		}
	}
	if ctx.Err() != nil {
		r.Error = "scenario timeout"
		return r
	}
	if err != nil {
		r.Error = "scenario process failed"
		fmt.Fprintf(os.Stderr, "%s: %v\n%s", s.name, err, stderr.String())
		return r
	}
	if err = validate(s, out.Bytes()); err != nil {
		r.Error = err.Error()
		return r
	}
	r.OK = true
	r.Output = append([]byte(nil), out.Bytes()...)
	return r
}
func run() error {
	repeats := flag.Int("repeats", 1, "fixed passes, 1..3")
	output := flag.String("output", ".cache/performance.json", "JSON report")
	summary := flag.String("summary", ".cache/performance.md", "Markdown report")
	flag.Parse()
	if *repeats < 1 || *repeats > 3 {
		return fmt.Errorf("repeats must be 1..3")
	}
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return fmt.Errorf("RSS measurement requires Linux or macOS")
	}
	rows := []row{}
	failed := false
	var md bytes.Buffer
	fmt.Fprintf(&md, "# Fixed-scenario performance replay\n\n%s/%s · %s · %d logical CPUs · commit %s\n\n", runtime.GOOS, runtime.GOARCH, runtime.Version(), runtime.NumCPU(), os.Getenv("GITHUB_SHA"))
	md.WriteString("CPU only; fresh process per case. Wall time includes model load; embedded native-warm timings exclude load. Peak RSS includes native allocations, not GPU memory. Fixed sequential replay, not an arrival-rate load test. Public development fixtures do not prove production accuracy or Codex savings.\n\n| Case | Pass | OK | Wall ms | CPU seconds | Peak RSS MiB |\n|---|---:|---|---:|---:|---:|\n")
	for i := 1; i <= *repeats; i++ {
		for _, s := range scenarios() {
			r := measure(s, i)
			rows = append(rows, r)
			failed = failed || !r.OK
			fmt.Fprintf(&md, "| %s | %d | %t | %.1f | %.2f | %.1f |\n", r.Name, i, r.OK, r.WallMS, r.CPUSeconds, float64(r.PeakRSSBytes)/(1<<20))
			fmt.Fprintf(os.Stderr, "%s pass=%d ok=%t\n", s.name, i, r.OK)
		}
	}
	report := map[string]any{"os": runtime.GOOS, "arch": runtime.GOARCH, "go": runtime.Version(), "logical_cpus": runtime.NumCPU(), "commit": os.Getenv("GITHUB_SHA"), "mode": "fixed_sequential_replay", "rows": rows}
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(*output, b, 0600); err != nil {
		return err
	}
	if err = os.WriteFile(*summary, md.Bytes(), 0600); err != nil {
		return err
	}
	if failed {
		return fmt.Errorf("one or more scenarios failed; inspect report")
	}
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
