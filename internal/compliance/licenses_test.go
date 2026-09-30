package compliance

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// This gate catches dependency/version drift and lost notices. It does not
// determine legal compatibility or prove upstream ownership/training rights.
func TestDistributedLicenseInventory(t *testing.T) {
	root := filepath.Join("..", "..")
	b, err := os.ReadFile(filepath.Join(root, "licenses", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var inventory struct {
		Modules map[string]string `json:"modules"`
		Files   map[string]string `json:"files"`
	}
	if err = json.Unmarshal(b, &inventory); err != nil {
		t.Fatal(err)
	}
	for name, want := range inventory.Files {
		b, err := os.ReadFile(filepath.Join(root, "licenses", name))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != want {
			t.Fatalf("license changed: %s; review source and update inventory", name)
		}
	}
	for _, name := range []string{"LICENSE", "PATENTS"} {
		b, err := os.ReadFile(filepath.Join(runtime.GOROOT(), name))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != inventory.Files["go."+name] {
			t.Fatalf("Go %s changed; update distributed notice after review", name)
		}
	}
	cmd := exec.Command("go", "list", "-deps", "-json", "./cmd/riidolaya")
	cmd.Dir = root
	b, err = cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	d := json.NewDecoder(bytes.NewReader(b))
	actual := map[string]string{}
	for {
		var p struct {
			Module *struct {
				Path, Version string
				Main          bool
			}
		}
		err := d.Decode(&p)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if p.Module != nil && !p.Module.Main {
			actual[p.Module.Path] = p.Module.Version
		}
	}
	if !reflect.DeepEqual(actual, inventory.Modules) {
		t.Fatalf("runtime dependency inventory changed: got %v, audited %v; review licenses", actual, inventory.Modules)
	}
}

func TestModelNoticesRequired(t *testing.T) {
	b, err := os.ReadFile("../assets/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var artifacts map[string]struct {
		Files map[string]string `json:"files"`
	}
	if err = json.Unmarshal(b, &artifacts); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"base", "code"} {
		for _, name := range []string{"LICENSE", "NOTICE", "MODIFICATIONS.md", "PROVENANCE.json", "MODEL_CARD.md"} {
			if len(artifacts[kind].Files[name]) != 64 {
				t.Fatalf("%s lacks pinned %s", kind, name)
			}
		}
	}
}
