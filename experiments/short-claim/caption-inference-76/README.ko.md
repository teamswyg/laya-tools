# 짧은 주장 힌트: 캡션을 바꿔서 무엇을 배웠나

[English](README.en.md) · [저장 결과](observations/results.json) · [독립 저장 결과 대조](saved-qa/SAVED-QA.ko.md) · [상세 분석](analysis/ANALYSIS.v1.ko.md)

자세한 설명을 주면 작은 학습 모델도 더 잘 정렬할 것이라는 기대는 이번 실험에서 지지되지 않았다. 소스에서 만든 행동 설명은 기존 단어 비교 방식에 도움이 됐지만, 고정한 두 모델의 결과는 나빠졌다. 현재 사용 경로는 **모델 없이 후보 확인 순서를 제안하는 Go 도구**이며, 이번 실패 모델을 기본 설정에 적용하지 않는다.

실험은 크기 파싱의 오류와 receiver 상태, struct의 query 값 순서·생략 조건, 닫히지 않은 인용문·끝의 escape 오류 앞에서 완성한 토큰 보존이라는 세 요청을 다뤘다. 요청마다 같은 후보 세 개를 두고 A에는 기존 원문 캡션, B에는 소스에 근거해 작성한 행동 캡션을 넣었다. 요청·후보 ID와 순서·개발 평가기준·모델은 고정했다. 모델 입력은 요청과 캡션뿐이며 코드·평가기준·관측 정답을 숨겨 넣지 않았다. 두 FP32 모델 × 두 캡션 버전 × 아홉 후보로 점수 36개를 계산했고 새 학습은 하지 않았다.

`checks`는 고정된 개발 평가기준에서 **첫 허용 후보의 0부터 시작하는 순위 + 1**을 요청별로 합한 모의 확인 수다. 예를 들어 허용 후보가 세 번째면 3을 센다. 실제 코드 검사기·LLM을 그 횟수만큼 호출한 것이 아니다. 작을수록 먼저 확인할 후보를 잘 찾았다는 뜻이다.

| 고정 방식 | A → B checks | A → B Top1 |
|---|---:|---:|
| 모델 71, 후보별 정답 학습 | 3 → 8 | 3/3 → 0/3 |
| 모델 72, 후보별 학습 + 상대 순서 학습 | 5 → 6 | 1/3 → 1/3 |
| 원래 순서 | 6 → 6 | 1/3 → 1/3 |
| BM25 / lexical_ordered | 8 → 3 | 0/3 → 3/3 |

`narrow_rule`도 8 → 3이지만 이번 일반 문장에서는 전부 BM25 fallback을 사용했다. 규칙이 문장의 의미를 이해했다는 결과가 아니다. 모든 방식의 Top3는 3/3이다. 후보가 세 개라 전부 포함하는 수치여서 순서 품질을 구분하지 못한다. B는 내용뿐 아니라 길이·문장 구성도 바뀌었으므로 정보가 추가된 효과만을 원인으로 확정할 수 없다. 모델이 학습한 표현과의 차이, 구별 조건 부족 등은 후속 가설이다.

두 모델은 이전 validation에서 이미 효용에 실패한 연구 보관본이다. [모델 71의 고정 HF 보관본](https://huggingface.co/JooYoon/riidolaya-shortclaim-fp32-failed-71/tree/32b8f4579065247be0f71c83e1e7c143845e7b33)과 [모델 72의 고정 HF 보관본](https://huggingface.co/JooYoon/riidolaya-shortclaim-rank-bce-failed-72/tree/d090b00e9a5dab00d5372dfd6412c9aee0c60b7b)은 실패를 포함해 보존하기 위한 자료다. 각 FP32 파일은 32,792 B이고 정확한 파일 SHA는 [고정 계획](plan/PLAN.v1.json)에 있다. 공개 보관이나 이번 작은 결과가 모델 승격을 뜻하지 않는다. `qualification`, `production_ready`, `training_ready`, `protected_final`은 모두 false다. 원래 후보의 null 라벨·역할·가중치는 별도 개발용 source rubric으로 덮어쓰지 않았다.

**지금 써보려면** 저장소에서 Go 1.27.1로 기존 도구를 빌드한다. Python·모델 다운로드·API 키 없이 실행된다.

```sh
CGO_ENABLED=0 go build -trimpath -o bin/riido-shortclaim ./cmd/riido-shortclaim
./bin/riido-shortclaim --baseline lexical_ordered < examples/shortclaim/order.json
```

사람은 짧은 요청과 후보 설명을 입력하고, 에이전트는 출력의 `verification_order`에 따라 실제 소스와 검사를 확인한다. 모든 후보를 유지하며 `status`는 `unverified_heuristic`이다. 점수는 성공 확률이나 실행 승인이 아니다. 요청·캡션은 각각 최대 512 B·정규화 후 32단어, 후보는 1~8개다. 여러 요청에는 `--stream`을 사용한다. 입력 schema·예제·네 비교 방식은 [사용법](../USAGE-56.ko.md)에 있다. 이 명령은 두 HF 모델을 불러오거나 Codex 라우팅을 자동 변경하지 않는다.

Mac에서 자신의 모델 없는 실행 전체를 재려면 `/usr/bin/time -l ./bin/riido-shortclaim --baseline lexical_ordered < examples/shortclaim/order.json`을 사용할 수 있다. 이는 별도 사용자 실행의 수치다. 이번 연구 worker의 [저장된 자원 표본](RESOURCES.v1.json)은 최대 RSS **10.0625 MiB**, real 2.23 s, user 0.05 s, system 0.17 s였다. 시작·소스/파일 SHA 확인·모델 Decode·비교 방식·체크포인트·fsync가 포함된 전체 프로세스이며 Go heap과 다르다. 2.23초를 점수 36개로 나눠 순수 추론 지연이나 LLM 절감량으로 해석하지 않는다. GPU/Laya 인코더 실행도 아니다.

연구 실행기 소스는 `runner/source/*.go.txt` 보관 자료이며 바로 실행되는 공개 CLI가 아니다. 개인 경로가 있는 maintainer `go.mod`는 생략했다. 원시 로그·모델 본체는 Git에 넣지 않는다.

다음은 이 세 요청을 반복 학습하는 것보다 새 비교 문제를 준비한다. B의 단어 비교 방식은 이미 최소 3 checks라 이 범위에서는 추가 개선 여지가 없다. 기존 점수를 보기 전에 새로운 공개 동작의 오류 반환과 panic, 조건부 상태 변경, 부정·예외 조건을 구별하는 요청을 고정하고 소스에 충실한 후보 설명을 만든다. 공유 소스·helper·별칭은 같은 가족으로 묶어 train/validation 누출을 막고, 평가기준과 기존 비교 규칙을 먼저 봉인한다. 반복 실패가 확인되면 의미 조건을 표현하는 특징을 별도 비교 실험으로 검토한다. 이 세 개발 요청은 독립 최종 자료가 아니며 작성자들은 자료와 이전 결과에 노출돼 있었다. 도메인별 새 2,400개 보호 final 검증 목표는 아직 충족되지 않았다.
