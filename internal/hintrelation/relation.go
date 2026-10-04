// SPDX-License-Identifier: Apache-2.0
// Package hintrelation extracts bounded, explicitly recognized interval hints.
// It is an experimental representation and an unlearned comparison control,
// not a truth checker, model loader, router, or approval mechanism.
package hintrelation

import (
	"errors"
	"strings"
	"unicode/utf8"
)

const (
	Schema        = "riido-interval-relations8-v1"
	Dimension     = 8
	MaxCandidates = 8
	MaxBytes      = 512
	MaxWords      = 32
	MaxTokens     = 64
)

var ErrInput = errors.New("hintrelation_input_bounds_or_utf8")

type Direction uint8

const (
	DirectionUnknown Direction = iota
	Ascending
	Descending
)

type Bound uint8

const (
	BoundUnknown Bound = iota
	Unbounded
	Open
	Closed
)

type Coverage uint8

const (
	HasDirection Coverage = 1 << iota
	HasLower
	HasUpper
	HasStop
)

// Attributes reports grammar recognition only. A recognized candidate may be
// false for a fixture; an unrecognized one may be true. Coverage must never be
// used to change labels, loss weights, evaluation eligibility, or role masks.
type Attributes struct {
	Direction Direction
	Lower     Bound
	Upper     Bound
	CanStop   bool
	Coverage  Coverage
}

const (
	DirectionMatch = iota
	DirectionConflict
	LowerMatch
	LowerConflict
	UpperMatch
	UpperConflict
	StopMatch
	StopConflict
)

// Columns owns fixed SoA storage. The query occupies slot zero in the attribute
// columns; candidates occupy slots 1..count. No source strings, maps, pooled
// scratch, locks, labels, source IDs, or model coefficients are retained.
// The feature payload is 8*8*4 = 256 bytes, not a total-memory/RSS measurement.
type Columns struct {
	direction [MaxCandidates + 1]Direction
	lower     [MaxCandidates + 1]Bound
	upper     [MaxCandidates + 1]Bound
	stop      [MaxCandidates + 1]bool
	coverage  [MaxCandidates + 1]Coverage
	values    [Dimension][MaxCandidates]float32
	count     uint8
}

func pair(q, d Attributes) [Dimension]float32 {
	var v [Dimension]float32
	compare := func(mask Coverage, same bool, match int) {
		if q.Coverage&d.Coverage&mask != 0 {
			if same {
				v[match] = 1
			} else {
				v[match+1] = 1
			}
		}
	}
	compare(HasDirection, q.Direction == d.Direction, DirectionMatch)
	compare(HasLower, q.Lower == d.Lower, LowerMatch)
	compare(HasUpper, q.Upper == d.Upper, UpperMatch)
	compare(HasStop, q.CanStop == d.CanStop, StopMatch)
	return v
}

// Extract accepts raw text only. It neither accepts nor consults supervision.
// Unsupported, valid text yields zero coverage rather than guessed semantics.
func Extract(request, candidate string) ([Dimension]float32, Attributes, Attributes, error) {
	q, err := ParseRequest(request)
	if err != nil {
		return [Dimension]float32{}, Attributes{}, Attributes{}, err
	}
	d, err := ParseCandidate(candidate)
	if err != nil {
		return [Dimension]float32{}, Attributes{}, Attributes{}, err
	}
	return pair(q, d), q, d, nil
}

// Build parses the query once and each live candidate once. On any bounds error
// it returns a zero value, never partially prepared columns. Unused slots are
// zero. Existing 8192-dimensional weights are incompatible with this schema.
func Build(request string, candidates [MaxCandidates]string, count int) (Columns, error) {
	if count < 1 || count > MaxCandidates {
		return Columns{}, ErrInput
	}
	q, err := ParseRequest(request)
	if err != nil {
		return Columns{}, err
	}
	var out Columns
	out.count = uint8(count)
	out.set(0, q)
	for i := 0; i < count; i++ {
		d, err := ParseCandidate(candidates[i])
		if err != nil {
			return Columns{}, err
		}
		out.set(i+1, d)
		v := pair(q, d)
		for column := range v {
			out.values[column][i] = v[column]
		}
	}
	return out, nil
}

func (c *Columns) set(i int, a Attributes) {
	c.direction[i], c.lower[i], c.upper[i] = a.Direction, a.Lower, a.Upper
	c.stop[i], c.coverage[i] = a.CanStop, a.Coverage
}

func (c Columns) Count() int { return int(c.count) }

// Values returns a copy, so modifying it cannot affect this or another batch.
func (c Columns) Values() [Dimension][MaxCandidates]float32 { return c.values }

func (c Columns) Attributes(index int) (Attributes, error) {
	if c.count == 0 || index < -1 || index >= int(c.count) {
		return Attributes{}, ErrInput
	}
	i := index + 1 // -1 denotes the query.
	return Attributes{c.direction[i], c.lower[i], c.upper[i], c.stop[i], c.coverage[i]}, nil
}

// AgreementOrder is an unlearned control: +1 per recognized match and -1 per
// recognized conflict. Ties retain current candidate display order. All live
// candidates remain available; the result does not deny or certify any action.
func (c Columns) AgreementOrder() ([MaxCandidates]int, [MaxCandidates]int, error) {
	var order, scores [MaxCandidates]int
	for i := range order {
		order[i] = -1
	}
	if c.count < 1 || c.count > MaxCandidates {
		return order, scores, ErrInput
	}
	for i := 0; i < int(c.count); i++ {
		order[i] = i
		for column := 0; column < Dimension; column += 2 {
			scores[i] += int(c.values[column][i] - c.values[column+1][i])
		}
		for j := i; j > 0 && scores[order[j]] > scores[order[j-1]]; j-- {
			order[j], order[j-1] = order[j-1], order[j]
		}
	}
	return order, scores, nil
}

// lexer borrows token substrings only while parsing. ASCII word runs have a
// 32-word limit before camel-case splitting; punctuation and split tokens have
// a separate 64-token limit. Valid non-ASCII prose is unsupported, not invalid
// UTF-8. Keeping it out of this narrow grammar avoids a translation heuristic.
type lexer struct {
	tokens [MaxTokens]string
	n      int
}

func lex(raw string) (lexer, error) {
	var out lexer
	if len(raw) > MaxBytes || !utf8.ValidString(raw) {
		return out, ErrInput
	}
	for i := range len(raw) {
		if raw[i] >= utf8.RuneSelf {
			return out, nil
		}
	}
	add := func(s string) bool {
		if out.n == MaxTokens {
			return false
		}
		out.tokens[out.n] = s
		out.n++
		return true
	}
	words := 0
	for i := 0; i < len(raw); {
		if raw[i] == ' ' || raw[i] == '\t' || raw[i] == '\r' || raw[i] == '\n' {
			i++
			continue
		}
		start := i
		if word(raw[i]) {
			words++
			if words > MaxWords {
				return lexer{}, ErrInput
			}
			i++
			for i < len(raw) && word(raw[i]) {
				if upper(raw[i]) && (lower(raw[i-1]) || (upper(raw[i-1]) && i+1 < len(raw) && lower(raw[i+1]))) {
					if !add(raw[start:i]) {
						return lexer{}, ErrInput
					}
					start = i
				}
				i++
			}
		} else {
			i++
			if (raw[start] == '<' || raw[start] == '>') && i < len(raw) && raw[i] == '=' {
				i++
			}
		}
		if !add(raw[start:i]) {
			return lexer{}, ErrInput
		}
	}
	return out, nil
}

func lower(c byte) bool { return c >= 'a' && c <= 'z' }
func upper(c byte) bool { return c >= 'A' && c <= 'Z' }
func word(c byte) bool  { return lower(c) || upper(c) || c >= '0' && c <= '9' }

type parser struct {
	lex lexer
	i   int
}

func (p *parser) take(s string) bool {
	if p.i < p.lex.n && strings.EqualFold(p.lex.tokens[p.i], s) {
		p.i++
		return true
	}
	return false
}
func (p *parser) phrase(words ...string) bool {
	for _, w := range words {
		if !p.take(w) {
			return false
		}
	}
	return true
}
func (p *parser) end() bool {
	p.take(".")
	return p.i == p.lex.n
}

// ParseRequest recognizes complete v1 English interval templates. An omitted
// bound means unbounded only inside a complete supported filter or explicit
// "all keys" template, never from arbitrary missing text. Empty-tree,
// negation, quotations, conditional prose, nested and contradictory filters
// are unsupported. Stop count is not encoded; CanStop is only a soft hint.
func ParseRequest(raw string) (Attributes, error) {
	l, err := lex(raw)
	if err != nil {
		return Attributes{}, err
	}
	p := parser{lex: l}
	if !p.take("return") {
		return Attributes{}, nil
	}
	all, stop := p.take("all"), false
	if !all && p.take("the") {
		stop = true
		if !p.phrase("first", "two", "qualifying") {
			return Attributes{}, nil
		}
	}
	if !p.phrase("keys", "in") {
		return Attributes{}, nil
	}
	var out Attributes
	if p.take("increasing") {
		out.Direction = Ascending
	} else if p.take("decreasing") {
		out.Direction = Descending
	} else {
		return Attributes{}, nil
	}
	if !p.take("order") {
		return Attributes{}, nil
	}
	out.Coverage = HasDirection | HasLower | HasUpper
	out.Lower, out.Upper = Unbounded, Unbounded
	if all {
		if !p.end() {
			return Attributes{}, nil
		}
		return out, nil
	}
	p.take(",")
	if !p.take("keeping") && !p.take("with") {
		return Attributes{}, nil
	}
	axis, bound, ok := p.predicate()
	if !ok {
		return Attributes{}, nil
	}
	if axis == HasLower {
		out.Lower = bound
	} else {
		out.Upper = bound
	}
	if p.take("and") {
		next, b, ok := p.predicate()
		if !ok || next == axis {
			return Attributes{}, nil
		}
		if next == HasLower {
			out.Lower = b
		} else {
			out.Upper = b
		}
	}
	if stop {
		if !p.phrase(";", "stop", "the", "callback", "after", "two") {
			return Attributes{}, nil
		}
		out.CanStop = true
		out.Coverage |= HasStop
	}
	if !p.end() {
		return Attributes{}, nil
	}
	return out, nil
}

func (p *parser) predicate() (Coverage, Bound, bool) {
	if p.i+3 > p.lex.n {
		return 0, BoundUnknown, false
	}
	lhs, op, rhs := p.lex.tokens[p.i], p.lex.tokens[p.i+1], p.lex.tokens[p.i+2]
	p.i += 3
	if strings.EqualFold(rhs, "x") {
		lhs, rhs = rhs, lhs
		switch op {
		case "<":
			op = ">"
		case "<=":
			op = ">="
		case ">":
			op = "<"
		case ">=":
			op = "<="
		default:
			return 0, BoundUnknown, false
		}
	}
	if !strings.EqualFold(lhs, "x") {
		return 0, BoundUnknown, false
	}
	if strings.EqualFold(rhs, "lower") {
		switch op {
		case ">":
			return HasLower, Open, true
		case ">=":
			return HasLower, Closed, true
		}
	} else if strings.EqualFold(rhs, "upper") {
		switch op {
		case "<":
			return HasUpper, Open, true
		case "<=":
			return HasUpper, Closed, true
		}
	}
	return 0, BoundUnknown, false
}

// ParseCandidate recognizes v1 traversal comments whose leading method is an
// Ascend/Descend word with one of the documented lexical suffixes. Bounds come
// from the actual interval in the text, not a lookup from method ID to label.
// Descending intervals reverse positional lower/upper roles. Included first/
// last extrema mean unbounded only in their correctly oriented positions.
// Unsupported endpoints keep direction/stop recognition but no bound claim.
func ParseCandidate(raw string) (Attributes, error) {
	l, err := lex(raw)
	if err != nil {
		return Attributes{}, err
	}
	p := parser{lex: l}
	var out Attributes
	if p.take("ascend") {
		out.Direction = Ascending
	} else if p.take("descend") {
		out.Direction = Descending
	} else {
		return Attributes{}, nil
	}
	if p.take("range") {
		// Lexical suffix; interval endpoints still provide all bounds.
	} else if p.take("less") {
		if !p.take("than") && !p.phrase("or", "equal") {
			return Attributes{}, nil
		}
	} else if p.take("greater") {
		if !p.take("than") && !p.phrase("or", "equal") {
			return Attributes{}, nil
		}
	}
	if !p.phrase("calls", "the", "iterator", "for", "every", "value", "in", "the", "tree", "within", "the", "range") {
		return Attributes{}, nil
	}
	leftClosed := p.take("[")
	if !leftClosed && !p.take("(") {
		return Attributes{}, nil
	}
	leftStart := p.i
	for p.i < p.lex.n && !strings.EqualFold(p.lex.tokens[p.i], ",") {
		p.i++
	}
	leftEnd := p.i
	if !p.take(",") {
		return Attributes{}, nil
	}
	rightStart := p.i
	for p.i < p.lex.n && p.lex.tokens[p.i] != ")" && p.lex.tokens[p.i] != "]" {
		p.i++
	}
	rightEnd := p.i
	// Partial recognition applies to an unknown endpoint word, not malformed
	// interval syntax such as an empty endpoint, nesting, or extra commas.
	if !endpointSyntax(l.tokens[leftStart:leftEnd]) || !endpointSyntax(l.tokens[rightStart:rightEnd]) {
		return Attributes{}, nil
	}
	rightClosed := p.take("]")
	if !rightClosed && !p.take(")") {
		return Attributes{}, nil
	}
	if !p.phrase(",", "until", "iterator", "returns", "false") || !p.end() {
		return Attributes{}, nil
	}
	out.Coverage, out.CanStop = HasDirection|HasStop, true
	leftAxis, rightAxis := HasLower, HasUpper
	if out.Direction == Descending {
		leftAxis, rightAxis = HasUpper, HasLower
	}
	left := endpoint(l.tokens[leftStart:leftEnd], leftAxis, leftClosed)
	right := endpoint(l.tokens[rightStart:rightEnd], rightAxis, rightClosed)
	assign := func(axis Coverage, bound Bound) {
		if bound == BoundUnknown {
			return
		}
		out.Coverage |= axis
		if axis == HasLower {
			out.Lower = bound
		} else {
			out.Upper = bound
		}
	}
	assign(leftAxis, left)
	assign(rightAxis, right)
	return out, nil
}

func endpointSyntax(tokens []string) bool {
	if len(tokens) < 1 || len(tokens) > 3 {
		return false
	}
	for _, token := range tokens {
		for i := range len(token) {
			if !word(token[i]) {
				return false
			}
		}
	}
	return true
}

func endpoint(tokens []string, axis Coverage, closed bool) Bound {
	eq := func(words ...string) bool {
		if len(tokens) != len(words) {
			return false
		}
		for i := range words {
			if !strings.EqualFold(tokens[i], words[i]) {
				return false
			}
		}
		return true
	}
	if closed && (axis == HasLower && eq("first") || axis == HasUpper && eq("last")) {
		return Unbounded
	}
	if eq("pivot") ||
		axis == HasLower && (closed && eq("greater", "or", "equal") || !closed && eq("greater", "than")) ||
		axis == HasUpper && (closed && eq("less", "or", "equal") || !closed && eq("less", "than")) {
		if closed {
			return Closed
		}
		return Open
	}
	return BoundUnknown
}
