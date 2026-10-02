// SPDX-License-Identifier: Apache-2.0
package storedaudit

import (
	_ "embed"
	"path"
)

// Compiled source bytes are audit metadata only. They are never Text or ranking
// features. The source list is closed; callers cannot supply files or functions.
//
//go:embed binding.go
var compiledBindingSource string

//go:embed binding_provenance.go
var compiledBindingProvenanceSource string

type FilePin struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int    `json:"bytes"`
}

// CompiledFiles reports the two exact Go source files embedded in this binary.
// A caller supplies its known repository-relative directory prefix. Each call
// returns fresh records, so changing a returned slice cannot change provenance.
func CompiledFiles(prefix string) []FilePin {
	return []FilePin{
		{Path: path.Join(prefix, "binding.go"), SHA256: sum([]byte(compiledBindingSource)), Bytes: len(compiledBindingSource)},
		{Path: path.Join(prefix, "binding_provenance.go"), SHA256: sum([]byte(compiledBindingProvenanceSource)), Bytes: len(compiledBindingProvenanceSource)},
	}
}
