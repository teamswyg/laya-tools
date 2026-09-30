// riido-taskoutcome reads existing local telemetry; it never launches Codex.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/teamswyg/laya-tools/pkg/taskoutcome"
)

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	f := flag.NewFlagSet("riido-taskoutcome", flag.ContinueOnError)
	// Flag errors can contain supplied paths or labels. Replace them with a code.
	f.SetOutput(io.Discard)
	input := f.String("input", "-", "JSONL file or - for stdin")
	task := f.String("task-label", "", "public authored task identifier (optional)")
	model := f.String("requested-model", "", "requested model identifier; observed model stays unknown")
	reasoning := f.String("requested-reasoning", "", "requested reasoning identifier (optional)")
	const help = `riido-taskoutcome — summarize existing private Codex JSONL

Usage: riido-taskoutcome --input trace.jsonl
       riido-taskoutcome --input - < trace.jsonl

Options: --task-label ID --requested-model ID --requested-reasoning ID
Labels must be short safe identifiers. They are caller-provided requests,
not observed model evidence. Output is aggregate JSON; raw content is discarded.
Missing usage, process exit, observed model and task acceptance stay unknown.
No Codex execution, task approval, token-to-fee conversion or upload occurs.
Limits: 64MiB trace, 1MiB line, 100000 events, 4096 turns.
`
	f.Usage = func() {} // Parse errors must not write non-JSON help to stdout.
	if err := f.Parse(args); err != nil {
		if err == flag.ErrHelp {
			fmt.Fprint(stdout, help)
			return 0
		}
		fmt.Fprintln(stderr, "invalid_arguments")
		return 2
	}
	if f.NArg() != 0 || *input == "" {
		fmt.Fprintln(stderr, "invalid_arguments")
		return 2
	}
	reader := stdin
	if *input != "-" {
		info, err := os.Lstat(*input)
		if err != nil {
			fmt.Fprintln(stderr, "input_open_failed")
			return 1
		}
		if !info.Mode().IsRegular() {
			fmt.Fprintln(stderr, "input_not_regular")
			return 1
		}
		file, err := os.Open(*input)
		if err != nil {
			fmt.Fprintln(stderr, "input_open_failed")
			return 1
		}
		defer file.Close()
		opened, err := file.Stat()
		if err != nil || !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
			fmt.Fprintln(stderr, "input_not_regular")
			return 1
		}
		reader = file
	}
	summary, err := taskoutcome.Summarize(reader, taskoutcome.Metadata{PublicTaskLabel: *task, RequestedModel: *model, RequestedReasoning: *reasoning})
	if err != nil {
		// Summarize exposes only its fixed ErrorCode values.
		if code, ok := err.(taskoutcome.ErrorCode); ok {
			fmt.Fprintln(stderr, code)
		} else {
			fmt.Fprintln(stderr, "summary_failed")
		}
		return 1
	}
	if json.NewEncoder(stdout).Encode(summary) != nil {
		fmt.Fprintln(stderr, "output_write_failed")
		return 1
	}
	return 0
}

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
