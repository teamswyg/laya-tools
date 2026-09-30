# Recovering oversized recursive catalogs through subtrees

The initial 2,000-request run acquired 1,995 catalogs; cache-preserving continuation reached 2,395/2,400. All five remaining snapshots belonged to Babel. After adding error classification, all five were confirmed to exceed the **local 8MiB single-response bound**. Do not infer authentication failure or timeout from the former generic error.

Freeze the [repair plan](subtree-plan-33.json) first. GitHub recommends non-recursive subtree retrieval when recursive trees are truncated. Use this structure to obtain complete components instead of accepting an arbitrarily larger network response. [Official Git trees documentation](https://docs.github.com/en/rest/git/trees#get-a-tree)

`--repair-subtrees` handles only confirmed response-size overflow or explicit `truncated=true`. Discard partial responses. Obtain child identities from validated parent entries; reuse complete recursive child catalogs and expand only where needed. Ordinary request/auth/identity failures and corrupt caches do not silently enter this path.

Keep network responses at 8MiB and non-recursive validation at 2MiB. Preserve the overall 100,000-entry, 4,096-byte-per-path and16MiB total-path ceilings. Separately permit up to32MiB decompressed normalized assembled cache JSON. Bound manual expansion to depth64 and1,000 non-recursive trees per root, with30-second individual requests, minimum300ms spacing and the invocation request budget.

A failed child prevents root persistence; already verified child caches remain reusable. Recheck duplicate paths, parents, types, identities and size on the assembled root before atomic storage. Never dereference symlinks or submodule targets.

```sh
go run ./cmd/riido-filecatalog --repair-subtrees --request-budget 1500 --out .cache/file-catalogs-repair-new
```

Tests cover trigger classification, ordinary-error non-expansion, missing-child non-persistence, complete assembly, cache reuse and cycle/depth bounds. Full local Go race/vet, formatting and redacted checks passed. Actual repair, final all-2,400 coverage and offline replay are complete. This is not retrieval quality or LLM cost evidence.

Follow-up: recovery terminated successfully after763 requests with final coverage2,400/2,400. [Final aggregate](results-33.json).
