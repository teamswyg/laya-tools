package filelabels

import (
	"reflect"
	"strings"
	"testing"
)

func TestOldFileLabelsAndBodyIsolation(t *testing.T) {
	patch := "diff --git a/src/a.go b/src/a.go\n--- a/src/a.go\n+++ b/src/a.go\n@@ -1 +1 @@\n--- a/not-a-header.go\n+new\ndiff --git a/new.go b/new.go\nnew file mode 100644\n--- /dev/null\n+++ b/new.go\n@@ -0,0 +1 @@\n+new\ndiff --git a/before.go b/after.go\nsimilarity index 100%\nrename from before.go\nrename to after.go\n"
	got, e := Parse(patch)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(got.OldPaths, []string{"before.go", "src/a.go"}) || got.Blocks != 3 || got.NewFiles != 1 || got.UnsupportedBlocks != 0 {
		t.Fatalf("%+v", got)
	}
}
func TestQuotedSpaceBinaryAndUnsupported(t *testing.T) {
	for _, p := range []string{"src/a file.go", "src/한글.go"} {
		patch := "diff --git a/x b/x\n--- a/" + p + "\n+++ b/" + p + "\n@@ -1 +1 @@\n-a\n+b\n"
		got, e := Parse(patch)
		if e != nil || !reflect.DeepEqual(got.OldPaths, []string{p}) {
			t.Fatalf("%+v %v", got, e)
		}
	}
	got, e := Parse("diff --git \"a/a file\" \"b/a file\"\nBinary files differ\n")
	if e != nil || !reflect.DeepEqual(got.OldPaths, []string{"a file"}) {
		t.Fatalf("%+v %v", got, e)
	}
	got, e = Parse("diff --git a/a file b/a file\nBinary files differ\n")
	if e != nil || got.UnsupportedBlocks != 1 || len(got.OldPaths) != 0 {
		t.Fatalf("guessed ambiguous path: %+v %v", got, e)
	}
	got, e = Parse("diff --git a/x b/x\n--- a/../escape\n+++ b/x\n@@ -1 +1 @@\n-a\n+b\n")
	if e != nil || got.UnsupportedBlocks != 1 {
		t.Fatal("accepted traversal")
	}
}
func TestLabelBoundsAndDuplicateAccounting(t *testing.T) {
	for _, s := range []string{"", strings.Repeat("x", MaxPatchBytes+1), string([]byte{0xff}), "not a diff"} {
		if _, e := Parse(s); e == nil {
			t.Fatal("accepted malformed patch")
		}
	}
	got, e := Parse(strings.Repeat("diff --git a/x b/x\nold mode 100644\nnew mode 100755\n", 2))
	if e != nil || len(got.OldPaths) != 1 || got.DuplicatePaths != 1 {
		t.Fatalf("%+v %v", got, e)
	}
}

func TestBinaryBodyAndQuotedUnicode(t *testing.T) {
	got, err := Parse("diff --git a/image.bin b/image.bin\nGIT binary patch\nliteral 1\nx\n--- a/fake\n")
	if err != nil || !reflect.DeepEqual(got.OldPaths, []string{"image.bin"}) {
		t.Fatalf("%+v %v", got, err)
	}
	got, err = Parse("diff --git a/x b/x\n--- \"a/\\303\\251.go\"\n+++ \"b/\\303\\251.go\"\n@@ -1 +1 @@\n-a\n+b\n")
	if err != nil || !reflect.DeepEqual(got.OldPaths, []string{"é.go"}) {
		t.Fatalf("%+v %v", got, err)
	}
	got, err = Parse("diff --git a/x b/x\n--- a/x\n+++ /invalid\n@@ -1 +1 @@\n-a\n+b\n")
	if err != nil || got.UnsupportedBlocks != 1 || len(got.OldPaths) != 0 {
		t.Fatal("accepted invalid new path")
	}
}
