package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyRejectsTampering(t *testing.T) {
	d := t.TempDir()
	b := []byte("public fixture")
	sum := sha256.Sum256(b)
	files := map[string]string{"model.onnx": hex.EncodeToString(sum[:])}
	if err := os.WriteFile(filepath.Join(d, "model.onnx"), b, 0600); err != nil {
		t.Fatal(err)
	}
	if err := Verify(d, files); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "model.onnx"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if Verify(d, files) == nil {
		t.Fatal("tampering accepted")
	}
	if Verify(d, map[string]string{}) == nil {
		t.Fatal("empty manifest accepted")
	}
}
