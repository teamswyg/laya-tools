# 소수→정수 변환에서 주장의 차이를 검증하기

[English](README.en.md) · [선택형 Go API](../../../pkg/hintprepared/README.ko.md)

작은 모델이 확인 순서를 제안하려면 비슷한 설명의 차이를 구별해야 합니다. 이 개발 감사는 공개 decimal 코드를 실제 실행해 “소수부터 거부”, “범위부터 확인”, “잘라낸 뒤 범위 확인”, “그냥 정수 부분 반환”을 비교합니다. 생성 모델 학습이나 새 모델 성능 평가가 아닙니다.

요구사항은 **소수이면 먼저 non_integer, 정수이면서 int64 범위를 넘으면 out_of_range, 나머지는 정확한 정수 반환**입니다. 입력의 coefficient와 exponent도 보존해야 합니다. 이 오류 이름은 우리가 만든 wrapper의 코드이며 upstream 라이브러리 오류와 구분합니다.

| 처리 방식 | 기대값 통과 |
|---|---:|
| strict_fraction_first: 소수부터 거부 | 8/8 |
| strict_range_first: 범위부터 확인 | 7/8 |
| truncate_checked: 소수 절삭 후 범위 확인 | 5/8 |
| intpart_unchecked: 검사 없이 IntPart | 3/8 |

`9223372036854775808.1`은 소수이면서 범위도 넘습니다. 요청대로라면 non_integer가 우선입니다. 범위를 먼저 확인하면 out_of_range가 나와 요구와 어긋납니다. `1.00`은 정수이므로 1을 반환하면서 coefficient=100, exponent=-2를 보존해야 합니다.

실제 결과는 입력8개×후보4개=32회 모두 정상 반환, 명시적으로 감싼 API298회 모두 반환, API 오류·패닉0회였습니다. 후보가 의도적으로 반환한 거부 코드13개를 API 장애와 혼동하지 않습니다. 모든 입력의 coefficient/exponent 전후 값도 보존됐습니다. [전체 비교](evidence/COMPARISON.actual.public.v1.json)와 [완전 관측 gzip](evidence/OBSERVATIONS.actual.public.v1.json.gz)을 제공합니다. 원본 JSON74,491B를 2,253B로 보관했고 CRC·EOF·UTF-8·전체 JSON·카운트를 검증했습니다.

## 재현

Go1.27.1이 있는 저장소 루트에서 실행합니다.

```sh
go run ./experiments/short-claim/next-cohort-decimal-audit/replay
```

고정 소스·입력·고지를 확인하고 임시 모듈에서 실제 decimal 코드를 컴파일합니다. 전체 관측과 literal Wants를 다시 비교한 뒤 임시 파일을 정리합니다. `-go`, `-root`로 실행 파일과 공개 packet 위치를 지정할 수 있습니다. 모델·로그인·학습은 필요하지 않습니다. [공개 Go 재현 실행](replay/EXECUTION.actual.public.v1.json)은 로컬 macOS에서 통과했고 임시 디렉터리도 남기지 않았습니다. 성공은 실패한 후보까지 포함한 전체 관측을 재현했다는 뜻입니다. 모든 후보가 요구를 통과하거나 모델의 추천 품질이 좋아졌다는 뜻은 아닙니다. Linux·macOS CI 상태는 PR 검사를 확인하세요.

## 관측 범위와 다음 단계

호출 ledger는 wrapper가 명시적으로 감싼 decimal/big.Int 호출만 셉니다. 패키지 시작 때 ln10 파싱·부동소수 초기화도 일어날 수 있으며 숨은 내부 호출은298회에 포함하지 않았습니다. 입력은 정확히 고정한8개만 허용합니다. 짧은 과학 표기에도 큰 지수가 들어갈 수 있어 무제한 입력/지수 탐색으로 확장하지 않았습니다.

[실행 기록](evidence/EXECUTION.actual.public.v1.json)의 시간·자원 보고는 초기화부터32변환·snapshot·JSON/gzip 출력까지 포함한 한 자식 전체입니다. 함수 한 번의 비용이나 모델 비용이 아닙니다. Darwin 원시 보고 단위를 변환하지 않았고 GPU·전체 시스템 메모리를 측정하지 않았습니다. GOMEMLIMIT은 Go의 소프트 목표이며 RSS 상한이 아닙니다.

이것은 노출된 개발 요청1개입니다. 기존 정수 변환 과제와 의미가 겹치고 source/family 관계도 검토 중이므로 독립 Golden 추가0개, 학습 라벨0개, 역할 부여0개, Fit0회입니다. 이번 후보의 정답을 보고 기존 모델·threshold를 튜닝하지 않았습니다. 32실행이나 후보4개를 독립 표본으로 세지 않습니다. 모델의 추천 품질·독립2400 평가·LLM/전체 작업 비용 절감은 아직 입증되지 않았습니다.

다음에는 다른 출처의 검토된 요청과 관계 그룹을 확보하고, 고정 순서·BM25·lexical·모델 힌트를 같은 실제 검증 비용으로 비교합니다. 모든 후보를 남기며 실패/fallback도 기록합니다. 노출된 한 사례에 맞춰 확신 점수나 조기 중단 기준을 만들지 않습니다.

## 출처·고지

원본은 [shopspring/decimal의 고정 revision](https://github.com/shopspring/decimal/tree/ca4740823783f3bc026235a5ed3515aca626da92)입니다. 실제 컴파일 선택은121패키지/843소스 파일로 보관했습니다. require v1.0.0은 로컬 replace를 위한 graph sentinel이며 upstream release 주장으로 쓰지 않습니다.

[Decimal/fpd 전체 MIT 고지](source/decimal/LICENSE.txt)와 Go2009 header를 유지했습니다. [전체 Go BSD 고지](source/GO-BSD-NOTICE.txt)는 고정된 과거 Go 소스의 notice comparator이며 실제 복사 선행 revision·extent를 확정하지 않습니다. 소유 observer는 Apache-2.0이고 upstream의 보증이나 지지를 뜻하지 않습니다. 유한 감사의 소스 게시와 전체 family의 학습 편입은 별도입니다.

[Freeze](evidence/FREEZE.public.v1.json)의 pending은 생성 당시 상태입니다. 실제 완료는 EXECUTION/COMPARISON/CLEANUP을 보세요. private controller·호스트 경로·원시 stderr·모델 본체는 공개하지 않습니다. [PACKET](PACKET.public.v1.json)과 SHA256SUMS가 공개 별칭·완전 파일 지문을 연결합니다.
