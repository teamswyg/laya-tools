package filelabels

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

func devMembers(ids ...string) []sweaudit.TaskRole {
	var out []sweaudit.TaskRole
	for _, id := range ids {
		r := sweaudit.TaskRole{Role: "train"}
		r.Task.ID = id
		out = append(out, r)
	}
	return out
}
func devProjection(t *testing.T, rows ...[2]string) (string, string) {
	t.Helper()
	var b []byte
	for _, r := range rows {
		row := struct {
			ID         string  `json:"instance_id"`
			Patch      *string `json:"patch"`
			PatchBytes int     `json:"patch_bytes"`
			Oversized  bool    `json:"oversized"`
		}{ID: r[0], Patch: &r[1], PatchBytes: len(r[1]), Oversized: len(r[1]) > MaxPatchBytes}
		if row.Oversized {
			row.Patch = nil
		}
		v, e := json.Marshal(row)
		if e != nil {
			t.Fatal(e)
		}
		b = append(b, v...)
		b = append(b, '\n')
	}
	p := filepath.Join(t.TempDir(), "labels.jsonl")
	if e := os.WriteFile(p, b, 0600); e != nil {
		t.Fatal(e)
	}
	h := sha256.Sum256(b)
	return p, hex.EncodeToString(h[:])
}

const devPatch = "diff --git a/a.go b/a.go\n--- a/a.go\n+++ b/a.go\n@@ -1 +1 @@\n-old\n+new\n"

func TestDevelopmentReaderIsolation(t *testing.T) {
	members := devMembers("b", "a")
	before := append([]sweaudit.TaskRole(nil), members...)
	p, h := devProjection(t, [2]string{"a", devPatch}, [2]string{"b", strings.ReplaceAll(devPatch, "a.go", "b.go")})
	got, e := ReadDevelopment(p, h, members)
	if e != nil {
		t.Fatal(e)
	}
	if len(got) != 2 || got[0].ID != "a" || !reflect.DeepEqual(got[0].Result.OldPaths, []string{"a.go"}) || got[1].Result.OldPaths[0] != "b.go" {
		t.Fatalf("unexpected labels: %+v", got)
	}
	if !reflect.DeepEqual(members, before) {
		t.Fatal("mutated members")
	}
	for _, tc := range []struct {
		name    string
		rows    [][2]string
		members []sweaudit.TaskRole
	}{
		{"unknown", [][2]string{{"final-task", devPatch}}, devMembers("a")},
		{"duplicate", [][2]string{{"a", devPatch}, {"a", devPatch}}, devMembers("a", "b")},
		{"missing", [][2]string{{"a", devPatch}}, devMembers("a", "b")},
		{"duplicate member", [][2]string{{"a", devPatch}}, devMembers("a", "a")},
		{"final member", [][2]string{{"a", devPatch}}, func() []sweaudit.TaskRole { r := devMembers("a"); r[0].Role = "final"; return r }()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, h := devProjection(t, tc.rows...)
			if _, e := ReadDevelopment(p, h, tc.members); e == nil {
				t.Fatal("accepted invalid membership/labels")
			}
		})
	}
	if _, e := ReadDevelopment(p, strings.Repeat("0", 64), members); e == nil {
		t.Fatal("accepted wrong hash")
	}
}
func TestDevelopmentBoundsAndFailures(t *testing.T) {
	p, h := devProjection(t, [2]string{"a", devPatch})
	stat, _ := os.Stat(p)
	if _, e := readDevelopment(p, h, devMembers("a"), stat.Size(), stat.Size()); e != nil {
		t.Fatal(e)
	}
	if _, e := readDevelopment(p, h, devMembers("a"), stat.Size()-1, stat.Size()); e == nil {
		t.Fatal("byte bound")
	}
	if _, e := readDevelopment(p, h, devMembers("a"), stat.Size(), stat.Size()-1); e == nil {
		t.Fatal("line bound")
	}
	p, h = devProjection(t, [2]string{"a", strings.Repeat("x", MaxPatchBytes+1)}, [2]string{"b", "not a patch"})
	got, e := ReadDevelopment(p, h, devMembers("a", "b"))
	if e != nil {
		t.Fatal(e)
	}
	if len(got) != 2 || !got[0].ParseError || !got[1].ParseError {
		t.Fatal("parse errors lost from denominator")
	}
}
func TestDevelopmentSchema(t *testing.T) {
	for _, body := range []string{"{}\n", "{\"instance_id\":\"a\"}\n", "{\"instance_id\":\"a\",\"patch\":null}\n", "{\"instance_id\":\"a\",\"patch\":\"\",\"extra\":1}\n", "{\"instance_id\":\"a\",\"patch\":\"\"} {}\n", "not json\n"} {
		p := filepath.Join(t.TempDir(), "labels")
		os.WriteFile(p, []byte(body), 0600)
		h := sha256.Sum256([]byte(body))
		if _, e := ReadDevelopment(p, hex.EncodeToString(h[:]), devMembers("a")); e == nil {
			t.Fatal("invalid schema accepted")
		}
	}
}

func TestDevelopmentParserOwnsPaths(t *testing.T) {
	raw := devPatch + strings.Repeat("+padding\n", 8192)
	result, e := Parse(raw)
	if e != nil || len(result.OldPaths) != 1 {
		t.Fatal("fixture parse", e)
	}
	start := uintptr(unsafe.Pointer(unsafe.StringData(raw)))
	end := start + uintptr(len(raw))
	for _, path := range result.OldPaths {
		ptr := uintptr(unsafe.Pointer(unsafe.StringData(path)))
		if ptr >= start && ptr < end {
			t.Fatal("retained raw patch backing storage")
		}
	}
}

func TestDevelopmentOversizeMetadataConsistency(t *testing.T) {
	for _, body := range []string{
		`{"instance_id":"a","patch":null,"patch_bytes":1,"oversized":true}`,
		`{"instance_id":"a","patch":"text","patch_bytes":2097153,"oversized":true}`,
		`{"instance_id":"a","patch":null,"patch_bytes":2097153,"oversized":false}`,
		`{"instance_id":"a","patch":"text","patch_bytes":3,"oversized":false}`,
		`{"instance_id":"a","patch":"","patch_bytes":-1,"oversized":false}`,
	} {
		p := filepath.Join(t.TempDir(), "labels")
		raw := []byte(body + "\n")
		if e := os.WriteFile(p, raw, 0600); e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(raw)
		if _, e := ReadDevelopment(p, hex.EncodeToString(h[:]), devMembers("a")); e == nil {
			t.Fatal("inconsistent metadata accepted")
		}
	}
	p, h := devProjection(t, [2]string{"a", strings.Repeat("x", MaxPatchBytes+1)}, [2]string{"b", "preamble\n" + devPatch})
	rows, e := ReadDevelopment(p, h, devMembers("a", "b"))
	if e != nil {
		t.Fatal(e)
	}
	if rows[0].FailureKind != "oversized" || !rows[0].Oversized || rows[1].FailureKind != "non_git_preamble" {
		t.Fatal("failure categories lost")
	}
}

func TestBlankTargetRemainsFailure(t *testing.T) {
	p, h := devProjection(t, [2]string{"a", " \n\t"})
	rows, e := ReadDevelopment(p, h, devMembers("a"))
	if e != nil {
		t.Fatal(e)
	}
	if !rows[0].ParseError || rows[0].FailureKind != "blank_patch" {
		t.Fatal("blank patch treated as a valid negative")
	}
}
