이 실험은 RIIDOH01 가중치를 FP64 배열로 펼치지 않고 원래 압축 바이트에서 필요한 계수만 읽는 Go 커널이다. 모델의 판단 품질을 개선하거나 새로운 3진 모델을 학습하는 실험과 분리한다. 기존 모델·파일 형식·features·학습 방식은 바꾸지 않으며 HF 모델과 보호 데이터는 읽지 않았다.

기존 Decode는 FP32·INT8·3진 형식 모두 8192개 FP64 계수를 만들어 계수 배열만 65,536 B를 사용한다. `New(raw)`는 같은 형식 검사를 수행한 뒤 입력을 복사해 소유한다. 반환 후 호출자가 원본 byte slice를 바꿔도 결과가 변하지 않는다. constructor 실행 중 입력을 동시에 바꾸는 것은 허용하지 않는다. `Coefficient(index)`와 `Score(features)`가 원래 계수와 동일한 순서의 곱·합을 계산한다.

FP32는 4 B를 읽어 FP64로 변환한다. INT8은 해당 q와 header scale을 먼저 곱한다. 3진은 존재 bitmap을 읽고, 129개 uint16 접두 popcount 표로 해당 계수의 sign bit 위치를 찾는다. 이 표는 3진에만 생성하며 258 B다. map·공유 lock·가변 전역 cache는 없다. `math/bits` 호출이 있다는 사실만으로 SIMD나 특정 hardware intrinsic을 사용한다고 주장하지 않는다.

| 형식 | 소유 byte payload | 추가 접두 표 | 합계 payload |
|---|---:|---:|---:|
| FP32 | 32,792 B | 0 | 32,792 B |
| INT8 | 8,216 B | 0 | 8,216 B |
| 3진, 13 nonzero 합성 예 | 1,050 B | 258 B | 1,308 B |
| 3진, 8192 nonzero | 2,072 B | 258 B | 2,330 B |

표는 **소유 데이터 길이**다. struct/slice header, allocator padding, 호출자가 보존한 입력, feature buffers, Go heap·OS RSS는 제외한다. 입력 byte 파일 자체는 줄지 않았다. FP32에서도 복사본과 호출자의 원본을 동시에 보유할 수 있다. packed 접근이 계수 저장을 줄이는 대신 매 feature마다 변환·popcount 작업을 추가할 수 있으므로 실제 속도·할당·RSS는 후속 측정이 필요하다.

형식 허용/거절은 source7cea의 Decode와 맞춘다. 특히 FP32 header scale을 계수 최대값과 같게 강제하지 않는다. INT8 scale ±0은 기존처럼 nonzero q를 ±0으로 만들 수 있고 count 0이면 허용하지만 q=-128은 항상 거절한다. 3진의 absent bit는 +0이며, present bit에 scale 0은 거절한다. count·finite·header reserved·길이·sign padding도 유지한다. 점수는 feature 순서와 중복을 유지하고 **sign들의 합을 먼저 구해 마지막에 scale을 곱하지 않는다**. 그렇게 재배열하면 반올림 비트가 달라지는 합성 예를 검사했다. 잘못된 feature index는 panic 대신 고정 오류를 반환한다. 이 추가 안전 진단은 유효 index의 계산 계약과 별개이며, NaN/Inf feature의 IEEE 산술을 숨겨 바꾸지 않는다.

공개 합성 9형식/경계 사례에서 모든 8192계수 비트와 36개 reference score를 비교했다. 산술 재배열을 구별하는 별도 합성 사례의 score 1개도 확인했다. 빈 feature·중복·비정렬 index·큰/작은 수·signed zero·비유한 feature, prefix 64-bit 경계와 dense/empty 3진, 원본 수정 후 불변성·동시 읽기를 포함한다. 29개 bytes 허용/거절 사례도 기존 Decode와 대조했다. 첫 검사에는 차원을 무효화하지 못한 테스트 오류와 unkeyed literal의 vet 오류가 있었다. 원래 기록을 보존하고 테스트만 수정했으며, 해당 3개 검사를 다시 race로 실행해 통과했다. 두 검사 실행을 합친 synthetic reference Decode/Score 호출은 각각 78/74회다. 반복 검사가 독립 사례 수를 늘린 것은 아니다. production `view.go`는 그 수정 전후 동일하다.

benchmark는 **코드만 준비했고 실행 0회**다. 후속 root 검토 후 constructor의 기존 Decode와 owned New, 준비된 features의 기존 직접 Score와 packed 직접 Score를 나누어 비교한다. 원래 control에 callback 비용을 넣지 않는다. 생성/Decode/준비는 score 타이머 밖에서 수행하고 CPU 1·고정 합성 fixture·동일 feature 순서를 유지한다. 단일 microbenchmark가 파일 I/O나 whole-process RSS·서비스 latency를 증명하지 않는다.

이 private module에는 로컬 replace 경로가 있으므로 그대로 public runtime으로 게시하지 않는다. 기준 source commit은 `7cea49090727555924bf22679ffeccc4f38c9865`, decoder SHA `f27e18883f073555775ff15de4f2e72d4fcff9bb1b49dfe7b11007779ec95aff`, feature/scorer SHA `d0c906d1283617d0778039ff8bc4c6fff6cb0af3a813c48df088424d78bebb71`다. FP64 bit parity는 현재 Go1.27.1/darwin arm64의 합성 범위에서 확인했으며 모든 compiler/architecture에 대한 증명은 아니다. 실제 HF Decode/Score·Features·Fit·native observer·보호 데이터·업로드·공유 수정은 0이다.
