// SPDX-License-Identifier: Apache-2.0
package roleplan

import (
	"crypto/sha256"
	_ "embed"
)

// Embed the same immutable algorithm source file used by the Go build. This is
// maintainer provenance, not scorer input or an independent compiler proof.
// Only optional role/fit preparation tools import this package.
//
//go:embed role.go
var compiledRoleSource string

// CompiledSourceSHA256 lets a frozen driver compare the role source included in
// its build with the pinned on-disk source. A trusted build recipe, the complete
// source closure and executable digest still need separate caller verification.
func CompiledSourceSHA256() [32]byte {
	return sha256.Sum256([]byte(compiledRoleSource))
}
