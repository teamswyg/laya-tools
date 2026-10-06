package statehintcli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestTypedRequestBoundary(t *testing.T) {
	for _, s := range []string{`null`, `{"text":null}`, `{"text":"one","text":"two"}`, `{"text":"one","Text":"two"}`, `{"text":"one","context":null}`, `{"text":"one","Context":null}`, `{"text":"one"} {"text":"two"}`, `{"text":"one","context":{"work_id":"a","work_id":"b"}}`} {
		if validRequest([]byte(s)) {
			t.Errorf("accepted ambiguous request %s", s)
		}
	}
	if !validRequest([]byte(`{"text":"질문이 있습니다"}`)) {
		t.Fatal("valid original request rejected")
	}
}

func TestContextEmojiIsAuthoritative(t *testing.T) {
	for _, current := range []string{"", "1f4ac"} {
		payload := `{"text":"질문이 있습니다","context":{"work_id":"demo-work","command_id":"demo-command","expected_version":1,"current_state":"todo","emojis":[{"intent":"question","code":"1f4ac"}],"current_emoji_code":"` + current + `"}}` + "\n"
		var out, errout bytes.Buffer
		if e := Run([]string{"--jsonl"}, strings.NewReader(payload), &out, &errout); e != nil {
			t.Fatal(e)
		}
		var r response
		if e := json.Unmarshal(out.Bytes(), &r); e != nil {
			t.Fatal(e)
		}
		if r.Plan == nil || r.EmojiCode != r.Plan.EmojiCode {
			t.Fatal("default display bypassed supplied emoji plan")
		}
		if current == "" && r.EmojiCode != "1f4ac" {
			t.Fatal("configured icon was not used")
		}
		if current != "" && r.EmojiCode != "" {
			t.Fatal("existing icon was proposed again")
		}
	}
}
func TestJSONLUnknownDoesNotEchoText(t *testing.T) {
	var out, errout bytes.Buffer
	in := strings.NewReader("null\n{\"text\":null}\n{\"text\":\"질문이 있습니다\"}\n")
	if err := Run([]string{"--jsonl"}, in, &out, &errout); err != nil {
		t.Fatal(err)
	}
	if strings.Count(out.String(), `"status":"unknown"`) != 2 || strings.Count(out.String(), `"status":"suggested"`) != 1 || strings.Contains(out.String(), "질문이 있습니다") {
		t.Fatalf("unexpected request isolation output %s", out.String())
	}
}
