// Package statehintbody derives bounded model text from an explicitly declared
// body format. It does not establish authorization, owner revisions or UI parity.
package statehintbody

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Format string

const (
	PlainText        Format = "plain_text"
	NeutralHTML      Format = "neutral_html_fragment"
	ExtractorVersion        = "statehint-body-neutral-xml-v1"
	MaxRawBytes             = 16 * 1024
	MaxTextBytes            = 4096
	MaxDepth                = 32
	MaxTokens               = 2048
	MaxAttributes           = 8
)

var (
	ErrInput     = errors.New("statehintbody: invalid input")
	ErrFormat    = errors.New("statehintbody: explicit supported format required")
	ErrRawLimit  = errors.New("statehintbody: raw body limit exceeded")
	ErrTextLimit = errors.New("statehintbody: model text limit exceeded")
	ErrBudget    = errors.New("statehintbody: parser budget exceeded")
	ErrMarkup    = errors.New("statehintbody: unsupported or malformed markup")
)

// Text is private model input. The hashes are derived observations, never owner
// revisions. Callers must not include Text or hashes in public telemetry.
type Result struct {
	Text                       string `json:"-"`
	Format                     Format `json:"format"`
	ExtractorVersion           string `json:"extractor_version"`
	RawBodySHA256              string `json:"-"`
	ModelTextSHA256            string `json:"-"`
	RawBytes                   int    `json:"raw_bytes"`
	TextBytes                  int    `json:"text_bytes"`
	Tokens                     int    `json:"tokens"`
	RenderedVisibilityVerified bool   `json:"rendered_visibility_verified"`
	OwnerRevisionVerified      bool   `json:"owner_revision_verified"`
}

// Workspace belongs to one caller/in-flight extraction. No shared cache or
// lock is used. Scratch text is cleared before Extract returns.
type Workspace struct {
	text  [MaxTextBytes]byte
	n     int
	stack [MaxDepth]tag
	depth int
}

type tag uint8

const (
	unknown tag = iota
	paragraph
	division
	linebreak
	span
	strong
	bold
	emphasis
	italic
	underline
	anchor
)

func knownTag(name string) tag {
	switch name {
	case "p":
		return paragraph
	case "div":
		return division
	case "br":
		return linebreak
	case "span":
		return span
	case "strong":
		return strong
	case "b":
		return bold
	case "em":
		return emphasis
	case "i":
		return italic
	case "u":
		return underline
	case "a":
		return anchor
	default:
		return unknown
	}
}

func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	var encoded [64]byte
	hex.Encode(encoded[:], sum[:])
	return string(encoded[:])
}

func (w *Workspace) append(value []byte) error {
	if len(value) > MaxTextBytes-w.n {
		return ErrTextLimit
	}
	copy(w.text[w.n:], value)
	w.n += len(value)
	return nil
}

func (w *Workspace) boundary() error {
	if w.n == 0 || w.text[w.n-1] == '\n' {
		return nil
	}
	return w.append([]byte{'\n'})
}

func (w *Workspace) clear() {
	clear(w.text[:w.n])
	clear(w.stack[:w.depth])
	w.n, w.depth = 0, 0
}

// Extract accepts plain text exactly, or a deliberately small neutral rich-text
// subset. Unknown/malformed/oversize input produces no partial Result. It never
// guesses formats, strips unsupported subtrees or repairs unbalanced markup.
func Extract(ctx context.Context, raw string, format Format, w *Workspace) (Result, error) {
	if ctx == nil || w == nil {
		return Result{}, ErrInput
	}
	w.clear()
	defer w.clear()
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if format != PlainText && format != NeutralHTML {
		return Result{}, ErrFormat
	}
	if len(raw) > MaxRawBytes {
		return Result{}, ErrRawLimit
	}
	if !utf8.ValidString(raw) || strings.ContainsRune(raw, 0) {
		return Result{}, ErrInput
	}
	text, tokens := raw, 0
	if format == PlainText {
		if len(raw) > MaxTextBytes {
			return Result{}, ErrTextLimit
		}
	} else {
		// Reject declarations, comments, CDATA and processing instructions before
		// normalization; CDATA otherwise appears as ordinary xml.CharData.
		if strings.Contains(raw, "<!") || strings.Contains(raw, "<?") || !validNumericEntities(raw) {
			return Result{}, ErrMarkup
		}
		// These two exact lexical rewrites preserve this supported subset's
		// semantics. Other HTML void forms/entities are explicitly unsupported.
		// Escaped &lt;br&gt; and &amp;nbsp; are never rewritten or decoded twice.
		normalized := strings.ReplaceAll(raw, "<br>", "<br/>")
		normalized = strings.ReplaceAll(normalized, "&nbsp;", "&#160;")
		decoder := xml.NewDecoder(strings.NewReader(normalized))
		decoder.Strict = true
		for {
			if err := ctx.Err(); err != nil {
				return Result{}, err
			}
			beforeToken := decoder.InputOffset()
			// Balance and namespace handling belong to the fixed caller stack.
			// RawToken avoids allocating a second decoder-owned element stack.
			token, err := decoder.RawToken()
			if err == io.EOF {
				if w.depth != 0 || decoder.InputOffset() != int64(len(normalized)) || decoder.InputOffset() != beforeToken {
					return Result{}, ErrMarkup
				}
				break
			}
			if err != nil {
				return Result{}, ErrMarkup
			}
			tokens++
			if tokens > MaxTokens {
				return Result{}, ErrBudget
			}
			switch t := token.(type) {
			case xml.StartElement:
				kind := knownTag(t.Name.Local)
				rawTag := normalized[beforeToken:decoder.InputOffset()]
				if t.Name.Space != "" || kind == unknown || !validAttributes(kind, t.Attr) || !strictStartTag(rawTag) {
					return Result{}, ErrMarkup
				}
				selfClosing := strings.HasSuffix(rawTag, "/>")
				if (kind == linebreak) != selfClosing {
					return Result{}, ErrMarkup
				}
				if w.depth >= MaxDepth {
					return Result{}, ErrBudget
				}
				if w.depth > 0 {
					parent := w.stack[w.depth-1]
					if parent == linebreak || ((kind == paragraph || kind == division) && parent != division) {
						return Result{}, ErrMarkup
					}
				}
				if kind == anchor {
					for _, parent := range w.stack[:w.depth] {
						if parent == anchor {
							return Result{}, ErrMarkup
						}
					}
				}
				if kind == linebreak {
					if err := w.append([]byte{'\n'}); err != nil {
						return Result{}, err
					}
				} else if kind == paragraph || kind == division {
					if err := w.boundary(); err != nil {
						return Result{}, err
					}
				}
				w.stack[w.depth], w.depth = kind, w.depth+1
			case xml.EndElement:
				kind := knownTag(t.Name.Local)
				if t.Name.Space != "" || w.depth == 0 || kind != w.stack[w.depth-1] {
					return Result{}, ErrMarkup
				}
				w.depth--
				w.stack[w.depth] = unknown
				if kind == paragraph || kind == division {
					if err := w.boundary(); err != nil {
						return Result{}, err
					}
				}
			case xml.CharData:
				if w.depth > 0 && w.stack[w.depth-1] == linebreak && len(t) != 0 {
					return Result{}, ErrMarkup
				}
				if err := w.append(t); err != nil {
					return Result{}, err
				}
			default:
				return Result{}, ErrMarkup
			}
		}
		text = string(w.text[:w.n])
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	return Result{Text: text, Format: format, ExtractorVersion: ExtractorVersion, RawBodySHA256: digest(raw), ModelTextSHA256: digest(text), RawBytes: len(raw), TextBytes: len(text), Tokens: tokens}, nil
}

// encoding/xml accepts adjacent attributes even in strict mode. This narrow
// lexical check requires separators and quoted values before trusting its token.
func strictStartTag(raw string) bool {
	if len(raw) < 3 || raw[0] != '<' || raw[len(raw)-1] != '>' {
		return false
	}
	end, i := len(raw)-1, 1
	for i < end && raw[i] >= 'a' && raw[i] <= 'z' {
		i++
	}
	if i == 1 {
		return false
	}
	space := func(c byte) bool { return c == ' ' || c == '\t' || c == '\r' || c == '\n' }
	for i < end {
		if raw[i] == '/' {
			return i+1 == end
		}
		if !space(raw[i]) {
			return false
		}
		for i < end && space(raw[i]) {
			i++
		}
		if i == end {
			return true
		}
		if raw[i] == '/' {
			return i+1 == end
		}
		start := i
		for i < end && (raw[i] >= 'a' && raw[i] <= 'z' || raw[i] == '-') {
			i++
		}
		if i == start {
			return false
		}
		for i < end && space(raw[i]) {
			i++
		}
		if i == end || raw[i] != '=' {
			return false
		}
		i++
		for i < end && space(raw[i]) {
			i++
		}
		if i == end || (raw[i] != '\'' && raw[i] != '"') {
			return false
		}
		quote := raw[i]
		i++
		for i < end && raw[i] != quote {
			if raw[i] == '<' {
				return false
			}
			i++
		}
		if i == end {
			return false
		}
		i++
	}
	return true
}

func validAttributes(kind tag, attrs []xml.Attr) bool {
	if len(attrs) > MaxAttributes || (kind == linebreak && len(attrs) != 0) {
		return false
	}
	var names [MaxAttributes]string
	mention, hasMentionField := false, false
	for i, attr := range attrs {
		if attr.Name.Space != "" || strings.Contains(attr.Name.Local, ":") {
			return false
		}
		name := attr.Name.Local
		for _, prior := range names[:i] {
			if name == prior {
				return false
			}
		}
		names[i] = name
		switch name {
		case "title":
		case "href":
			if kind != anchor {
				return false
			}
		case "data-type":
			if kind != span || attr.Value != "mention" {
				return false
			}
			mention = true
		case "data-id", "data-label", "data-mention-type":
			if kind != span {
				return false
			}
			hasMentionField = true
		default:
			// Includes namespaces, style, class, hidden, event handlers, and any
			// other attribute whose visual/semantic meaning is not supported.
			return false
		}
	}
	return !hasMentionField || mention
}

func validNumericEntities(raw string) bool {
	for start := 0; start < len(raw); start++ {
		if raw[start] != '&' || start+1 >= len(raw) || raw[start+1] != '#' {
			continue
		}
		end := start + 2
		for end < len(raw) && raw[end] != ';' && end-start <= 13 {
			end++
		}
		if end == len(raw) || raw[end] != ';' || end-start > 13 {
			return false
		}
		digits, base := raw[start+2:end], 10
		if strings.HasPrefix(digits, "x") {
			digits, base = digits[1:], 16
		}
		if digits == "" {
			return false
		}
		for _, c := range digits {
			if !(c >= '0' && c <= '9' || base == 16 && (c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F')) {
				return false
			}
		}
		value, err := strconv.ParseUint(digits, base, 32)
		if err != nil || !(value == 9 || value == 10 || value == 13 || value >= 0x20 && value <= 0xD7FF || value >= 0xE000 && value <= 0xFFFD || value >= 0x10000 && value <= 0x10FFFF) {
			return false
		}
		start = end
	}
	return true
}
