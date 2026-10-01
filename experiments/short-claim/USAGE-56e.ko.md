# 유한 입력 주장 감사 사용법

`riido-propertyaudit`는 유지보수자가 작은 주장 학습의 **입력 문구와 정답 근거를 맞추는 오프라인 Go 도구**다. Laya나 모델 가중치를 로드하지 않는다. 사람이든 에이전트든 원문 결과의 요청→후보 소스→지정 입력→기대/관측을 따라갈 수 있다. 함수 전체의 안전성·일반 정확성을 승인하는 도구는 아니다.

가장 쉬운 확인은 저장소에서 Go1.27.1로 회귀 검사를 실행하는 것이다.

```sh
go test ./internal/finiteproperty ./cmd/riido-propertyaudit
```

공개 기록을 읽고 같은 소스 동작을 replay한다. CI의 Linux/macOS 검사도 이 관측 일치를 확인한다. 이는 새 공식 실험 또는 모델 호출이 아니며, 보존된 Darwin 실행 파일이 Linux에서 실행됐다는 뜻도 아니다.

## 공식 수집을 만든 순서

아래는 이미 수행한56e 수집 절차다. 새 실험을 진행할 때는 새로운 계획/버전으로 원문을 보존한다. 파일 본체를 Git에 올리지 않고 빌드 방법과 SHA만 기록한다. 실제 source/input Git commit이 필요하며 준비한 binary를 audit까지 유지해야 한다.

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -p=1 -o PRIVATE_BINARY ./cmd/riido-propertyaudit
PRIVATE_BINARY --stage prepare --source-commit SOURCE_COMMIT --out NEW_PREPARATION
# dataset.json과 plan.json의 bytes를 별도 입력 commit에 고정한 후
PRIVATE_BINARY --stage audit --input-commit INPUT_COMMIT --in NEW_PREPARATION --out NEW_AUDIT
```

공식 실행은 [고정 build recipe](build-recipe-56e.json)의 환경까지 적용했다. `prepare`는 후보 실행0으로 dataset/plan을 만든다. `audit`는 `experiments/short-claim/probes-56e.json`, `execution-plan-56e.json`에 고정된 입력의 실제 Git blob과 동일한지 확인한다. 기존 directory·변조·다른 binary·잘못된 input commit을 받으면 거절한다. CLI가 모든 새 directory의 중복 실행을 막는 것은 아니므로 실제 공식1회/실패 원장을 유지해야 한다.

## 결과를 어떻게 쓰나

[원문 결과](results-56e.json)의 `source_results`에는 각 원천과 입력의 expected/got 전체 관측이 있다. `rows`는 원래 후보 순서와 그 근거 index를 유지한다. `matches_finite_literals`는 정해진 입력에서 일치한다는 뜻이고, 다른 입력·일반 함수·모델의 확률 판단으로 확장할 수 없다. `finite_no_answer`는 후보를 삭제하거나 덜 검사하라는 지시가 아니다. 모든 후보와 fallback을 유지한다.

이 기록은 후속 효용·학습 계획의 정답 재료일 뿐이다. 현재 학습 적격false, fit/model/ranking0이고 독립 부모 증가0이다. [한글 결과](RESULTS-56e.ko.md), [English](RESULTS-56e.en.md), [다양한 공개 원천의 다음 계약](SOURCE-NEXT-56e.ko.md)을 함께 읽으면 범위와 남은 일을 확인할 수 있다.
