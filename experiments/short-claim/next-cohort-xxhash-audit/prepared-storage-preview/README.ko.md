# 같은 요청의 특징을 준비해서 재사용하기

공개 요청1개·후보5개의 특징을 배열로 한 번 준비하여 기존 Go packed 모델에 연결했습니다. **15개 점수 bit와 모든 후보 순서는 그대로이며, 잘못된 legacy-first 추천도 유지됩니다.** 이번 결과는 반복 계산 비용 개선입니다. 모델 품질·학습·Codex 비용 절감으로 해석하지 않습니다.

비교 방식은 세 가지입니다. fresh는 사용할 때마다 기존 Features를 생성하고 재사용 변환 버퍼로 Score를 호출합니다. AoS는 index·value 쌍을 한 배열에, SoA는 uint16 index와 float64 value를 각각 배열에 저장합니다. 모든 방식은 같은 검증·후보 순서 계산·호출 단위 계측을 거칩니다. 공유 map·lock·전역 캐시는 없고 각 호출자가 준비본을 소유합니다. 모델 View는 공통입니다.

| 준비본에 계속 보유하는 논리 데이터 | fresh 변환 버퍼 | AoS | SoA |
| --- | ---: | ---: | ---: |
| 데이터 len/cap bytes | 20,976B | 89,168B | 55,730B |
| 저장 객체 고정 부분 | 32B | 48B | 72B |
| interface handle | 16B | 16B | 16B |

객체 고정 부분에 slice header·offset·포인터가 이미 포함되므로 다시 더하지 않습니다. SoA 특징 payload는 AoS보다33,438B(37.5%) 작지만 fresh 버퍼보다34,754B 더 보유합니다. 실제 할당 단위·View 구조체·입력과 결과·보고서·전체 heap/RSS·GPU는 이 표에 포함되지 않습니다. 공통 encoded View payload32,792B와 원본 모델 버퍼 len32,792/cap40,960은 별도이며 원본도 실행 중 유지됩니다. Decode 참고 계수는 초기 검증에만 사용하고 측정 전 해제합니다. 초기 공존 peak가 줄었다고 주장하지 않습니다.

## 준비 비용까지 포함한 결과

고정 입력 하나를4회 교대 순서로 측정했습니다. 아래 시간은 각 구간 중앙값이며 입력 검증·원본 모델 로딩을 제외합니다. 준비를 먼저 끝낸 reuse 구간은1,024번 사용하고, construct_repeat 구간은 준비1회와 N번 사용의 전체 비용을 포함합니다. 준비된 CPU cache를 보장하는 실험은 아닙니다.

| 준비+사용 횟수 | fresh 전체 시간 / Go 할당 | AoS 전체 시간 / Go 할당 | SoA 전체 시간 / Go 할당 |
| --- | ---: | ---: | ---: |
| 1 | 490µs / 149,360B | 478µs / 217,728B | 478µs / 189,088B |
| 2 | 910µs / 276,912B | 459µs / 217,728B | 471µs / 189,088B |
| 8 | 3.65ms / 1,042,224B | 467µs / 217,728B | 523µs / 189,088B |
| 64 | 29.71ms / 약8,185,160B | 752µs / 217,728B | 1.42ms / 189,088B |
| 이미 준비한 상태: 사용1회 환산 | 453µs / 약127,552B | 4.76µs / 0B | 12.52µs / 0B |

한 번만 쓰면 준비본이 더 많은 메모리를 할당합니다. 이 사례에서 두 번부터 총 할당이 줄었고 반복64회에서는 효과가 컸습니다. AoS가 빠르게 관측됐고 SoA가 더 작은 특징 저장소를 보유했습니다. SoA는 계수마다 API의 범위 검사를 거치므로 이 속도 차이를 배열 배치·캐시 히트만의 효과로 단정하지 않습니다. 시간 차이의 일반화·생산 지연·전체 작업 성공·LLM 호출량 개선은 별도 검증이 필요합니다.

재사용은 **요청·후보 텍스트·후보 순서가 모두 동일할 때**만 합니다. 하나라도 바뀌면 새 준비본을 만듭니다. 요청마다 준비본을 무제한 저장하지 않습니다. 장기 캐시가 필요해지면 명시적인 용량·수명 제한을 먼저 설계합니다.

## 검증과 선택적 재실행

Go1.27.1에서 저장 결과만 확인합니다. 모델 다운로드·모델 읽기·학습은 없습니다. 추가 CI 검사도 이 저장 검증만 수행합니다. 기존 별도 native Laya CI는 그대로입니다.

```sh
bash experiments/short-claim/next-cohort-xxhash-audit/prepared-storage-preview/verify.sh
```

기존 정확한 FP32 모델 파일을 보유하면 아래 명령으로 다시 측정합니다. 시각·할당·heap snapshot은 달라질 수 있지만 점수 bit·순위·호출 일정·논리 저장소는 유지해야 합니다.

```sh
MODEL_FILE=/absolute/path/data-only8192-fp32-seed1729.hbin \
  bash experiments/short-claim/next-cohort-xxhash-audit/prepared-storage-preview/reproduce.sh
```

[사전 고정](FREEZE.public.v1.json), [원래 관측](observations.actual.public.v1.json), [검증·집계](READBACK.actual.public.v1.json)를 각각 보존합니다. Source-derived setup/nested counts와 실제 build/five-score attempted/normal-success counters를 구분합니다. 전체 per-call panic ledger는 없습니다. Go heap snapshot은 모든 공통 객체를 포함하는 두 시점의 값이며 peak/RSS·방식별 메모리가 아닙니다. 새 pprof/GPU 측정은 하지 않았습니다.

부모1개·비맹검 사후 자료이며 반복은 독립 사례가 아닙니다. 새 역할·라벨·Fit·보호 최종 평가·기본 활성화는0입니다. 기존 실패 모델은 Apache-2.0 scratch 연구 참조이며 기존 출처 고지를 유지합니다. 전체 미해결 계보의 적격화를 주장하지 않습니다. 모델·원본 프로파일·비공개 입력은 GitHub에 올리지 않습니다.
