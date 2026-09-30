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
