package shortclaim

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
)

func validInput() Input {
	return Input{Schema: Schema, Request: "KeepActive entries; remove EXPIRED entries.", Candidates: []Candidate{{ID: "keep-1", Text: "retainActive entries"}, {ID: "keep-2", Text: "removeActive entries"}}, Provenance: "original.contract-56"}
}

func wire(t *testing.T, in Input) string {
	t.Helper()
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestInputPreservesRawOrderAndOwnsCandidateArray(t *testing.T) {
	in := validInput()
	p, err := Load(strings.NewReader(wire(t, in)))
	if err != nil {
		t.Fatal(err)
	}
	if p.Request != in.Request || p.NormalizedRequest != "keep active entries remove expired entries" || p.Provenance != in.Provenance || p.Count != 2 {
		t.Fatal("request or metadata not preserved")
	}
	if p.Candidates[0] != (PreparedCandidate{"keep-1", "retainActive entries", "retain active entries"}) || p.Candidates[1].ID != "keep-2" || p.Candidates[1].NormalizedText != "remove active entries" {
		t.Fatal("candidate text or order changed")
	}
	direct, err := Validate(in)
	if err != nil || p != direct {
		t.Fatal("wire and direct contracts differ")
	}
	in.Candidates[0] = Candidate{ID: "changed", Text: "changed"}
	if p.Candidates[0].ID != "keep-1" || direct.Candidates[0].ID != "keep-1" {
		t.Fatal("prepared aliases caller slice")
	}
	copy := p
	copy.Candidates[0].Text = "changed"
	if p.Candidates[0].Text != "retainActive entries" {
		t.Fatal("prepared copies share mutable array")
	}
	if err = ValidatePrepared(p); err != nil {
		t.Fatal(err)
	}
}

func TestStrictDecodedKeysAndTypes(t *testing.T) {
	base := `{"schema":"riido-short-behavior-claim-v1","request":"keep entries","candidates":[{"id":"a","text":"keep entries"}],"provenance":"own"}`
	cases := []struct {
		name, input string
		want        Error
	}{
		{"top_duplicate", strings.Replace(base, `"request":"keep entries"`, `"request":"keep entries","request":"remove entries"`, 1), ErrDuplicateField},
		{"escaped_top_duplicate", strings.Replace(base, `"request":"keep entries"`, `"request":"keep entries","requ\u0065st":"remove entries"`, 1), ErrDuplicateField},
		{"candidate_duplicate", strings.Replace(base, `"id":"a"`, `"id":"a","\u0069d":"b"`, 1), ErrDuplicateField},
		{"candidate_text_duplicate", strings.Replace(base, `"text":"keep entries"`, `"text":"keep entries","text":"remove entries"`, 1), ErrDuplicateField},
		{"unknown_top", strings.Replace(base, `"provenance":"own"`, `"provenance":"own","extra":{"a":1}`, 1), ErrUnknownField},
		{"top_case_alias", strings.Replace(base, `"request"`, `"Request"`, 1), ErrUnknownField},
		{"candidate_case_alias", strings.Replace(base, `"id"`, `"ID"`, 1), ErrUnknownField},
		{"candidate_unknown", strings.Replace(base, `"id":"a"`, `"id":"a","target":true`, 1), ErrUnknownField},
		{"request_null", strings.Replace(base, `"request":"keep entries"`, `"request":null`, 1), ErrJSON},
		{"request_object", strings.Replace(base, `"request":"keep entries"`, `"request":{"x":1}`, 1), ErrJSON},
		{"candidate_null", strings.Replace(base, `{"id":"a","text":"keep entries"}`, `null`, 1), ErrJSON},
		{"candidates_null", strings.Replace(base, `[{"id":"a","text":"keep entries"}]`, `null`, 1), ErrJSON},
		{"candidates_object", strings.Replace(base, `[{"id":"a","text":"keep entries"}]`, `{}`, 1), ErrJSON},
		{"empty_candidates", strings.Replace(base, `[{"id":"a","text":"keep entries"}]`, `[]`, 1), ErrCandidateCount},
		{"missing_schema", strings.Replace(base, `"schema":"riido-short-behavior-claim-v1",`, "", 1), ErrSchema},
		{"missing_request", strings.Replace(base, `"request":"keep entries",`, "", 1), ErrEmptyText},
		{"missing_provenance", strings.Replace(base, `,"provenance":"own"`, "", 1), ErrIdentifier},
		{"missing_candidate_id", strings.Replace(base, `"id":"a",`, "", 1), ErrIdentifier},
		{"missing_candidate_text", strings.Replace(base, `,"text":"keep entries"`, "", 1), ErrEmptyText},
		{"schema_case_value", strings.Replace(base, Schema, strings.ToUpper(Schema), 1), ErrSchema},
		{"array_root", "[" + base + "]", ErrJSON},
		{"trailing_object", base + "{}", ErrJSON},
		{"trailing_scalar", base + " true", ErrJSON},
		{"truncated", base[:len(base)-1], ErrJSON},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, err := Load(strings.NewReader(c.input))
			if err != c.want || p != (Prepared{}) {
				t.Fatalf("got %v, want %v or nonzero prepared", err, c.want)
			}
		})
	}
	for _, field := range []string{"schema", "request", "candidates", "provenance", "id", "text"} {
		t.Run("every_case_alias_"+field, func(t *testing.T) {
			_, err := Load(strings.NewReader(strings.Replace(base, `"`+field+`"`, `"`+strings.ToUpper(field)+`"`, 1)))
			if err != ErrUnknownField {
				t.Fatalf("case alias accepted: %v", err)
			}
		})
	}
	// Exact field spelling after JSON escape decoding is valid on its own.
	if _, err := Load(strings.NewReader(strings.Replace(base, `"request"`, `"requ\u0065st"`, 1))); err != nil {
		t.Fatal(err)
	}
}

func TestUnicodeScalarValidationAcrossEntireJSON(t *testing.T) {
	base := `{"schema":"riido-short-behavior-claim-v1","request":"keep entries","candidates":[{"id":"a","text":"keep entries"}],"provenance":"own"}`
	for _, escaped := range []string{`\ud800`, `\uDBFF`, `\udc00`, `\uDFFF`, `\ud800\u0041`, `\ud800\ud800`, `\ud800x`, `\ud800\\udc00`} {
		for _, where := range []string{"schema", "request", "provenance", "id", "text", "unknown_value", "unknown_key"} {
			s := base
			switch where {
			case "schema":
				s = strings.Replace(s, `"schema":"riido-short-behavior-claim-v1"`, `"schema":"`+escaped+`"`, 1)
			case "request":
				s = strings.Replace(s, `"request":"keep entries"`, `"request":"`+escaped+`"`, 1)
			case "provenance":
				s = strings.Replace(s, `"provenance":"own"`, `"provenance":"`+escaped+`"`, 1)
			case "id":
				s = strings.Replace(s, `"id":"a"`, `"id":"`+escaped+`"`, 1)
			case "text":
				s = strings.Replace(s, `"text":"keep entries"`, `"text":"`+escaped+`"`, 1)
			case "unknown_value":
				s = strings.TrimSuffix(s, "}") + `,"extra":{"nested":["` + escaped + `"]}}`
			case "unknown_key":
				s = strings.TrimSuffix(s, "}") + `,"` + escaped + `":1}`
			}
			if p, err := Load(strings.NewReader(s)); err != ErrUnicode || p != (Prepared{}) {
				t.Fatalf("unpaired surrogate in %s: %v", where, err)
			}
		}
	}
	for _, text := range []string{`keep \ud83d\ude00 entries`, `keep \uD800\uDC00 entries`, `keep \\ud800 entries`, `keep \ufffd entries`, "keep � entries", `keep \"quoted\" entries`} {
		s := strings.Replace(base, `"request":"keep entries"`, `"request":"`+text+`"`, 1)
		if _, err := Load(strings.NewReader(s)); err != nil {
			t.Fatalf("valid scalar or literal escape rejected: %v", err)
		}
	}
	bad := []byte(base)
	bad[bytes.Index(bad, []byte("keep"))] = 0xff
	if _, err := Load(bytes.NewReader(bad)); err != ErrUnicode {
		t.Fatal("invalid wire UTF-8 accepted")
	}
	for _, field := range []string{"schema", "request", "provenance", "id", "text"} {
		in := validInput()
		invalid := string([]byte{'a', 0xed, 0xa0, 0x80}) // UTF-8 encoding of a surrogate is invalid.
		switch field {
		case "schema":
			in.Schema = invalid
		case "request":
			in.Request = invalid
		case "provenance":
			in.Provenance = invalid
		case "id":
			in.Candidates[0].ID = invalid
		case "text":
			in.Candidates[0].Text = invalid
		}
		if p, err := Validate(in); err != ErrUnicode || p != (Prepared{}) {
			t.Fatalf("invalid direct UTF-8 in %s accepted: %v", field, err)
		}
	}
}

func TestTextRawAndNormalizedBoundaries(t *testing.T) {
	cases := []struct {
		name, text string
		want       error
	}{
		{"raw_512", strings.Repeat("x", 512), nil},
		{"raw_513", strings.Repeat("x", 513), ErrTextBounds},
		{"utf8_512_bytes", strings.Repeat("가", 170) + "aa", nil},
		{"utf8_513_bytes", strings.Repeat("가", 171), ErrTextBounds},
		{"words_32", strings.TrimSpace(strings.Repeat("word ", 32)), nil},
		{"words_33", strings.TrimSpace(strings.Repeat("word ", 33)), ErrNormalizedBounds},
		{"camel_32_words", strings.TrimSpace(strings.Repeat("itemOne ", 16)), nil},
		{"camel_34_words", strings.TrimSpace(strings.Repeat("itemOne ", 17)), ErrNormalizedBounds},
		// U+023A lowercases to U+2C65, expanding two UTF-8 bytes to three.
		{"normalized_512_bytes", strings.Repeat("Ⱥ", 170) + "11", nil},
		{"normalized_513_bytes", strings.Repeat("Ⱥ", 171), ErrNormalizedBounds},
		{"empty", "", ErrEmptyText},
		{"punctuation_only", " ! _ -- \t", ErrEmptyText},
	}
	for _, c := range cases {
		for _, candidate := range []bool{false, true} {
			t.Run(c.name+"_"+string(rune('0'+btoi(candidate))), func(t *testing.T) {
				in := validInput()
				if candidate {
					in.Candidates[0].Text = c.text
				} else {
					in.Request = c.text
				}
				p, err := Validate(in)
				if err != c.want || err != nil && p != (Prepared{}) {
					t.Fatalf("direct: got %v, want %v", err, c.want)
				}
				loaded, loadErr := Load(strings.NewReader(wire(t, in)))
				if loadErr != c.want || loaded != p {
					t.Fatalf("wire: got %v, want %v", loadErr, c.want)
				}
			})
		}
	}
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func TestIdentifiersCountsAndUnusedSlots(t *testing.T) {
	in := validInput()
	in.Candidates = make([]Candidate, 8)
	for i := range in.Candidates {
		in.Candidates[i] = Candidate{ID: string(rune('a' + i)), Text: "keep entries"}
	}
	if p, err := Validate(in); err != nil || p.Count != 8 {
		t.Fatal("eight candidates rejected")
	}
	in.Candidates = append(in.Candidates, Candidate{ID: "ninth", Text: "keep entries"})
	if _, err := Validate(in); err != ErrCandidateCount {
		t.Fatal(err)
	}
	if _, err := Load(strings.NewReader(wire(t, in))); err != ErrCandidateCount {
		t.Fatal(err)
	}
	in.Candidates = nil
	if _, err := Validate(in); err != ErrCandidateCount {
		t.Fatal(err)
	}
	for _, invalid := range []string{"", ".hidden", "path/name", "path\\name", "bad:kind", "has space", "한글", "id\x00end", strings.Repeat("a", MaxIDBytes+1)} {
		in = validInput()
		in.Candidates[0].ID = invalid
		if p, err := Validate(in); err != ErrIdentifier || p != (Prepared{}) {
			t.Fatal("invalid candidate identifier accepted")
		}
	}
	in = validInput()
	in.Candidates[0].ID = strings.Repeat("a", MaxIDBytes)
	in.Provenance = strings.Repeat("b", MaxProvenanceBytes)
	if _, err := Validate(in); err != nil {
		t.Fatal("identifier inclusive boundary rejected")
	}
	in.Provenance += "x"
	if _, err := Validate(in); err != ErrIdentifier {
		t.Fatal("provenance overflow accepted")
	}
	in = validInput()
	in.Candidates[1].ID = in.Candidates[0].ID
	if _, err := Validate(in); err != ErrDuplicateID {
		t.Fatal("duplicate candidate ID accepted")
	}
	in.Candidates[1].ID = strings.ToUpper(in.Candidates[0].ID)
	if _, err := Validate(in); err != nil {
		t.Fatal("case-sensitive IDs merged")
	}
	p, _ := Validate(validInput())
	for _, mutate := range []func(*Prepared){
		func(p *Prepared) { p.Count = 0 },
		func(p *Prepared) { p.Count = 9 },
		func(p *Prepared) { p.Schema = "other" },
		func(p *Prepared) { p.NormalizedRequest = "forged" },
		func(p *Prepared) { p.Candidates[0].NormalizedText = "forged" },
		func(p *Prepared) { p.Candidates[7].ID = "hidden" },
		func(p *Prepared) { p.Candidates[7].Text = "hidden" },
		func(p *Prepared) { p.Provenance = "private/path" },
	} {
		bad := p
		mutate(&bad)
		if ValidatePrepared(bad) == nil {
			t.Fatal("forged prepared accepted")
		}
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("authored-read-error-marker")
}

type countingReader struct {
	r io.Reader
	n int
}

func (r *countingReader) Read(b []byte) (int, error) {
	n, err := r.r.Read(b)
	r.n += n
	return n, err
}

func TestWireBudgetAndFixedErrors(t *testing.T) {
	base := wire(t, validInput())
	atLimit := base + strings.Repeat(" ", MaxJSONBytes-len(base))
	if _, err := Load(strings.NewReader(atLimit)); err != nil {
		t.Fatal("inclusive JSON wire limit rejected")
	}
	r := &countingReader{r: strings.NewReader(atLimit + strings.Repeat(" ", 10000))}
	if p, err := Load(r); err != ErrJSONBounds || p != (Prepared{}) || r.n != MaxJSONBytes+1 {
		t.Fatal("wire overflow was accepted or read without bound")
	}
	if _, err := Load(failingReader{}); err != ErrRead || strings.Contains(err.Error(), "marker") {
		t.Fatal("reader diagnostic leaked")
	}
	if _, err := Load(nil); err != ErrRead {
		t.Fatal(err)
	}
	in := validInput()
	in.Request = strings.Repeat("authored-input-marker ", 100)
	if p, err := Validate(in); err != ErrTextBounds || p != (Prepared{}) || strings.Contains(err.Error(), "marker") {
		t.Fatal("input error leaked text")
	}
	// Every prepared string is derived only from decoded input, not decoder text.
	if _, err := Load(strings.NewReader(`{"private-marker":1}`)); err != ErrUnknownField || strings.Contains(err.Error(), "marker") {
		t.Fatal("unknown-key error leaked text")
	}
}

func TestConcurrentInputPreparationIsIndependent(t *testing.T) {
	raw := wire(t, validInput())
	var wg sync.WaitGroup
	for range 12 {
		wg.Go(func() {
			for range 20 {
				p, err := Load(strings.NewReader(raw))
				if err != nil || p.Candidates[0].ID != "keep-1" || ValidatePrepared(p) != nil {
					t.Error("independent preparation failed")
					return
				}
			}
		})
	}
	wg.Wait()
}
