# Packed score78 소스 및 저장된 합성 기록 검토

검토 범위에서 구체적인 실행 차단 문제를 찾지 못했다. 기존 64-bit Decode의 허용 범위, 계수 읽기, Score의 순차 계산, 저장 공간 소유권이 소스상 일치한다. 이는 실제 모델 품질이나 속도·프로세스 메모리 개선을 승인하는 결과가 아니다. 모델 본문, Features, Fit, 벤치마크, 새 테스트 및 원본 런타임 API는 이 검토에서 실행·열람하지 않았다.

검토자는 packed78 코드 작성자가 아니다. 이전 role62 구현, 입력·관찰기 준비 및 다른 실험 검토에 참여한 nonblind 검토자이며, 원천 독립성이나 blind 품질 평가를 주장하지 않는다. 코드 작성자는 semantic_review60_prep이다.

## 정확성 및 소유권

`New`는 RIIDOH01 magic·8192 dimension·예약 바이트·kind·scale·nonzero count·정확한 길이를 검사한다. FP32 scale은 기존 Decode처럼 계수에 적용되지 않지만 finite/nonnegative 여부는 검사한다. scale의 +0와 -0는 모두 허용한다. INT8은 scale0에서도 q=-128을 먼저 거절하고, count는 저장된 q의 개수가 아니라 q*scale의 nonzero 개수다. 따라서 nonzero q/scale±0/count0인 기존 입력을 허용한다. Ternary는 presence 개수와 count를 맞추고, presence가 있으면 scale±0를 거절하며, 마지막 sign 바이트의 사용하지 않는 비트를 거절한다. 이 일치 판단은 동결 Decode의 64-bit 타깃 계약에 한정한다. 오류 문자열까지 같은 계약은 아니다.

검증 후 입력 바이트를 소유 복사한다. 외부로 내부 slice를 반환하지 않으며, 읽기 메서드는 저장 공간을 쓰지 않는다. 생성 중에는 caller가 입력을 안정적으로 유지해야 한다. 반환 후 caller의 입력 변경과 View 값 복사는 내부 immutable 저장 공간을 바꾸지 않는다. 저장된 race 기록의 ternary 입력 변경·동시 읽기 테스트는 처음 실행에서 통과했고, 검토자는 이를 재실행하지 않았다.

Ternary의 prefix는 129개 uint16이며, 각 누적값은 최대8192라 overflow가 없다. 유효 인덱스0..8191에서 word는0..127, sign rank는0..n-1 범위다. bit0의 하위 mask0과 bit63 경계도 소스상 맞다. 빈 ternary와 dense8192의 마지막 rank8191까지 저장된 합성 coefficient-bit 검사가 다룬다.

Score는 feature 순서와 중복을 유지하면서 매 항목의 FP64 계수를 읽고 `sum += coefficient * value`를 수행한다. INT8도 q*scale을 먼저 계산하며, ternary의 없는 계수는 기존 Decode의 할당되지 않은 +0다. sign을 먼저 합한 뒤 scale을 곱하는 재배열은 없다. 저장된 재배열 방지 테스트는 소수 scale 및 큰 수의 취소에서 차이를 구분한다. 잘못된 feature 인덱스는 기존 panic 대신 고정 오류를 반환하며, 이는 유효 인덱스 계산 일치 범위 밖의 명시적 차이다. NaN/Inf feature 산술을 새로 정상화하지 않는다. 저장된 Go1.27.1/darwin arm64 합성 결과는 모든 컴파일러·아키텍처의 FP bit 일치를 증명하지 않는다.

## 저장된 검증 증거

메타데이터 전용 stdlib helper를 한 번 실행해 handoff 자체, 준비 자산11개, 동결 reference 소스2개의14개 bytes/SHA와 저장된 JSONL 두 개의 pass/fail 집계를 확인했다. 첫 race는 top4 pass/1 fail, sub37 pass/1 fail이다. dimension fixture가 이미0인 바이트에0을 쓴 실패이며, 별도 첫 vet의 unkeyed literal13건도 작성자 장부에 보존되어 있다. 테스트만 고친 뒤 affected race는 top3/sub38 pass, fail0이다. production view.go는 수정 전후 같은 SHA다. 검토자는 새 테스트나 Decode/Score를 호출하지 않았다. 전체 작성자 합성 Decode78/Score74는 반복 실행을 포함한 수이며, 새로운 독립 사례 수가 아니다.

향후 회귀 검사를 보강한다면 다음 작은 사례가 유용하다. 지금 source에서 발견한 오류나 새 실행 조건은 아니다.

- q=-128을 scale+0/-0, count0와 함께 넣어도 거절하는지, nonempty ternary의 scale-0도 거절하는지 확인하는 조합 사례. 현재 개별 분기와 zero-scale 성공 사례는 이미 있다.
- ternary nonzero count7/8/9처럼 sign 바이트 경계를 직접 비교하는 작은 사례. 현재13개 edge와 dense8192도 경계를 통과하지만, 이 작은 조합은 실패 위치를 쉽게 드러낸다.

## 측정이 아직 말해 주지 않는 것

65,536 B는 기존 float64 계수8192개의 값 공간이다. packed FP32 32,792 B, INT8 8,216 B, ternary13 1,308 B, dense8192 2,330 B는 소유 데이터 길이 계산값이다. wire header24 B와 ternary prefix258 B는 포함하며, Go struct/slice header·allocator rounding·caller input·feature buffers·전체 heap·OS RSS는 제외한다. Handoff의 짧은 “excludes headers”는 README와 코드에서 설명한 Go header 제외라는 뜻으로 읽어야 한다. wire header를 제외한 숫자가 아니다.

준비된 벤치마크는 같은 공개 합성 입력과266개 순서가 같은 feature로 직접 Decode/New 및 Score/View.Score를 비교하고, Score의 준비 비용은 측정 밖에 둔다. 이는 커널 비교로 적합하다. 생성자는 서로 다른 소유 표현을 만들고 packed scorer는 인덱스·view 검사와 coefficient 변환 비용을 포함한다. 둘이 같은 연산 수를 수행한다는 주장은 하지 않는다. 같은 프로세스에서 control→packed 순서이며 decode 배열과 view가 함께 존재하므로 cold-start·실서비스·프로세스 RSS 비교가 아니다. 추후 count5 결과도 독립 모델 사례 수로 세지 않는다. 계획된 전체 timeout60초는 벤치마크가 그 안에 끝난다는 보장이 아니다.

현재 benchmark0이므로 ns/op·alloc/op·실제 heap·RSS 개선 결과는 없다. 저장 공간 계산상 가능성과 실제 속도/메모리 결과를 구분해야 한다. 새 format, 압축 품질, GPU/SIMD 구현, 학습 성공, serving 준비, LLM 호출/요금 절감도 검증하지 않았다.

근거: [MECHANICS.v1.json](MECHANICS.v1.json), [RECEIPT.v1.json](RECEIPT.v1.json), [ATTEMPT-LEDGER.v1.json](ATTEMPT-LEDGER.v1.json).
