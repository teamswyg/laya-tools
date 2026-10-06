// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Materializes original synthetic development examples; no private Riido data.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const origin = "original_authored_synthetic_development"

var intents = [8]string{"question", "blocker", "reference", "progress", "completion_report", "cancel_request", "planned", "unclear"}

type template struct {
	FamilyID string `json:"family_id"`
	Locale   string `json:"locale"`
	Intent   string `json:"intent"`
	Split    string `json:"split"`
	Template string `json:"template"`
}
type input struct {
	Schema    string     `json:"schema"`
	Origin    string     `json:"origin"`
	License   string     `json:"license"`
	Intents   []string   `json:"intents"`
	Variants  int        `json:"variants_per_family"`
	Templates []template `json:"templates"`
}
type row struct {
	ID     string `json:"id"`
	Group  string `json:"group_id"`
	Locale string `json:"locale"`
	Intent string `json:"intent"`
	Split  string `json:"split"`
	Text   string `json:"text"`
	Origin string `json:"origin"`
}

func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func materialize(b []byte) ([]byte, error) {
	var in input
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(&in); err != nil {
		return nil, err
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return nil, fmt.Errorf("trailing template data")
	}
	if len(b) > 1<<20 || !utf8.Valid(b) || in.Schema != "riido-statehint-templates-v1" || in.Origin != origin || in.License != "Apache-2.0" || in.Variants != 25 || len(in.Intents) != 8 || len(in.Templates) != 96 {
		return nil, fmt.Errorf("unsupported template contract")
	}
	for i, s := range intents {
		if in.Intents[i] != s {
			return nil, fmt.Errorf("intent order changed")
		}
	}
	var counts [2][8][4]int
	seenGroups := make(map[string]bool)
	seenTexts := make(map[string]bool)
	var out bytes.Buffer
	e := json.NewEncoder(&out)
	e.SetEscapeHTML(false)
	for _, t := range in.Templates {
		li, ci, si := -1, -1, -1
		for i, s := range []string{"ko", "en"} {
			if t.Locale == s {
				li = i
			}
		}
		for i, s := range intents {
			if t.Intent == s {
				ci = i
			}
		}
		for i, s := range []string{"train", "validation", "calibration", "test"} {
			if t.Split == s {
				si = i
			}
		}
		if li < 0 || ci < 0 || si < 0 || len(t.FamilyID) < 1 || len(t.FamilyID) > 12 || seenGroups[t.FamilyID] || len(t.Template) > 3072 || !strings.Contains(t.Template, "{project}") || !strings.Contains(t.Template, "{item}") {
			return nil, fmt.Errorf("invalid family, locale, intent or template")
		}
		seenGroups[t.FamilyID] = true
		counts[li][ci][si]++
		for j := 0; j < 25; j++ {
			project, item := fmt.Sprintf("FictionalProject-%02d", j), fmt.Sprintf("FictionalItem-%02d", j)
			if li == 0 {
				project, item = fmt.Sprintf("가상프로젝트-%02d", j), fmt.Sprintf("가상항목-%02d", j)
			}
			text := strings.NewReplacer("{project}", project, "{item}", item).Replace(t.Template)
			if strings.ContainsAny(text, "{}") || seenTexts[text] || len(text) > 4096 {
				return nil, fmt.Errorf("duplicate or unsupported substituted text")
			}
			seenTexts[text] = true
			if err := e.Encode(row{fmt.Sprintf("%s-%02d", t.FamilyID, j), t.FamilyID, t.Locale, t.Intent, t.Split, text, origin}); err != nil {
				return nil, err
			}
		}
	}
	for _, locale := range counts {
		for _, class := range locale {
			if class != [4]int{3, 1, 1, 1} {
				return nil, fmt.Errorf("family split imbalance")
			}
		}
	}
	return out.Bytes(), nil
}
func run() error {
	f := flag.NewFlagSet("riido-statehint-data", flag.ContinueOnError)
	source := f.String("templates", "experiments/state-hints/templates.json", "original template source")
	dest := f.String("out", "experiments/state-hints/corpus.jsonl", "public synthetic corpus")
	check := f.Bool("check", false, "verify an existing exact materialization")
	if err := f.Parse(os.Args[1:]); err != nil {
		return err
	}
	b, err := os.ReadFile(*source)
	if err != nil {
		return err
	}
	result, err := materialize(b)
	if err != nil {
		return err
	}
	if *check {
		existing, e := os.ReadFile(*dest)
		if e != nil {
			return e
		}
		if !bytes.Equal(existing, result) {
			return fmt.Errorf("corpus differs from original templates")
		}
	} else {
		if err := os.MkdirAll(filepath.Dir(*dest), 0755); err != nil {
			return err
		}
		file, e := os.OpenFile(*dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if e != nil {
			return e
		}
		_, e = file.Write(result)
		ce := file.Close()
		if e != nil {
			return e
		}
		if ce != nil {
			return ce
		}
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Schema       string `json:"schema"`
		Rows, Groups int
		TemplateSHA  string `json:"templates_sha256"`
		CorpusSHA    string `json:"corpus_sha256"`
		CorpusBytes  int    `json:"corpus_bytes"`
		Scope        string `json:"scope"`
	}{"riido-statehint-corpus-v1", 2400, 96, hash(b), hash(result), len(result), "Original bilingual synthetic development rows; 96 template families, not 2400 independent product observations. Train1200/validation400/calibration400/test400; all partitions remain synthetic development."})
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "riido-statehint-data:", err)
		os.Exit(1)
	}
}
