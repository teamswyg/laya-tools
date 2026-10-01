# PDCA54: 공개 Go 후보 2개의 수락 계약 준비

[English](public-go-contracts-54.en.md) · [기계 판독 기록](../benchmarks/training/public-go-contracts-54.json)

공개 Go 원천 후보 120개 중 **두 요청에 대해 로컬에서 검증한 수락기를 준비했습니다.** 모델이 작업을 수행한 결과나 난이도 정답을 얻은 것은 아닙니다. 수락기는 원본을 거절하고, 직접 작성한 올바른 수정을 통과시키며, 알려진 잘못된 수정을 거절하는 검사입니다.

이 문서는 작성 시점의 개발 기록입니다. CI를 통과한 공개 실행 지원과 배포 증거는 별도로 확인해야 합니다. 로컬 registry와 오프라인 검증 연결은 구현됐지만, 이 기록에는 아직 해당 공개 CI·병합 증거가 없습니다.

| 구분 | 이 기록의 수 |
| --- | ---: |
| 원천 조사에 기록된 후보 | 120 |
| 그중 로컬에서 검증한 수락 계약 | 2 |
| 이 두 요청의 실제 모델 실행 | 0 |
| 이 두 요청의 학습 실행·학습 라벨 | 0 |
| 보호된 최종 평가에 적격인 요청 | 0 |

두 계약은 120개에 포함된 요청을 구체화한 것입니다. 후보가 122개로 늘어난 것이 아닙니다. 원본 [PDCA53 조사](public-go-acquisition-53.ko.md)와 [120개 inventory](../benchmarks/training/public-go-acquisition-53.json)는 수정하지 않았습니다. 그 파일의 SHA256은 `e79cf90505fff4c636de99f322ab82001afd22516216cebcd8249b9b3d61cf67`입니다. 별도 laya-tools 과제·모델 파일럿 결과도 이 두 요청의 실행 수에 합치지 않습니다.

## 1. 부호를 유지하는 전체 int64 서수

`go53-humanize-ordinal64`는 [dustin/go-humanize의 고정 원본](https://github.com/dustin/go-humanize/tree/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e)에 `Ordinal64`를 추가하는 요청입니다. 예를 들어 `-21`은 `-21st`, `-11`은 `-11th`로 표시합니다. 최소값 `-9223372036854775808`도 숫자를 잃지 않아야 합니다.

원본에는 `Ordinal(int)`만 있습니다. 기존 함수는 음수에 `th`를 붙이므로 `Ordinal(-21)`의 `-21th` 동작을 유지하고, 새 함수만 절댓값 기준 접미사를 사용합니다. 호출 가능한 타입 `func(int64) string`을 확인하며, 선언된 함수와 함수 값 변수의 차이를 보장하는 계약은 아닙니다.

복사한 원본은 `go.mod`, `ordinals.go`, `ordinals_test.go`, 테스트 도우미 `common_test.go`, `LICENSE`의 5개 파일입니다. 전체 upstream 패키지를 복사한 것이 아니라, 의존성이 닫힌 서수 부분입니다. 수정 가능한 파일은 `ordinals.go` 하나입니다.

[원문 MIT 라이선스](https://github.com/dustin/go-humanize/blob/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e/LICENSE)를 그대로 보존합니다. 별도 WTFPL 원전이 있는 `number.go`는 복사하거나 빌드하지 않습니다. 실제 원본에 `NOTICE`는 없으며 만들지 않았습니다. 이 저장소의 Apache-2.0 표시로 upstream MIT를 대체하지 않습니다.

검사 기준은 임의 정밀도 `math/big` 산술과 명시적인 100개 나머지별 접미사 표입니다. 올바른 대조 구현은 문자열 끝자리로 처리하므로 검사 구현을 그대로 복제하지 않습니다. 양·음수, `11·12·13` 예외, 모든 접미사 나머지, int64 양 끝, 32비트·실수 정밀도 경계와 고정된 넓은 입력을 검사합니다.

실제 오프라인 검증에서 원본은 거절되고, 대조 구현은 원본 테스트와 독립 테스트를 합친 필수 terminal pass 6개를 기록했습니다. 잘못된 후보 13개도 거절됐습니다. 이 13개에는 새 API가 없는 주석 수정과 잘못된 타입의 컴파일 실패도 포함되므로 전부를 행동 오류 검출로 세지 않습니다. 독립 검토자는 별도의 안전한 unsigned 산술 구현을 통과시키고, 큰 수의 teen 접미사와 최소값 접미사를 잘못 처리한 두 후보를 거절했습니다. 후보가 작성한 고장 난 테스트는 수락 근거로 사용되지 않았습니다.

## 2. 엄격한 소문자 UUID 텍스트 파서

`go53-uuid-canonical-parse`는 [google/uuid의 고정 원본](https://github.com/google/uuid/tree/2d3c2a9cc518326daf99a383f07c4d3c44317e4d)에 `ParseCanonical`을 추가하는 요청입니다. 정확히 36바이트의 소문자 ASCII hex와 정해진 네 하이픈만 허용합니다. 잘못된 입력은 오류와 모두 0인 UUID를 반환해야 하며, 정상적인 앞부분을 해석한 뒤 실패해도 부분 값을 남기면 안 됩니다. 0·최대 UUID와 모든 version/variant 비트 패턴을 허용합니다.

기존 `Parse`와 `ParseBytes`의 대문자·URN·32자리 raw hex·래퍼 허용과 오류 종류는 유지합니다. 원천 조사에 있던 짧은 요청을 그대로 덮어쓰지 않고, 이 개발 계약에서 호환성과 지원 가능한 구현 범위를 별도 prompt로 명시했습니다.

원본 Go 파일 23개에 `go.mod`와 [BSD-3-Clause LICENSE](https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/LICENSE)를 더한 25개 파일을 고정했습니다. 원본 Google 표시를 유지하며 실제 `NOTICE`는 없습니다. 소스 재배포에는 저작권·조건·면책 조항을 유지하고, 바이너리 재배포에도 라이선스가 요구하는 표시를 제공해야 합니다. 다른 파일·의존성·미래 모델 배포에 대한 포괄적 허가는 주장하지 않습니다.

이 수락기는 원본 `uuid.go`의 정확한 10,254바이트 prefix 또는 고정한 Go1.27.1 formatter 결과인 10,251바이트 prefix 뒤에 함수 하나만 추가하는 범위를 지원합니다. 정규화는 확인된 formatter 공백 차이만 허용합니다. 지역 값, 루프, 고정 원본의 순수 `Parse`·`xtob`와 선정한 순수 표준 호출만 사용할 수 있습니다. 전역 쓰기, 직접 난수 읽기, 함수 별칭 호출, 포인터·전역 배열 별칭, 추가 도우미, 함수 리터럴, 동시성, import·init·directive·저작권 변경 등은 지원 밖입니다. 내부 지역 선언과 실제 전역 쓰기를 이름만으로 혼동하지 않도록 선언 객체를 구분합니다.

**지원 밖의 코드는 올바르게 동작해도 `verifier_unknown`입니다. 모델 실패나 낮은 능력의 라벨로 바꾸면 안 됩니다.** 일반 Go 프로그램 전체의 효과를 증명하는 수락기는 아닙니다.

원본 테스트는 현재 호스트에서 terminal pass 212개를 기록했습니다. 새 API가 없는 원본은 독립 계약에 실패했고, 대조 구현은 원본 212개와 독립 6개를 합쳐 218개를 통과했습니다. 지원 범위 안의 잘못된 후보 14개는 각각 독립 테스트 6개를 실제 종료한 뒤 행동 오류로 거절됐습니다. 지원 밖의 후보 21개는 별도 shape 검사로 구분했습니다.

독립 CLI 검토에서는 별도의 `hex.DecodeString` 대조 구현이 218개를 통과하고, 대문자를 허용한 후보는 거절됐습니다. 직접 난수 읽기·숨긴 전역 풀 쓰기·Google 표시 삭제는 실행 전에 unknown으로 처리됐습니다. 고정 벡터, 36위치×256바이트 검사, 반복·동시 호출, 난수 reader 및 풀 상태 검사는 한 요청의 검사 입력입니다. 9,216개 검사나 218개 terminal pass를 별도 과제 수로 세지 않습니다. reader 계측만으로 모든 entropy 경로의 부재를 주장하지 않으며, 직접 호출은 제한된 소스 gate에서도 차단합니다.

더 자세한 UUID 기록은 [별도 원천·계약 문서](UPSTREAM-54-UUID.ko.md)에 있습니다.

## 확인 방법과 고정 해시

개발 checkout에서 실행 파일을 한 번 만들고, 모델을 실행하지 않고 spec만 읽을 수 있습니다.

```sh
mkdir -p .cache/bin
go build -trimpath -o .cache/bin/riido-taskverify ./cmd/riido-taskverify
.cache/bin/riido-taskverify --task go53-humanize-ordinal64 --spec
.cache/bin/riido-taskverify --task go53-uuid-canonical-parse --spec
```

후보를 검사할 때는 고정 공개 원본과 그 복사본을 지정합니다. 권한과 실제 동작이 확인된 Go1.27.1 설치가 필요합니다. 이 기록의 실행 호스트는 macOS arm64이며 모든 플랫폼을 시험한 것은 아닙니다.

```sh
.cache/bin/riido-taskverify --task TASK --base-dir PUBLIC_PINNED_BASE --candidate-dir PUBLIC_CANDIDATE_COPY --go-root TRUSTED_GO_1_27_1
```

위 실행 파일의 종료 코드 0은 accepted, 1은 rejected, 2는 잘못된 입력, 3은 검증 불가입니다. 원본 module 파일은 그대로 고정하고, 임시 compiler module만 원래 module 이름과 Go1.27.1로 구성합니다. 외부 의존성을 가져오지 않고, 인증 정보·사용자 홈·네트워크를 상속하지 않습니다. 테스트 시간은 모델 속도나 비용 측정치가 아닙니다.

| 고정 항목 | Humanize | UUID |
| --- | --- | --- |
| Spec SHA256 | `5c6fd6058a48bd0e6441a68c351e7b1ed491aec20baff78c6942e377ae79f5d5` | `c8e3a2edcc2024ce68ba7ea7b5fed00554b2e1c36a12174f942f32b5a5a5eec3` |
| Definition SHA256 | `99a7c0723e3c9f40a069003ccf3f9c2e290c3cbe8e67be3a4ced02d4a31a29b8` | `9413ed54a10ee9a1576f71a68cf664134570cf84e735f8daeb88c66c1f4ad7c8` |
| Contract SHA256 | `3b1351162b746fa61f7837af2923644eb2f4df100de643114c9d9f0eb064fd06` | `259f3fcd56ac9bf0d4d307cc1b8ed77968b6b3bf21327160e09efd440f71ffd5` |
| Prompt SHA256 | `107ae4d6da023114cc615ebdef6709053be8471bb0ee37f9763af73f279a94a9` | `5c6456db8da12a5e9510cf9f39a448298df4dd73bead5706a41288f567f30cb2` |

원본 파일별 SHA256·Git blob과 구현 파일 참조는 별도 JSON 기록에 있습니다. 계약 검사는 유한한 개발 증거이며, 전체 int64·모든 프로그램·모든 환경의 정형 증명은 아닙니다. 두 복사 범위의 라이선스 검토를 모델 학습·가중치·데이터셋 배포의 적격성으로 확대하지 않습니다.

## 2,400개 목표까지 남은 일

목표는 도메인마다 최소 2,400개의 **서로 다른 보호된 최종 요청**입니다. 현재 120개 원천 후보와 두 개발 계약은 정답 집합이나 달성 수가 아닙니다. 이미 개발 중 읽은 이 자료는 보호된 최종 평가에서 제외합니다.

먼저 CI 공개 증거와 새 모델 실행 계획을 고정한 뒤, 실제 모델의 전체 요청 수행·독립 완료 검사·사용량과 실패 상태를 기록해야 능력 라벨을 검토할 수 있습니다. 비용이나 최적 라우팅 모델은 아직 알 수 없습니다. repository·복사 원전·공유 코드·부모/자식/형제 요청을 연결 그룹으로 묶고, 훈련·선택·보정·최종 자료를 별도로 확보해야 합니다. 번역, 반복 실행, 접미사·바이트 검사 입력은 원래 요청 그룹을 유지합니다.

모델 상향·하향 라우팅, 저장소 선택, 작업 분할, 작은 비결정적 힌트는 각각 별도 요청과 기준이 필요합니다. 이 두 Go 코드 작업의 수락기를 그 도메인들의 학습 라벨로 대신 쓰지 않습니다.
