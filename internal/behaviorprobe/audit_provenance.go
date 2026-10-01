package behaviorprobe

import _ "embed"

// Expose compiled provenance before running legacy candidate controls. The
// original v1 implementation and its embedded source remain unchanged.
//
//go:embed audit_provenance.go
var compiledProvenanceSource string

type AuditSourceArtifact struct {
	Name, SHA256 string
}

func AuditSourceArtifacts() [2]AuditSourceArtifact {
	return [2]AuditSourceArtifact{
		{"data.go", digest([]byte(authoredSource))},
		{"audit_provenance.go", digest([]byte(compiledProvenanceSource))},
	}
}
