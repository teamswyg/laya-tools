package taskverify

import (
	"bytes"
	"encoding/json"
	"slices"
)

// The owned Go process already exited successfully. Still require every pinned
// package and independently supplied test to have actually started and passed;
// a matching text fragment, a skipped test or a package-only pass cannot suffice.
func versionedTerminalPasses(d Definition, output []byte) testOutcome {
	r := testOutcome{isolated: true, code: "independent_tests_failed"}
	type packageState struct{ started, passed bool }
	type testState struct{ ran, passed bool }
	packages := make([]packageState, len(d.Packages))
	tests := make([]testState, len(d.RequiredTests))
	for _, line := range bytes.Split(output, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		var event struct{ Action, Package, Test string }
		if json.Unmarshal(line, &event) != nil {
			r.unknown, r.code = true, "invalid_test_output"
			return r
		}
		pi := slices.Index(d.Packages, event.Package)
		if pi < 0 {
			r.unknown, r.code = true, "unexpected_test_package"
			return r
		}
		if event.Test == "" {
			switch event.Action {
			case "start":
				if packages[pi].started || packages[pi].passed {
					r.unknown, r.code = true, "invalid_test_output"
					return r
				}
				packages[pi].started = true
			case "pass":
				if !packages[pi].started || packages[pi].passed {
					r.unknown, r.code = true, "invalid_test_output"
					return r
				}
				packages[pi].passed = true
			case "fail", "skip":
				return r
			}
			continue
		}
		if !packages[pi].started || packages[pi].passed {
			r.unknown, r.code = true, "invalid_test_output"
			return r
		}
		if event.Action == "pass" {
			r.tests++
		}
		for ti, required := range d.RequiredTests {
			if event.Package != required.Package || event.Test != required.Test {
				continue
			}
			switch event.Action {
			case "run":
				if tests[ti].ran || tests[ti].passed {
					r.unknown, r.code = true, "invalid_test_output"
					return r
				}
				tests[ti].ran = true
			case "pass":
				if !tests[ti].ran || tests[ti].passed {
					r.unknown, r.code = true, "invalid_test_output"
					return r
				}
				tests[ti].passed = true
			case "fail", "skip":
				return r
			}
		}
	}
	for _, p := range packages {
		if !p.started || !p.passed {
			return r
		}
	}
	for _, t := range tests {
		if !t.ran || !t.passed {
			return r
		}
	}
	r.passed, r.code = true, "independent_pinned_tests_passed"
	return r
}
