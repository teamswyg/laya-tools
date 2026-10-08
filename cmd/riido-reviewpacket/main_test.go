package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixture = "package fixture\n"
const fixturePin = "c9f09e0a76fbbf83847013f7d79aa09c1566778489cc499e76e3e5d72b1c8db6"

type outputGate struct {
	t    *testing.T
	path string
	bytes.Buffer
}

func (out *outputGate) Write(data []byte) (int, error) {
	for _, path := range []string{out.path, out.path + ".result"} {
		b, err := os.ReadFile(path)
		if err != nil {
			out.t.Fatalf("stdout before receipt completion: %v", err)
		}
		var receipt struct {
			State string `json:"state"`
		}
		if err := json.Unmarshal(b, &receipt); err != nil {
			out.t.Fatal(err)
		}
		if path == out.path+".result" && receipt.State != "verified" {
			out.t.Fatal("stdout before verified terminal receipt")
		}
	}
	return out.Buffer.Write(data)
}

func TestCLIEmitsOnlyVerifiedBytesAfterTerminalReceipt(t *testing.T) {
	dir := t.TempDir()
	input, receipt := filepath.Join(dir, "public.go"), filepath.Join(dir, "read.json")
	if err := os.WriteFile(input, []byte(fixture), 0600); err != nil {
		t.Fatal(err)
	}
	stdout := &outputGate{t: t, path: receipt}
	var stderr bytes.Buffer
	status := run([]string{"--input", input, "--sha256", fixturePin, "--receipt", receipt, "--actor", "public-cli-test"}, stdout, &stderr)
	if status != 0 || stdout.String() != fixture || strings.Contains(stderr.String(), fixture) {
		t.Fatalf("status=%d output=%q stderr=%q", status, stdout.String(), stderr.String())
	}
}

func TestCLIFailuresNeverEmitSource(t *testing.T) {
	for _, kind := range []string{"mismatch", "oversize", "missing", "existing", "actor-empty", "bound-zero", "pin-missing"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			input, receipt := filepath.Join(dir, "public.go"), filepath.Join(dir, "read.json")
			if err := os.WriteFile(input, []byte(fixture), 0600); err != nil {
				t.Fatal(err)
			}
			pin := fixturePin
			if kind == "mismatch" {
				pin = strings.Repeat("0", 64)
			}
			if kind == "pin-missing" {
				pin = ""
			}
			if kind == "missing" {
				input += ".missing"
			}
			if kind == "existing" {
				if err := os.WriteFile(receipt, []byte("public sentinel\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{"--input", input, "--sha256", pin, "--receipt", receipt}
			if kind == "oversize" {
				args = append(args, "--max-bytes", "3")
			}
			if kind == "bound-zero" {
				args = append(args, "--max-bytes", "0")
			}
			if kind == "actor-empty" {
				args = append(args, "--actor", "")
			}
			var stdout, stderr bytes.Buffer
			if status := run(args, &stdout, &stderr); status == 0 || stdout.Len() != 0 || strings.Contains(stderr.String(), fixture) {
				t.Fatalf("failed CLI released source: status=%d stdout=%q stderr=%q", status, stdout.String(), stderr.String())
			}
		})
	}
}

func TestCLIParserErrorsDoNotEchoSuppliedText(t *testing.T) {
	for _, args := range [][]string{{"--max-bytes", "ORIGINAL_PUBLIC_SOURCE_SENTINEL"}, {"--ORIGINAL_PUBLIC_SOURCE_SENTINEL"}} {
		var stdout, stderr bytes.Buffer
		if status := run(args, &stdout, &stderr); status != 2 || stdout.Len() != 0 || strings.Contains(stderr.String(), "ORIGINAL_PUBLIC_SOURCE_SENTINEL") {
			t.Fatalf("unsafe option error: status=%d stdout=%q stderr=%q", status, stdout.String(), stderr.String())
		}
	}
	var stdout, stderr bytes.Buffer
	if status := run([]string{"--help"}, &stdout, &stderr); status != 0 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "sha256") {
		t.Fatal("safe parser handling removed explicit help")
	}
}
