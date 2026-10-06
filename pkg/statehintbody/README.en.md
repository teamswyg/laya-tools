# Deriving bounded model text

[한국어](README.ko.md)

Derive input for the small Go classifier from a body already held by an authenticated application. The caller must explicitly declare `plain_text` or `neutral_html_fragment`. Unknown formats produce no model input. Actual application format, permission and deployment contracts require separate verification.

```go
var bodyWork statehintbody.Workspace
result, err := statehintbody.Extract(ctx, heldBody, statehintbody.NeutralHTML, &bodyWork)
if err != nil {
    // Defer this complete body; do not predict from a truncated remainder.
    return err
}
var modelWork statehint.Workspace
prediction, err := localModel.Predict(result.Text, &modelWork)
```

The local model and body format are supplied explicitly. Extraction performs no download, body fetch, training or state write. A prediction is a textual hypothesis; it does not replace current catalog, revision and permission checks.

| Boundary | Limit |
|---|---:|
| Raw body | 16 KiB |
| Model text | 4,096 bytes including inserted newlines |
| Tag depth | 32 |
| Parser tokens | 2,048 |
| Attributes per tag | 8 |

`plain_text` preserves exact UTF-8 bytes. The neutral fragment subset supports only `p/div/br/span/strong/em/b/i/u/a`. It separates blocks with newlines and preserves every `br` as one newline. Non-void tags require balanced explicit closing tags. Block-inside-inline and nested-link combinations that browsers may restructure are rejected.

The five built-in XML entities, `&nbsp;`, and valid numeric references are decoded once. Before parsing, only exact `<br>` is rewritten to `<br/>` and `&nbsp;` to a numeric reference. Escaped `&lt;br&gt;` and `&amp;nbsp;` are not interpreted again. The raw hash is computed from bytes before these rewrites.

Allowed attributes are `title`, link `href`, and mention span `data-type="mention"` with `data-id/data-label/data-mention-type`. Model text uses inner span text; identifiers, link destinations and attribute labels are not appended. Class/style/hidden, scripts/images/iframes, quotation/code/deletion markup, namespaces, duplicate attributes, CDATA/comments/declarations/processing instructions, and other tags or attributes defer the entire body. Unsupported context is never removed to make the remainder appear complete.

The Go standard [`encoding/xml`](https://pkg.go.dev/encoding/xml#Decoder) decoder runs in strict mode, with additional namespace, structure, duplicate-attribute, numeric-scalar and complete-consumption checks. This is a narrow supported fragment contract, not general HTML parsing or browser-rendering equivalence.

`RawBodySHA256`, `ModelTextSHA256` and `ExtractorVersion` identify different derived representations. They do not establish owner revisions, authorization, atomic snapshots or actual UI equivalence. `OwnerRevisionVerified` and `RenderedVisibilityVerified` remain false. Text and both hashes are omitted from default JSON. Actual work evidence stays private.

Workspace is a caller-owned fixed array; do not share one across concurrent calls. There is no shared mutable cache or lock. Scratch text is cleared before return; the caller controls the returned Text's lifetime and private handling. Errors do not echo source content or return partial results.

[Local costs for one original fictional 149-byte body](COSTS.development.json) include only extraction and provenance hashing. Removing the duplicate decoder stack and temporary hash buffers changed allocation from 2,248 to 1,544 bytes and from 43 to 30 allocations per call. Single-condition time observations were 2.016→1.886µs. This excludes classifier, JSON, owner reads, total RSS and GPU; it is not an operational speedup claim. Raw pprof stays local.

Public tests use original fictional text and malformed inputs to check structure, budgets, cancellation and private-output boundaries. They establish neither actual Riido format coverage, model quality nor live-service permissions.
