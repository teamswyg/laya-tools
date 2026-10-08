package familycohort

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

func fakeFile(name string) File {
	data := []byte("public software fixture only: " + name)
	return File{Path: name, SHA256: sourcecohort.Hash(data), Bytes: int64(len(data))}
}

// No files or Source/frame/comment material are opened or generated. These
// strings and declared pins are owned software fixtures only.
func fakeBundle() Bundle {
	b := Bundle{Schema: Schema, FrameSchema: fakeFile("fake-frame-schema.json"), FrameSchemaVersion: "fake-version-v1"}
	frame, review := fakeFile("fake-shared-frames.json"), fakeFile("fake-shared-reviews.json")
	for i := 0; i < SourceCount; i++ {
		split, style := "train", ""
		switch {
		case i >= 360:
			split = "test"
		case i >= 320:
			split = "cal"
		case i >= 280:
			split, style = "dev", "short"
			if i >= 310 {
				style = "general"
			}
		}
		source := SourceBinding{fmt.Sprintf("fake-source-%04d", i), split, style, fakeFile(fmt.Sprintf("fake-sources/%04d.txt", i)), review}
		b.Sources = append(b.Sources, source)
		for j := 0; j < 3; j++ {
			family := FamilyBinding{fmt.Sprintf("fake-family-%04d", i*3+j), source.ID, frame, b.FrameSchema}
			b.Families = append(b.Families, family)
			for _, locale := range []string{"ko", "en"} {
				b.Comments = append(b.Comments, CompleteComment{
					ID: family.ID + "-" + locale, FamilyID: family.ID, SourceID: source.ID,
					Source: source.Source, Frame: frame, Locale: locale,
					Text: "public fake UTF-8 software fixture 한글 " + family.ID + " " + locale,
				})
			}
		}
	}
	return b
}

func clone(b Bundle) Bundle {
	b.Sources = append([]SourceBinding{}, b.Sources...)
	b.Families = append([]FamilyBinding{}, b.Families...)
	b.Comments = append([]CompleteComment{}, b.Comments...)
	return b
}

func TestWholeStructureAndInheritedCountsWithoutQA(t *testing.T) {
	b := fakeBundle()
	// A complete multibyte comment at the exact limit must remain intact.
	b.Comments[0].Text = strings.Repeat("한", 1365) + "a"
	for i, j := 0, len(b.Sources)-1; i < j; i, j = i+1, j-1 {
		b.Sources[i], b.Sources[j] = b.Sources[j], b.Sources[i]
	}
	for i, j := 0, len(b.Families)-1; i < j; i, j = i+1, j-1 {
		b.Families[i], b.Families[j] = b.Families[j], b.Families[i]
	}
	for i, j := 0, len(b.Comments)-1; i < j; i, j = i+1, j-1 {
		b.Comments[i], b.Comments[j] = b.Comments[j], b.Comments[i]
	}
	before := clone(b)
	s, err := Validate(b)
	if err != nil || s.State != "STRUCTURAL_ONLY_QA_PENDING" || s.Sources != 400 || s.Families != 1200 || s.Comments != 2400 || s.KORows != 1200 || s.ENRows != 1200 {
		t.Fatalf("full declared structure failed: %+v %v", s, err)
	}
	if !reflect.DeepEqual(before, b) {
		t.Fatal("validation mutated supplied order or complete comment bytes")
	}
	for _, c := range s.Splits {
		if c.Families != c.Sources*3 || c.Comments != c.Sources*6 {
			t.Fatal("source partition was not inherited", c)
		}
	}
	if s.DEVStyles[0] != (Counts{"short", 30, 90, 180}) || s.DEVStyles[1] != (Counts{"general", 10, 30, 60}) {
		t.Fatal("source DEV style was not inherited", s.DEVStyles)
	}
	if s.MeaningProven || s.TrainingEligible || s.ProviderIndependence != "unknown" || len(s.PendingChecks) != 12 {
		t.Fatal("structure promoted an unverified claim", s)
	}
	// Accessor data must be fresh; mutating one report cannot affect the next.
	s.PendingChecks[0] = "changed"
	again, _ := Validate(b)
	if again.PendingChecks[0] != "whole_source_qa" {
		t.Fatal("summary retained shared mutable pending flags")
	}
}

func TestGraphCountsBindingsLocalesPinsAndTextReject(t *testing.T) {
	base := fakeBundle()
	tests := []struct {
		name, code string
		edit       func(*Bundle)
	}{
		{"source-missing", "whole_count", func(b *Bundle) { b.Sources = b.Sources[:399] }},
		{"family-missing", "whole_count", func(b *Bundle) { b.Families = b.Families[:1199] }},
		{"comment-missing", "whole_count", func(b *Bundle) { b.Comments = b.Comments[:2399] }},
		{"comment-extra", "whole_count", func(b *Bundle) { b.Comments = append(b.Comments, b.Comments[0]) }},
		{"source-duplicate", "source_id", func(b *Bundle) { b.Sources[1].ID = b.Sources[0].ID }},
		{"source-invalid-id", "source_id", func(b *Bundle) { b.Sources[0].ID = "bad/source" }},
		{"family-duplicate", "family_id", func(b *Bundle) { b.Families[1].ID = b.Families[0].ID }},
		{"comment-duplicate", "comment_id", func(b *Bundle) { b.Comments[1].ID = b.Comments[0].ID }},
		{"unknown-source", "family_source", func(b *Bundle) { b.Families[0].SourceID = "fake-missing" }},
		{"wrong-source-cardinality", "source_family_count", func(b *Bundle) { b.Families[0].SourceID = b.Sources[1].ID }},
		{"unknown-family", "comment_family", func(b *Bundle) { b.Comments[0].FamilyID = "fake-missing" }},
		{"comment-source-id", "comment_binding", func(b *Bundle) { b.Comments[0].SourceID = b.Sources[1].ID }},
		{"comment-source-pin", "comment_binding", func(b *Bundle) { b.Comments[0].Source = b.Sources[1].Source }},
		{"comment-frame-pin", "comment_binding", func(b *Bundle) { b.Comments[0].Frame = fakeFile("different-frame.json") }},
		{"frame-schema-pin", "family_pin", func(b *Bundle) { b.Families[0].FrameSchema = fakeFile("different-schema.json") }},
		{"invalid-frame-pin", "family_pin", func(b *Bundle) { b.Families[0].Frame.SHA256 = "invalid" }},
		{"review-pin", "source_pin", func(b *Bundle) { b.Sources[0].SourceReview.SHA256 = "invalid" }},
		{"source-pin-case", "source_pin", func(b *Bundle) { b.Sources[0].Source.SHA256 = strings.ToUpper(b.Sources[0].Source.SHA256) }},
		{"source-empty-byte-pin", "source_pin", func(b *Bundle) { b.Sources[0].Source.Bytes = 0 }},
		{"path-traversal", "source_pin", func(b *Bundle) { b.Sources[0].Source.Path = "../outside.txt" }},
		{"path-alias", "source_pin", func(b *Bundle) { b.Sources[0].Source.Path = "fake-sources/./0000.txt" }},
		{"path-backslash", "source_pin", func(b *Bundle) { b.Sources[0].Source.Path = `fake-sources\0000.txt` }},
		{"conflicting-path-pins", "pin_alias", func(b *Bundle) { b.Sources[0].SourceReview.Bytes++ }},
		{"split", "source_split", func(b *Bundle) { b.Sources[0].Split = "validation" }},
		{"nondev-style", "dev_style", func(b *Bundle) { b.Sources[0].DEVStyle = "short" }},
		{"missing-dev-style", "dev_style", func(b *Bundle) { b.Sources[280].DEVStyle = "" }},
		{"unknown-locale", "comment_locale", func(b *Bundle) { b.Comments[0].Locale = "KO" }},
		{"locale-duplicate", "comment_locale", func(b *Bundle) { b.Comments[1].Locale = "ko" }},
		{"family-pair-displaced", "comment_locale", func(b *Bundle) { b.Comments[0].FamilyID = b.Comments[2].FamilyID }},
		{"invalid-utf8", "comment_text", func(b *Bundle) { b.Comments[0].Text = string([]byte{0xff}) }},
		{"empty-text", "comment_text", func(b *Bundle) { b.Comments[0].Text = " \t\n" }},
		{"nul-text", "comment_text", func(b *Bundle) { b.Comments[0].Text = "fake\x00" }},
		{"oversize-text", "comment_text", func(b *Bundle) { b.Comments[0].Text = strings.Repeat("한", 1365) + "ab" }},
		{"missing-schema-version", "bundle_schema", func(b *Bundle) { b.FrameSchemaVersion = "" }},
		{"invalid-schema-pin", "bundle_schema", func(b *Bundle) { b.FrameSchema.SHA256 = "invalid" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := clone(base)
			tc.edit(&b)
			s, err := Validate(b)
			if err == nil || err.Error() != tc.code || s.Code != tc.code || s.State != "blocked" || s.ExpectedComments != CommentCount || s.MeaningProven || s.TrainingEligible {
				t.Fatalf("wrong failure: %+v %v, want %s", s, err, tc.code)
			}
		})
	}
}

func TestClosedJSONAndMandatoryFields(t *testing.T) {
	b := fakeBundle()
	b.Sources, b.Families, b.Comments = b.Sources[:1], b.Families[:1], b.Comments[:1]
	raw, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Decode(raw); err != nil {
		t.Fatal("small shape cannot be decoded independently of whole-count validation", err)
	}
	tests := []struct{ old, new string }{
		{`"schema":`, `"Schema":`},
		{`"schema":`, `"unknown":false,"schema":`},
		{`"schema":`, `"schema":"duplicate","schema":`},
		{`"source_id":`, `"Source_id":`},
		{`"source_id":`, `"source_id":"duplicate","source_id":`},
		{`"frame_schema_version":"fake-version-v1",`, ``},
		{`"source_review":{`, `"source_review":null,"removed":{`},
		{`"comments":[`, `"comments":null,"removed":[`},
		{`"text":`, `"expected_head_labels":[true],"text":`},
		{`"text":`, `"truncated":false,"text":`},
		{`"locale":"ko"`, `"locale":null`},
		{`"frame_schema_version":"fake-version-v1"`, `"frame_schema_version":null`},
	}
	for _, tc := range tests {
		changed := bytes.Replace(raw, []byte(tc.old), []byte(tc.new), 1)
		if bytes.Equal(changed, raw) {
			t.Fatal("test replacement did not change shape", tc.old)
		}
		if _, err := Decode(changed); err == nil {
			t.Fatalf("invalid closed shape accepted: %s", tc.old)
		}
	}
	text := b.Comments[0].Text
	for _, replacement := range [][]byte{[]byte(`"\ud800"`), []byte("\"\xff\""), []byte("null")} {
		quoted, _ := json.Marshal(text)
		changed := bytes.Replace(raw, quoted, replacement, 1)
		if _, err := Decode(changed); err == nil {
			t.Fatal("invalid comment Unicode/null accepted")
		}
	}
	if _, err := Decode(append(raw, []byte(`{}`)...)); err == nil {
		t.Fatal("trailing input accepted")
	}
}

func TestWholeBundleBudgetIsSeparateFromPerCommentBudget(t *testing.T) {
	b := fakeBundle()
	text := strings.Repeat("x", MaxCommentBytes)
	for i := range b.Comments {
		b.Comments[i].Text = text
	}
	if s, err := Validate(b); err != nil || s.Comments != CommentCount {
		t.Fatal("individual 4096-byte allowance was lowered", err)
	}
	raw, err := json.Marshal(b)
	if err != nil || len(raw) <= MaxInputBytes {
		t.Fatal("fixture did not exceed encoded transport budget", err)
	}
	if _, err := Decode(raw); err == nil || err.Error() != "bundle_bytes" {
		t.Fatal("oversize bundle was admitted or partially selected", err)
	}
}
