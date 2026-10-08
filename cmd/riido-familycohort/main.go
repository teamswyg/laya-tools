package main

import (
	"encoding/json"
	"flag"
	"io"
	"os"

	"github.com/teamswyg/laya-tools/internal/familycohort"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout)) }

func run(args []string, out io.Writer) int {
	f := flag.NewFlagSet("riido-familycohort", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	var root, dir string
	var input familycohort.File
	f.StringVar(&root, "root", "", "input bundle root directory")
	f.StringVar(&input.Path, "input", "", "root-relative pinned bundle path")
	f.StringVar(&input.SHA256, "sha256", "", "independently supplied bundle SHA-256")
	f.Int64Var(&input.Bytes, "bytes", 0, "independently supplied encoded bundle byte count")
	f.StringVar(&dir, "out", "", "new private output directory")
	invalid := func() int {
		json.NewEncoder(out).Encode(struct {
			Code string `json:"code"`
		}{"arguments_invalid"})
		return 2
	}
	if err := f.Parse(args); err != nil {
		if err == flag.ErrHelp {
			f.SetOutput(out)
			f.PrintDefaults()
			return 0
		}
		return invalid()
	}
	if f.NArg() != 0 || root == "" || dir == "" || input.Path == "" {
		return invalid()
	}
	s, err := familycohort.Run(root, input, dir)
	if json.NewEncoder(out).Encode(s) != nil {
		return 2
	}
	if err != nil {
		return 1
	}
	return 0
}
