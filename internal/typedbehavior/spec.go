// Package typedbehavior audits authored, finite Go behavioral contracts.
// Its truth and provenance are offline development evidence, never features.
package typedbehavior

import _ "embed"

//go:embed spec.go
var sourceSpecText string

// The closed source registry binds IDs to compiled implementations. Runtime
// input cannot supply Go code, callbacks, truth tables or expected labels.
type sourceSpec struct {
	id, prototype, core, function string
	correct                       bool
}

// Candidate checkers return actual vector observations. A safely captured
// candidate panic violates these no-panic contracts; an observer failure is
// unknown. Every correct/wrong registry control must agree with literal truth.
type controlCheck struct {
	Checked int
	Failed  int
	Unknown bool
}

// ContractSpec describes the independently authored literal observation table.
// Scope remains finite even when its prose states the intended general rule.
type ContractSpec struct {
	Prototype        string `json:"prototype"`
	Core             string `json:"core_template"`
	InputSemantics   string `json:"input_semantics"`
	Vectors          int    `json:"literal_vectors"`
	TruthTableSHA256 string `json:"truth_table_sha256"`
}
