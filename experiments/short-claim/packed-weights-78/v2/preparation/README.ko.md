# Packed Score 78 v2 준비

이 초안은 기존 packed 모델 형식의 점수 계산에서 **형식 선택을 feature 루프 밖으로 이동**한 작은 수정입니다. FP32·INT8·ternary에 각각 직렬 루프를 두었습니다. 기존 `New`, 계수 읽기 API, 소유 바이트 복사, wire 형식, prefix 표, 유효성·기본 상태·메모리 계수 방식은 그대로입니다. 원 v1 폴더와 당시의 실패 로그·테스트 수정 이력도 변경하지 않았습니다.

각 루프는 feature 순서와 중복을 유지하며, 계수를 FP64로 읽고 `coefficient * value` 뒤 누적합니다. INT8/ternary의 scale을 누적 뒤로 옮기지 않습니다. ternary에서 계수가 0이어도 곱셈을 생략하지 않아 `0 * Inf`·`0 * NaN` 동작을 보존합니다. signed zero와 NaN payload도 기존 합성 reference와 비트 단위로 대조했습니다. 잘못된 index가 유효한 prefix 뒤에 나타나도 `ErrIndex`와 +0을 반환합니다. map·lock·전역 캐시·SIMD는 추가하지 않았습니다.

Go 1.27.1 Darwin arm64에서 기존 5개 주요 검사와 38개 하위 사례가 한 번의 race 검사에서 통과했습니다. vet와 library build도 각각 한 번 통과했습니다. 새 테스트 실패는 0입니다. nil/zero View, 모든 형식의 index 오류, Inf/-Inf/NaN, signed zero, 중복, caller 입력 변경 이후의 소유권, 동시 읽기를 확인했습니다. 검사는 공개 합성 바이트에만 reference `Decode`/`Score`를 사용했습니다. HF 모델 본체나 학습 자료는 읽지 않았습니다.

| 자산 | 상태 |
|---|---|
| `view.go` | Score만 형식별 루프로 변경 |
| `view_test.go` | 기존 5개 주요 검사 안에서 경계 사례 보강 |
| `view_bench_test.go` | v1와 동일한 원문; 여기서는 실행 0 |
| `go.mod` | v1와 동일; 호스트 경로가 있어 공개 원문에서 제외 |
| 이전 v1 원본과 실패 기록 | 기존 위치에서 바이트와 해시 보존 |

이 단계에서는 **새 성능 측정이 없습니다**. 상위 작업자가 앞선 v1 실측에서 packed 경로의 CPU 증가를 관측했지만, 이번 변경이 그 원인을 분리하거나 해결했다고 주장하지 않습니다. 다음 비교는 그대로 둔 네 형식·266개 합성 feature·count 5·200 ms benchmark 계획으로 상위 작업자가 한 번 수행할 예정입니다. 라이브러리 build 결과는 실행 파일이나 새 모델이 아닙니다.

소유 payload 크기는 기존과 같고, 그것을 Go heap·RSS 감소로 해석하지 않습니다. 입력 바이트와 packed 소유 바이트가 동시에 남을 수 있습니다. Go 1.27.1에서 합성 reference와의 비트 일치는 모든 플랫폼·컴파일러의 수학적 증명이 아닙니다. Laya/GPU·원 native 관찰·Fit·Features·Project·새 가중치·학습 라벨·공유 저장소·원격 수정은 0이며 production/default activation은 false입니다.

최종 핀은 [HANDOFF.v2.json](HANDOFF.v2.json), 준비 시도와 이전 실패 이력은 [ATTEMPT-LEDGER.v2.json](ATTEMPT-LEDGER.v2.json)에 있습니다. 원 테스트 JSONL과 호스트 경로가 있는 module 파일은 공개하지 않습니다.
