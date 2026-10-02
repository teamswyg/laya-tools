# Packed78 v2 소스 변경 및 저장된 결과 검토

검토 범위에서 구체적인 의미·정확성 차단 문제를 찾지 못했다. v2의 packed scoring은 같은 실행의 기존 배열 scorer보다 여전히1.21~2.68배 느렸다. 원래 v1 코드·실패·metrics는 그대로 유지했다. 검토자는 v1/v2 코드 작성자가 아니지만 이전 role62·입력·관찰기 작업에 노출된 nonblind 검토자다. v2 작성자는 task_expansion_52이며, 저장된 작성자 테스트를 검토자가 재실행하지 않았다.

## 변경 범위와 계산 계약

kind 선택을 feature마다 호출한 coefficient 메서드에서 Score 앞의 switch로 옮겼다. 세 branch 모두 원래 feature 순서와 중복을 그대로 순회하고, 유효 인덱스를 먼저 확인한 뒤 같은 계수 값을 FP64로 얻고 `sum += weight * value`를 수행한다. INT8은 여전히 q*scale을 먼저 구한다. Ternary absence의 weight는+0이며, absence일 때도 곱셈을 생략하지 않는다. 따라서0*Inf/0*NaN을 정상적인0으로 바꾸지 않는다. scale을 합산 뒤로 옮기거나 feature를 재정렬·합치지 않는다.

잘못된 인덱스는 부분 합이 finite/NaN인지와 관계없이 기존 v1처럼 ErrIndex와+0를 반환한다. nil/zero View는 여전히 ErrView다. New·validation·소유 복사·private fields·Coefficient·uint16 prefix·OwnedPayloadBytes 및 format은 변경되지 않았다. 메타데이터 helper가 Score 전후 나머지 원문 bytes, fixtures/features 함수가 들어 있는 test prefix, benchmark source의 exact equality를 확인했다. 따라서 이전 validation/ownership/prefix 검토를 해당 변경 없는 범위에 연결할 수 있다. 이것이 모든 컴파일러의 bit 일치 증명은 아니다.

저장된 작성자 race 기록은5 top tests/38 subtests/1 package pass, fail0이다. 강화된 테스트 소스에는 absent ternary×Inf, NaN payload, signed-zero feature, 양·음 Inf의 취소, 모든9개 fixture에서 invalid index 전 finite/Inf 항목과+0 오류 반환이 들어 있다. 이는 작성자의 Go1.27.1/darwin arm64 합성 검증이며, 검토자는 테스트·Decode·Score·실제 모델을 호출하지 않았다.

## v2 실행 안의 점수 비교

각 값은5개 sample의 중앙값이다. Ratio는 packed÷legacy이며, 양의 시간 변화율은 더 느리다는 뜻이다. 같은266개 공개 synthetic feature를 사용했고 setup은 score timing 밖이다.

| 형식 | Legacy ns/op | Packed v2 ns/op | 비율 | 시간 증가율 |
|---|---:|---:|---:|---:|
| FP32 | 155.9 | 213.1 | 1.3669× | +36.69% |
| INT8 | 156.4 | 189.6 | 1.2123× | +21.23% |
| Ternary13 | 155.9 | 248.9 | 1.5965× | +59.65% |
| Ternary8192 | 156.1 | 418.8 | 2.6829× | +168.29% |

모든40개 score sample에서 양쪽 B/op·allocs/op는0이다. v1의4.08~5.10배보다 격차가 작아진 관측이지만, v1/v2는 서로 다른 실행이며 background load·시작 상태를 독립 통제하지 않았다. 이를 kind hoist만의 인과 효과나 다른 CPU에서의 보장으로 제시하지 않는다.

생성 중앙값 legacy→packed는 FP32 15849→6756 ns, INT8 8919→4223 ns, ternary13 9562→365.5 ns, dense12832→413.7 ns다. B/op는 기존65536에 대해 packed41024/9536/1504/2656이며 allocs/op는1에 대해2/2/3/3이다. New 원문은 변경되지 않았으므로 생성 숫자의 미세한 v1/v2 차이도 코드 최적화 결과로 귀속하지 않는다.

## 실행·자원·provenance 한계

원래 Root invocation1/Start1/joined Wait/exit0/retry0,16cases×5=80행과 원 iteration count를 확인했다. stdlib metadata helper 실행1회 성공/실패0로19개 exact pins, test-binary Go1.27.1/darwin arm64/CGO0/trimpath 및 VCS metadata 없음, CPU1·soft heap256MiB·바깥60초·testing50초·run `^$`·count5/200ms/bounds를 연결했다. Helper 준비에는 patch-context 실패1·Perl 편집 문법 실패1·Go compile 실패1이 있었고, 실행 전 실패 소스를 보존했다. 실제 helper는 두 번째 Go-run에서 한 번 실행됐으며 원본 실험은 재실행하지 않았다.

전체 process는 real19.47/user18.55/sys0.25초, controller wall19.47491875초, peak RSS11,288,576 B/footprint8,684,000 B였다. 같은 process에 두 표현·setup·testing runtime이 함께 있다. 소유 data length(legacy65536/packed32792·8216·1308·2330), cumulative allocated B/op, 살아 있는 heap와 전체 RSS는 서로 다르다. 한 RSS로 표현별 RAM 절감을 입증하지 못하며 soft heap은 RSS hard cap이 아니다.

이 결과는 Apple M4 Pro의 고정 legacy→packed 순서·반복 warm loop만 다룬다. 새로운 모델 품질·학습·LLM 사용량/요금 절감·GPU/SIMD·serving/cold-start 또는 production/default 활성화 승인은 없다. Private raw logs·binary·helper source·host paths와 HF/훈련 모델 본문은 공개 자료에 넣지 않는다.

[전체 숫자 및 원문 SHA](METRICS.v2.json) · [영수증](RECEIPT.v2.json) · [장부](ATTEMPT-LEDGER.v2.json).
