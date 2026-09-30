# Move to at least 2,400 evaluation requests

[한국어](README.ko.md) · [Source audit](source-audit.json)

The user requires **at least 2,400 final evaluation requests** before judging usefulness. The previous 24-case synthetic evaluation is a learning smoke test, not evidence for a production recommendation. Local execution and artifact size were measured; practical usefulness remains unverified.

Count distinct requests, not candidates, pairwise comparisons or seed repetitions. Group paraphrases and shared code/repository/template cases across splits. Report unique source/code/group counts alongside requests; correlated cases are not independent samples. Prefer group-bootstrap confidence intervals. Once final data informs changes it becomes development data; two seeds do not double the sample size.

## Source already downloaded and audited

[CoSQA authors](https://github.com/Jun-jie-Huang/CoCLR) provide human-annotated natural-language query/Python-code pairs. Their [paper](https://arxiv.org/abs/2105.13239) describes at least three annotators per pair. Raw content remains local.

- Revision `14ebcacf9e9bc3e7109102632bc63047876f27d2`, `data/qa/cosqa-all.json`, 13,998,805 bytes.
- SHA256 `4d1a227fe3652992d1e98f43fa9dc1b8f54a6c27a1a24b9d182bdfb8cbc3274c`.
- 20,604 pairs: 10,020 positive /10,584 negative.
- 20,604 unique lowercased/whitespace-normalized queries; 6,267 unique outer-whitespace-trimmed code snippets.
- 6,267 connected groups sharing query or code; largest group20 pairs.
- Zero exact duplicate query/code pairs or conflicting labels among such duplicates.
- 19,054 pairs fit the current scorer's input limits. Do not silently discard the remainder; report out-of-scope/fallback rates.

This exact-duplicate audit does not establish semantic deduplication, clone detection or label correctness. Repository metadata is absent, so repository-level separation cannot yet be verified. Connected code/query groups are the minimum splitting unit to avoid seeing the same code in training and final.

## Evaluation scope

CoSQA labels **one query/code pair's relevance**. Passing 2,400 pair evaluations does not establish 64/256/1024-candidate retrieval or repository selection. Use it for relevance discrimination and scope handling; evaluate candidate retrieval separately with relevance judgments. Custom holdouts must not be presented as official benchmark scores.

CodeSearchNet offers larger code/documentation pairs, but documentation-as-query proxy labels differ from human relevance judgments. Its [official README](https://github.com/github/CodeSearchNet) assigns licenses per source repository. The official S3 Go archive returned403 during this audit; no unverified mirror was substituted.

## License and publication

CoSQA's authors distinguish MIT code from C-UDA data. The [C-UDA text](https://github.com/microsoft/Computational-Use-of-Data-Agreement/blob/master/C-UDA-1.0.md) distinguishes computational use/results from source-data redistribution and does not warrant upstream rights. Publish only aggregate audit results for now. Do not relabel this data as project Apache-2.0 or copy raw content to HF. Model publication requires separate provenance/terms/source-content checks; a root MIT LICENSE does not license every dataset as MIT.

## Next sequence and actual status

1. Extend exact/near-duplicate and connected-group auditing.
2. Freeze ≥2,400 final requests and separate validation/calibration/training groups without selecting easy or short examples based on performance.
3. Freeze hashes, metrics and missing/out-of-scope treatment before fitting. The old smoke-test gates are not new deployment gates.
4. Compare BM25/nonlearned/FP32/INT8/ternary controls on identical samples. Separate classification from retrieval; report worst costs, fallback and length/language breakdowns.
5. Establish actual candidate-search quality and repeated-use costs before recommending deployment. Sample count alone does not qualify a model.

**Completed: source audit of20,604 pairs. Pending: sealing and evaluating the≥2,400 final set.** Research artifact publication and production recommendation are distinct; current models are not recommended for production.

Reproduce after downloading the pinned source into local cache:

```sh
go run ./cmd/riido-corpusaudit \
  --input .cache/semantic-scale/cosqa-all.json \
  --sha256 4d1a227fe3652992d1e98f43fa9dc1b8f54a6c27a1a24b9d182bdfb8cbc3274c
```

The Go auditor prints no query/code text and checks file bounds/hash, connected leakage groups and conflicting duplicates.

Follow-up completed: [actual2,400-pair results](PAIR-01.en.md). Current models did not establish improvement over BM25. The pending status above records the earlier source-audit stage; see the follow-up for evaluation status.
