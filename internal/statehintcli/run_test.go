package statehintcli

import (
	"bytes"
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
