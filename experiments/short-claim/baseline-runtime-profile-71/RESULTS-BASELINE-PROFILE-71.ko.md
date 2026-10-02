# 공개 기준 도구의 1회 CPU·메모리 측정

기존 공개 `riido-shortclaim`을 Go1.27.1/CGO0/trimpath/buildvcs=false로 한 번 빌드하고, 공개 8후보 fixture에 `lexical_ordered`, 각 단계 10,000회, CPU profile을 사용해 한 번 실행했다. 재시도·모델·fit·새 label·paid 호출은 0이다. 원래 71 fit의 카운터와 결과는 바꾸지 않았다. 이 측정은 실제 76 corpus의 검증 성능이나 학습 모델의 추론 속도를 측정하지 않는다.

| 단계 | p50 | p95 | Go 할당 횟수/op |
|---|---:|---:|---:|
| JSON 파싱·검증·정규화 | 19.708µs | 45.542µs | 184 |
| 준비 입력 재검증·특징·점수·순서 | 21.334µs | 31.666µs | 81 |
| 입력 digest | 0.875µs | 1.167µs | 23 |
| ranking 직렬화 | 0.458µs | 1.250µs | 2 |
| 전체 Go 요청 | 42.500µs | 83.042µs | 291 |

각 단계는 10,000 latency sample 외에 warmup20회, 할당 검사101회(Go testing의 warmup1+측정100회)를 실제로 실행한다. 단계 callback은 각각10,121회, 전체50,605회다. source 성공 경로의 계산상 Load/Rank/ValidatePrepared는 각각20,243회, Validate40,486회, 9개 텍스트의 NormalizeText364,374회, 선택 기준 도구 tokenization20,243회다. 이 값은 소스에서 계산한 별도 benchmark 작업량이며 동적 호출 계측이 아니다. `Baselines`의 네 도구 전체 실행과 학습용 `Features` 함수는 호출하지 않는다. 단계 이름의 features는 비학습 lexical 처리까지 포함하는 기존 명칭이다.

전체 자식 OS 측정은 real1.47초, user0.99초, sys0.01초, 최대 RSS **13,189,120바이트(약12.58MiB)**, peak footprint10,568,184바이트다. 감시기 경과 시간은1.476222375초, CLI 내부 측정은1.009231333초다. 내부 측정에는 profile 종료·flush/close가 포함되지 않고 바깥 측정은 시작·종료 비용을 포함하므로 두 시간은 측정 범위가 다르다. GOMAXPROCS1, Go heap soft256MiB, 바깥30초 SIGKILL guard를 사용했고 timeout은 없었다. guard는 kill 후 Wait를 하며 정확한 전체 감시기 종료시간30초 보장을 뜻하지 않는다.

Go HeapAlloc 종료 snapshot은3,221,544바이트, 전체 누적 할당 증가량은1,009,991,512바이트다. 약1GB 누적 할당은 동시에1GB를 점유했다는 뜻이 아니다. snapshot은 수집되지 않은 garbage를 포함하고 강제 GC를 하지 않았으며 OS peak RSS와도 다르다. 이 단일 실행은 profile을 켰으므로 profiler 오버헤드는 분리해 측정하지 않았다.

CPU pprof는 한 번 읽었다. 표시된 top에서 `runtime.kevent`는 flat730ms/77.66%, `internal/lexicalhint.tokens`는 flat40ms/4.26%, cumulative140ms/14.89%였다. 프로파일은 warmup·할당 검사·다섯 단계를 함께 포함하고 샘플의 함수 귀속을 보여준다. 이 순위만으로 실제 lexical 수학 연산이나 SIMD의 이득을 단정하지 않는다. 현재 증거는 재검증·정규화와 full request의 할당 비용을 따로 개선 가설로 검토할 이유를 제공하지만, 기능·정규화 경계를 바꾸는 코드는 이번에 수정하지 않았다.

GPU와 학습 encoder/model은 사용하지 않았다. CPU pprof·Go heap·OS RSS는 GPU 사용량 계측기가 아니며 GPU 시간·메모리를 측정한 결과도 아니다. 공개 fixture 반복 측정이고 quality·토큰 비용 절감·일반화·배포 승인은 주장하지 않는다.

안전한 수치 요약은 `COMPACT-RUNTIME-PROFILE-71.v1.json` 7,493바이트/SHA256 `b0d0fd85fb082b50a564b0974177b857bda8e62856966518f7459c54796ef414`, 기호만 남긴 CPU top은 `CPU-TOP-AGGREGATE-71.v1.txt` 672바이트/SHA256 `d65e066a980195e2044a32a593af6446306b4caba1c91047076e5917a4d28576`다. 원본 입력 문구·profile·raw logs·비공개 경로는 Git/HF에 게시하지 않았다. 처음 안내된 `benchmark.go`는 실제 파일명 `bench.go`와 달라 조회1회가 실패했고 올바른 소스를 읽은 후 진행했다. build1/benchmark1/pprofread1/metadata summary1은 모두 성공했다.
