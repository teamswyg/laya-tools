// verify reads captured public evidence only. It never opens a network connection.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
)

const modelPin = "cfd35ee70a23f94a8d470b7dea596244a91ac7f1410475e8a42959693c99864b"
const sourceInputPin = "7c241bca2facb9eb15b8bcd1785b54aa5171d92b3e677beea4f8139e92a583d3"

var heads = []string{"response_requested", "current_activity_claimed", "completion_claimed"}
var states = []string{"true", "false", "unknown"}

type counts map[string]int
type input struct {
	ID, Locale, Text string
	Expected         map[string]string
}
type prediction struct {
	Head, Winner, State string
	Probabilities       [3]float64
	Confidence, Margin  float64
	UnknownReason       string `json:"unknown_reason"`
}
type response struct {
	Schema, State string
	Prediction    struct{ Heads []prediction }
	InferenceUS   int `json:"inference_us"`
	InputBytes    int `json:"input_bytes"`
	Model         struct {
		ArtifactSHA string `json:"artifact_sha256"`
	}
}
type record struct {
	ID, Locale     string
	Attempts       int
	RequestNumber  int             `json:"request_number"`
	RequestBytes   int             `json:"request_bytes"`
	InputUTF8Bytes int             `json:"input_utf8_bytes"`
	InputSHA       string          `json:"input_text_sha256"`
	RequestDate    string          `json:"request_date_utc"`
	ResponseDate   string          `json:"response_date_utc"`
	HTTPStatus     int             `json:"http_status"`
	HTTPMS         float64         `json:"http_elapsed_ms"`
	ResponseBody   string          `json:"response_body"`
	ParsedResponse json.RawMessage `json:"parsed_response"`
}
type row struct {
	Locale, Head       string
	InputCount         int            `json:"input_count"`
	ExpectedCounts     counts         `json:"expected_counts"`
	RawCounts          counts         `json:"raw_counts"`
	FinalCounts        counts         `json:"final_counts"`
	Accepted           int            `json:"accepted_positive_count"`
	RawDiscrepancies   int            `json:"expected_vs_raw_discrepancies"`
	FinalDiscrepancies int            `json:"expected_vs_final_discrepancies"`
	RawFinalChanges    int            `json:"raw_vs_final_changes"`
	UnknownReasons     map[string]int `json:"unknown_reason_counts"`
}
type totals struct {
	HeadOutcomes       int    `json:"head_outcomes"`
	Accepted           int    `json:"accepted_positive_count"`
	RawDiscrepancies   int    `json:"expected_vs_raw_discrepancies"`
	FinalDiscrepancies int    `json:"expected_vs_final_discrepancies"`
	RawFinalChanges    int    `json:"raw_vs_final_changes"`
	FinalCounts        counts `json:"final_counts"`
}
type timingStats struct {
	N      int     `json:"n"`
	Min    float64 `json:"min"`
	Median float64 `json:"median"`
	Mean   float64 `json:"mean"`
	P95    float64 `json:"p95_nearest_rank"`
	Max    float64 `json:"max"`
	Total  float64 `json:"total"`
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
func require(ok bool, message string) {
	if !ok {
		fail("evidence check failed: %s", message)
	}
}
func read(dir, name string) []byte {
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		fail("read %s: %v", name, err)
	}
	return b
}
func decode(b []byte, target any) {
	if err := json.Unmarshal(b, target); err != nil {
		fail("invalid JSON: %v", err)
	}
}
func digest(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}
func newCounts() counts        { return counts{"true": 0, "false": 0, "unknown": 0} }
func validState(s string) bool { return s == "true" || s == "false" || s == "unknown" }
func date(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		fail("invalid evidence date: %v", err)
	}
	return t
}
func describe(values []float64) timingStats {
	x := append([]float64(nil), values...)
	sort.Float64s(x)
	total := 0.0
	for _, v := range values {
		total += v
	}
	n := len(x)
	return timingStats{n, x[0], (x[n/2-1] + x[n/2]) / 2, total / float64(n), x[int(math.Ceil(float64(n)*0.95))-1], x[n-1], total}
}

func main() {
	dir := flag.String("dir", "experiments/live-claims-behavior-check-v1", "public evidence directory")
	flag.Parse()
	manifest := bufio.NewScanner(strings.NewReader(string(read(*dir, "SHA256SUMS"))))
	listed := map[string]bool{}
	for manifest.Scan() {
		parts := strings.SplitN(manifest.Text(), "  ", 2)
		require(len(parts) == 2 && len(parts[0]) == 64 && filepath.Base(parts[1]) == parts[1], "manifest format")
		require(!listed[parts[1]], "duplicate manifest entry")
		require(digest(read(*dir, parts[1])) == parts[0], "manifest digest: "+parts[1])
		listed[parts[1]] = true
	}
	if err := manifest.Err(); err != nil {
		fail("manifest: %v", err)
	}
	files, err := os.ReadDir(*dir)
	if err != nil {
		fail("public directory: %v", err)
	}
	bytes := int64(0)
	for _, file := range files {
		require(!file.IsDir(), "unexpected evidence subdirectory")
		require(file.Name() == "SHA256SUMS" || listed[file.Name()], "unlisted file: "+file.Name())
		info, err := file.Info()
		if err != nil {
			fail("public file metadata: %v", err)
		}
		bytes += info.Size()
	}
	require(bytes <= 512*1024, "public bundle exceeds 512 KiB")
	var inputs struct{ Entries []input }
	decode(read(*dir, "inputs.json"), &inputs)
	var evidence struct{ Records []record }
	decode(read(*dir, "responses.json"), &evidence)
	var lock struct {
		Date  string `json:"date_utc"`
		SHA   string `json:"locked_inputs_sha256"`
		Model string `json:"expected_model_pin"`
	}
	decode(read(*dir, "pre-inference-lock.json"), &lock)
	require(lock.SHA == sourceInputPin && lock.Model == modelPin, "original lock pins")
	var provenance struct {
		InputsSHA        string `json:"public_inputs_sha256"`
		ResponsesSHA     string `json:"public_responses_sha256"`
		PublicationCalls int    `json:"publication_model_inferences"`
	}
	decode(read(*dir, "provenance.json"), &provenance)
	require(provenance.InputsSHA == digest(read(*dir, "inputs.json")), "public input pin")
	require(provenance.ResponsesSHA == digest(read(*dir, "responses.json")), "public response pin")
	require(provenance.PublicationCalls == 0, "publication inference count")
	var status struct {
		Before, After struct {
			HTTPStatus int `json:"http_status"`
			Body       string
		}
	}
	decode(read(*dir, "status-receipts.json"), &status)
	for _, s := range []struct {
		HTTPStatus int `json:"http_status"`
		Body       string
	}{status.Before, status.After} {
		var body struct {
			State string
			Model struct {
				SHA       string `json:"artifact_sha256"`
				Qualified bool   `json:"semantic_quality_qualified"`
			}
			Thresholds struct {
				Confidence  float64 `json:"confidence_floor"`
				Margin      float64 `json:"margin_floor"`
				Temperature float64
			}
		}
		decode([]byte(s.Body), &body)
		require(s.HTTPStatus == 200 && body.State == "ready" && body.Model.SHA == modelPin && !body.Model.Qualified, "status pin/readiness")
		require(body.Thresholds.Confidence == .9 && body.Thresholds.Margin == .05 && body.Thresholds.Temperature == 1, "fixed status thresholds")
	}
	var exclusion struct {
		Variants bool `json:"variants_included"`
		PerText  []struct {
			ID  string
			SHA string `json:"sha256"`
		} `json:"per_text_sha256"`
		Destinations []string `json:"excluded_destinations"`
	}
	decode(read(*dir, "exclusion.json"), &exclusion)
	require(exclusion.Variants && len(exclusion.PerText) == 24, "24 originals and variant exclusion")
	require(reflect.DeepEqual(exclusion.Destinations, []string{"final400", "final1200", "final2400", "future_model_selectors"}), "exclusion destinations")
	require(len(inputs.Entries) == 24 && len(evidence.Records) == 24, "24 original inputs and responses")
	rows := []row{}
	for _, locale := range []string{"ko", "en"} {
		for _, head := range heads {
			rows = append(rows, row{Locale: locale, Head: head, ExpectedCounts: newCounts(), RawCounts: newCounts(), FinalCounts: newCounts(), UnknownReasons: map[string]int{}})
		}
	}
	seen := map[string]bool{}
	localeCounts := map[string]int{}
	httpTimes, inferenceTimes := []float64{}, []float64{}
	for i, in := range inputs.Entries {
		r := evidence.Records[i]
		require(!seen[in.ID] && r.ID == in.ID && r.Locale == in.Locale && r.Attempts == 1 && r.RequestNumber == i+1 && r.HTTPStatus == 200, "unique ordered exactly-once response: "+in.ID)
		seen[in.ID] = true
		localeCounts[in.Locale]++
		require(r.InputUTF8Bytes == len([]byte(in.Text)) && r.InputSHA == digest([]byte(in.Text)), "input byte/digest receipt: "+in.ID)
		require(exclusion.PerText[i].ID == in.ID && exclusion.PerText[i].SHA == r.InputSHA, "exclusion text digest: "+in.ID)
		require(!date(r.RequestDate).Before(date(lock.Date)) && !date(r.ResponseDate).Before(date(r.RequestDate)), "lock/request chronology: "+in.ID)
		request, err := json.Marshal(map[string]string{"text": in.Text})
		if err != nil {
			fail("request encoding: %v", err)
		}
		require(len(request) == r.RequestBytes, "request byte receipt: "+in.ID)
		var rawJSON, parsedJSON any
		decode([]byte(r.ResponseBody), &rawJSON)
		decode(r.ParsedResponse, &parsedJSON)
		require(reflect.DeepEqual(rawJSON, parsedJSON), "complete body/parsed evidence match: "+in.ID)
		var p response
		decode([]byte(r.ResponseBody), &p)
		require(p.Schema == "riidolaya-live-claims-v1" && p.State == "ready" && p.Model.ArtifactSHA == modelPin && p.InputBytes == len([]byte(in.Text)) && len(p.Prediction.Heads) == 3, "response schema/pin: "+in.ID)
		require(r.HTTPMS >= 0 && p.InferenceUS >= 0, "recorded timing: "+in.ID)
		httpTimes = append(httpTimes, r.HTTPMS)
		inferenceTimes = append(inferenceTimes, float64(p.InferenceUS))
		seenHeads := map[string]bool{}
		for _, h := range p.Prediction.Heads {
			rowIndex := -1
			for j := range rows {
				if rows[j].Locale == in.Locale && rows[j].Head == h.Head {
					rowIndex = j
				}
			}
			require(rowIndex >= 0 && !seenHeads[h.Head] && validState(h.Winner) && validState(h.State) && validState(in.Expected[h.Head]), "three valid heads: "+in.ID)
			seenHeads[h.Head] = true
			probs := append([]float64(nil), h.Probabilities[:]...)
			winner, sum := 0, 0.0
			for j, value := range probs {
				require(value >= 0 && value <= 1 && !math.IsNaN(value), "probability bounds")
				sum += value
				if value > probs[winner] {
					winner = j
				}
			}
			sort.Sort(sort.Reverse(sort.Float64Slice(probs)))
			require(math.Abs(sum-1) < 1e-9 && h.Winner == states[winner] && math.Abs(h.Confidence-probs[0]) < 1e-9 && math.Abs(h.Margin-(probs[0]-probs[1])) < 1e-9, "score arithmetic: "+in.ID)
			final := h.Winner
			if final == "unknown" || h.Confidence < .9 || h.Margin < .05 {
				final = "unknown"
			}
			require(h.State == final, "fixed threshold application: "+in.ID)
			x := &rows[rowIndex]
			x.InputCount++
			x.ExpectedCounts[in.Expected[h.Head]]++
			x.RawCounts[h.Winner]++
			x.FinalCounts[h.State]++
			if h.State == "true" {
				x.Accepted++
			}
			if h.Winner != in.Expected[h.Head] {
				x.RawDiscrepancies++
			}
			if h.State != in.Expected[h.Head] {
				x.FinalDiscrepancies++
			}
			if h.Winner != h.State {
				x.RawFinalChanges++
			}
			if h.UnknownReason != "" {
				x.UnknownReasons[h.UnknownReason]++
			}
			if in.ID == "EN04" && h.Head == "current_activity_claimed" {
				require(in.Expected[h.Head] == "false" && h.State == "true" && h.Confidence == 0.946169690250911 && h.Margin == 0.9008599783475525, "negated-activity limitation score")
			}
		}
	}
	require(localeCounts["ko"] == 12 && localeCounts["en"] == 12, "12 KO and 12 EN inputs")
	total := totals{FinalCounts: newCounts()}
	for _, x := range rows {
		total.HeadOutcomes += x.InputCount
		total.Accepted += x.Accepted
		total.RawDiscrepancies += x.RawDiscrepancies
		total.FinalDiscrepancies += x.FinalDiscrepancies
		total.RawFinalChanges += x.RawFinalChanges
		for _, state := range states {
			total.FinalCounts[state] += x.FinalCounts[state]
		}
	}
	var published struct {
		Rows    []row `json:"by_locale_head"`
		Totals  totals
		Timings struct {
			HTTP      timingStats `json:"client_http_elapsed_ms"`
			Inference timingStats `json:"server_reported_inference_us"`
		}
	}
	decode(read(*dir, "results.json"), &published)
	require(reflect.DeepEqual(rows, published.Rows) && reflect.DeepEqual(total, published.Totals), "published aggregate readback")
	require(reflect.DeepEqual(describe(httpTimes), published.Timings.HTTP) && reflect.DeepEqual(describe(inferenceTimes), published.Timings.Inference), "original timing aggregate readback")
	fmt.Printf("PASS: %d captured responses; %d shared-input head outputs; accepted=%d; expected!=raw=%d; expected!=final=%d; final T/F/U=%d/%d/%d; bundle=%d bytes; no network calls\n", len(evidence.Records), total.HeadOutcomes, total.Accepted, total.RawDiscrepancies, total.FinalDiscrepancies, total.FinalCounts["true"], total.FinalCounts["false"], total.FinalCounts["unknown"], bytes)
}
