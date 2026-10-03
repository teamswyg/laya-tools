package main

import (
	"encoding/json"
	"os"
)

// Future Root-only entry point. No invocation occurred at this source seal.
func main() {
	if len(os.Args) != 5 || os.Args[1] != "--config" || os.Args[3] != "--config-sha256" {
		os.Stderr.WriteString("archive_config_args\n")
		os.Exit(2)
	}
	result, err := run(os.Args[2], os.Args[4])
	// Captured stdout is a control artifact; reserve 2 KiB before execution.
	b, marshalErr := json.Marshal(result)
	if marshalErr != nil || len(b)+1 > 2048 {
		os.Stderr.WriteString("archive_result_encoding\n")
		os.Exit(1)
	}
	if _, writeErr := os.Stdout.Write(append(b, '\n')); writeErr != nil {
		os.Exit(1)
	}
	if err != nil {
		os.Exit(1)
	}
}
