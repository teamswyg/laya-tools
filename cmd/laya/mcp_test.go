package main

import (
	"bytes"
	"encoding/json"
	"github.com/teamswyg/laya-tools/internal/app"
	"strings"
	"testing"
)

func TestMCPProtocol(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize"}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"missing","arguments":{}}}
`
	var out bytes.Buffer
	if err := serveMCP(&app.App{}, strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 3 {
		t.Fatal(out.String())
	}
	for _, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatal(err)
		}
		if m["jsonrpc"] != "2.0" {
			t.Fatal(m)
		}
	}
}
