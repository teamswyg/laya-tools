package statehint

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func TestFrozenFeatureSchemaAndWorkspaceReuse(t *testing.T) {
	var workspace Workspace
	if err := extract("HELLO", &workspace); err != nil {
		t.Fatal(err)
	}
	// Independently calculated FNV-1a buckets for the frozen schema's word,
	// he/el/ll/lo bigrams, and hel/ell/llo trigrams.
	want := [8]uint16{134, 454, 308, 245, 60, 439, 197, 679}
	if workspace.count != len(want) || workspace.wordCount != 1 {
		t.Fatalf("unexpected feature count %d", workspace.count)
	}
	for i, index := range want {
		if workspace.indices[i] != index || math.Abs(float64(workspace.values[index])-1/math.Sqrt(8)) > 1e-7 {
			t.Fatalf("schema bucket %d drifted", i)
		}
	}
	if err := extract("", &workspace); err != nil {
		t.Fatal(err)
	}
	for _, value := range workspace.values {
		if value != 0 {
			t.Fatal("workspace retained earlier features")
		}
	}
	if err := extract("한글 42", &workspace); err != nil || workspace.wordCount != 2 {
		t.Fatalf("Unicode words lost: %v", err)
	}
	var sum float64
	for _, value := range workspace.values {
		sum += float64(value) * float64(value)
	}
	if math.Abs(sum-1) > 1e-6 {
		t.Fatalf("feature norm = %f", sum)
	}
}

func TestBoundedUTF8Input(t *testing.T) {
	var workspace Workspace
	for _, input := range []string{strings.Repeat("x", MaxTextBytes+1), "\xff", "text\x00text"} {
		if err := extract(input, &workspace); !errors.Is(err, ErrInput) {
			t.Fatalf("invalid input accepted: %v", err)
		}
	}
	if err := extract(strings.Repeat("x", MaxTextBytes), &workspace); err != nil {
		t.Fatal(err)
	}
	if err := extract("text", nil); !errors.Is(err, ErrInput) {
		t.Fatal("nil workspace accepted")
	}
}
