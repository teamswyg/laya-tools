# Connect agents and Go applications

[한국어](https://github.com/teamswyg/laya-tools/wiki/Agents-and-Go-KO) · [Home](https://github.com/teamswyg/laya-tools/wiki)

Choose the smallest interface you need. No registration or background service is required for CLI use.

## One request: JSON

```sh
riidolaya repo-preview --catalog examples/repositories/catalog.json --json 'refund invoices'
```

Parse JSON rather than human text. Check preview/status/reason, not just exit code. Keep stderr separate from stdout. Catalogs and real requests can contain private data: keep your own logs private.

## Several requests: persistent JSONL

Search/model recommendations:

```sh
riidolaya serve --root /path/to/repository
```

```json
{"id":1,"op":"search","query":"redirect authorization","lexical":true}
{"id":2,"op":"route","query":"Fix a typo"}
```

Repository preview uses its own process/catalog:

```sh
riidolaya repo-serve --catalog examples/repositories/catalog.json --laya
```

```json
{"id":1,"query":"Fix login sessions"}
{"id":2,"query":"refund invoices"}
```

Send one JSON object per line; receive one response per line with the request ID. Repository responses wrap the preview in `result`, or report `error`. Drop `--laya` for model-free operation. The repo process retains its index and loads the model lazily when eligible. Separate processes can each consume model RAM; they do not share weights in this implementation. End stdin or stop your child process when done.

`serve` does not accept `plan` or repo-preview operations; do not mix protocols. Search rebuilds code candidates for each request even when its model is warm.

## Optional MCP

```sh
riidolaya mcp --root /path/to/repository
```

Tools are `search_code` and `route_model`. Repository preview and planner are not currently MCP tools. Optional Codex registration:

```sh
codex mcp add riidolaya -- /absolute/path/to/riidolaya mcp --root /absolute/path/to/repository
```

This does not switch the current Codex model or relax its permission settings. Tool recommendations are not permission to execute work.

## Import the public Go packages

The repository has one Go module. In an existing Go module:

```sh
go get github.com/teamswyg/laya-tools@v0.3.0
```

A model-free repository preview:

```go
package main

import (
    "fmt"
    "log"
    "github.com/teamswyg/laya-tools/pkg/reporouter"
)

func main() {
    idx, err := reporouter.New([]reporouter.Repository{
        {Name: "example/billing", Summary: "Invoices and refunds"},
    })
    if err != nil { log.Fatal(err) }
    result, err := idx.Preview("refund invoices", reporouter.DefaultConfig(), nil)
    if err != nil { log.Fatal(err) }
    fmt.Println(result.Status, result.Suggested)
}
```

Expected: `candidate example/billing`. This is not a work assignment.

| Package | Your application supplies |
|---|---|
| `pkg/reporouter` | Access-filtered catalog; optional judge adapter |
| `pkg/catalog` | Capability/quality/prices and budget usage snapshot |
| `pkg/switchpolicy` | Current/target profiles, confidence, context, switch history |
| `pkg/planner` | Catalog, policy and request; inspect recommend/hold/blocked |

These packages use no CGO/Python/model files. Native inference remains internal. Your application owns authorization, state persistence, atomic budget reservations, execution, and outcome evaluation. Riido-daemon is not already connected.

[Policy reference](https://github.com/teamswyg/laya-tools/blob/main/docs/ecosystem.en.md) · [Repository reference](https://github.com/teamswyg/laya-tools/blob/main/docs/repository-routing-preview.en.md)
