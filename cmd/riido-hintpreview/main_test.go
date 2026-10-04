// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestPreviewBoundedProtocolAndNoSupervision(t *testing.T) {
	valid := `{"request":"Return all keys in increasing order.","candidates":["ordinary unsupported text"]}`
	var b bytes.Buffer
	if err := run(strings.NewReader(valid), &b); err != nil {
		t.Fatal(err)
	}
	var got output
	if err := json.Unmarshal(b.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if !got.PreviewOnly || got.ModelCalls != 0 || got.FitCalls != 0 || len(got.Candidates) != 1 || got.Scores[0] != 0 || got.Order[0] != 0 {
		t.Fatalf("unexpected preview: %+v", got)
	}
	for _, invalid := range []string{
		`{"request":"q","candidates":[]}`,
		`{"candidates":["d"]}`,
		`{"request":null,"candidates":["d"]}`,
		`{"request":"q","candidates":[null]}`,
		`{"request":"q","request":"other","candidates":["d"]}`,
		`{"request":"q","candidates":["d"],"candidates":["other"]}`,
		`{"request":"q","candidates":["\ud800"]}`,
		`{"request":"q","candidates":["\udc00"]}`,
		`{"request":"\ud800\ud800","candidates":["d"]}`,
		`{"request":"` + string([]byte{0xff}) + `","candidates":["d"]}`,
		`{"request":"q","candidates":["d"],"label":true}`,
		`{"request":"q","candidates":["d"],"role":"development_train"}`,
		valid + valid,
		strings.Repeat(" ", maxJSONBytes+1),
		`{"request":"` + strings.Repeat("x", 513) + `","candidates":["d"]}`,
		`{"request":"` + strings.Repeat("a ", 33) + `","candidates":["d"]}`,
		`{"request":"q","candidates":["a","b","c","d","e","f","g","h","i"]}`,
	} {
		b.Reset()
		if err := run(strings.NewReader(invalid), &b); !errors.Is(err, errInput) || b.Len() != 0 {
			t.Fatalf("expected clean protocol rejection: err=%v output=%q", err, b.String())
		}
	}
}

func TestScalarEscapeAndFieldOrder(t *testing.T) {
	for _, valid := range []string{
		`{"candidates":["\ud83d\ude42"],"request":"한글"}`,
		`{"request":"\"quoted\" \\ud800","candidates":["\u00e9"]}`,
		`{"request":"","candidates":[""]}`,
	} {
		var out bytes.Buffer
		if err := run(strings.NewReader(valid), &out); err != nil || out.Len() == 0 {
			t.Fatalf("explicit valid scalar text rejected: %v", err)
		}
	}
}

func TestSymbolFingerprintSeparatesOperatorLoss(t *testing.T) {
	preview := func(q string) output {
		t.Helper()
		v := input{q, []string{"limit value"}}
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		if err := run(bytes.NewReader(b), &out); err != nil {
			t.Fatal(err)
		}
		var result output
		if err := json.Unmarshal(out.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	a, b := preview("value < limit"), preview("value <= limit")
	if a.Candidates[0].Legacy.SHA256 != b.Candidates[0].Legacy.SHA256 || a.Candidates[0].Symbol.SHA256 == b.Candidates[0].Symbol.SHA256 {
		t.Fatal("operator information-loss control failed")
	}
	if a.Request.Coverage != 0 || b.Request.Coverage != 0 {
		t.Fatal("unscoped prose must remain unsupported")
	}
}
