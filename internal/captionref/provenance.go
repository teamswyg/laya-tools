// SPDX-License-Identifier: Apache-2.0
package captionref

import (
	_ "embed"
	"path"
)

//go:embed references.go
var compiledReferences []byte

//go:embed provenance.go
var compiledProvenance []byte

// CompiledFiles returns fresh metadata records. Embedded source is never used
// as a scoring feature, a literal oracle, or a candidate implementation.
func CompiledFiles(prefix string) []Artifact {
	return []Artifact{
		{Path: path.Join(prefix, "references.go"), SHA256: digest(compiledReferences), Bytes: len(compiledReferences)},
		{Path: path.Join(prefix, "provenance.go"), SHA256: digest(compiledProvenance), Bytes: len(compiledProvenance)},
	}
}
