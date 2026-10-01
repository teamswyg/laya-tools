package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/finiteproperty"
)

func TestStrictArtifactRejectsSilentJSONChanges(t *testing.T) {
	d, err := finiteproperty.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := finiteproperty.JSON(d)
	if err != nil {
		t.Fatal(err)
	}
	var decoded finiteproperty.Dataset
	if err := decode(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	var generic map[string]json.RawMessage
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatal(err)
	}
	var parents []json.RawMessage
	if err := json.Unmarshal(generic["parents"], &parents); err != nil {
		t.Fatal(err)
	}
	parents = append(parents, parents[0])
	generic["parents"], _ = json.Marshal(parents)
	extra, _ := finiteproperty.JSON(generic)
	for _, bad := range [][]byte{
		extra,
		bytes.Replace(raw, []byte(`"schema":`), []byte(`"schema": "wrong", "schema":`), 1),
		bytes.Replace(raw, []byte(`"schema":`), []byte(`"unknown": 1, "schema":`), 1),
		append(append([]byte{}, raw...), []byte("{}")...),
	} {
		if err := decode(bad, &decoded); err == nil {
			t.Fatal("noncanonical/malleable JSON accepted")
		}
	}
}

func TestCompiledManifestAndPackageDirectoryGuard(t *testing.T) {
	entries, err := sources()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 21 {
		t.Fatal("compiled closure count")
	}
	root, err := os.OpenRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := verifyFiles(root, entries); err != nil {
		t.Fatal(err)
	}
	if err := verifyDirectories(root); err != nil {
		t.Fatal(err)
	}
	changed := append([]artifact{}, entries...)
	changed[0].SHA256 = strings.Repeat("0", 64)
	if err := verifyFiles(root, changed); err == nil {
		t.Fatal("stale compiled source accepted")
	}
}

func TestGitCommitRequiresVerifiedBlobsAndBoundedOutput(t *testing.T) {
	if err := verifyGit("invalid", nil); err == nil {
		t.Fatal("invalid commit accepted")
	}
	var b boundedBuffer
	if _, err := b.Write(make([]byte, maxBytes)); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Write([]byte{1}); err == nil {
		t.Fatal("unbounded git output")
	}
}
