// SPDX-License-Identifier: Apache-2.0
package claimfit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/laya-tools/internal/roleplan"
)

// Verify immutable, publicly safe preparation history. Reads only these small
// synthetic/provenance archives; no corpus, source inventory or model execution.
func TestFrozenProjectionPreparation64(t *testing.T) {
	const maxArchiveBytes = 128 << 10
	type pin struct {
		path  string
		bytes int
		sha   string
	}
	pins := [...]pin{
		{"prototype-projection-64.go.txt", 12922, PrototypeSourceSHA256},
		{"prototype-projection-tests-64.go.txt", 12123, "c312a1ba1aab690ff613510a572d4604b74cedc4e243b10c7a666d8ed6ef82c8"},
		{"independent-prototype-tests-64.go.txt", 8258, "a1f97396d22434dd0deb012bc58824dc36bedbc2217d56ab6dd24434d980850c"},
		{"PREPARATION-LEDGER-64.json", 4340, "e6420e1c56dd50bb5d50da5a485ef83a40be348dda4942795c753fb0fb5018f1"},
		{"API-LIMITS-64.json", 2508, "f719fe003ff7e8f4294c92481018a723baf73f05d96c1083ffa70b35bb10e14a"},
		{"INDEPENDENT-AUDIT-LEDGER-64.json", 1858, "dc2e98697a1df2defab8af6c53ca0c7463f4b47735b0cede105a99f728b78497"},
		{"INDEPENDENT-FINDINGS-64.ko.md", 5597, "9db8f52145c662c5b3f80a366c8cdbcca44529f122e1fc923dcac2d5b6ef3b40"},
		{"INDEPENDENT-FINDINGS-64.en.md", 5608, "8a3c728e6aff4f3697b05e61a2557c3d3f5d92dc45e060473f96170637186be5"},
	}
	root := filepath.Join("..", "..", "experiments", "short-claim", "claimfit-preparation")
	for _, p := range pins {
		f, err := os.Open(filepath.Join(root, p.path))
		if err != nil {
			t.Fatal(err)
		}
		b, readErr := io.ReadAll(io.LimitReader(f, maxArchiveBytes+1))
		closeErr := f.Close()
		h := sha256.Sum256(b)
		if readErr != nil || closeErr != nil || len(b) > maxArchiveBytes || len(b) != p.bytes || hex.EncodeToString(h[:]) != p.sha {
			t.Fatal("frozen preparation archive changed", p.path)
		}
		if p.path == "API-LIMITS-64.json" {
			var bounds struct {
				Ready     bool `json:"training_ready"`
				Approved  bool `json:"semantic_eligibility_approved"`
				Projected bool `json:"actual_corpus_projection"`
				Labels    int  `json:"new_corpus_truth_labels"`
				Models    int  `json:"models_or_weights_created"`
				Gates     int  `json:"new_readiness_gates"`
			}
			if json.Unmarshal(b, &bounds) != nil || bounds.Ready || bounds.Approved || bounds.Projected || bounds.Labels != 0 || bounds.Models != 0 || bounds.Gates != 0 {
				t.Fatal("prototype limits/readiness changed")
			}
		}
		if p.path == "INDEPENDENT-AUDIT-LEDGER-64.json" {
			var qa struct {
				Ready    bool `json:"training_readiness_approved"`
				Proven   bool `json:"caller_source_binding_or_eligibility_proven"`
				Failures int  `json:"race_test_failures"`
				Original int  `json:"original_corps_read_or_projection"`
				Fit      int  `json:"fit_AUC_scoring_model_weights_calls"`
			}
			if json.Unmarshal(b, &qa) != nil || qa.Ready || qa.Proven || qa.Failures != 0 || qa.Original != 0 || qa.Fit != 0 {
				t.Fatal("historical QA scope changed")
			}
		}
	}
}

func TestPublicRoleEnumBridge(t *testing.T) {
	if uint8(roleplan.DevelopmentTrain) != 0 || uint8(roleplan.DevelopmentValidation) != 1 || uint8(roleplan.DevelopmentCalibration) != 2 || RoleEnumContract != "riido-claimfit-role-bridge-v1:train=0,validation=1,calibration=2" {
		t.Fatal("caller enum bridge changed")
	}
	for i, r := range [3]roleplan.Role{roleplan.DevelopmentTrain, roleplan.DevelopmentValidation, roleplan.DevelopmentCalibration} {
		if n, ok := roleIndex(r); !ok || n != i {
			t.Fatal("role bridge mismatch")
		}
	}
	if _, ok := roleIndex(roleplan.Role(255)); ok {
		t.Fatal("unknown role accepted")
	}
}
