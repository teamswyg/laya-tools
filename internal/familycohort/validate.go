package familycohort

import (
	"encoding/hex"
	"path"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

func identifier(s string) bool {
	if len(s) == 0 || len(s) > 128 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

func relative(s string) bool {
	return len(s) > 0 && len(s) <= 4096 && utf8.ValidString(s) && !strings.ContainsAny(s, "\\:\x00") && !path.IsAbs(s) && path.Clean(s) == s && s != "." && s != ".." && !strings.HasPrefix(s, "../")
}

func validFile(f File) bool {
	b, err := hex.DecodeString(f.SHA256)
	return relative(f.Path) && err == nil && len(b) == 32 && strings.ToLower(f.SHA256) == f.SHA256 && f.Bytes >= 0 && (f.Bytes == 0) == (f.SHA256 == sourcecohort.Hash(nil))
}

func sourceIndex(sources []SourceBinding, id string) int {
	i := sort.Search(len(sources), func(i int) bool { return sources[i].ID >= id })
	if i == len(sources) || sources[i].ID != id {
		return -1
	}
	return i
}

func familyIndex(families []FamilyBinding, id string) int {
	i := sort.Search(len(families), func(i int) bool { return families[i].ID >= id })
	if i == len(families) || families[i].ID != id {
		return -1
	}
	return i
}

func countIndex(counts []Counts, name string) int {
	for i, c := range counts {
		if c.Name == name {
			return i
		}
	}
	return -1
}

// Decode checks exactly one closed JSON shape; nullable or optional fields are
// deliberately absent from this bundle. It preserves decoded comment text.
func Decode(data []byte) (Bundle, error) {
	var b Bundle
	if len(data) == 0 || len(data) > MaxInputBytes {
		return b, Error("bundle_bytes")
	}
	if sourcecohort.Decode(data, &b) != nil {
		return Bundle{}, Error("bundle_json")
	}
	return b, nil
}

// Validate joins the complete declared graph without reading any pinned file,
// modifying caller slices, generating content or interpreting frame semantics.
// The encoded transport cap is checked by Decode and Run, not by this in-memory
// API; every individual decoded comment still has its full 4096-byte allowance.
func Validate(b Bundle) (Summary, error) {
	s := initial()
	s.Sources, s.Families, s.Comments = len(b.Sources), len(b.Families), len(b.Comments)
	fail := func(code string) (Summary, error) { s.Code = code; return s, Error(code) }
	if b.Schema != Schema || !validFile(b.FrameSchema) || len(b.FrameSchemaVersion) == 0 || len(b.FrameSchemaVersion) > 128 || !utf8.ValidString(b.FrameSchemaVersion) || strings.TrimSpace(b.FrameSchemaVersion) != b.FrameSchemaVersion || strings.ContainsRune(b.FrameSchemaVersion, 0) {
		return fail("bundle_schema")
	}
	if s.Sources != SourceCount || s.Families != FamilyCount || s.Comments != CommentCount {
		return fail("whole_count")
	}
	sources := append([]SourceBinding{}, b.Sources...)
	families := append([]FamilyBinding{}, b.Families...)
	comments := append([]CompleteComment{}, b.Comments...)
	sort.Slice(sources, func(i, j int) bool { return sources[i].ID < sources[j].ID })
	sort.Slice(families, func(i, j int) bool { return families[i].ID < families[j].ID })
	sort.Slice(comments, func(i, j int) bool { return comments[i].ID < comments[j].ID })
	pins := make([]File, 0, 1+2*SourceCount+2*FamilyCount+2*CommentCount)
	pins = append(pins, b.FrameSchema)
	for i, source := range sources {
		if !identifier(source.ID) || i > 0 && sources[i-1].ID == source.ID {
			return fail("source_id")
		}
		if !validFile(source.Source) || !validFile(source.SourceReview) {
			return fail("source_pin")
		}
		k := countIndex(s.Splits, source.Split)
		if k < 0 {
			return fail("source_split")
		}
		s.Splits[k].Sources++
		if source.Split == "dev" {
			k := countIndex(s.DEVStyles, source.DEVStyle)
			if k < 0 {
				return fail("dev_style")
			}
			s.DEVStyles[k].Sources++
		} else if source.DEVStyle != "" {
			return fail("dev_style")
		}
		pins = append(pins, source.Source, source.SourceReview)
	}
	var perSource [SourceCount]int
	for i, family := range families {
		if !identifier(family.ID) || i > 0 && families[i-1].ID == family.ID {
			return fail("family_id")
		}
		j := sourceIndex(sources, family.SourceID)
		if j < 0 {
			return fail("family_source")
		}
		if !validFile(family.Frame) || family.FrameSchema != b.FrameSchema {
			return fail("family_pin")
		}
		perSource[j]++
		source := sources[j]
		s.Splits[countIndex(s.Splits, source.Split)].Families++
		if source.Split == "dev" {
			s.DEVStyles[countIndex(s.DEVStyles, source.DEVStyle)].Families++
		}
		pins = append(pins, family.Frame, family.FrameSchema)
	}
	for _, n := range perSource {
		if n != 3 {
			return fail("source_family_count")
		}
	}
	var pairs [FamilyCount][2]bool
	for i, comment := range comments {
		if !identifier(comment.ID) || i > 0 && comments[i-1].ID == comment.ID {
			return fail("comment_id")
		}
		j := familyIndex(families, comment.FamilyID)
		if j < 0 {
			return fail("comment_family")
		}
		family := families[j]
		source := sources[sourceIndex(sources, family.SourceID)]
		if comment.SourceID != source.ID || comment.Source != source.Source || comment.Frame != family.Frame {
			return fail("comment_binding")
		}
		locale := -1
		switch comment.Locale {
		case "ko":
			locale = 0
		case "en":
			locale = 1
		}
		if locale < 0 || pairs[j][locale] {
			return fail("comment_locale")
		}
		if len(comment.Text) > MaxCommentBytes || !utf8.ValidString(comment.Text) || strings.TrimSpace(comment.Text) == "" || strings.ContainsRune(comment.Text, 0) {
			return fail("comment_text")
		}
		pairs[j][locale] = true
		if locale == 0 {
			s.KORows++
		} else {
			s.ENRows++
		}
		s.Splits[countIndex(s.Splits, source.Split)].Comments++
		if source.Split == "dev" {
			s.DEVStyles[countIndex(s.DEVStyles, source.DEVStyle)].Comments++
		}
		pins = append(pins, comment.Source, comment.Frame)
	}
	for _, pair := range pairs {
		if !pair[0] || !pair[1] {
			return fail("family_comment_pair")
		}
	}
	// One root-relative name cannot silently declare conflicting bytes. Equal
	// pins and equal text do not establish source or semantic independence.
	sort.Slice(pins, func(i, j int) bool { return pins[i].Path < pins[j].Path })
	for i := 1; i < len(pins); i++ {
		if pins[i].Path == pins[i-1].Path && pins[i] != pins[i-1] {
			return fail("pin_alias")
		}
	}
	s.State = "STRUCTURAL_ONLY_QA_PENDING"
	return s, nil
}
