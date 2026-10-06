// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/teamswyg/laya-tools/internal/statehintfit"
)

func main() {
	runtime.GOMAXPROCS(2)
	if err := statehintfit.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "statehint study failed; inspect local preparation and pins")
		os.Exit(1)
	}
}
