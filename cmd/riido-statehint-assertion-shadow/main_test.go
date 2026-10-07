// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func ownedText(text, role string) observation {
	return observation{ID: "owned-observation", Type: "text_observation", Role: role, Text: text, Actor: "owned-actor", Unit: "owned widget", Timestamp: "2026-01-02T03:04:05Z", Version: "owned-v1", Source: "owned arithmetic/pipeline fixture; no human semantic truth"}
}
func hasReason(e evidence, s string) bool {
	for _, x := range e.Reasons {
		if x == s {
			return true
		}
	}
	return false
}
func evidenceKind(t *testing.T, o observation, kind string) evidence {
	t.Helper()
	for _, e := range extract(o) {
		if e.Kind == kind {
			return e
		}
	}
	t.Fatal("owned literal absent", kind)
	return evidence{}
}
func TestReplySpanKeepsOriginalUTF8ByteOffsets(t *testing.T) {
	text := "안녕하세요. Please explain the owned widget boundary."
	e := evidenceKind(t, ownedText(text, "comment"), "reply_claim")
	start := strings.Index(text, "Please explain")
	if e.Start != start || e.End != start+len("Please explain") || text[e.Start:e.End] != e.Literal || e.Literal != "Please explain" || !e.Unlearned || e.SemanticJudgment || e.Confidence != nil || !e.DisplayCandidate {
		t.Fatal("unlearned literal/span/provenance", e)
	}
	// ASCII-only HTML case handling cannot shift a Unicode byte boundary.
	html := "İ <P>Please explain the owned widget.</P>"
	e = evidenceKind(t, ownedText(html, "comment"), "reply_claim")
	if e.Start != strings.Index(html, "Please explain") || html[e.Start:e.End] != e.Literal {
		t.Fatal("HTML byte offsets changed", e)
	}
}
func TestQuotesCodeAttributesAndUnsupportedMarkupCannotDisplay(t *testing.T) {
	for _, text := range []string{`"Please explain the owned widget."`, "`Please explain the owned widget.`", "```\nPlease explain the owned widget.\n```", "<blockquote>Please explain the owned widget.</blockquote>", "<code>Please explain the owned widget.</code>", "> Please explain the owned widget.", "<script>Please explain the owned widget.</script>"} {
		e := evidenceKind(t, ownedText(text, "comment"), "reply_claim")
		if e.DisplayCandidate || len(e.Reasons) == 0 {
			t.Fatal("quoted/code message displayed", e)
		}
	}
	if got := extract(ownedText(`<a href="https://example.invalid/?x=Please explain">Owned link</a>`, "comment")); len(got) != 0 {
		t.Fatal("attribute/URL became reply", got)
	}
}
func TestUnclosedQuoteCodeAndMarkupTailsCannotDisplay(t *testing.T) {
	for _, tc := range []struct{ prefix, reason string }{
		{"<code>", "quoted_or_code"},
		{"<pre\nclass=\"owned\">", "quoted_or_code"},
		{"<blockquote\tclass=\"owned\">", "quoted_or_code"},
		{"<script>", "unsupported_markup"},
		{"<style>", "unsupported_markup"},
		{"<code\n", "unsupported_markup"},
		{"`", "quoted_or_code"},
		{"```\n", "quoted_or_code"},
		{"\"", "quoted_or_code"},
		{"“", "quoted_or_code"},
	} {
		t.Run(tc.prefix, func(t *testing.T) {
			text := "İ 한글 " + tc.prefix + "Please explain the owned widget."
			e := evidenceKind(t, ownedText(text, "comment"), "reply_claim")
			if e.DisplayCandidate || !hasReason(e, tc.reason) || e.Start != strings.Index(text, "Please explain") || text[e.Start:e.End] != e.Literal {
				t.Fatal("unresolved tail displayed or changed original offsets", e)
			}
		})
	}
	text := "Please explain the owned widget. `Please confirm its boundary."
	got := extract(ownedText(text, "comment"))
	if len(got) != 2 || !got[0].DisplayCandidate || got[1].DisplayCandidate || !hasReason(got[1], "quoted_or_code") {
		t.Fatal("unclosed tail changed the preceding unquoted candidate", got)
	}
}
func TestProtectedHTMLTagsRecognizeWhitespaceAndPreserveOffsets(t *testing.T) {
	for _, space := range []string{" ", "\t", "\n", "\r", "\f"} {
		text := "İ <PRE" + space + "class=\"owned\">Please explain the owned widget.</PRE> Please confirm its boundary."
		got := extract(ownedText(text, "comment"))
		if len(got) != 2 || got[0].DisplayCandidate || !hasReason(got[0], "quoted_or_code") || !got[1].DisplayCandidate {
			t.Fatal("HTML whitespace did not delimit the protected region", got)
		}
		for _, e := range got {
			if text[e.Start:e.End] != e.Literal {
				t.Fatal("HTML whitespace changed original byte offsets", e)
			}
		}
	}
}
func TestNegatedFutureClosureAndRequiredRemainderAlwaysWithheld(t *testing.T) {
	for _, tc := range []struct{ text, reason string }{{"I have not completed the owned widget.", "negated_context"}, {"Tomorrow I will mark the owned widget completed.", "future_or_conditional_context"}, {"I completed the owned widget. Required integration is still remaining.", "required_remainder_scope_unresolved"}} {
		e := evidenceKind(t, ownedText(tc.text, "comment"), "closure_claim")
		if e.DisplayCandidate || !e.CompletionWithheld || !hasReason(e, tc.reason) || !hasReason(e, "completion_display_always_withheld") || e.Confidence != nil {
			t.Fatal("completion escaped withholding", e)
		}
	}
	e := evidenceKind(t, ownedText("I completed the owned widget.", "comment"), "closure_claim")
	if e.DisplayCandidate || !e.CompletionWithheld {
		t.Fatal("clear literal improperly promoted", e)
	}
}
func TestActivityScopeAndPoliteReplyAreSeparateCandidates(t *testing.T) {
	o := ownedText("I am testing the owned widget. Please confirm its boundary.", "comment")
	activity := evidenceKind(t, o, "activity_claim")
	reply := evidenceKind(t, o, "reply_claim")
	if !activity.DisplayCandidate || !reply.DisplayCandidate || activity.Confidence != nil || reply.Confidence != nil {
		t.Fatal("independent unlearned candidates", activity, reply)
	}
	e := evidenceKind(t, ownedText("The reference describes testing.", "comment"), "activity_claim")
	if e.DisplayCandidate || !hasReason(e, "literal_activity_scope_unresolved") {
		t.Fatal("reference activity promoted", e)
	}
	e = evidenceKind(t, ownedText("Could you explain the owned widget?", "comment"), "reply_claim")
	if hasReason(e, "future_or_conditional_context") {
		t.Fatal("polite request modal became a plan", e)
	}
}
func TestBodyOnlyGapAndObservedCreatedNeverMeanDone(t *testing.T) {
	body := ownedText("Please explain the owned widget. I completed it.", "body_snapshot")
	in := input{Schema: "riido-assertion-shadow-input-v1", Observations: []observation{body}}
	for i, kind := range []string{"CommentObserved", "CommentCreated", "ProgressObserved", "CommandSucceeded", "TurnEnded"} {
		in.Observations = append(in.Observations, observation{ID: string(rune('a' + i)), Type: "observed_event", Kind: kind, Source: "owned typed-event replay", Unit: "owned widget"})
	}
	r := shadow(in, strings.Repeat("a", 64))
	if r.CommentTexts != 0 || r.BodySnapshots != 1 || len(r.InputGap) != 1 || len(r.Facts) != 5 || !r.CompletionWithheld || r.StateWrites != 0 || r.ModelCalls != 0 || r.FitCalls != 0 || r.Qualified || r.GateApplied || r.ExistingGates != [2]float64{.9, .05} {
		t.Fatal("body/event truth or existing criteria changed", r)
	}
	for _, e := range r.Evidence {
		if e.DisplayCandidate || !hasReason(e, "body_snapshot_is_not_a_comment") {
			t.Fatal("body counted as live comment", e)
		}
	}
	for _, f := range r.Facts {
		if !f.InputAttestationOnly || f.CompletionAuthority {
			t.Fatal("typed event acquired completion authority", f)
		}
	}
}
func TestBoundedSchemaPinnedInputAndPrivateExclusiveReport(t *testing.T) {
	in := input{Schema: "riido-assertion-shadow-input-v1", Observations: []observation{ownedText("Please explain the owned widget.", "comment")}}
	data, e := json.Marshal(in)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := decode(data); e != nil {
		t.Fatal(e)
	}
	for _, change := range []func(*input){func(i *input) { i.Observations[0].Role = "task_status" }, func(i *input) { i.Observations[0].Text = strings.Repeat("x", 4097) }, func(i *input) { i.Observations = append(i.Observations, i.Observations[0]) }, func(i *input) { i.Observations[0].Timestamp = "unknown-date" }} {
		var bad input
		json.Unmarshal(data, &bad)
		change(&bad)
		b, _ := json.Marshal(bad)
		if _, e := decode(b); e == nil {
			t.Fatal("bad input accepted")
		}
	}
	if _, e := decode(append(append([]byte(nil), data...), []byte(" {}")...)); e == nil {
		t.Fatal("trailing metadata")
	}
	dir := t.TempDir()
	root, e := os.OpenRoot(dir)
	if e != nil {
		t.Fatal(e)
	}
	defer root.Close()
	os.WriteFile(filepath.Join(dir, "owned.json"), data, 0600)
	if got, e := readInput(root, "owned.json", digest(data)); e != nil || !bytes.Equal(got, data) {
		t.Fatal("exact pin")
	}
	if _, e := readInput(root, "owned.json", strings.Repeat("0", 64)); e == nil {
		t.Fatal("wrong pin")
	}
	os.Symlink("owned.json", filepath.Join(dir, "link.json"))
	if _, e := readInput(root, "link.json", digest(data)); e == nil {
		t.Fatal("symlink input")
	}
	out, e := privateOutput(root, anchor+"/owned-run")
	if e != nil {
		t.Fatal(e)
	}
	out.Close()
	if _, e := privateOutput(root, anchor+"/owned-run"); e == nil {
		t.Fatal("run overwritten")
	}
	st, e := root.Stat(anchor + "/owned-run")
	if e != nil || st.Mode().Perm() != 0700 {
		t.Fatal("nonprivate output")
	}
}
func TestActualCLIUsesOnlyOwnedInputAndNoAppOrModelPaths(t *testing.T) {
	t.Chdir(t.TempDir())
	in := input{Schema: "riido-assertion-shadow-input-v1", Observations: []observation{ownedText("Please explain the owned widget.", "comment")}}
	data, e := json.Marshal(in)
	if e != nil {
		t.Fatal(e)
	}
	if os.WriteFile("owned.json", data, 0600) != nil {
		t.Fatal("owned fixture")
	}
	args := []string{"--input", "owned.json", "--input-sha256", digest(data)}
	var out, errOut bytes.Buffer
	if e := run(append(append([]string(nil), args...), "--check"), &out, &errOut); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(anchor); !os.IsNotExist(e) {
		t.Fatal("check created output")
	}
	out.Reset()
	if e := run(append(append([]string(nil), args...), "--out", anchor+"/owned-cli"), &out, &errOut); e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(anchor + "/owned-cli/report.json")
	var r report
	if e != nil || json.Unmarshal(raw, &r) != nil || r.CommentTexts != 1 || len(r.Evidence) != 1 || r.StateWrites != 0 || r.ModelCalls != 0 || r.FitCalls != 0 || !r.CompletionWithheld || r.Qualified {
		t.Fatal("actual owned CLI report")
	}
	st, e := os.Stat(anchor + "/owned-cli/report.json")
	if e != nil || st.Mode().Perm() != 0600 {
		t.Fatal("nonprivate report")
	}
	for _, bad := range [][]string{nil, {"--model", "private-marker"}, {"--train", "private-marker"}, {"--status", "private-marker"}} {
		var stdout, stderr bytes.Buffer
		e := run(bad, &stdout, &stderr)
		if e == nil || stdout.Len() != 0 || strings.Contains(stderr.String()+e.Error(), "private-marker") {
			t.Fatal("unsupported model/app operation", e)
		}
	}
}
