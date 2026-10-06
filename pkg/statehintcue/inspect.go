// Package statehintcue observes bounded question punctuation in explicitly
// declared prose. It is an experimental lexical cue, not a question classifier.
package statehintcue

import (
	"context"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Role string

const (
	Prose    Role = "prose"
	Metadata Role = "metadata"
	Code     Role = "code"

	MaxTextBytes = 4096
	Source       = "unlearned_question_punctuation_v1"

	GuardNone            = ""
	GuardRoleNotProse    = "role_not_prose"
	GuardInvalidInput    = "invalid_input"
	GuardUnsupportedRole = "unsupported_role"
	GuardTextLimit       = "text_limit_exceeded"
	GuardUnbalancedCode  = "unbalanced_code"
	GuardCancelled       = "context_cancelled"
)

var (
	ErrInput          = errors.New("statehintcue: invalid input")
	ErrRole           = errors.New("statehintcue: explicit supported role required")
	ErrTextLimit      = errors.New("statehintcue: text byte limit exceeded")
	ErrUnbalancedCode = errors.New("statehintcue: unbalanced backtick delimiters")
)

// Result contains only fixed-source punctuation observations. It contains no
// input, hashes, probability, semantic intent or authority evidence.
type Result struct {
	QuestionPunctuation bool   `json:"question_punctuation"`
	Guard               string `json:"guard"`
	Source              string `json:"source"`
	ModelUsed           bool   `json:"model_used"`
	ActualVerified      bool   `json:"actual_verified"`
	StateChangeProposed bool   `json:"state_change_proposed"`
	MutationExecuted    bool   `json:"mutation_executed"`
}

func guarded(ctx context.Context, guard string, err error) (Result, error) {
	if ctx != nil {
		if cancelled := ctx.Err(); cancelled != nil {
			return Result{Source: Source, Guard: GuardCancelled}, cancelled
		}
	}
	return Result{Source: Source, Guard: guard}, err
}

// Inspect checks the complete bounded input before returning any cue. Metadata
// and code roles are explicitly excluded. In prose, equal-length literal
// backtick runs delimit code; unmatched runs reject the entire observation.
// This deliberately small lexical contract is not a Markdown/URI parser.
func Inspect(ctx context.Context, text string, role Role) (Result, error) {
	if ctx == nil {
		return guarded(nil, GuardInvalidInput, ErrInput)
	}
	if err := ctx.Err(); err != nil {
		return guarded(ctx, GuardCancelled, err)
	}
	if role != Prose && role != Metadata && role != Code {
		return guarded(ctx, GuardUnsupportedRole, ErrRole)
	}
	if len(text) > MaxTextBytes {
		return guarded(ctx, GuardTextLimit, ErrTextLimit)
	}
	if !utf8.ValidString(text) || strings.IndexByte(text, 0) >= 0 {
		return guarded(ctx, GuardInvalidInput, ErrInput)
	}
	if err := ctx.Err(); err != nil {
		return guarded(ctx, GuardCancelled, err)
	}
	if role != Prose {
		return guarded(ctx, GuardRoleNotProse, nil)
	}

	found, tokenStart, inURI, ticks := false, true, false, 0
	for i := 0; i < len(text); {
		if err := ctx.Err(); err != nil {
			return guarded(ctx, GuardCancelled, err)
		}
		r, size := utf8.DecodeRuneInString(text[i:])
		if ticks != 0 {
			if r == '`' {
				end, err := backtickEnd(ctx, text, i)
				if err != nil {
					return guarded(ctx, GuardCancelled, err)
				}
				if end-i == ticks {
					ticks = 0
					tokenStart = true
				}
				i = end
			} else {
				i += size
			}
			continue
		}
		if inURI {
			// Adjoining punctuation stays in a URI span conservatively. A
			// backtick starts a separately checked code delimiter instead.
			if !unicode.IsSpace(r) && r != '`' {
				i += size
				continue
			}
			inURI = false
		}
		if r == '`' {
			end, err := backtickEnd(ctx, text, i)
			if err != nil {
				return guarded(ctx, GuardCancelled, err)
			}
			ticks, tokenStart, i = end-i, false, end
			continue
		}
		if tokenStart && startsURI(text[i:]) {
			inURI, tokenStart = true, false
			continue
		}
		if r == '?' || r == '？' {
			found = true
		}
		tokenStart = tokenBoundary(r)
		i += size
	}
	if ticks != 0 {
		return guarded(ctx, GuardUnbalancedCode, ErrUnbalancedCode)
	}
	if err := ctx.Err(); err != nil {
		return guarded(ctx, GuardCancelled, err)
	}
	return Result{Source: Source, QuestionPunctuation: found}, nil
}

func backtickEnd(ctx context.Context, text string, start int) (int, error) {
	end := start
	for end < len(text) && text[end] == '`' {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		end++
	}
	return end, nil
}

func tokenBoundary(r rune) bool {
	if unicode.IsSpace(r) {
		return true
	}
	switch r {
	case '(', '[', '{', '<', '\'', '"', ':', '=':
		return true
	default:
		return false
	}
}

func startsURI(text string) bool {
	return text[0] == '/' || foldPrefix(text, "http://") ||
		foldPrefix(text, "https://") || foldPrefix(text, "ws://") ||
		foldPrefix(text, "wss://")
}

func foldPrefix(text, prefix string) bool {
	if len(text) < len(prefix) {
		return false
	}
	for i := range len(prefix) {
		b := text[i]
		if b >= 'A' && b <= 'Z' {
			b += 'a' - 'A'
		}
		if b != prefix[i] {
			return false
		}
	}
	return true
}
