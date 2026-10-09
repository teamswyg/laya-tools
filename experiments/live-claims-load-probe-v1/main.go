// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// This public probe records timing only; no predictions are saved.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"
)

const warmup = 16
const requests = 1000
const modelSHA = "cfd35ee70a23f94a8d470b7dea596244a91ac7f1410475e8a42959693c99864b"

func main() {
	client := &http.Client{Timeout: 2 * time.Second}
	defer client.CloseIdleConnections()
	body := []byte(`{"text":"로컬 예제 문장을 입력합니다."}`)
	latency := make([]int64, requests)
	inference := make([]int64, requests)
	var started time.Time
	for i := -warmup; i < requests; i++ {
		if i == 0 {
			started = time.Now()
		}
		at := time.Now()
		r, err := client.Post("http://127.0.0.1:8877/api/hints", "application/json", bytes.NewReader(body))
		if err != nil {
			fail("local request failed")
		}
		data, err := io.ReadAll(io.LimitReader(r.Body, 32769))
		r.Body.Close()
		if err != nil || r.StatusCode != http.StatusOK || len(data) > 32768 {
			fail("local response failed")
		}
		var metrics struct {
			Schema      string `json:"schema"`
			State       string `json:"state"`
			InferenceUS int64  `json:"inference_us"`
			Model       struct {
				SHA string `json:"artifact_sha256"`
			} `json:"model"`
		}
		if json.Unmarshal(data, &metrics) != nil || metrics.State != "ready" || metrics.Schema != "riidolaya-live-claims-v1" || metrics.Model.SHA != modelSHA || metrics.InferenceUS < 0 {
			fail("invalid local response")
		}
		if i >= 0 {
			latency[i] = time.Since(at).Nanoseconds()
			inference[i] = metrics.InferenceUS
		}
	}
	duration := time.Since(started)
	sort.Slice(latency, func(i, j int) bool { return latency[i] < latency[j] })
	sort.Slice(inference, func(i, j int) bool { return inference[i] < inference[j] })
	percentile := func(values []int64, n int) int64 { return values[(len(values)*n+99)/100-1] }
	result := struct {
		Schema            string  `json:"schema"`
		ModelSHA          string  `json:"model_sha256"`
		UTC               string  `json:"measured_utc"`
		Warmup            int     `json:"warmup_requests"`
		Requests          int     `json:"timed_requests"`
		WallNS            int64   `json:"wall_ns"`
		RequestsPerSecond float64 `json:"requests_per_second"`
		HTTPP50NS         int64   `json:"http_p50_ns"`
		HTTPP95NS         int64   `json:"http_p95_ns"`
		HTTPP99NS         int64   `json:"http_p99_ns"`
		InferenceP50US    int64   `json:"inference_p50_us"`
		InferenceP95US    int64   `json:"inference_p95_us"`
		InferenceP99US    int64   `json:"inference_p99_us"`
		Limits            string  `json:"limits"`
	}{
		"riidolaya-local-serial-load-probe-v1", modelSHA, time.Now().UTC().Format(time.RFC3339Nano),
		warmup, requests, duration.Nanoseconds(), float64(requests) / duration.Seconds(),
		percentile(latency, 50), percentile(latency, 95), percentile(latency, 99),
		percentile(inference, 50), percentile(inference, 95), percentile(inference, 99),
		"One repeated newly authored sentence, serial warm local HTTP; predictions discarded. Not semantic accuracy, diverse workloads, cold start, peak server memory, GPU use, or cost savings. HTTP time includes client JSON decoding; server inference time excludes HTTP and browser debounce. Other work may run concurrently.",
	}
	if json.NewEncoder(os.Stdout).Encode(result) != nil {
		fail("cannot write metrics")
	}
}

func fail(message string) { fmt.Fprintln(os.Stderr, message); os.Exit(1) }
