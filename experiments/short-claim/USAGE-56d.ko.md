# 상주 힌트 비용을 확인하는 방법56d

[English](USAGE-56d.en.md) · [실제 결과](RESULTS-56d.ko.md) · [구현 경계](IMPLEMENTATION-RESIDENT-56d.ko.md)

에이전트가 작은 힌트 도구를 계속 켜두고 요청할 때 얼마나 가벼운지 확인하는 개발용 도구다. 입력은 공개 자체 작성 자료이고, 이번56d는 **학습 모델 없이 동작하는 Go 기준선**의 실제 JSONL 처리 비용을 기록했다. 추천 순서는 검증할 후보를 먼저 보여주는 힌트이며, 후보를 삭제하거나 작업 실행을 승인하지 않는다.

## 먼저 힌트 처리기를 직접 써보기

저장소 루트에서 Go1.27.1로 빌드한다. Python이나 GPU는 필요하지 않다.

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -p=1 -o bin/riido-shortclaim ./cmd/riido-shortclaim
bin/riido-shortclaim --baseline lexical_ordered < examples/shortclaim/order.json
bin/riido-shortclaim --stream --baseline lexical_ordered < examples/shortclaim/resident.jsonl
```

첫 명령은 보기 좋은 JSON 파일 한 건을 처리하고, 두 번째 명령은 실제 한 줄 JSONL을 스트림으로 처리한다. 스트림은 JSON 한 줄을 보내면 응답 한 줄을 받는다. 같은 프로세스에 다음 줄을 계속 보낼 수 있다. `verification_order`에 원래 후보가 모두 남고 상태는 `unverified_heuristic`이다. 스트림 종료는 입력 종료로 알린다. 잘못된 JSON이나 후보 제한 위반을 성공한 힌트로 바꾸지 않는다.

이 비용 측정이 Laya ONNX 추론이나 Codex 토큰 감소를 증명하지는 않는다. Codex·riido-daemon 연동은 별도 선택이며 이 명령이 자동 등록하지 않는다.

## 측정 자료 준비해 보기

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -p=1 -o bin/riido-residentperf ./cmd/riido-residentperf
bin/riido-residentperf --stage prepare --out .cache/my-resident-preparation
```

새 출력 디렉터리를 사용한다. `corpus.json`에는 정확한 전송 bytes와 원래 부모 연결, `preparation.json`에는 선택 수·해시·소스 목록, `build-recipe.json`에는 고정 빌드 조건이 있다. 준비는 child 자원 측정·새 정답 감사·학습·모델 호출을 하지 않는다. 정규화 특징 해시는 중복 선택 기준이며, 원문 문법이 다른 요청의 안전한 결과 캐시 키가 아니다.

## 게시된 공식 실행을 읽거나 재현하기

[기계 실행 계획](execution-plan-56d.json), [원본 연결 corpus](wire-corpus-56d.json), [실행 전 봉인](freeze-56d.json), [원문 결과](results-56d.json), [수치 요약](summary-56d.json)을 함께 읽는다. 소스 커밋 `60b03500fab417470d48d1a096be4ca1f43fafc2`와 실행 전 계획 커밋 `891e52d83e9aaf39c87cb5ef55c6d492724bebf1`을 분리해 고정했다.

게시된 공식 계획은 **darwin/arm64와 실제 Go1.27.1 두 바이너리 SHA에 묶여 있다.** 다른 OS·Go 버전·추가 빌드 옵션은 그대로 통과하지 않는다. [고정 recipe](build-recipe-56d.json) 환경을 적용해 빌드하고 해시를 확인한 경우 다음 형태로 실행한다.

```sh
bin/riido-residentperf --stage replay \
  --plan experiments/short-claim/execution-plan-56d.json \
  --plan-sha256 8c0b7a5b697e52d4afffbb5e4ca2f9b5d3c992a31d4eae4f9767bf57e2dbfed2 \
  --corpus experiments/short-claim/wire-corpus-56d.json \
  --binary bin/riido-shortclaim \
  --out .cache/my-separate-resident-replay
```

새 실행은 새 개발 관측이다. 기존56d 결과에 합치거나 최초 관측을 바꾸지 않는다. 바이트·환경이 다르면 억지로 해시를 수정해 기존 계획으로 실행하지 않고, 별도 계획과 결과 버전을 먼저 만든다. CLI는 소스 commit의 형식을 검사하며 Git blob31개와 그 commit의 실제 관계는 게시된 별도 봉인 기록으로 확인한다.

성공 횟수만 읽지 말고 단계별 수, EOF·종료·reader 정리, 오류, 입출력 해시도 확인한다. 실패 뒤 후속 행은 미시작으로 남고 자동 재시도하지 않는다. 기존 출력도 덮어쓰지 않는다. 학습이나 운영 승인 도구가 아니다.

## 숫자로 알 수 있는 것

이 Mac에서 child peak RSS는9.21875~10.515625MiB였고 timed wire RTT의 행별 p95 범위는0.021875~0.050333ms였다. 첫 응답577.13ms는 별도로 보존했다. Controller의33.109375MiB peak는 준비와 측정 작업을 포함하는 별도 값이다. 이 두 peak를 더해 동시에 사용한 메모리로 보지 않는다.

72개 원본 중3후보48개만 선택했고 나머지24개는 제외했다. 이48개를25,080회 반복했으며 독립적인2400개 골든셋이 아니다. 이번 corpus의 `narrow_rule`은 모두 범위 밖 요청이라 BM25 fallback으로 처리됐다. 따라서 좁은 문법 성공 경로의 성능이나 의미 정확도를 얻었다고 해석하지 않는다.

다음 캐시·SIMD·동시성·모델 실험은 별도 비교 계획이 필요하다. 현재 자료는 작은 비학습 기준선의 자원 비용을 제공하며, LLM 절감 효과와 모델의 유용함은 별도 실제 작업 검증으로 판단한다.
