package shortclaim

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
)

// Bind the compiled input policy and linked baseline package without changing
// their frozen v1 source bytes. This API is only for offline audit provenance.
//
//go:embed input.go
var compiledInputSource string

//go:embed baseline.go
var compiledBaselineSource string

//go:embed audit_provenance.go
var compiledProvenanceSource string

type AuditSourceArtifact struct {
	Name, SHA256 string
}

func AuditSourceArtifacts() [3]AuditSourceArtifact {
	hash := func(raw string) string {
		h := sha256.Sum256([]byte(raw))
		return hex.EncodeToString(h[:])
	}
	return [3]AuditSourceArtifact{
		{"input.go", hash(compiledInputSource)},
		{"baseline.go", hash(compiledBaselineSource)},
		{"audit_provenance.go", hash(compiledProvenanceSource)},
	}
}
