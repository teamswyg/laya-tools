package main

import (
	"strings"
	"testing"
)

func TestRequestBoundary(t *testing.T) {
	for _, s := range []string{`{}`, `{"text":null}`, `{"text":"x","extra":1}`, `{"text":"x"}{}`, `{"text":"x"}` + strings.Repeat(" ", 65536) + `{}`, "{\"text\":\"\xff\"}"} {
		if _, e := readText(strings.NewReader(s)); e == nil {
			t.Fatal("accepted malformed or oversized request")
		}
	}
	got, e := readText(strings.NewReader(`{"text":""}`))
	if e != nil || got != "" {
		t.Fatal("empty text must remain supported", e)
	}
}
