// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"github.com/teamswyg/laya-tools/internal/statehintcli"
	"os"
)

func main() {
	if err := statehintcli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "riido-statehint:", err)
		os.Exit(1)
	}
}
