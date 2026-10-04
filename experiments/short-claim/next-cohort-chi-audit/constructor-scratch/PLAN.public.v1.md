# Constructor-local scratch / 생성자 임시 배열 재사용

이 실험은 같은 공개 Chi 개발 입력 한 개의 준비 비용을 줄이는 가설을 검증합니다. 이전 [준비 비용 프로파일](../setup-reuse/PROFILES-AGGREGATE.actual.public.v1.json)에서 특징 생성과 최종 저장 복사에 할당이 집중됐습니다. 새 구현은 distinct index 개수 스캔 후, 임시 cross-feature 배열 하나를 재사용하고 정확한 길이의 최종 배열에 각 행을 복사합니다. 원래 `hintlearn.Features`는 수정하지 않고 비교 기준으로 남깁니다.

성능 측정 전에 고정할 채택 기준:

- 모든 특징 Index·float64 비트·순서, 상쇄된 0, 마지막 overlap 특징, 점수·동점·후보 순위가 일치해야 합니다. 입력 검증·오류·텍스트 소유권·동시 읽기 계약도 유지합니다.
- 최종 저장소는 같은 길이·값과 tight capacity를 유지합니다. 생성자 전용 scratch에 pool, 전역 cache, lock을 추가하지 않습니다.
- N=1과 N=8에서 기존/새 준비를 각각 6쌍, 양쪽 실행 순서를 3번씩 배치합니다. 24개 측정 구간 모두 준비 전체와 N번 Rank를 포함합니다. 추가 토큰화와 cardinality prepass도 유료 구간입니다.
- 각 N의 6쌍 모두 Go 누적 할당 바이트가 20% 이상 줄어야 하고, 쌍별 `새 시간/기존 시간`의 중앙값이 1.25 이하여야 합니다. 통과하지 않으면 첫 결과를 보존하고 기존 Prepare를 유지합니다. 측정 후 기준을 낮추지 않습니다.

This is a source and constructor-cost experiment on one already exposed development parent. Keep the original feature extractor as an independent oracle and compare a byte-pinned prior constructor with the new constructor using the same unchanged Rank method. Charge the entire cardinality prepass and construction in every timed interval. The adoption gates above are fixed before actual model execution.

Go `TotalAlloc` and `Mallocs` describe cumulative Go allocations, not peak heap, retained memory, process RSS, native/GPU memory, or GPU execution. Shared input/model/View setup and untimed anchors remain alive consistently and are counted separately. This is not cold startup or an end-to-end agent trial. Repeated measurements do not add independent Golden parents. No new training, labels, roles, protected evaluation or default model activation is authorized by this experiment. The failed79 RIIDOH01 FP32 hashed-feature scorer remains inactive and is not the Laya encoder.
