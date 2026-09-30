# Choose a workflow

[한국어](https://github.com/teamswyg/laya-tools/wiki/Workflows-KO) · [Home](https://github.com/teamswyg/laya-tools/wiki)

Examples with `examples/…` paths run from a laya-tools clone. Flags precede queries. Synthetic prices and catalogs are examples, not your real configuration.

## Find the code responsible for a behavior

```sh
riidolaya search --root . --lexical --json 'redirect authorization'
riidolaya search --root . --candidates 8 --limit 3 'where are redirect headers removed?'
```

Use keywords/identifiers first. The second command additionally uses Laya if installed, or reports a warning and falls back to keywords. Results contain paths, line ranges, and excerpts for inspection. Missing candidates cannot be recovered by reranking. Reranking can add substantial latency; compare it with `--lexical` on your own tasks.

For the optional code-relevance checkpoint:

```sh
riidolaya setup --checkpoint code
riidolaya search --checkpoint code 'redirect authentication'
```

Use base, not code, for model/repository choice experiments. Korean retrieval is unvalidated; `--candidate-query 'English identifiers'` can help keyword candidate generation.

## Compare model selection before execution

```sh
riidolaya plan --config examples/planner/config.json --request examples/planner/request.json --json
```

Think of this as a calculator: the catalog describes capability/context/prices; the request supplies estimated tokens, assessment, current model, and switching history. These inputs are not discovered from your subscription.

| `plan.status` | What to do |
|---|---|
| `recommend` | Inspect the proposed model and rejection reasons for alternatives. Nothing ran. |
| `hold` | The current model satisfies constraints and a switch is withheld. |
| `blocked` | No permissible recommendation. Check required capabilities, prices, budget usage, and reasons. Do not execute an empty model ID. |

A valid blocked plan exits with code 0. Invalid JSON/input exits nonzero. A trailing task adds local Laya classification; it does not call a paid coding model. This plan is not automatically applied by `codex`.

## Preview which repository owns a task

```sh
riidolaya repo-preview --catalog examples/repositories/catalog.json --json 'refund invoices'
riidolaya repo-preview --catalog examples/repositories/catalog.json --laya --json 'refund invoices'
```

Supply a catalog filtered to repositories the caller may access. Use short role summaries and aliases; do not include whole source trees. See the [example catalog](https://github.com/teamswyg/laya-tools/blob/main/examples/repositories/catalog.json).

| `status` | Meaning |
|---|---|
| `candidate` | Keyword candidate only; not a calibrated selection. Inspect `reason`. |
| `suggest` | Experimental Laya suggestion passed thresholds; still not execution authority. |
| `abstain` | No safe suggestion; inspect candidates/reason, or resolve manually. |

Every result has `preview: true`. No GitHub access, clone, repository modification, or work assignment occurs. No permission is granted. `--laya` is optional and English-only in the tested configuration. Current measurements did not establish an accuracy benefit.

## Optionally launch a new Codex task

Replace placeholder IDs with models available in your Codex installation:

```sh
riidolaya route --fast-model YOUR_FAST_MODEL --standard-model YOUR_STANDARD_MODEL --strong-model YOUR_STRONG_MODEL --json 'Fix a spelling mistake'
riidolaya codex --model YOUR_MODEL --dry-run 'Implement the feature'
```

The first only recommends. The second shows the proposed command. Removing `--dry-run` starts the installed Codex and can consume your normal Codex usage. Existing login/permissions/approval settings remain in force. No active conversation is switched.

In route output, `suggested_tier` is the model's proposal; `tier` is the applied policy; `abstained: true` explains retaining stronger capability. A high classification probability is not a coding-success guarantee. An explicit `--model` override wins.

Next: [Agent/Go integration](https://github.com/teamswyg/laya-tools/wiki/Agents-and-Go-EN) or [troubleshooting](https://github.com/teamswyg/laya-tools/wiki/Performance-and-Troubleshooting-EN).

## Recommend upgrades as well as downgrades

Easy work can move down; difficult work can move to a stronger model. Set the current model with request JSON `current` and capability ordering with catalog `rank`. A higher price alone does not mean an upgrade.

```sh
riidolaya plan --config examples/planner/config.json --request examples/planner/upgrade.json --json
```

This example recommends `example-fast` → `example-strong` with `direction: upgrade` and `reason: quality_upgrade`. The existing `request.json` demonstrates a downgrade. This example uses a supplied assessment without invoking Laya. Append an English task description to let local Laya assess complexity instead.

`direction` is `upgrade`, `downgrade`, `lateral` (same rank), `initial` (first selection), or `unchanged` (hold); blocked plans omit it. Upgrades prioritize quality over cache payback and cooldown, but still respect budgets, capabilities, context limits, and manual pins. Uncertain Laya output is not upgrade evidence; if the current model also fails the constraints, the result is `blocked`. With default configuration, catalog confidence must also reach 0.9: the switch upgrade threshold of 0.5 alone is insufficient.

These are recommendations only. No active Codex conversation is switched and no failure is automatically detected or retried. An integrating agent must update the current model and assessment at each task stage.

## Deciding whether to split work — design preview

Decomposition is not an executable feature yet. A generative agent proposes a plan; Laya will be evaluated/trained to choose keep atomic, split sequentially, split with parallel work, or request more context. Go checks dependencies, declared shared-resource conflicts, and budgets. Laya does not generate subtask descriptions or code.

For example, establishing a contract before implementation needs ordering; independent modules with fixed contracts may have concurrent work. Different files alone do not establish independence. Measure savings using success rates and total cost including planning, integration, and retries.

[Decomposition design and external review](https://github.com/teamswyg/laya-tools/blob/main/docs/decomposition-preview.en.md) · [MPS training preparation](https://github.com/teamswyg/laya-tools/blob/main/docs/mps-training.en.md) · [Tracking issue #11](https://github.com/teamswyg/laya-tools/issues/11)
