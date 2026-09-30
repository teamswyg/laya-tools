# Semantic hints / 의미 탐색 힌트

[한국어 기획](PLAN.ko.md) · [English plan](PLAN.en.md) · [Raw development results](results.json)

```sh
go run ./cmd/riido-hintbench
```

A Go-only, encoder-free **nonlearned baseline**, independent of the production router. No downloads or credentials. It preserves every candidate for fallback. The executable includes all original public synthetic fixtures. Run on Apple M4 Pro, macOS arm64, Go 1.27.1. Results are a single local run, not a stable cross-machine performance claim.

학습된 모델이 아닌 Go 기준선입니다. 모델 다운로드 없이 모든 후보를 보존하며 순서만 바꿉니다. 16개 가상 후보·52개 영어 요청의 단일 작성자 개발 실험이며, 기존 라우터와 독립적입니다.

| Method | Recall@1 (48 present) | Recall@4 | Mean oracle checks (52 total) | Warm p95 |
|---|---:|---:|---:|---:|
| Catalog order | 6.25% | 25.00% | 9.077 | 0.292 µs |
| Fixed shuffle | 6.25% | 25.00% | 9.077 | 1.750 µs |
| Exact tokens | 52.08% | 75.00% | 4.615 | 1.458 µs |
| Hash 256 | 52.08% | 75.00% | 4.173 | 0.833 µs |
| Hash 4096 | 52.08% | 72.92% | 4.654 | 1.042 µs |

Warm timing covers query features, scoring and sorting over 16 candidates; excludes catalog preparation and verification. 1,000 warmups then 10,000 timed queries per method, sequential fixed method order, no statistical confidence interval. Five allocations/query for token/hash methods. Fixed-shuffle baseline regenerates hash priorities each call; its timing is not an optimized random baseline. Hash paths currently reserve 512 bytes per document per representation even for the 256-bit variant; logical bit width is not allocated size. Both representations are resident in this comparative harness. No model file, SIMD kernel or GPU inference is involved.

OS `/usr/bin/time -l` for the built executable: maximum RSS **11,321,344 bytes (10.8 MiB)**, peak footprint 9,208,312 bytes, wall 0.42s, user CPU 0.06s, system CPU rounded 0.00s, swaps 0. This includes all methods/fixtures/reporting; it is not individual-method RSS or GPU memory. Raw profiles remain local.

**The important failure:** exact tokens and both hashes get direct requests 16/16, paraphrases only 1/16, and contrasts 8/16 at rank 1. Hash-256's fewer mean checks can come from accidental collisions/tie changes; it is not evidence of learned semantics. All absent-answer queries require 16 checks, and worst case is 16 for every method. Sorting symmetric targets gives catalog/shuffle the same mean by construction. The oracle knows authored labels; saved checks are **not measured LLM calls or token savings**.

**핵심 실패:** 세 특징 기준선 모두 직접 표현 16/16, 바꿔 말하기 1/16, 대조 조건 8/16만 첫 후보로 맞췄습니다. 256-bit의 평균 검증 횟수 이득은 해시 충돌·동점 순서 영향일 수 있으며 의미 학습 성과가 아닙니다. 정답 없음은 모두 16개를 끝까지 확인합니다. 실제 LLM 호출·토큰 절감은 측정하지 않았습니다. 전체 프로세스 최대 RSS는 10.8MiB였으며 GPU를 쓰지 않았습니다. 256-bit도 현재 구현에서는 고정 배열 때문에 문서당 512B를 확보하므로 논리적 비트 수를 실제 메모리로 오해하면 안 됩니다.

Next PDCA: create new family-separated labels and negatives; train a single encoder-free relevance head with FP32/INT8/ternary controls; keep this fixture as development-only; evaluate against exact-token and BM25 baselines on a new sealed final set. Repeated identical predictions are not independent evidence. No Hugging Face model release is justified by this baseline run.

다음 PDCA는 새 가족 분리 데이터와 어려운 음성을 만들고, 인코더 없는 관련성 헤드 하나를 FP32/INT8/3진 대조군으로 학습하는 것입니다. BM25도 비교군에 추가합니다. 이번 자료는 개발용으로만 사용하며 최종 세트는 새로 봉인합니다. 새 학습 모델이 없어 이번에는 Hugging Face 모델 버전을 만들지 않습니다.

PDCA 02 adds a reusable Go BM25 index, a JSON hint CLI, and a fallback scheduler tested against bad hints: [한국어](CYCLE-02.ko.md) · [English](CYCLE-02.en.md). The original results above remain the cycle-01 record; [cycle-02 results](results-02.json) are separate.

Encoder-free learning cycle 01 now compares FP32/INT8/ternary on a separate domain-transfer probe: [한국어](../semantic-learning/README.ko.md) · [English](../semantic-learning/README.en.md). INT8 passes the narrow probe; ternary fails the inspection-cost gate. This is not real-code validation.
