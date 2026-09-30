package main

import (
	"encoding/json"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/ternarytrain"
	"github.com/teamswyg/laya-tools/internal/tinyhead"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func bench(dev ternarytrain.Dataset, finalPath, selectionPath, out string) error {
	final, _, e := ternarytrain.Load(finalPath)
	if e != nil {
		return e
	}
	if final.Schema != "riido-qat-final-v1" {
		return fmt.Errorf("expected final feature schema")
	}
	b, e := os.ReadFile(selectionPath)
	if e != nil {
		return e
	}
	var s selection
	if e = json.Unmarshal(b, &s); e != nil {
		return e
	}
	if len(s.Models) != 2 {
		return fmt.Errorf("two sealed models required")
	}
	names := []string{"parent-fp32", "previous-ptq-660b"}
	models := []*tinyhead.Model{}
	m, e := tinyhead.Build(dev.Source, tinyhead.Float32, 0)
	if e != nil {
		return e
	}
	models = append(models, m)
	m, e = tinyhead.Build(dev.Source, tinyhead.Ternary, .5)
	if e != nil {
		return e
	}
	models = append(models, m)
	for _, v := range s.Models {
		if v.File != fmt.Sprintf("seed-%d.rdh", v.Seed) {
			return fmt.Errorf("invalid model path")
		}
		b, e := os.ReadFile(filepath.Join(filepath.Dir(selectionPath), v.File))
		if e != nil {
			return e
		}
		if ternarytrain.Hash(b) != v.SHA {
			return fmt.Errorf("model hash mismatch")
		}
		m, e := tinyhead.Decode(b)
		if e != nil {
			return e
		}
		models = append(models, m)
		names = append(names, v.File)
	}
	type timing struct {
		Name         string    `json:"name"`
		Bytes        int       `json:"bytes"`
		ZeroFraction float64   `json:"zero_fraction"`
		Samples      []float64 `json:"ns_per_op_samples"`
		Allocations  float64   `json:"allocations_per_op"`
	}
	values := make([]timing, len(models))
	for i, m := range models {
		values[i] = timing{Name: names[i], Bytes: m.Bytes(), ZeroFraction: m.ZeroFraction()}
	}
	for repeat := 0; repeat < 5; repeat++ {
		for step := range models {
			i := (step + repeat) % len(models)
			m := models[i]
			scratch := make([]float64, m.Width())
			for j := 0; j < 100; j++ {
				sink, _ = m.Predict(final.Rows[j%len(final.Rows)].Feature, scratch)
			}
			values[i].Allocations = testing.AllocsPerRun(100, func() { sink, _ = m.Predict(final.Rows[0].Feature, scratch) })
			start := time.Now()
			for j := 0; j < 30000; j++ {
				sink, _ = m.Predict(final.Rows[j%len(final.Rows)].Feature, scratch)
			}
			values[i].Samples = append(values[i].Samples, float64(time.Since(start).Nanoseconds())/30000)
		}
	}
	return write(filepath.Join(out, "benchmark.json"), struct {
		Scope  string   `json:"scope"`
		Go     string   `json:"go"`
		Arch   string   `json:"arch"`
		Values []timing `json:"models"`
	}{"paired cached features; five rotated-order repeats of 30000 calls; no encoder/tokenizer/JSON I/O inside timings", runtime.Version(), runtime.GOARCH, values})
}
