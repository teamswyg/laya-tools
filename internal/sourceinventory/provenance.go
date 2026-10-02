// SPDX-License-Identifier: Apache-2.0
// Compiled source-byte provenance, separate from semantic source closure.
package sourceinventory

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
)

// Embedding proves which package-source bytes were supplied to this build.
// It does not discover behavior closures or bind runtime source objects.
//
//go:embed inventory.go
var compiledInventory []byte

//go:embed registry.go
var compiledRegistry []byte

//go:embed provenance.go
var compiledProvenance []byte

type SourceArtifact struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int    `json:"bytes"`
}

// SourceArtifacts returns owned descriptors of the compiled package sources.
// Return fresh descriptors; never expose the mutable embedded byte slices.
func SourceArtifacts() []SourceArtifact {
	paths := [3]string{"internal/sourceinventory/inventory.go", "internal/sourceinventory/registry.go", "internal/sourceinventory/provenance.go"}
	bytes := [3][]byte{compiledInventory, compiledRegistry, compiledProvenance}
	out := make([]SourceArtifact, len(paths))
	for i, p := range paths {
		h := sha256.Sum256(bytes[i])
		out[i] = SourceArtifact{Path: p, SHA256: hex.EncodeToString(h[:]), Bytes: len(bytes[i])}
	}
	return out
}
