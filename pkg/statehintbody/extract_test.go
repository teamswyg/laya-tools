package statehintbody

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestSupportedRepresentations(t *testing.T) {
	tests := []struct {
		name, body, want string
		format           Format
	}{
		{"plain_exact", " &lt;br&gt;<br>\r\n", " &lt;br&gt;<br>\r\n", PlainText},
		{"plain_numeric_literal", "&#xD800;", "&#xD800;", PlainText},
		{"br", "A<br>B", "A\nB", NeutralHTML},
		{"br_self_closed", "A<br/>B", "A\nB", NeutralHTML},
		{"consecutive_br", "<br>A<br><br>B", "\nA\n\nB", NeutralHTML},
		{"decode_once", "&amp;lt;br&amp;gt; &amp;nbsp;", "&lt;br&gt; &nbsp;", NeutralHTML},
		{"escaped_tag", "<p>&lt;br&gt; &amp; &#x1F6E0; &nbsp;</p>", "<br> & 🛠 \u00a0\n", NeutralHTML},
		{"paragraph_negation", "<p>완료한</p><p>것이 아닙니다.</p>", "완료한\n것이 아닙니다.\n", NeutralHTML},
		{"neutral_markup", "<div><p><strong>Still</strong> <em>checking</em> <a href=\"https://example.com/\">results</a>.</p></div>", "Still checking results.\n", NeutralHTML},
		{"mention_visible_text", "<p><span data-type=\"mention\" data-id=\"fictional-member\" data-label=\"Example\">@Example</span> 확인 중입니다.</p>", "@Example 확인 중입니다.\n", NeutralHTML},
	}
	var w Workspace
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Extract(context.Background(), tc.body, tc.format, &w)
			if err != nil {
				t.Fatal(err)
			}
			if result.Text != tc.want {
				t.Fatalf("text = %q, want %q", result.Text, tc.want)
			}
			if result.RawBodySHA256 != digest(tc.body) || result.ModelTextSHA256 != digest(tc.want) {
				t.Fatal("representation hashes must match their exact bytes")
			}
			if result.ExtractorVersion != ExtractorVersion || result.RawBytes != len(tc.body) || result.TextBytes != len(tc.want) || result.OwnerRevisionVerified || result.RenderedVisibilityVerified {
				t.Fatal("invalid provenance or verification claim")
			}
			for _, b := range w.text {
				if b != 0 {
					t.Fatal("caller scratch retained derived text")
				}
			}
		})
	}
}

func TestRejectWholeUnsupportedBody(t *testing.T) {
	for _, body := range []string{
		"<p>visible prefix</p><script>hidden completion</script>",
		"<p>visible</p><img alt=\"required context\"/>",
		"<p hidden=\"hidden\">negation</p>", "<p style=\"display:none\">negation</p>", "<p class=\"hidden\">negation</p>",
		"<blockquote>completed</blockquote>", "<code>completed</code>", "<del>completed</del>",
		"<![CDATA[A<br>B]]>", "<!--A<br>B--><p>visible</p>", "<!DOCTYPE p><p>visible</p>", "<?xml version=\"1.0\"?><p>visible</p>",
		"<p title=\"first\" title=\"second\">visible</p>", "<p xmlns=\"urn:fictional\">visible</p>", "<x:p>visible</x:p>", "<p x:title=\"value\">visible</p>",
		"<p>A<div>B</div></p>", "<span><p>block in inline</p></span>", "<p><p>nested</p></p>", "<a><a>nested link</a></a>",
		"<p title=unquoted>visible</p>", "<p title=\"<br>\">visible</p>", "<p>A</p>tail<b", "<p>not closed", "<p>mismatch</div>",
		"<a href=\"x\"title=\"y\">visible</a>", "<span data-type=\"mention\"data-id=\"fictional\">visible</span>",
		"A<", "<p>A</p><", "A</",
		"<br>child</br>", "<br></br>", "<br title=\"break\"/>", "<span/>",
		"<span data-id=\"fictional\">missing explicit mention kind</span>", "<span data-type=\"unknown\">visible</span>",
		"<p>&unknown;</p>", "<p>&#0;</p>", "<p>&#xD800;</p>", "<p>&#x110000;</p>", "<p>&#xFFFF;</p>", "<p>&#1;</p>", "<p>&#-1;</p>",
	} {
		t.Run(body, func(t *testing.T) {
			var w Workspace
			result, err := Extract(context.Background(), body, NeutralHTML, &w)
			if err == nil || result != (Result{}) {
				t.Fatal("unsupported body returned partial or successful text")
			}
			if strings.Contains(err.Error(), "visible") || strings.Contains(err.Error(), "fictional") {
				t.Fatal("error echoed source content")
			}
		})
	}
}

func TestExactBudgetsAndCancellation(t *testing.T) {
	var w Workspace
	if _, err := Extract(context.Background(), strings.Repeat("a", MaxTextBytes), PlainText, &w); err != nil {
		t.Fatal(err)
	}
	if result, err := Extract(context.Background(), strings.Repeat("a", MaxTextBytes+1), PlainText, &w); !errors.Is(err, ErrTextLimit) || result != (Result{}) {
		t.Fatal("oversized model text accepted")
	}
	if _, err := Extract(context.Background(), "<p>"+strings.Repeat("a", MaxTextBytes-1)+"</p>", NeutralHTML, &w); err != nil {
		t.Fatal(err)
	}
	if result, err := Extract(context.Background(), "<p>"+strings.Repeat("a", MaxTextBytes)+"</p>", NeutralHTML, &w); !errors.Is(err, ErrTextLimit) || result != (Result{}) {
		t.Fatal("inserted boundary escaped text budget")
	}
	stem := "<span title=\"\">x</span>"
	exactRaw := "<span title=\"" + strings.Repeat("a", MaxRawBytes-len(stem)) + "\">x</span>"
	if len(exactRaw) != MaxRawBytes {
		t.Fatal("test budget construction")
	}
	if _, err := Extract(context.Background(), exactRaw, NeutralHTML, &w); err != nil {
		t.Fatal(err)
	}
	if _, err := Extract(context.Background(), exactRaw+" ", NeutralHTML, &w); !errors.Is(err, ErrRawLimit) {
		t.Fatal("raw limit escaped")
	}
	if _, err := Extract(context.Background(), strings.Repeat("<span>", MaxDepth)+"x"+strings.Repeat("</span>", MaxDepth), NeutralHTML, &w); err != nil {
		t.Fatal(err)
	}
	if result, err := Extract(context.Background(), strings.Repeat("<span>", MaxDepth+1)+"x"+strings.Repeat("</span>", MaxDepth+1), NeutralHTML, &w); !errors.Is(err, ErrBudget) || result != (Result{}) {
		t.Fatal("depth escaped budget")
	}
	if _, err := Extract(context.Background(), strings.Repeat("<span></span>", MaxTokens/2), NeutralHTML, &w); err != nil {
		t.Fatal(err)
	}
	if result, err := Extract(context.Background(), strings.Repeat("<span></span>", MaxTokens/2+1), NeutralHTML, &w); !errors.Is(err, ErrBudget) || result != (Result{}) {
		t.Fatal("token budget escaped")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if result, err := Extract(ctx, "private fictional text", PlainText, &w); !errors.Is(err, context.Canceled) || result != (Result{}) {
		t.Fatal("cancelled extraction returned content")
	}
}

func TestExplicitFormatAndPrivateJSON(t *testing.T) {
	var w Workspace
	for _, format := range []Format{"", "html", "editor_json"} {
		if result, err := Extract(context.Background(), "private fictional text", format, &w); !errors.Is(err, ErrFormat) || result != (Result{}) {
			t.Fatal("format guessed")
		}
	}
	result, err := Extract(context.Background(), "private fictional text", PlainText, &w)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{result.Text, result.RawBodySHA256, result.ModelTextSHA256} {
		if strings.Contains(string(encoded), value) {
			t.Fatal("default JSON exposes private text or fingerprint")
		}
	}
	for _, invalid := range []string{"\xff", "a\x00b"} {
		if result, err := Extract(context.Background(), invalid, PlainText, &w); !errors.Is(err, ErrInput) || result != (Result{}) {
			t.Fatal("invalid input accepted")
		}
	}
}

func FuzzNoPartialExtraction(f *testing.F) {
	for _, seed := range []string{"<p>Example &amp; content.</p>", "<p>prefix</p><script>tail</script>", "<![CDATA[A<br>B]]>", "&#xD800;", "<p x:title='x'>x</p>", "A<br><br>B"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > MaxRawBytes+1 {
			t.Skip()
		}
		var w Workspace
		result, err := Extract(context.Background(), raw, NeutralHTML, &w)
		if err != nil {
			if result != (Result{}) {
				t.Fatal("partial result")
			}
			return
		}
		if len(result.Text) > MaxTextBytes || result.OwnerRevisionVerified || result.RenderedVisibilityVerified || result.RawBodySHA256 != digest(raw) || result.ModelTextSHA256 != digest(result.Text) {
			t.Fatal("unbounded or unqualified successful result")
		}
	})
}

func BenchmarkNeutralFragment(b *testing.B) {
	raw := "<div><p><span data-type=\"mention\" data-id=\"fictional-member\">@Example</span> I am checking the build results.</p><p>It is not complete yet.</p></div>"
	var w Workspace
	b.ReportAllocs()
	b.SetBytes(int64(len(raw)))
	for b.Loop() {
		if _, err := Extract(context.Background(), raw, NeutralHTML, &w); err != nil {
			b.Fatal(err)
		}
	}
}
