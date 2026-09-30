# Routing policies adapted from the Laya ecosystem

[한국어](ecosystem.ko.md) · English

[laya.tools](https://laya.tools/) is an independent community directory by Nielogiczny. It helped us discover projects; we are not an official product of the directory or Laya's creators.

On 2026-09-30 we reviewed its [routing](https://laya.tools/laya-for-routing), [coding tools](https://laya.tools/laya-for-coding-tools), and [search/RAG](https://laya.tools/laya-for-search-and-rag) lists and the following public sources/licenses. Upstream performance claims are not our measurements.

| Project / pinned revision | What we examined | Use here |
|---|---|---|
| [system-one-router](https://github.com/mmornati/system-one-router/tree/a437d00bca33a4c10038b37a5efe04dc0e3d40bb), Apache-2.0 | Go gateway: capability, estimated prices, context, tools, locality, load, budget | Selection structure from `internal/router/score.go` adapted into `pkg/catalog` |
| [pi-pignon](https://github.com/siiick/pi-pignon/tree/4d97d1a35134b813bac6a8e345a1bf9b0bbf0bdb), MIT | TypeScript Pi extension: asymmetric confidence, cooldown, cache payback | Portions of `src/policy.ts` and its formula ported into `pkg/switchpolicy` |
| [laya_router](https://github.com/glukicov/laya_router/tree/a2278d9690676a6232098f1c0cbaf44e9f727e95), Apache-2.0 | Tier-specific evaluation and wording sensitivity | Evaluation reference: measure incorrect downgrades separately from aggregate accuracy; no code/data copied |
| [laya-codex](https://github.com/pilotspace/laya-codex/tree/580bc73c2ed95fd319db93ef725f30bf35047428), Apache-2.0 | Rust retrieval; RRF in `crates/laya-rank/src/fusion.rs` | Future retrieval comparison; no implementation copied or search algorithm changed here |
| [keel](https://github.com/codejunkie99/keel/tree/5cef4569e7f850b3a85dc998cc64c64e944c80a0), MIT | `crates/engine/src/jev_routing.rs`: eligible candidates before new-session routing | Boundary-design reference; no app/runtime code copied |
| [fast-laya-compaction](https://github.com/ShinyDataTech/fast-laya-compaction) | Compaction candidate; GitHub metadata did not establish a license | Not ported; lossless/savings claims not verified |

## What changed in the adaptation

`catalog` checks quality, tools, vision, locality, and context before selecting the lowest estimated cost adjusted for load. Cost uses input and estimated output tokens. Ties prefer quality, then ID for reproducibility.

Instead of system-one-router's topic-weighted capability, this version uses a single operator-supplied quality score, not measured performance. Unlike its best-available fallback, we abstain when no model meets the floor. Uncertainty applies the highest quality floor. Both input and expected output count against context capacity.

Budgets compare **existing spend + this request's estimated cost**. Missing spend for a configured budget excludes that model. Missing prices are not free: omitted/null `price` is unknown; `{"input":0,"output":0}` explicitly declares zero prices. Load influences ranking, not the amount charged against budget.

`switchpolicy` ports pi-pignon's formula. Prices use the same USD-per-million-token units, which cancel in the ratio:

```text
cache premium = context tokens × max(0, max(target input, target cache write) - target cache read)
recurring savings = context tokens × (current cache read - target cache read)
                  + expected output tokens × (current output - target output)
payback requests = cache premium / recurring savings
```

Missing prices/history hold downward and lateral switches. We did not port the upstream fallback allowing switches with unknown prices in small contexts. Lateral changes also need high confidence and evidence of savings. Capability upgrades can bypass cooldown/cache economics after their confidence gate, but cannot bypass the planner's preceding quality/budget constraints.

A new session (no `current`, `context_tokens: 0`) has no cache to lose and checks entry confidence. Unknown current models in active contexts hold. `pinned` prevents switching but cannot authorize keeping a model that violates required constraints.

## Public Go packages and responsibilities

One root `go.mod` contains three policy packages importable by other programs. They use only the standard library; no CGO, Python, or weights are needed.

- `github.com/teamswyg/laya-tools/pkg/catalog`: constraints and estimated cost.
- `github.com/teamswyg/laya-tools/pkg/switchpolicy`: switch guards and payback.
- `github.com/teamswyg/laya-tools/pkg/planner`: combined pre-execution plan.

```go
// cfg and req are your planner.Config / planner.Request.
plan, err := planner.Build(cfg, req)
if err != nil {
    return err
}
// plan.Status: recommend / hold / blocked
// plan.RecommendedModel: recommended or eligible current model; empty if blocked.
// plan.Selection.Candidates: rejection reasons and cost estimates.
```

`recommend` proposes a candidate, `hold` means the current model still meets constraints, and `blocked` means no permissible recommendation. None executes anything. A valid blocked request exits successfully; invalid input exits nonzero. Agents must inspect `plan.status`, not just the process exit code.

`plan` can use supplied assessment values. A trailing task replaces tier, confidence, and uncertainty using local Laya. Token estimates, budget usage, and capability requirements remain caller inputs. Korean, inference failures, and uncertainty retain conservative behavior; abstention is never reinterpreted as confidence.

The output's optional `classification` distinguishes model-assisted plans. This policy is not automatically wired into `codex`; `serve`/MCP still expose existing search/routing. riido-daemon can import the packages and add persistence/execution adapters.

## Examples

Run from the repository root. Model names, prices, quality, and confidence are **fictional examples**, not vendor prices or measured accuracy.

```sh
go build -o bin/riidolaya ./cmd/riidolaya
./bin/riidolaya plan --config examples/planner/config.json --request examples/planner/request.json --json
cat examples/planner/request.json | ./bin/riidolaya plan --config examples/planner/config.json --request - --json
./bin/riidolaya plan --config examples/planner/config.json --request examples/planner/request.json --json 'Fix a spelling mistake in a comment'
```

The example estimates payback at about `0.1385` requests when moving from example-strong to example-fast. This is arithmetic over fictional inputs, not observed savings. Setting `prompts_since_switch` to 1 causes a cooldown hold. Setting `assessment.local_only` to true blocks all example remote models.

IDs must uniquely identify a provider/model/reasoning execution profile; distinct reasoning settings need distinct IDs. `rank` is the operator's capability ordering, not an official vendor ranking.

## Measurements and limits

Single local microbenchmark run, M4 Pro, Go 1.27, 2026-09-30:

| Function | Fixture | ns/op | B/op | allocs/op |
|---|---|---:|---:|---:|
| catalog.Select | 2 fictional models | 132.0 | 296 | 6 |
| switchpolicy.Decide | 1 fictional switch | 28.92 | 8 | 1 |
| planner.Build | 3 example models | 246.9 | 432 | 9 |

```sh
CGO_ENABLED=0 go test ./pkg/...
go test ./pkg/... -run '^$' -bench . -benchmem
```

These numbers cover **policy functions only**, excluding CLI startup, JSON, model loading/inference, coding work, and process RSS. Enabling Laya still needs roughly 1.4 GiB of native-model memory.

Payback assumes reusable full context and repeated outputs of the same size. It omits cache hit rates, prefix changes, new input, latency, retries, and actual billing. Budget snapshots are neither concurrent reservations nor a spending ledger; hard service budgets need atomic reservation/settlement.

API-price estimates do not translate into ChatGPT/Codex subscription allowances or remaining quota. Preserved quality and real cost savings have not been demonstrated.

## Attribution and licenses

The table pins upstream commits. The [MIT text and copyright](../licenses/pi-pignon.LICENSE) and [Apache-2.0 text](../licenses/system-one-router.LICENSE) are preserved. Adapted switchpolicy portions retain upstream MIT terms; original project code uses Apache-2.0. Source headers and this document record modifications.
