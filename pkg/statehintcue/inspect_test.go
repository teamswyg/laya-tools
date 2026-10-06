package statehintcue

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// All fixtures are invented lexical cases, not private text or an independent
// question-intent accuracy evaluation.
func TestInspectOriginalLexicalFixtures(t *testing.T) {
	tests := []struct {
		name, text string
		want       bool
	}{
		{"english", "Is the fictional lantern ready?", true},
		{"fullwidth", "가상의 등불이 준비됐나요？", true},
		{"statement", "The fictional lantern is ready.", false},
		{"marker_only", "?", true},
		{"empty", "", false},
		{"inline_code", "Use `ready?` for the sample.", false},
		{"double_inline", "Use ``a ` ?`` for the sample.", false},
		{"code_then_prose", "After `ready?`, can we begin？", true},
		{"closing_code_adjacent_uri", "`sample`https://example.invalid/queue?limit=2", false},
		{"closing_code_adjacent_relative", "`sample`/lookup?limit=2", false},
		{"closing_code_uri_then_prose", "`sample`https://example.invalid/queue?limit=2 Ready?", true},
		{"closing_code_prose_question", "`sample`Ready?", true},
		{"fence", "```text\nCould this run?\n```", false},
		{"fence_different_inner_run", "```text\n```` ?\n```", false},
		{"fence_then_prose", "```\n?\n```\nMay the lantern glow?", true},
		{"http_query", "http://example.invalid/find?topic=lantern", false},
		{"https_query", "https://example.invalid/find?topic=lantern", false},
		{"ws_query", "ws://example.invalid/feed?topic=lantern", false},
		{"wss_query", "wss://example.invalid/feed?topic=lantern", false},
		{"uppercase_scheme", "HTTPS://example.invalid/find?topic=lantern", false},
		{"relative_query", "/search?topic=lantern", false},
		{"quoted_relative", "See (/search?topic=lantern).", false},
		{"assigned_uri", "Address=https://example.invalid/find?topic=lantern", false},
		{"uri_inside_code", "Use `https://example.invalid/find?topic=lantern`.", false},
		{"uri_then_word_question", "Can https://example.invalid/find?topic=lantern answer?", true},
		{"uri_then_standalone_marker", "https://example.invalid/find?topic=lantern ？", true},
		{"unicode_space_boundary", "https://example.invalid/find?topic=lantern\u2003Ready?", true},
		{"ambiguous_uri_ending", "Is this https://example.invalid?", false},
		{"relative_and_prose", "At /search?topic=lantern, is it ready?", true},
		{"prose_before_uri", "Ready? https://example.invalid/find?topic=lantern", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Inspect(context.Background(), tt.text, Prose)
			if err != nil || got.QuestionPunctuation != tt.want || got.Guard != GuardNone {
				t.Fatalf("got cue=%v guard=%q error=%v; want cue=%v", got.QuestionPunctuation, got.Guard, err, tt.want)
			}
			assertObservationOnly(t, got)
		})
	}
}

func TestInspectRoleIsExplicitAndDoesNotProveOwnership(t *testing.T) {
	for _, role := range []Role{Metadata, Code} {
		got, err := Inspect(context.Background(), "Could the fictional lantern run? `", role)
		if err != nil || got.QuestionPunctuation || got.Guard != GuardRoleNotProse {
			t.Fatalf("excluded role %q: result=%+v error=%v", role, got, err)
		}
		assertObservationOnly(t, got)
	}
	for _, role := range []Role{"", "Prose", "private-role-marker"} {
		got, err := Inspect(context.Background(), "Ready?", role)
		if !errors.Is(err, ErrRole) || got.Guard != GuardUnsupportedRole || got.QuestionPunctuation {
			t.Fatalf("unsupported role: result=%+v error=%v", got, err)
		}
		assertObservationOnly(t, got)
		if strings.Contains(err.Error(), string(role)) && role != "" {
			t.Fatal("error echoed caller role")
		}
	}
}

func TestInspectDiscardsCandidateOnUnbalancedTail(t *testing.T) {
	for _, text := range []string{
		"Ready? `unfinished",
		"Ready？ ```\nunfinished",
		"Ready? ``a `?` b",
		"Ready? https://example.invalid/path`unfinished",
		"Ready? \\`literal escape is not parsed",
	} {
		got, err := Inspect(context.Background(), text, Prose)
		if !errors.Is(err, ErrUnbalancedCode) || got.Guard != GuardUnbalancedCode || got.QuestionPunctuation {
			t.Fatalf("unbalanced delimiter returned candidate: result=%+v error=%v", got, err)
		}
		assertObservationOnly(t, got)
	}
}

func TestInspectBoundsAndInvalidTailsApplyToEveryRole(t *testing.T) {
	valid := strings.Repeat("가", (MaxTextBytes-1)/3) + "?"
	if len(valid) != MaxTextBytes {
		t.Fatal("fixture does not meet exact UTF-8 byte boundary")
	}
	for _, role := range []Role{Prose, Metadata, Code} {
		got, err := Inspect(context.Background(), valid, role)
		if err != nil || got.QuestionPunctuation != (role == Prose) {
			t.Fatalf("exact boundary role=%q: result=%+v error=%v", role, got, err)
		}
		for _, bad := range []struct {
			text  string
			err   error
			guard string
		}{
			{valid + "x", ErrTextLimit, GuardTextLimit},
			{"Ready?\x00tail", ErrInput, GuardInvalidInput},
			{"Ready?\xff", ErrInput, GuardInvalidInput},
			{"Ready?\xe2\x82", ErrInput, GuardInvalidInput},
		} {
			got, err = Inspect(context.Background(), bad.text, role)
			if !errors.Is(err, bad.err) || got.Guard != bad.guard || got.QuestionPunctuation {
				t.Fatalf("invalid tail role=%q: result=%+v error=%v", role, got, err)
			}
			assertObservationOnly(t, got)
		}
	}
	got, err := Inspect(nil, "Ready?", Prose)
	if !errors.Is(err, ErrInput) || got.Guard != GuardInvalidInput || got.QuestionPunctuation {
		t.Fatalf("nil context: result=%+v error=%v", got, err)
	}
}

// This synthetic context cancels after a fixed number of Err checks, making
// tail cancellation deterministic without scheduling or sleeps.
type cancellationContext struct {
	context.Context
	cancelAt, checks int
}

func (c *cancellationContext) Err() error {
	c.checks++
	if c.checks >= c.cancelAt {
		return context.Canceled
	}
	return nil
}

func TestInspectCancellationBeforeAndAfterCandidate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, role := range []Role{Prose, Metadata, Code} {
		got, err := Inspect(ctx, "Ready?", role)
		if !errors.Is(err, context.Canceled) || got.Guard != GuardCancelled || got.QuestionPunctuation {
			t.Fatalf("cancelled role=%q: result=%+v error=%v", role, got, err)
		}
	}
	for _, text := range []string{
		"?" + strings.Repeat(" fictional tail", 20),
		"? `" + strings.Repeat("a", 100) + "`",
		"? https://example.invalid/" + strings.Repeat("a", 100),
		"? " + strings.Repeat("`", 100),
	} {
		ctx := &cancellationContext{Context: context.Background(), cancelAt: 10}
		got, err := Inspect(ctx, text, Prose)
		if !errors.Is(err, context.Canceled) || got.Guard != GuardCancelled || got.QuestionPunctuation {
			t.Fatalf("tail cancellation returned candidate: result=%+v error=%v", got, err)
		}
		assertObservationOnly(t, got)
	}
	ctxDeadline, cancelDeadline := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancelDeadline()
	got, err := Inspect(ctxDeadline, "Ready?", Prose)
	if !errors.Is(err, context.DeadlineExceeded) || got.Guard != GuardCancelled || got.QuestionPunctuation {
		t.Fatalf("deadline: result=%+v error=%v", got, err)
	}
}

func TestInspectSerializationDoesNotEchoInput(t *testing.T) {
	const secret = "fictional-private-marker"
	for _, text := range []string{
		secret + "? https://example.invalid/" + secret + "?token=sample",
		secret + "? `unfinished",
		secret + "?\x00",
	} {
		got, err := Inspect(context.Background(), text, Prose)
		b, marshalErr := json.Marshal(got)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		for _, forbidden := range []string{secret, "example.invalid", "token=", "http", "text", "hash", "confidence", "probabilities"} {
			if strings.Contains(string(b), forbidden) {
				t.Fatalf("result serialization included forbidden data category %q", forbidden)
			}
		}
		if err != nil && strings.Contains(err.Error(), secret) {
			t.Fatal("error echoed input")
		}
		assertObservationOnly(t, got)
	}
}

func TestInspectConcurrentCallersHaveNoSharedState(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			text, want := "Does the fictional lantern glow?", true
			if i%2 == 0 {
				text, want = "https://example.invalid/find?topic=lantern", false
			}
			for range 20 {
				got, err := Inspect(context.Background(), text, Prose)
				if err != nil || got.QuestionPunctuation != want {
					t.Errorf("concurrent observation mismatch: result=%+v error=%v", got, err)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func assertObservationOnly(t *testing.T, got Result) {
	t.Helper()
	if got.Source != Source || got.ModelUsed || got.ActualVerified || got.StateChangeProposed || got.MutationExecuted {
		t.Fatalf("unexpected model or authority claim: %+v", got)
	}
}

var benchmarkResult Result

func BenchmarkInspectOriginalBoundedProse(b *testing.B) {
	text := strings.Repeat("A fictional lantern is described. `sample?` ", 70) + "Is it ready?"
	if len(text) > MaxTextBytes {
		b.Fatal("oversize benchmark fixture")
	}
	ctx := context.Background()
	b.ReportAllocs()
	b.SetBytes(int64(len(text)))
	for b.Loop() {
		got, err := Inspect(ctx, text, Prose)
		if err != nil || !got.QuestionPunctuation {
			b.Fatal("unexpected benchmark observation")
		}
		benchmarkResult = got
	}
}
