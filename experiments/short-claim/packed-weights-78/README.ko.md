# 78: 작은 가중치 표현의 비용

작은 가중치 표는 저장 크기를 줄였지만, 이 실험의 반복 점수 계산은 기존 FP64 배열보다 느렸습니다. v1은 4.08–5.10배, 형식 선택을 반복문 밖으로 옮긴 v2도 1.21–2.68배의 시간이 걸렸습니다. 그래서 [공개 Go 패키지 hintweights](../../../pkg/hintweights/README.ko.md)는 개발자가 선택해서 호출하는 옵션입니다. 기존 모델과 기본 Decode/Score 경로는 바꾸지 않습니다.

대상은 RIIDOH01 형식의 **8,192차원 작은 head**입니다. FP32·INT8·3진 계수를 읽고, 이미 준비된 특징과 내적을 계산합니다. Laya 전체 문장 encoder, 새 학습 모델, GPU 추론을 측정한 것이 아닙니다. 3진 저장은 네이티브 1.58-bit 연산을 뜻하지 않습니다.

## 무엇을 측정했나

각 버전에서 원래 벤치마크를 한 번씩 실행했습니다. 공개 자체 작성 합성 입력의 형식 4개, 생성/점수 계산 2개, 기존/packed 구현 2개, 반복 5회로 **각 80행**입니다. 점수 계산은 순서와 중복을 유지한 같은 특징 266개를 사용하며, 생성과 준비는 점수 측정 밖에 있습니다. 80행이나 반복 횟수는 독립 학습 문제 수가 아닙니다.

Apple M4 Pro, Go 1.27.1 darwin/arm64, CGO 0, trimpath 조건입니다. CPU 설정은 1, Go heap soft limit는 256 MiB이며 RSS 하드 제한은 아닙니다. 실행마다 Start 1, joined Wait, 종료 0, 재시도 0, timeout/overflow 0이 저장돼 있습니다. 기존→packed 순서와 따뜻해진 반복 루프를 사용했고, 시작 시 상태와 백그라운드 부하는 따로 통제하지 않았습니다.

**공개 pkg/hintweights 이름으로 포팅한 뒤 벤치마크 실행은 0입니다.** 아래 숫자는 private prototype kernel의 기존 측정입니다. 포팅에서는 공개 합성 race/vet/build와 비트·소유권 검사를 했습니다. 이 기록 자체는 상위 저장소의 전체 CI 완료를 주장하지 않습니다.

## 점수 계산은 여전히 느리다

중앙값은 버전별 같은 실행 안의 5개 샘플입니다. 비율은 packed/기존 시간이며, 1보다 크면 더 느립니다. 모든 점수 샘플은 두 구현 모두 0 B/op, 0 allocs/op입니다.

| 형식 | v1 기존 → packed ns/op | v1 비율 | v2 기존 → packed ns/op | v2 비율 |
|---|---:|---:|---:|---:|
| FP32 | 155.8 → 635.9 | 4.08배 | 155.9 → 213.1 | 1.37배 |
| INT8 | 156.7 → 680.6 | 4.34배 | 156.4 → 189.6 | 1.21배 |
| 3진, nonzero 13개 | 155.8 → 794.1 | 5.10배 | 155.9 → 248.9 | 1.60배 |
| 3진, nonzero 8,192개 | 155.7 → 776.2 | 4.99배 | 156.1 → 418.8 | 2.68배 |

v2는 Score의 형식 switch를 특징별 호출 밖으로 옮긴 세 직렬 반복문입니다. v1과 v2는 서로 다른 실행이므로, 차이가 전부 이 수정 때문이라는 인과관계나 다른 CPU에서의 개선을 증명하지 않습니다. 자세한 샘플·원문 해시는 [v1 숫자](v1/saved-qa/METRICS.v1.json), [v2 숫자](v2/saved-qa/METRICS.v2.json)에 그대로 있습니다.

## 저장 바이트, 생성 할당, 프로세스 메모리는 다르다

기존 표는 8,192개 FP64 값으로 payload 65,536 B를 갖습니다. Packed payload는 자체 소유 wire bytes와 3진 prefix 데이터의 길이입니다. 생성 B/op는 실제 관찰한 **한 연산의 누적 할당량**이며, live heap이나 RSS가 아닙니다. 아래 B/op·할당 횟수는 두 버전에서 같았습니다.

| 형식 | Packed owned payload B | 생성 기존 → packed B/op | 생성 기존 → packed allocs/op |
|---|---:|---:|---:|
| FP32 | 32,792 | 65,536 → 41,024 | 1 → 2 |
| INT8 | 8,216 | 65,536 → 9,536 | 1 → 2 |
| 3진, nonzero 13개 | 1,308 | 65,536 → 1,504 | 1 → 3 |
| 3진, nonzero 8,192개 | 2,330 | 65,536 → 2,656 | 1 → 3 |

3진 payload는 `24 + 1,024 + ceil(nonzero/8) + 258` B입니다. Struct/slice header, allocator padding, 원본 입력, 특징 배열은 제외합니다. 원본 입력과 New의 복사본이 함께 존재할 수도 있습니다. 생성 시간은 이 입력에서 더 짧았지만, 준비된 표를 반복 읽는 점수 계산의 이득으로 해석할 수 없습니다.

| 버전 | 프로세스 전체 peak RSS B | peak footprint B | OS real / user / system 초 |
|---|---:|---:|---:|
| v1 | 11,173,888 | 8,520,184 | 19.76 / 18.73 / 0.24 |
| v2 | 11,288,576 | 8,684,000 | 19.47 / 18.55 / 0.25 |

이 RSS는 **두 표현·준비 코드·Go testing runtime가 공존하는 프로세스 전체** 값입니다. 표현별 분리 측정이 아니어서 실사용 RAM 절감을 입증하지 않습니다. Controller wall은 v1 19.764369208초, v2 19.47491875초입니다. 개별 constructor의 전체 메모리나 cold start 값으로 사용하지 않습니다.

## 사용할 때의 계약

New가 읽는 동안 호출자는 입력 bytes를 안정적으로 유지해야 합니다. 반환 후에는 원본을 바꿔도 View의 소유 복사본이 변하지 않습니다. View의 값 복사는 같은 불변 데이터를 공유하며 읽기만 하는 동시 호출을 지원합니다. 호출 중 특징 배열도 외부에서 변경하지 않아야 합니다.

점수는 특징의 원래 순서와 중복을 보존해 FP64 계수 변환·곱셈·덧셈을 수행합니다. Scale을 합계 뒤로 옮기거나 0을 건너뛰지 않아 `0*Inf`·`0*NaN` 및 signed zero 의미를 유지합니다. 합성 검사에서는 기존 구현과 계수/점수 bits가 같았습니다. 모든 입력·컴파일러에 대한 보편적 증명은 아닙니다. Nil/zero View, 잘못된 index/모델 bytes의 오류 계약과 sign byte 7/8/9 경계도 포팅 검사에 포함됐습니다. [API 설명](../../../pkg/hintweights/README.ko.md)을 먼저 읽어주세요.

SIMD·SoA 가속, 모델 품질, 학습 적격성, production/default 승격, LLM/Codex 사용량·요금 절감은 이 실험에서 입증하지 않았습니다. 실제 호출 방식에 따라 저장량과 CPU 비용의 교환이 유용한지 별도로 판단해야 합니다.

## 원본과 실패 이력을 읽는 법

- [v1 준비 장부](v1/preparation/ATTEMPT-LEDGER.v1.json)는 잘못된 합성 fixture로 인한 첫 race 실패와 unkeyed test literal의 vet 실패 및 수정 내역을 보존합니다. Production v1 View 변경으로 처리하지 않았습니다.
- [v1 소스 검토](v1/independent-review/RECEIPT.v1.json)와 [저장 결과 검토](v1/saved-qa/RECEIPT.v1.json), [v2 저장 결과 검토](v2/saved-qa/RECEIPT.v2.json)는 작성자와 다른 reader의 범위입니다. 이전 자료 노출이 있는 nonblind AI-assisted 검토이며 인간 blind 평가나 새 실행이 아닙니다.
- [v2 준비 장부](v2/preparation/ATTEMPT-LEDGER.v2.json)는 파일명 오인과 봉인 전 잘못된 v1 handoff pin의 정정을 보존합니다. `history/ATTEMPT-LEDGER.pre-handoff-pin-fix.v2.json`은 **수정 전 역사**이며 현재 핀의 근거가 아닙니다. [v2 QA 장부](v2/saved-qa/ATTEMPT-LEDGER.v2.json)의 helper compile·편집 실패도 그대로 남아 있습니다.
- [공개 포팅 handoff](public-port/HANDOFF.v1.json)와 [장부](public-port/ATTEMPT-LEDGER.v1.json)는 별도입니다. 5개 top test/46개 subtest, race/vet/build 통과 및 포팅 후 benchmark 0을 기록합니다.

준비 기록의 `benchmark=0`은 **그 기록 작성 시점** 상태입니다. 후속 root 실제 실행 1회씩은 [v1 plan](v1/root/plan.v1.json)·[actual ledger](v1/root/ROOT-ACTUAL-LEDGER.v1.json), [v2 plan](v2/root/plan.v2.json)·[actual ledger](v2/root/ROOT-ACTUAL-LEDGER.v1.json)에 따로 보존했습니다. 공용 controller 재사용 때문에 root ledger schema에 `76`이 남아 있지만, 연결된 plan/binary/원문 SHA는 이 78 실행의 것입니다.

[COPY-LEDGER](COPY-LEDGER.v1.json)는 안전 원본 40개의 byte-exact 복사와 비공개 17개의 원 SHA·제외 이유를 기록합니다. Go는 `.go.txt`로만 inert 보관합니다. Private go.mod, 바이너리, 원 벤치마크/OS/테스트 로그, 호스트 경로 helper·disassembly는 공개하지 않습니다. 새 README와 [archive manifest](ARCHIVE-MANIFEST.v1.json)는 별도 작성 자료이며, 복사 원문을 수정하지 않았습니다. 보관 작업은 벤치마크·모델·학습 실행을 추가하지 않습니다.

보관 검증 도구는 생성 전 manifest 링크를 먼저 확인해서 첫 metadata 실행이 실패했습니다. 실패 소스/장부 SHA를 manifest에 남겼고, 자기 링크만 파일 생성 후 확인하도록 수정했습니다. 복사 1회 성공과 최종 검증 2회 중 1회 실패/1회 성공을 구분하며, 원 실험은 재실행하지 않았습니다.

원본 비교 대상은 자체 공개 Apache-2.0 코드 [hintlearn model](https://github.com/teamswyg/laya-tools/blob/7cea49090727555924bf22679ffeccc4f38c9865/internal/hintlearn/model.go)과 [scorer](https://github.com/teamswyg/laya-tools/blob/7cea49090727555924bf22679ffeccc4f38c9865/internal/hintlearn/learn.go)입니다. 새 라이브러리/자체 기록의 라이선스는 저장소 [LICENSE](../../../LICENSE)를 따릅니다. 모델 본체·학습된 weights·외부 문서 전체를 이 보관에 포함하지 않았습니다.
