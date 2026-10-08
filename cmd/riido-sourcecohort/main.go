package main

import (
	"encoding/json"
	"flag"
	"github.com/teamswyg/laya-tools/internal/sourcecohort"
	"io"
	"os"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout)) }
func run(args []string, out io.Writer) int {
	f := flag.NewFlagSet("riido-sourcecohort", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	var root, dir string
	var plan sourcecohort.File
	var allocation, defaults bool
	f.StringVar(&root, "root", "", "input cohort directory")
	f.StringVar(&plan.Path, "plan", "", "relative pinned plan path")
	f.StringVar(&plan.SHA256, "sha256", "", "independent plan SHA256")
	f.Int64Var(&plan.Bytes, "bytes", 0, "independent plan byte count")
	f.StringVar(&dir, "out", "", "new private output directory")
	f.BoolVar(&allocation, "allocation-only", false, "verify slots before material creation")
	f.BoolVar(&defaults, "default-recipe", false, "print whole400 allocation recipe")
	if e := f.Parse(args); e != nil {
		if e == flag.ErrHelp {
			f.SetOutput(out)
			f.PrintDefaults()
			return 0
		}
		json.NewEncoder(out).Encode(struct {
			Code string `json:"code"`
		}{"arguments_invalid"})
		return 2
	}
	if f.NArg() != 0 {
		json.NewEncoder(out).Encode(struct {
			Code string `json:"code"`
		}{"arguments_invalid"})
		return 2
	}
	if defaults {
		if json.NewEncoder(out).Encode(sourcecohort.DefaultRecipe()) != nil {
			return 2
		}
		return 0
	}
	s, e := sourcecohort.Run(root, plan, dir, allocation)
	if json.NewEncoder(out).Encode(s) != nil {
		return 2
	}
	if e != nil {
		return 1
	}
	return 0
}
