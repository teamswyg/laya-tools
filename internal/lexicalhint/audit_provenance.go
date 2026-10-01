package lexicalhint

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
)

// These bytes are used only by offline audit tools, never ranking features.
//
//go:embed features.go
var compiledFeatureSource string

//go:embed audit_provenance.go
var compiledProvenanceSource string

type AuditSourceArtifact struct {
	Name, SHA256 string
}

func AuditSourceArtifacts() [2]AuditSourceArtifact {
	hash := func(raw string) string {
		h := sha256.Sum256([]byte(raw))
		return hex.EncodeToString(h[:])
	}
	return [2]AuditSourceArtifact{
		{"features.go", hash(compiledFeatureSource)},
		{"audit_provenance.go", hash(compiledProvenanceSource)},
	}
}
