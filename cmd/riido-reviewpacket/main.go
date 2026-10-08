package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/teamswyg/laya-tools/internal/reviewpacket"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("riido-reviewpacket", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var cfg reviewpacket.Config
	flags.StringVar(&cfg.InputPath, "input", "", "regular input file")
	flags.StringVar(&cfg.ExpectedSHA256, "sha256", "", "required independently pinned SHA-256")
	flags.StringVar(&cfg.ReceiptPath, "receipt", "", "new start receipt (terminal sibling adds .result)")
	flags.Int64Var(&cfg.MaxBytes, "max-bytes", reviewpacket.DefaultMaxBytes, "read limit, at most 8388608")
	flags.StringVar(&cfg.Actor, "actor", "", "optional nonempty public actor label, at most 256 bytes")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			flags.SetOutput(stderr)
			flags.PrintDefaults()
			return 0
		}
		fmt.Fprintln(stderr, "review packet failed: invalid command options; use --help")
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "review packet failed: unexpected positional arguments")
		return 2
	}
	if cfg.MaxBytes < 1 || cfg.MaxBytes > reviewpacket.MaximumMaxBytes {
		fmt.Fprintln(stderr, "review packet failed: max-bytes must be between 1 and 8388608")
		return 2
	}
	actorSupplied := false
	flags.Visit(func(f *flag.Flag) { actorSupplied = actorSupplied || f.Name == "actor" })
	if actorSupplied && strings.TrimSpace(cfg.Actor) == "" {
		fmt.Fprintln(stderr, "review packet failed: actor must be nonempty when supplied")
		return 2
	}
	payload, outcome, err := reviewpacket.Capture(cfg)
	if outcome.StartCreated {
		fmt.Fprintf(stderr, "start receipt: %q (durable=%t)\n", cfg.ReceiptPath, outcome.StartDurable)
	}
	if outcome.ResultCreated {
		fmt.Fprintf(stderr, "result receipt: %q (durable=%t)\n", cfg.ReceiptPath+".result", outcome.ResultDurable)
	}
	if err != nil {
		fmt.Fprintln(stderr, "review packet failed:", err)
		return 1
	}
	n, err := stdout.Write(payload)
	if err != nil || n != len(payload) {
		fmt.Fprintln(stderr, "review packet failed: verified payload delivery failed")
		return 1
	}
	fmt.Fprintln(stderr, "review packet verified: payload emitted")
	return 0
}
