# Packed78 저장된 벤치마크 결과 검토

이 공개 합성 입력에서 packed scoring 중앙값은 기존 FP64 배열 scoring보다 약4.08~5.10배 느렸다. 생성은 더 빠르고 operation당 할당 바이트는 줄었지만, 이미 생성한 계수를 반복 scoring하는 경로에서는 속도 이득이 관측되지 않았다. 이 측정만으로 기본 경로 변경이나 모델 품질을 승인하지 않는다.

검토자는 packed78 작성자가 아니며, 이전 role62·입력·관찰기 작업에 노출된 nonblind 검토자다. 원래 독립 소스 검토 파일은 수정하지 않았다. 이번에는 저장된 숫자와 실행 접합을 검토했으며, 원본 모델·Features·Fit·Decode·Score·벤치마크·테스트를 새로 실행하지 않았다.

## 점수 계산: 같은266개 공개 feature

각 항목은 하나의 원래 실행 안에서 얻은5개 sample의 중앙값이다. 비율은 packed ns/op ÷ 기존 ns/op이며, 양의 변화율은 더 느리다는 뜻이다.

| 형식 | 기존 ns/op | Packed ns/op | 시간 비율 | 시간 증가율 |
|---|---:|---:|---:|---:|
| FP32 | 155.8 | 635.9 | 4.0815× | +308.15% |
| INT8 | 156.7 | 680.6 | 4.3433× | +334.33% |
| Ternary13 nonzero | 155.8 | 794.1 | 5.0969× | +409.69% |
| Ternary8192 nonzero | 155.7 | 776.2 | 4.9852× | +398.52% |

모든 score sample에서 두 방식의 B/op와 allocs/op는0이다. Packed의 coefficient 변환·prefix/popcount·kind 분기·bounds 검사 등은 기존 배열 로드와 다른 비용을 가진다. 이번 실험은 그 요소들을 따로 제거하지 않았으므로 어느 하나가 느림의 원인이라고 입증하지 않는다.

## 생성: 서로 다른 소유 표현

| 형식 | 기존 ns/op | Packed ns/op | 시간 변화율 | 기존→Packed B/op | 할당 바이트 변화율 | 기존→Packed allocs/op |
|---|---:|---:|---:|---:|---:|---:|
| FP32 | 15,888 | 6,772 | -57.38% | 65,536→41,024 | -37.40% | 1→2 |
| INT8 | 8,883 | 4,240 | -52.27% | 65,536→9,536 | -85.45% | 1→2 |
| Ternary13 | 9,521 | 366.8 | -96.15% | 65,536→1,504 | -97.71% | 1→3 |
| Ternary8192 | 12,906 | 419.4 | -96.75% | 65,536→2,656 | -95.95% | 1→3 |

기존 생성자는8192개 float64 값을 소유하고, 새 생성자는 wire bytes의 복사본과 ternary prefix를 소유한다. 생성 뒤 같은 내용을 다른 표현으로 읽는 커널 비교이지, 동일 표현을 만드는 함수의 비교는 아니다. 적은 할당 바이트와 적은 할당 횟수도 서로 다르다. Packed는 바이트를 덜 할당했지만 할당 횟수는 늘었다.

소유 데이터 길이의 계산값은 기존65,536 B, packed FP32 32,792 B, INT8 8,216 B, ternary13 1,308 B, dense2,330 B다. 이는 wire header24 B/prefix258 B를 포함한 slice 길이이며 Go object header·allocator rounding·caller input·feature buffers는 제외한다. B/op는 operation당 누적 할당 바이트의 벤치마크 관측값이다. 예를 들어 FP32의 길이32,792 B와 실제41,024 B/op는 같은 값이 아니다. 둘 다 전체 살아 있는 heap이나 OS RSS가 아니다.

## 실행 접합과 한계

Root의 원래 benchmark invocation1회, Start1, joined Wait, exit0, retry0, timeout/overflow0를 저장된 장부와 before-start 기록에서 확인했다. Benchmark16개 case×5회=80행이 모두 있고 순서는 동결된 기존→packed 순서와 맞는다. 각 sample의 원래 iteration count를 METRICS에 보존했다. 이 count는 loop의 보고값이며, calibration·setup을 포함한 정확한 원 API 호출 수나 독립 데이터 사례 수를 재구성한 값이 아니다.

메타데이터 stdlib helper1회 성공/실패0로14개 파일 접합을 확인했다. 동결 plan, source/test/benchmark 소스, 기존 Decoder/Score 소스, handoff, 이전 독립 영수증, 실제 test binary bytes/SHA 및 buildInfo, 실제 ledger/before-start, raw benchmark·OS record SHA가 일치한다. BuildInfo는 Go1.27.1/darwin arm64/CGO0/trimpath 및 VCS metadata 미포함과 맞는다. CPU1, soft Go heap256MiB, 바깥 timeout60초, Go testing timeout50초, run `^$`, count5/200ms/benchmem 인자도 계획과 맞다. BuildInfo·기록 접합은 별도 런타임 환경 동적 계측이나 컴파일러 신뢰 증명이 아니다.

전체 결합 프로세스는 real19.76초/user18.73초/sys0.24초, controller wall19.764369208초, 최대 RSS11,173,888 B, peak footprint8,520,184 B였다. 이 한 RSS에는 두 방식·setup·Go testing runtime이 같이 들어간다. 형식별 별도 프로세스 측정이 아니므로 RAM 절감은 입증하지 못했다. Soft heap256MiB는 OS RSS hard cap이 아니다.

타깃은 Apple M4 Pro와 Go1.27.1/darwin arm64다. 같은 프로세스에서 고정 순서로 반복된 warm loop이며, 파일/시작 상태는 독립 통제하지 않았고 background load도 측정하지 않았다. 중앙값5개는 여러 독립 호스트나 데이터의 통계가 아니다. 다른 CPU·실서비스·파일 I/O·콜드 시작·GPU/SIMD로 일반화하지 않는다. 모델 품질·학습 성공·LLM 사용량/요금 절감·production/default 활성화는 검증하지 않았다. 첫71 학습 장부의 카운터도 덮어쓰지 않았다.

공개 가능한 것은 [숫자 집계와 원문 SHA](METRICS.v1.json), [검토 영수증](RECEIPT.v1.json), [별도 장부](ATTEMPT-LEDGER.v1.json) 및 이 한영 설명이다. Private raw stdout/time·binary·helper source·host paths·모델 본문은 여기에 복사하지 않는다.
