package main

import (
	"fmt"
	"os"

	"github.com/teamswyg/laya-tools/internal/statehintpilotcli"
)

func main() {
	if err := statehintpilotcli.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
