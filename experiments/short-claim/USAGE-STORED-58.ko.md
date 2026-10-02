# 저장 정답 효용 도구 사용법

[English](USAGE-STORED-58.en.md) · [결과와 해석](RESULTS-STORED-58.ko.md)

`riido-storedutility`는 유지보수자가 후보 확인 순서의 개선 여지를 검토하는 Go 도구다. 모델 가중치나 Python 없이 공개 요청·후보 설명을 네 방식으로 정렬하고, 이미 저장된 정답과 비교한다. 후보 코드나 Laya를 실행하지 않는다. 모델 라우터를 활성화하거나 사람의 승인을 대신하는 기능은 없다.

## 가장 쉬운 확인

저장소 루트에서 Go 1.27.1로 실행한다.

```sh
go test ./internal/storedaudit ./cmd/riido-storedutility
```

검사는 고정 입력·계획·원본 결과의 해시를 확인하고 비학습 순위를 재현한다. CI도 Linux/macOS에서 같은 검사를 한다. 이 반복은 공식 실험이나 독립 요청을 추가하지 않는다. 과거 Darwin 실행 파일을 Linux에서 돌리는 것도 아니다.

원문 [results-58.json](results-58.json)의 `evaluation.metrics`는 방식별 합계, `group_metrics`는 그룹별 비용, `rows`는 요청별 순서·허용 후보·보류 이유다. `possible_relative_gain`은 정답을 아는 순서까지의 상한이며 달성한 성능이 아니다. `training_ready=false`와 `stop_reasons`를 함께 읽는다. `narrow_rule`의 fallback 72/72는 이번 자료 전체에서 BM25를 사용했다는 뜻이다.

## 공식 수집을 만든 절차

아래는 이미 끝난 수집 절차의 설명이다. 실제 소스와 입력의 서로 다른 40자리 commit이 필요하다. 계획 생성부터 평가까지 같은 실행 파일을 유지하고, 계획 파일은 입력 commit에 먼저 넣는다.

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -p=1 -o PRIVATE_BINARY ./cmd/riido-storedutility
PRIVATE_BINARY plan --repo . --source-commit SOURCE_COMMIT --out NEW_PLAN_FILE
# NEW_PLAN_FILE의 바이트를 execution-plan-58.json에 복사해 별도 input commit에 고정
PRIVATE_BINARY audit --repo . --input-commit INPUT_COMMIT --out NEW_RESULT_DIRECTORY
```

`PRIVATE_BINARY` 등 대문자 이름은 설명용 자리표시자다. 기존 계획은 실제 Darwin arm64 실행 파일의 해시까지 고정했으므로 다른 OS·빌드의 파일로 공식 평가를 다시 실행할 수 없다. 일반 재현은 위의 회귀 검사를 사용한다. 새로운 실험은 새 버전의 계획·입력·결과로 준비하며 58 원문을 덮어쓰지 않는다.

`plan`은 원문 연결만 검증하며 순위 계산은 0이다. `audit`는 원천 9개·support 13개·입력 3개·계획 1개의 실제 Git 바이트를 확인하고 새 폴더와 결과 파일을 예약한 뒤 평가한다. 실패 부분 기록도 보존한다. 출력 경로 재사용 방지는 전체 컴퓨터의 중복 실행을 막는 장치가 아니므로 공식 1회·재시도 0 원장도 별도로 필요하다.

모든 후보와 fallback을 유지한다. unknown은 틀림이나 답 없음으로 변환하지 않는다. 정답은 scorer에 전달하지 않으며 요청과 후보의 문구만 특징으로 사용한다. 자원 설정과 파일 상한은 [사전 계획](PLAN-STORED-58.ko.md)에 있다. 실제 CPU/GPU 메모리나 요금 절감을 이 기록에서 추정하지 않는다.
