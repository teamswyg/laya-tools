package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/app"
	"github.com/teamswyg/laya-tools/internal/assets"
	"github.com/teamswyg/laya-tools/internal/inference"
	"github.com/teamswyg/laya-tools/internal/questioncuecli"
	"github.com/teamswyg/laya-tools/internal/router"
	"github.com/teamswyg/laya-tools/internal/search"
	"github.com/teamswyg/laya-tools/internal/statehintcli"
	"github.com/teamswyg/laya-tools/internal/statehintpilotcli"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
)

var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "riidolaya:", err)
		os.Exit(1)
	}
}
func run(args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		fmt.Print(`riidolaya — local code retrieval and experimental content/state hints

Commands:
  question-cue  Preview an experimental punctuation clue without a model
  state-hint  Suggest content labels, emoji codes and guarded shadow state plans
  state-hint-pilot  Compare local models on bounded synthetic development cases
  setup     Download checksum-pinned model and native runtime
  search    Find line-addressable code excerpts (default: 8 candidates, 3 hits)
  route     Recommend a configured model; abstain on uncertain decisions
  repo-preview  Preview repository selection from a local catalog
  repo-serve    Warm JSONL repository preview (no execution)
  plan      Compare configured models, budgets, and cache switch costs (dry plan)
  codex     Route one new task, then launch the existing Codex CLI
  serve     Warm JSONL service: one {"op":"search|route","query":"..."} per line
  mcp       Warm stdio MCP server (search_code and route_model)
  bench     Measure cold load and repeated native inference; supports pprof
  doctor    Report platform and local asset locations
  version   Print version

Examples (flags precede the prompt):
  riidolaya state-hint --text "질문이 있습니다" --json
  riidolaya setup
  riidolaya search --root . --json "where are redirects handled?"
  riidolaya search --lexical "redirect headers"
  riidolaya route --fast-model MODEL --strong-model MODEL "fix a typo"
  riidolaya codex --model MODEL "implement a feature"
  riidolaya bench --iterations 10 --cpu-profile cpu.pprof --heap-profile heap.pprof

No credentials or source code are uploaded by search, route, serve, or mcp.
The codex command starts your installed Codex CLI with its existing settings.
`)
		return nil
	}
	cmd := args[0]
	if cmd == "question-cue" {
		return questioncuecli.Run(args[1:], os.Stdin, os.Stdout, os.Stderr)
	}
	if cmd == "state-hint" {
		return statehintcli.Run(args[1:], os.Stdin, os.Stdout, os.Stderr)
	}
	if cmd == "state-hint-pilot" {
		return statehintpilotcli.Run(args[1:], os.Stdout, os.Stderr)
	}
	if cmd == "version" {
		fmt.Println(version)
		return nil
	}
	f := flag.NewFlagSet(cmd, flag.ContinueOnError)
	root := f.String("root", ".", "repository root")
	modelName := f.String("checkpoint", "base", "base or code checkpoint")
	modelDir := f.String("model-dir", "", "custom converted model directory")
	lib := f.String("runtime", assets.Runtime(), "native ONNX Runtime library")
	provider := f.String("provider", "cpu", "cpu or experimental coreml")
	threads := f.Int("threads", 4, "native CPU threads (1..64)")
	tokens := f.Int("max-tokens", 512, "sequence token budget (256..512)")
	jsonOut := f.Bool("json", false, "machine-readable output")
	lexical := f.Bool("lexical", false, "skip model loading and reranking")
	candidates := f.Int("candidates", 8, "rerank candidate count (1..64)")
	limit := f.Int("limit", 3, "maximum non-duplicate results")
	candidateQuery := f.String("candidate-query", "", "optional lexical query, e.g. English identifiers")
	explicit := f.String("model", "", "explicit model override; wins over routing")
	fast := f.String("fast-model", os.Getenv("LAYA_FAST_MODEL"), "configured fast Codex model ID")
	standard := f.String("standard-model", os.Getenv("LAYA_STANDARD_MODEL"), "configured standard Codex model ID")
	strong := f.String("strong-model", os.Getenv("LAYA_STRONG_MODEL"), "strong model ID; empty preserves Codex default")
	threshold := f.Float64("threshold", .9, "minimum routing confidence")
	iterations := f.Int("iterations", 10, "warm benchmark iterations")
	cpu := f.String("cpu-profile", "", "write Go CPU pprof locally")
	heap := f.String("heap-profile", "", "write Go heap pprof locally")
	ortProfile := f.String("ort-profile", "", "native profiling file prefix (local)")
	repoCatalog := f.String("catalog", "", "local repository catalog JSON file")
	withLaya := f.Bool("laya", false, "opt in to English Laya repository selection")
	planConfig := f.String("config", "", "plan catalog and switch policy JSON")
	planRequest := f.String("request", "", "plan request JSON file or - for stdin")
	dry := f.Bool("dry-run", false, "show Codex argv without launching")
	if err := f.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *modelName != "base" && *modelName != "code" {
		return fmt.Errorf("checkpoint must be base or code")
	}
	if *modelDir == "" {
		*modelDir = assets.ModelDir(*modelName)
	}
	if *cpu != "" {
		file, err := os.Create(*cpu)
		if err != nil {
			return err
		}
		defer file.Close()
		if err = pprof.StartCPUProfile(file); err != nil {
			return err
		}
		defer pprof.StopCPUProfile()
	}
	a := &app.App{Root: *root, Options: inference.Options{ModelDir: *modelDir, Runtime: *lib, Provider: *provider, Threads: *threads, MaxTokens: *tokens, Profile: *ortProfile}, RouteConfig: router.Config{Fast: *fast, Standard: *standard, Strong: *strong, Threshold: *threshold}}
	defer a.Close()
	if *heap != "" {
		defer func() {
			file, err := os.Create(*heap)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
			defer file.Close()
			runtime.GC()
			defer runtime.KeepAlive(a)
			if err = pprof.WriteHeapProfile(file); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}()
	}

	switch cmd {
	case "repo-preview", "repo-serve":
		if *withLaya && *modelName != "base" {
			return fmt.Errorf("repository choice preview requires the base checkpoint")
		}
		return runRepos(a, *repoCatalog, strings.Join(f.Args(), " "), *withLaya, cmd == "repo-serve", *jsonOut, *threshold)
	case "plan":
		return runPlan(a, *planConfig, *planRequest, strings.Join(f.Args(), " "), *jsonOut)
	case "setup":
		return assets.Setup(context.Background(), *modelName, os.Stderr)
	case "doctor":
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"version": version, "os": runtime.GOOS, "arch": runtime.GOARCH, "model_dir": *modelDir, "model_exists": exists(filepath.Join(*modelDir, "model.onnx")), "runtime": *lib, "runtime_exists": exists(*lib), "gpu_measurement": "pprof excludes native/GPU memory; use ORT trace and macOS Instruments"})
	case "serve":
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 4096), 1<<20)
		for scanner.Scan() {
			var req app.Request
			if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
				_ = json.NewEncoder(os.Stdout).Encode(app.Response{Error: "invalid JSON request"})
				continue
			}
			if err := json.NewEncoder(os.Stdout).Encode(a.Process(req)); err != nil {
				return err
			}
		}
		return scanner.Err()
	case "mcp":
		return serveMCP(a, os.Stdin, os.Stdout)
	case "bench":
		if *iterations < 1 || *iterations > 1000 {
			return fmt.Errorf("iterations must be 1..1000")
		}
		e, err := a.Engine()
		if err != nil {
			return err
		}
		state := "The HTTP request follows a redirect. Remove authorization headers when the host changes."
		options := []string{"false: no, the statement does not hold", "true: yes, the statement holds"}
		if _, err = e.Predict(state, "noul", "Does this describe redirect security?", options); err != nil {
			return err
		}
		times := make([]float64, *iterations)
		var p inference.Prediction
		for i := range times {
			p, err = e.Predict(state, "noul", "Does this describe redirect security?", options)
			if err != nil {
				return err
			}
			times[i] = p.MS
		}
		sort.Float64s(times)
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"load_ms": e.LoadMS, "warm_median_ms": times[len(times)/2], "warm_min_ms": times[0], "warm_max_ms": times[len(times)-1], "iterations": *iterations, "tokens": p.Tokens, "probabilities": p.Probabilities, "go_heap_bytes": mem.HeapAlloc, "go_sys_bytes": mem.Sys, "provider_requested": *provider, "threads": *threads, "note": "Go heap excludes native model allocations; CoreML request does not prove GPU execution"})
	case "search", "route", "codex":
		query := strings.Join(f.Args(), " ")
		if query == "-" {
			b, err := io.ReadAll(io.LimitReader(os.Stdin, 65537))
			if err != nil {
				return err
			}
			query = string(b)
		}
		op := cmd
		if cmd == "codex" {
			op = "route"
		}
		r := a.Process(app.Request{Op: op, Query: query, CandidateQuery: *candidateQuery, Candidates: *candidates, Limit: *limit, Lexical: *lexical, Model: *explicit})
		if r.Error != "" {
			return errors.New(r.Error)
		}
		if cmd == "codex" {
			route := r.Result.(router.Result)
			argv := []string{}
			if route.Model != "" {
				argv = append(argv, "--model", route.Model)
			}
			argv = append(argv, "--", query)
			if *dry {
				return json.NewEncoder(os.Stdout).Encode(map[string]any{"executable": "codex", "args": argv, "route": route})
			}
			a.Close() // release model memory before Codex starts
			child := exec.Command("codex", argv...)
			child.Stdin = os.Stdin
			child.Stdout = os.Stdout
			child.Stderr = os.Stderr
			return child.Run()
		}
		if *jsonOut {
			return json.NewEncoder(os.Stdout).Encode(r)
		}
		if cmd == "search" {
			result := r.Result.(search.Result)
			for _, w := range result.Warnings {
				fmt.Fprintln(os.Stderr, "note:", w)
			}
			for _, c := range result.Results {
				fmt.Printf("%s:%d-%d  score=%.3f\n%s\n\n", c.Path, c.Start, c.End, c.Score, c.Text)
			}
			fmt.Fprintf(os.Stderr, "%s · %d candidates · %.1f ms\n", result.Engine, result.Candidates, r.MS)
		} else {
			route := r.Result.(router.Result)
			model := route.Model
			if model == "" {
				model = "(Codex default)"
			}
			fmt.Printf("%s → %s\n%s\n", route.Tier, model, route.Reason)
			if route.SuggestedTier != "" {
				fmt.Printf("Laya proposal: %s (%.3f); threshold: %.2f\n", route.SuggestedTier, route.Confidence, *threshold)
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown command %q; run riidolaya help", cmd)
	}
}
func exists(p string) bool { _, err := os.Stat(p); return err == nil }
