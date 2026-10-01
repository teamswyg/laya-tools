package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

func publicCorpusForTest(t *testing.T) (Corpus, []byte, []byte) {
	t.Helper()
	legacy, e := os.ReadFile(filepath.Join("..", "..", legacyPath))
	if e != nil {
		t.Fatal(e)
	}
	typed, e := os.ReadFile(filepath.Join("..", "..", typedPath))
	if e != nil {
		t.Fatal(e)
	}
	c, e := prepareCorpus(legacy, typed)
	if e != nil {
		t.Fatal(e)
	}
	return c, legacy, typed
}

func copyCorpusForTest(t *testing.T, c Corpus) Corpus {
	t.Helper()
	raw, e := json.Marshal(c)
	if e != nil {
		t.Fatal(e)
	}
	var out Corpus
	if e = json.Unmarshal(raw, &out); e != nil {
		t.Fatal(e)
	}
	return out
}

func TestPreparePublicOriginProjectionWithoutTruth(t *testing.T) {
	c, legacy, typed := publicCorpusForTest(t)
	if c.Schema != "riido-resident-public-wire-corpus-v2" || c.WireEncoding != "base64_exact_compact_json_without_lf_v1" || c.FeatureEncoding != featureEncoding || c.OriginalParents != 72 || len(c.Origins) != 72 || len(c.Payloads) < 2 {
		t.Fatal("public protocol-only identity/count gate failed")
	}
	// Independently decode metadata from the two original bytes, without using
	// FeatureInputs, source registries, candidate code, truth, or any model.
	type original struct {
		Parents []struct {
			ID         string `json:"id"`
			Request    string `json:"request"`
			Candidates []struct {
				ID   string `json:"id"`
				Text string `json:"text"`
			} `json:"candidates"`
		} `json:"parents"`
	}
	offset := 0
	for _, fixture := range []struct {
		name  string
		raw   []byte
		sha   string
		count int
	}{
		{"probes-56", legacy, legacySHA, 48}, {"probes-56b", typed, typedSHA, 24},
	} {
		var source original
		if hashBytes(fixture.raw) != fixture.sha || json.Unmarshal(fixture.raw, &source) != nil || len(source.Parents) != fixture.count {
			t.Fatal("original byte pin failed")
		}
		for i, parent := range source.Parents {
			link := c.Origins[offset+i]
			if link.Fixture != fixture.name || link.ParentID != parent.ID || link.OriginalIndex != i || link.CandidateCount != len(parent.Candidates) {
				t.Fatal("original parent/order/count metadata changed")
			}
			for j, candidate := range parent.Candidates {
				if link.OriginalCandidateIDs[j] != candidate.ID {
					t.Fatal("original candidate ID/order metadata changed")
				}
			}
			if len(parent.Candidates) != 3 {
				if link.PayloadIndex != -1 || link.ExcludedReason != "original_candidate_count_not_three" || link.FeatureSHA256 != "" {
					t.Fatal("non-3 original set was resized or selected")
				}
				continue
			}
			ids := candidateIDs()
			input := shortclaim.Input{Schema: textSchema, Request: parent.Request, Provenance: provenance, Candidates: make([]shortclaim.Candidate, 3)}
			for j, candidate := range parent.Candidates {
				input.Candidates[j] = shortclaim.Candidate{ID: ids[j], Text: candidate.Text}
			}
			p, e := shortclaim.Validate(input)
			if e != nil || link.FeatureSHA256 != featureSHA(p) || c.Payloads[link.PayloadIndex].FeatureSHA256 != link.FeatureSHA256 {
				t.Fatal("ordered original text projection lost its feature binding")
			}
		}
		offset += fixture.count
	}
	if validateCorpus(c) != nil {
		t.Fatal("prepared protocol-only corpus did not validate")
	}
	for _, changed := range [][]byte{nil, append(slices.Clone(legacy), ' '), bytes.Replace(legacy, []byte("request"), []byte("Request"), 1)} {
		got, e := prepareCorpus(changed, typed)
		if e == nil || got.Schema != "" || got.Payloads != nil || got.Origins != nil {
			t.Fatal("different original bytes produced a partial/accepted corpus")
		}
	}
	if _, e := prepareCorpus(legacy, append(slices.Clone(typed), '\n')); e == nil {
		t.Fatal("typed original bytes were silently reformatted")
	}
}

func TestCorpusBase64RoundTripPreservesExactWire(t *testing.T) {
	c, legacy, typed := publicCorpusForTest(t)
	raw, e := encodePublic(c, maxCorpusBytes)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Contains(raw, []byte(`"wire_base64": "`)) || bytes.Contains(raw, []byte(`"wire_base64": {`)) {
		t.Fatal("wire encoding does not preserve opaque bytes")
	}
	var decoded Corpus
	if strictJSON(raw, &decoded) != nil || validateCorpus(decoded) != nil {
		t.Fatal("base64 corpus round trip failed")
	}
	ready, e := bindCorpus(raw, legacy, typed, Plan{CorpusSHA256: hashBytes(raw)})
	if e != nil || len(ready.Payloads) != len(c.Payloads) {
		t.Fatal("exact original projection binding failed")
	}
	for i, payload := range c.Payloads {
		if !bytes.Equal(payload.Wire, decoded.Payloads[i].Wire) || !bytes.Equal(payload.Wire, ready.Payloads[i].Payload.Wire) || bytes.IndexAny(payload.Wire, "\r\n") >= 0 {
			t.Fatal("indent/base64 round trip altered compact input bytes")
		}
		line := ready.Payloads[i].Line
		if !bytes.Equal(line, append(slices.Clone(payload.Wire), '\n')) || len(line) > wireLimit || hashBytes(line) != payload.WireSHA256 {
			t.Fatal("line feed was changed/duplicated or not covered by byte cap")
		}
		for j, kind := range baselineKinds() {
			if checkResponse(ready.Payloads[i].Responses[j], payload, kind, payload.Expected[j]) != nil {
				t.Fatal("deterministic protocol-only response preparation failed")
			}
		}
	}
	// A plausible ID change passes shape checks but must fail the independent
	// exact-original projection replay. Shape is not provenance authentication.
	altered := copyCorpusForTest(t, c)
	altered.Origins[0].ParentID = "plausible-replacement"
	if validateCorpus(altered) != nil {
		t.Fatal("shape checker was incorrectly treated as an original ID oracle")
	}
	changed, e := encodePublic(altered, maxCorpusBytes)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := bindCorpus(changed, legacy, typed, Plan{CorpusSHA256: hashBytes(changed)}); e == nil {
		t.Fatal("altered origin identity escaped exact-original binding")
	}
	altered = copyCorpusForTest(t, c)
	altered.Origins[0].OriginalCandidateIDs[0], altered.Origins[0].OriginalCandidateIDs[1] = altered.Origins[0].OriginalCandidateIDs[1], altered.Origins[0].OriginalCandidateIDs[0]
	changed, e = encodePublic(altered, maxCorpusBytes)
	if e != nil || validateCorpus(altered) != nil {
		t.Fatal("plausible metadata order test failed before original replay")
	}
	if _, e := bindCorpus(changed, legacy, typed, Plan{CorpusSHA256: hashBytes(changed)}); e == nil {
		t.Fatal("changed original candidate order escaped exact-original binding")
	}
	ready.Payloads[0].Payload.Wire[0] = 'x'
	ready.Payloads[0].Line[0] = 'y'
	if decoded.Payloads[0].Wire[0] != '{' || c.Payloads[0].Wire[0] != '{' {
		t.Fatal("prepared byte buffers alias a corpus")
	}
}

func TestCorpusBindingRejectsDiscardedExtraArrayElements(t *testing.T) {
	c, legacy, typed := publicCorpusForTest(t)
	canonical, e := encodePublic(c, maxCorpusBytes)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := bindCorpus(canonical, legacy, typed, Plan{CorpusSHA256: hashBytes(canonical)}); e != nil {
		t.Fatal("canonical original projection must bind before mutation")
	}
	// Find the field's matching array end without depending on payload hashes
	// or bracket characters inside quoted JSON strings. Changes are confined
	// to one extra array element; the original fields and order are preserved.
	insertExtra := func(field, extra string) []byte {
		t.Helper()
		marker := []byte(`"` + field + `": [`)
		start := bytes.Index(canonical, marker)
		if start < 0 {
			t.Fatalf("fixed array field %s missing", field)
		}
		start += len(marker) - 1
		depth, inString, escaped, end := 0, false, false, -1
		for i := start; i < len(canonical); i++ {
			ch := canonical[i]
			if inString {
				if escaped {
					escaped = false
				} else if ch == '\\' {
					escaped = true
				} else if ch == '"' {
					inString = false
				}
				continue
			}
			if ch == '"' {
				inString = true
				continue
			}
			if ch == '[' {
				depth++
			}
			if ch == ']' {
				depth--
				if depth == 0 {
					end = i
					break
				}
			}
		}
		if end < 0 {
			t.Fatal("fixed array matching end missing")
		}
		lastValue := end - 1
		for lastValue >= start && (canonical[lastValue] == ' ' || canonical[lastValue] == '\n') {
			lastValue--
		}
		closingIndent := canonical[lastValue+1 : end]
		// Existing canonical arrays end with LF plus indentation. Indent the new
		// element two spaces deeper, including any nested object lines.
		indent := string(closingIndent) + "  "
		item := strings.ReplaceAll(extra, "\n", indent)
		out := append(slices.Clone(canonical[:lastValue+1]), ',')
		out = append(out, indent...)
		out = append(out, item...)
		out = append(out, closingIndent...)
		return append(out, canonical[end:]...)
	}
	for _, sample := range []struct{ field, extra string }{
		{"expected", "{\n  \"baseline\": \"discarded-extra\",\n  \"raw_sha256\": \"\",\n  \"raw_bytes\": 0,\n  \"fallback_reason\": \"\"\n}"},
		{"raw_text_bytes", "9999"},
		{"original_candidate_ids", `"discarded-ninth-id"`},
	} {
		t.Run(sample.field, func(t *testing.T) {
			raw := insertExtra(sample.field, sample.extra)
			var decoded Corpus
			if strictJSON(raw, &decoded) != nil || validateCorpus(decoded) != nil {
				t.Fatal("test must expose a silently discarded element, not another validation fault")
			}
			decodedBytes, e := encodePublic(decoded, maxCorpusBytes)
			if e != nil || !bytes.Equal(decodedBytes, canonical) || bytes.Equal(raw, canonical) {
				t.Fatal("fixed-array decoding did not reproduce the laundering boundary")
			}
			ready, e := bindCorpus(raw, legacy, typed, Plan{CorpusSHA256: hashBytes(raw)})
			if e == nil || e.Error() != "corpus_original_projection_mismatch" || len(ready.Payloads) != 0 {
				t.Fatal("updated hash allowed hidden extra array data to escape raw-byte binding")
			}
		})
	}
}

func TestProjectionStableTextOnlyDedup(t *testing.T) {
	var c Corpus
	baseIDs := []string{"original-a", "original-b", "original-c"}
	base := []string{"keep active rows", "remove active rows", "keep expired rows"}
	if e := appendProjection(&c, "authored-unit", "unit-0", 0, "keep active rows", base, baseIDs); e != nil {
		t.Fatal(e)
	}
	first := slices.Clone(c.Payloads[0].Wire)
	if e := appendProjection(&c, "authored-unit", "unit-1", 1, "keep active rows", base, []string{"other-a", "other-b", "other-c"}); e != nil {
		t.Fatal(e)
	}
	if e := appendProjection(&c, "authored-unit", "unit-2", 2, " keep  active rows ", []string{" keep active rows ", "remove  active rows", "keep expired  rows"}, baseIDs); e != nil {
		t.Fatal(e)
	}
	if len(c.Payloads) != 1 || len(c.Origins) != 3 || c.Origins[1].PayloadIndex != 0 || c.Origins[2].PayloadIndex != 0 || !bytes.Equal(first, c.Payloads[0].Wire) || c.Origins[1].OriginalCandidateIDs[0] != "other-a" {
		t.Fatal("ID/whitespace dedup changed first raw representative or metadata")
	}
	changed := slices.Clone(base)
	changed[1] = "retain ready rows"
	if e := appendProjection(&c, "authored-unit", "unit-3", 3, "keep active rows", changed, baseIDs); e != nil {
		t.Fatal(e)
	}
	reordered := []string{base[1], base[0], base[2]}
	if e := appendProjection(&c, "authored-unit", "unit-4", 4, "keep active rows", reordered, []string{baseIDs[1], baseIDs[0], baseIDs[2]}); e != nil {
		t.Fatal(e)
	}
	if len(c.Payloads) != 3 || c.Origins[3].PayloadIndex != 1 || c.Origins[4].PayloadIndex != 2 || c.Origins[4].OriginalCandidateIDs[0] != baseIDs[1] {
		t.Fatal("actual text/order change was collapsed")
	}
	if e := appendProjection(&c, "authored-unit", "unit-5", 5, "keep active rows", append(slices.Clone(base), "retain rows"), append(slices.Clone(baseIDs), "original-d")); e != nil {
		t.Fatal(e)
	}
	if len(c.Payloads) != 3 || c.Origins[5].CandidateCount != 4 || c.Origins[5].PayloadIndex != -1 || c.Origins[5].OriginalCandidateIDs[3] != "original-d" || c.Origins[5].FeatureSHA256 != "" {
		t.Fatal("non-3 set lost its original size/order")
	}
	beforeOrigins, beforePayloads := len(c.Origins), len(c.Payloads)
	if e := appendProjection(&c, "authored-unit", "unit-6", 6, strings.Repeat("x", 513), base, baseIDs); e == nil || len(c.Origins) != beforeOrigins || len(c.Payloads) != beforePayloads {
		t.Fatal("text rejection left a partial projection")
	}
	if e := appendProjection(&c, "authored-unit", "unit-6", 6, "keep active", base, []string{"id", "id", "other"}); e == nil {
		t.Fatal("duplicate original metadata IDs passed")
	}
	baseIDs[0], base[0] = "caller-mutated", "caller changed text"
	if c.Origins[0].OriginalCandidateIDs[0] != "original-a" || !bytes.Equal(first, c.Payloads[0].Wire) {
		t.Fatal("caller slices alias owned metadata/wire")
	}
}

func TestProjectionCountsLFInsideWireLimit(t *testing.T) {
	ids := candidateIDs()
	input := shortclaim.Input{Schema: textSchema, Request: "a", Provenance: provenance, Candidates: []shortclaim.Candidate{{ID: ids[0], Text: "a"}, {ID: ids[1], Text: "a"}, {ID: ids[2], Text: "a"}}}
	base, e := json.Marshal(input)
	if e != nil {
		t.Fatal(e)
	}
	// A valid 512-byte text may contain JSON-escaped controls. Build a decoded
	// bounded input whose encoded object is exactly wireLimit bytes; accepting
	// it would overflow the same limit when the required LF is appended.
	extra := wireLimit - len(base)
	zeros, plain := extra/6, extra%6
	var fields [4]string
	for i := range fields {
		n := min(zeros, 511)
		fields[i], zeros = strings.Repeat("\x00", n)+"a", zeros-n
	}
	if zeros != 0 {
		t.Fatal("authored boundary input exceeded decoded limits")
	}
	if plain == 0 {
		// Keep the encoded size unchanged but create an ordinary removable byte.
		fields[3] = strings.Replace(fields[3], "\x00", "xxxxxx", 1)
	} else {
		fields[3] = strings.Repeat("x", plain) + fields[3]
	}
	input.Request = fields[0]
	for i := range input.Candidates {
		input.Candidates[i].Text = fields[i+1]
	}
	raw, e := json.Marshal(input)
	if e != nil || len(raw) != wireLimit {
		t.Fatal("authored wire boundary is not exact")
	}
	if _, e := shortclaim.Validate(input); e != nil {
		t.Fatal("authored decoded boundary must remain valid")
	}
	texts := []string{fields[1], fields[2], fields[3]}
	var c Corpus
	if e := appendProjection(&c, "authored-unit", "boundary", 0, fields[0], texts, ids[:]); e == nil || len(c.Payloads) != 0 || len(c.Origins) != 0 {
		t.Fatal("the mandatory LF was not included in the input wire cap")
	}
	texts[2] = strings.Replace(texts[2], "x", "", 1)
	if e := appendProjection(&c, "authored-unit", "boundary", 0, fields[0], texts, ids[:]); e != nil || len(c.Payloads[0].Wire)+1 != wireLimit {
		t.Fatal("an exactly capped framed input was rejected")
	}
}

func TestCorpusRejectsBrokenBindings(t *testing.T) {
	c, _, _ := publicCorpusForTest(t)
	mutations := []struct {
		name   string
		change func(*Corpus)
	}{
		{"schema", func(c *Corpus) { c.Schema = "riido-resident-public-wire-corpus-v1" }},
		{"wire_encoding", func(c *Corpus) { c.WireEncoding = "json_raw_message" }},
		{"feature_encoding", func(c *Corpus) { c.FeatureEncoding = "unversioned" }},
		{"original_count", func(c *Corpus) { c.OriginalParents = 2400 }},
		{"source_hash", func(c *Corpus) { c.TypedInputSHA256 = strings.Repeat("0", 64) }},
		{"too_few_distinct", func(c *Corpus) { c.Payloads = c.Payloads[:1] }},
		{"wire_digest", func(c *Corpus) { c.Payloads[0].WireSHA256 = strings.Repeat("0", 64) }},
		{"input_digest", func(c *Corpus) { c.Payloads[0].InputDigest = strings.Repeat("0", 64) }},
		{"feature_digest", func(c *Corpus) { c.Payloads[0].FeatureSHA256 = strings.Repeat("0", 64) }},
		{"raw_text_size", func(c *Corpus) { c.Payloads[0].RawTextBytes[0]++ }},
		{"normalized_text_size", func(c *Corpus) { c.Payloads[0].NormalizedTextBytes[1]++ }},
		{"normalized_words", func(c *Corpus) { c.Payloads[0].NormalizedWords[2]++ }},
		{"response_bytes", func(c *Corpus) { c.Payloads[0].Expected[0].RawBytes++ }},
		{"response_baseline", func(c *Corpus) { c.Payloads[0].Expected[0].Baseline = "bm25" }},
		{"duplicate_feature", func(c *Corpus) { c.Payloads[1] = c.Payloads[0] }},
		{"origin_order", func(c *Corpus) { c.Origins[0], c.Origins[1] = c.Origins[1], c.Origins[0] }},
		{"origin_duplicate", func(c *Corpus) { c.Origins[1].ParentID = c.Origins[0].ParentID }},
		{"origin_unused_metadata", func(c *Corpus) { c.Origins[0].OriginalCandidateIDs[7] = "unused" }},
		{"origin_metadata_duplicate", func(c *Corpus) { c.Origins[0].OriginalCandidateIDs[1] = c.Origins[0].OriginalCandidateIDs[0] }},
		{"origin_feature_link", func(c *Corpus) { c.Origins[0].FeatureSHA256 = strings.Repeat("0", 64) }},
		{"selection_bounds", func(c *Corpus) { c.Origins[0].PayloadIndex = len(c.Payloads) }},
		{"selection_first_appearance", func(c *Corpus) {
			c.Origins[0].PayloadIndex, c.Origins[0].FeatureSHA256 = 1, c.Payloads[1].FeatureSHA256
		}},
		{"selection_orphan", func(c *Corpus) {
			last := len(c.Payloads) - 1
			for i := range c.Origins {
				if c.Origins[i].PayloadIndex == last {
					c.Origins[i].PayloadIndex, c.Origins[i].FeatureSHA256 = 0, c.Payloads[0].FeatureSHA256
				}
			}
		}},
		{"selected_exclusion", func(c *Corpus) { c.Origins[0].ExcludedReason = "original_candidate_count_not_three" }},
		{"framed_input_limit", func(c *Corpus) { c.Payloads[0].Wire = bytes.Repeat([]byte{' '}, wireLimit) }},
		{"wire_not_compact", func(c *Corpus) {
			c.Payloads[0].Wire = append([]byte{' '}, c.Payloads[0].Wire...)
			c.Payloads[0].WireSHA256 = hashBytes(append(slices.Clone(c.Payloads[0].Wire), '\n'))
		}},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			changed := copyCorpusForTest(t, c)
			mutation.change(&changed)
			if validateCorpus(changed) == nil {
				t.Fatal("broken corpus binding was accepted")
			}
		})
	}
}
