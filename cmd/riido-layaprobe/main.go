package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/assets"
	"github.com/teamswyg/laya-tools/internal/inference"
	"github.com/teamswyg/laya-tools/internal/layaprobe"
	"github.com/teamswyg/laya-tools/internal/paireval"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func run() error {
	input := flag.String("input", ".cache/semantic-scale/cosqa-all.json", "pinned CoSQA source")
	model := flag.String("model-dir", ".cache/models-v2/code", "pinned code checkpoint")
	native := flag.String("runtime", assets.Runtime(), "pinned ONNX Runtime library")
	flag.Parse()
	start := time.Now()
	plan, err := os.ReadFile("experiments/laya-relevance/plan-17.json")
	if err != nil {
		return err
	}
	if paireval.Hash(plan) != layaprobe.PlanSHA256 {
		return fmt.Errorf("plan mismatch")
	}
	if err := assets.Verify(*model, map[string]string{"model.onnx": layaprobe.ModelSHA256, "config.json": "febbcac345d8142d649567746a83446e458c49b0e61256811ac0b36e72a1169f", "tokenizer.json": "6c8aaa9a542084f2457eab775d4eeb51f92a70c0fd9de28d5edb0ddec3c08d30"}); err != nil {
		return err
	}
	// This local resource probe pins the measured macOS ARM64 runtime explicitly.
	if err := assets.Verify(filepath.Dir(*native), map[string]string{filepath.Base(*native): "bcc9110f9d638a119de2db7afb3ba9a1da8085f0cb3401e1c48ae1caf450b6fa"}); err != nil {
		return err
	}
	rows, err := paireval.Load(*input)
	if err != nil {
		return err
	}
	split := paireval.Partition(rows)
	if split.MembershipSHA256 != "d90eda8e8a423eff44847ff46ef19b468a57ac2791d45c57e7e8ba39a4ee9902" || split.Counts["validation"] != 607 {
		return fmt.Errorf("membership mismatch")
	}
	peak := int64(0)
	guard := func() error {
		if time.Since(start) > time.Hour {
			return fmt.Errorf("one-hour budget exceeded")
		}
		b, err := exec.Command("ps", "-o", "rss=", "-p", strconv.Itoa(os.Getpid())).Output()
		if err != nil {
			return fmt.Errorf("RSS monitor failed")
		}
		kb, err := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
		if err != nil || kb < 0 {
			return fmt.Errorf("invalid RSS observation")
		}
		rss := kb * 1024
		peak = max(peak, rss)
		if rss > 8<<30 {
			return fmt.Errorf("process RSS exceeds 8 GiB")
		}
		return nil
	}
	if err := guard(); err != nil {
		return err
	}
	engine, err := inference.New(inference.Options{ModelDir: *model, Runtime: *native, Provider: "cpu", Threads: 4, MaxTokens: 512})
	if err != nil {
		return err
	}
	defer engine.Close()
	last := time.Now()
	report, err := layaprobe.Run(rows, split, engine, guard, func(cases, calls int) {
		if time.Since(last) >= 20*time.Second {
			fmt.Fprintf(os.Stderr, "validation cases visited=%d native calls=%d elapsed_seconds=%.1f\n", cases, calls, time.Since(start).Seconds())
			last = time.Now()
		}
	})
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(struct {
		Report                 layaprobe.Report
		LoadMS, TotalSeconds   float64
		MaximumSampledRSSBytes int64
	}{report, engine.LoadMS, time.Since(start).Seconds(), peak})
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
