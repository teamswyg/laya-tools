# Existing hint model: post-hoc xxhash preview

We applied an existing inactive failed79 FP32 reference to the exact public request and five captions. This pilot performs no new training. The preceding75-trial behavior audit remains immutable.

| Method | Recommended order | Actual contract matches for top candidate |
| --- | --- | --- |
| Learned reference | legacy → roundtrip → reconstruct → digest → direct | 7/15 |
| BM25 | direct → digest → roundtrip → reconstruct → legacy | 15/15 |
| Ordered lexical | direct → reconstruct → digest → roundtrip → legacy | 15/15 |
| Fixed input order | direct → reconstruct → legacy → roundtrip → digest | 15/15 |

**The learned reference preferred a worse candidate here; keep it inactive.** This is one nonblind, post-hoc counterexample, not independent heldout accuracy. Fixed-order success depends on input ordering. Natural-language input is unsupported by narrow_rule, which falls back to BM25; it is not a separate successful rule baseline. Scores are not probabilities.

The model sees raw request/captions only, without IDs, Wants, execution results or roles. Learning roles/new labels/Fit/protected2,400 evaluation calls remain zero. Existing37 data, HF37 and default routing are unchanged.

## Verify saved evidence

From the repository root, using Go1.27.1:

```sh
bash experiments/short-claim/next-cohort-xxhash-audit/posthoc-hint-preview/verify.sh
```

This checks saved data without model reads/downloads. [Observations](observations.actual.public.v1.json) retain scores by original candidate index, ranking permutations and exact float64 bits. See the [pre-execution freeze](FREEZE.public.v1.json) and [readback](READBACK.actual.public.v1.json).

## Optional actual reproduction

Use an existing exactly pinned model file:

```sh
MODEL_FILE=/absolute/path/data-only8192-fp32-seed1729.hbin \
  bash experiments/short-claim/next-cohort-xxhash-audit/posthoc-hint-preview/reproduce.sh
```

The model is hosted at this [immutable HF revision](https://huggingface.co/JooYoon/riidolaya-shortclaim-data-effect-failed-79/tree/5bef215895b69d3f2ef4b82bb3f1970279d67f46). The script downloads nothing, references the existing file from temporary Go workspace and checks32,792B/SHA before decoding. Library sources must match the frozen revision. It compares every output byte; repetition is not a new independent parent. GitHub contains no weights or personal paths.

## Memory and next experiment

The32,792B file expands into8,192 float64 coefficients:65,536B logical coefficient payload. This is not measured whole heap/RSS/GPU memory. Current FP32/INT8/ternary packing is a disk format; Decode expands every kind to float64. The [existing optional packed reader](../../../../pkg/hintweights/README.en.md) can be connected in a separate experiment that should first preserve every score bit/order while avoiding this coefficient copy. Recommendation quality requires separate data/training evidence.

One successful Decode and Rank are explicit calls; five internal Features/Score operations are derived from pinned source, not a full attempt/panic ledger. Compile-inclusive process time is not an inference latency or speedup claim.

Owned code is Apache-2.0; the existing scratch asset and retained MIT/BSD notices follow prior provenance. No complete unresolved-ancestry license certification or new redistribution qualification is claimed. The source comment “not a trained model” means no new training in this pilot; the reference was previously trained.
