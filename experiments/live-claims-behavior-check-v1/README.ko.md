# 실시간 주장 동작 점검 v1

[English](README.md) · [실시간 데모 설치](https://github.com/teamswyg/laya-tools/blob/main/docs/live-claims-demo.ko.md)

이 제한된 개발 진단에는 **새로 작성한 공개 가능한 합성 메시지 24개**, 한국어 12개와 영어 12개를 사용했다. AI 작성자는 추론 전에 모든 문장의 세 헤드별 의미 가설을 고정했다. 이는 **사람이 판정한 정답, 최종 평가, 운영 정확도 추정이 아니다**. 헤드 출력 72개는 입력 24개를 공유한다. 명시된 일부 범위의 완료 가설은 다른 작업이 남아 있어도 true일 수 있으며, 전체 작업이 완료됐다는 뜻은 아니다.

문장 24개와 가설은 `2026-10-08T01:20:55.092Z`에 잠갔다. 각 문장에 실제 HTTP `POST /api/hints`를 정확히 한 번 요청했다. `2026-10-08T01:21:41.187Z`부터 `2026-10-08T01:21:41.211Z`까지 **24회 시도, HTTP 200 응답 24개, 재시도·대체 입력 없음**이었다. 공개 자료를 만드는 과정의 **추가 모델 호출은 0회**다. 현재 모델, 학습 스텝(4240), 임계값은 변경되지 않았으며, 학습·모델 선택·제품 상태 갱신을 하지 않았다.

전후 상태와 모든 응답의 모델 SHA-256은 `cfd35ee70a23f94a8d470b7dea596244a91ac7f1410475e8a42959693c99864b`다. confidence 하한은 **0.9**, margin 하한은 **0.05**, temperature는 **1**이다. 상태 응답은 `research_preview`, `semantic_quality_qualified=false`다.

## 기록된 결과

각 행은 출력 12개다. T/F/U는 true/false/unknown이다. “채택된 양성”은 고정된 수치 하한을 통과한 final=true다. 불일치는 기록된 출력과 고정된 **AI 가설**의 차이이며, final 불일치에는 판단 유보가 포함되므로 오류율·정확도로 해석할 수 없다.

| 언어 | 헤드 | 가설 T/F/U | Raw T/F/U | Final T/F/U | 채택된 양성 | 가설≠raw | 가설≠final |
|---|---|---:|---:|---:|---:|---:|---:|
| KO | response_requested | 2/9/1 | 3/9/0 | 0/3/9 | 0 | 2 | 8 |
| KO | current_activity_claimed | 2/9/1 | 3/9/0 | 0/2/10 | 0 | 2 | 9 |
| KO | completion_claimed | 3/8/1 | 3/9/0 | 0/3/9 | 0 | 2 | 8 |
| EN | response_requested | 2/9/1 | 3/9/0 | 1/3/8 | 1 | 2 | 7 |
| EN | current_activity_claimed | 2/9/1 | 4/8/0 | 1/6/5 | 1 | 3 | 5 |
| EN | completion_claimed | 3/8/1 | 5/7/0 | 0/4/8 | 0 | 2 | 7 |

출력 72개의 합계는 **true 2개, false 21개, unknown 49개, 채택된 양성 2개, 가설≠raw 13개, 가설≠final 44개**다. final unknown 49개는 모두 `low_confidence` 때문이며, raw 최댓값이 의미적 unknown인 경우는 없었다. 따라서 final unknown이 의미적 unknown 가설과 같다는 이유만으로 의미적 불확실성을 인식했다고 볼 수 없다. 채택된 양성은 EN09의 response_requested(가설 true)와 EN04의 current_activity_claimed(가설 false)였다.

EN04는 구체적인 한계를 보여 준다.

> I am not editing the walking-tour map now. That task is paused.

현재 활동 가설은 **false**였지만 raw와 final은 모두 **true**였다. confidence는 **0.946169690250911**, margin은 **0.9008599783475525**, T/F/U 확률은 **0.946169690250911 / 0.04530971190335854 / 0.008520597845730313**이었다. 현재 활동을 부정하는 보고가 양성으로 임계값을 통과했으므로, 임계값 통과는 의미 품질 검증이 아니다.

이번 **순차적인 로컬호스트 POST 24회에 한해**, HTTP 경과 시간의 최소/중앙값/p95/최대는 **0.289 / 0.455 / 1.354 / 3.249 ms**, 서버가 반환한 추론 시간은 **2 / 10 / 18 / 19 µs**였다. p95는 nearest rank다. HTTP 시간에는 응답 본문 읽기가 포함되며, `inference_us`는 더 좁은 서버 측 범위를 측정한다. 원래 기록은 Node 내장 `fetch`와 `performance.now`로 만들었다. 이는 해당 실행의 관측값이며 처리량이나 하드웨어 자원에 대한 주장, Go 시간 측정과의 동등성 주장이 아니다. 반올림하지 않은 수치는 증거 파일에 남아 있다.

## 증거와 통제

- [inputs.json](inputs.json): 변경되지 않은 원문 24개, 가설, 근거, 의미 정의, 작성 시각.
- [responses.json](responses.json): 원래 응답 본문 전체, 파싱된 예측, 헤더, 바이트 수, 시각, HTTP 경과 시간. [status-receipts.json](status-receipts.json)은 전후 모델 핀을 보존하며, [results.json](results.json)은 원문 없는 정확한 집계를 담는다.
- [pre-inference-lock.json](pre-inference-lock.json), [provenance.json](provenance.json), [exclusion.json](exclusion.json): 잠금 순서, 보존한 원본 자료의 SHA 핀, 명시된 공개 메타데이터 변환, 모든 원문과 변형을 **final400/final1200/final2400, 모든 최종 사람 판정 정답 평가, 향후 모델 선택기**에서 영구 제외하는 통제. 이 입력 24개는 개발 전용이다.
- [SHA256SUMS](SHA256SUMS): 공개본의 체크섬. 기록자의 식별 정보와 설명 메타데이터만 변환했으며, 문장·가설·실제 응답·수치 결과는 그대로다. 원본 자료를 보존했다. 공개 묶음에는 호스트 절대 경로, 비공개 내용, 모델 가중치가 없다.

저장소 루트에서 다음 Go 표준 라이브러리 검증기를 실행하면 저장된 증거만 읽고 **네트워크·모델 호출은 하지 않는다**.

```sh
go run ./experiments/live-claims-behavior-check-v1/verify.go
go vet ./experiments/live-claims-behavior-check-v1/verify.go
```

공개 체크섬, 고유한 순서의 기록 24개, 원본·공개 핀, 잠금 순서, 입력·제외 해시, 응답 본문 대조, 언어·헤드별 수, 고정 임계값 산술, 불일치 합계, 원래 시간 집계를 검증한다. 저장된 자료를 검증하며 원래 Node 기록기를 재실행하지 않는다.

**새로운** 수동 관측을 하려면 위 데모 문서의 고정된 다운로드·빌드·실행 절차를 사용하고, `/api/status`에서 위 모델 핀과 하한을 확인한 뒤 JSON `{"text":"..."}`를 재시도 없이 한 번 보낸다. 예시는 다음과 같다.

```sh
curl --fail --silent --show-error http://127.0.0.1:8877/api/status
curl --fail --silent --show-error \
  -H 'Content-Type: application/json' \
  --data-binary '{"text":"I am not editing the walking-tour map now. That task is paused."}' \
  http://127.0.0.1:8877/api/hints
```

수동 재실행은 위에 저장한 실행이 아니라 **새 출력과 시간 관측값**을 만든다. 전체 묶음을 다시 실행한다면 `inputs.json`의 변경되지 않은 원문 24개를 각각 정확히 한 번 사용하고, 오류를 재시도하거나 입력을 바꾸지 말고 그대로 보존한다.
