# Compute identical features with less repeated hashing

Cross features previously hashed the complete `query term + separator + document term` for every pair. Multiple document terms share each query prefix, so its bytes were repeatedly processed.

The change computes the FNV state of each query term and separator once, then appends each document term's bytes to that state. Byte order, feature order, collision merging, sign, dimension and scores remain unchanged. It adds no map, lock, shared cache, SIMD or model, and requires no retraining.

Two public synthetic fixtures were compared using Go1.27.1 on an Apple M4 Pro, CPU1 and a256MiB soft Go heap target. One benchmark invocation measured each path five times at200ms.

| Feature fixture | Original median | Prefix median | Observed time reduction | Identical allocation counts |
|---|---:|---:|---:|---:|
| Short request/caption | 4,240ns | 3,214ns | 24.20% | 2,216B /19 |
| 32words each | 351,653ns | 282,633ns | 19.63% | 70,792B /72 |

[Pinned metrics](METRICS.v1.json) preserve source hashes, all20 measurements and limitations. The control calls the original FNV loop statically, adding no per-pair indirect callback overhead. Ambient host load was unmeasured and path order was fixed; these samples do not guarantee the same improvement on other inputs, devices or deployments.

Independent pre-publication review caught a reporting error: the bounded32 prefix median had been picked from the middle sample rather than sorted values. It is corrected to282,633ns and the initial report is retained privately. No measured sample changed, and neither benchmark nor features were rerun.

An independent standard-library FNV check covers empty, UTF8, NUL and long byte compositions. Seven public fixtures separately compare exact sparse arrays and FP64 score bits with the original concatenating path. The final fixture uses a source-derived request already published in the draft corpus; it supplies no new training label or model execution.

The observation is limited to **feature extraction time**. Allocation counts did not improve; whole-process RSS, GPU usage, model quality and LLM savings were not measured. Both failed models, supervision, roles, objectives, thresholds and final data remain unchanged. Required CI gates automatic merge of the public Go change.

Reproduce from the repository: `go test ./internal/hintlearn -run '^$' -bench '^BenchmarkFeaturePrefix$' -benchmem -benchtime=200ms -count=5 -cpu=1`. Only public synthetic fixtures are used; no model download or upstream API execution is required.

[한국어](README.ko.md)
