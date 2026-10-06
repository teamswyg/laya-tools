package questioncuecli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestOriginalTypedCuePreviewAndPrivateOutput(t *testing.T) {
	input := "{\"text\":\"캐시를 다시 확인할까요?\",\"role\":\"prose\"}\n" + "{\"text\":\"private-original-test-canary?\",\"role\":\"metadata\"}\n" + "{\"text\":\"GET /status?format=plain\",\"role\":\"prose\"}\n"
	var output, stderr bytes.Buffer
	if err := Run([]string{"--jsonl"}, strings.NewReader(input), &output, &stderr); err != nil {
		t.Fatal(err)
	}
	var results []response
	for _, line := range strings.Split(strings.TrimSpace(output.String()), "\n") {
		var result response
		if json.Unmarshal([]byte(line), &result) != nil {
			t.Fatal("invalid JSON output")
		}
		if result.Mode != "cue_preview" || result.MutationExecuted || result.Cue.ModelUsed || result.Cue.ActualVerified || result.Cue.StateChangeProposed || result.Cue.MutationExecuted {
			t.Fatal("cue became a model or action")
		}
		results = append(results, result)
	}
	if len(results) != 3 || !results[0].Cue.QuestionPunctuation || results[1].Cue.QuestionPunctuation || results[2].Cue.QuestionPunctuation {
		t.Fatal("typed content roles or URI guard lost")
	}
	if strings.Contains(output.String(), "canary") || strings.Contains(output.String(), "캐시") || stderr.Len() != 0 {
		t.Fatal("private text echoed")
	}
}

func TestOriginalStrictRequestAndOptionPrivacy(t *testing.T) {
	for _, input := range []string{`{"text":"fixture?","role":"prose","role":"metadata"}`, `{"text":"fixture?","Role":"prose"}`, `{"text":"fixture?","role":null}`, `{"text":"fixture?","role":"prose"} trailing`, `{"text":"fixture?","role":"prose","context":{}}`} {
		var output bytes.Buffer
		if err := Run([]string{"--jsonl"}, strings.NewReader(input), &output, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		var result response
		if json.Unmarshal(output.Bytes(), &result) != nil || result.Status != "unknown" || result.Cue.QuestionPunctuation {
			t.Fatal("invalid typed request admitted")
		}
	}
	var output, stderr bytes.Buffer
	err := Run([]string{"--private-original-options-canary"}, strings.NewReader(""), &output, &stderr)
	if err == nil || strings.Contains(err.Error()+stderr.String()+output.String(), "canary") {
		t.Fatal("invalid option exposed its value")
	}
}
