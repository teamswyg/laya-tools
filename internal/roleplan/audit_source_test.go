// SPDX-License-Identifier: Apache-2.0
package roleplan

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"
)

func TestCompiledRoleSourceIdentity(t *testing.T) {
	b, err := os.ReadFile("role.go")
	if err != nil {
		t.Fatal(err)
	}
	got, disk := CompiledSourceSHA256(), sha256.Sum256(b)
	if got != disk || len(compiledRoleSource) != len(b) {
		t.Fatal("compiled role source differs from build input")
	}
	// Deliberately mutate only the owned disk copy to distinguish this source
	// identity check from a fixed constant that ignores actual embedded bytes.
	if len(b) == 0 {
		t.Fatal("empty role source")
	}
	b[0] ^= 1
	if sha256.Sum256(b) == got {
		t.Fatal("source mutation went undetected")
	}
	if hex.EncodeToString(got[:]) != "df8d53d0600e80c331c4912d90ff199aef753704f34c87e085fa738fe920ba0b" {
		t.Fatal("frozen role algorithm changed; update a separate versioned plan")
	}
}
