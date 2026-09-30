# 인코더 없는 학습 01: INT8은 가능성, 3진은 실패

[English](README.en.md) · [학습 전 계획](PLAN.md) · [선택 기록](selection.json) · [최종 결과](results.json)

이번에 실제로 Go에서 학습한 것은 **요청과 후보 문서의 관련성을 매기는 작은 선형 모델**이다. Laya 가중치·인코더·교사 출력은 사용하지 않는다. 기존 난이도 라우터 및 Laya 3진 헤드 실험과 별도다. 모델을 훈련하고 로컬 CLI에서 힌트로 사용하는 경로까지 구현했지만, 일반적인 코드 검색에 쓸 만하다는 결론은 아직 내릴 수 없다.

## 무엇을 배웠나

문서 두 개가 같은 단어를 다른 순서로 쓴다.

- `cache keeps active entries and removes inactive entries`
- `cache removes active entries and keeps inactive entries`

요청의 retain/discard와 문서의 keeps/removes 사이 관계를 단어·연속 두 단어의 조합으로 학습한다. 어순 없는 BM25는 위 문서의 점수가 같다. 20개 도메인에서 학습하고 다른 6개에서 선택, 또 다른 6개에서 최종 평가했다. 문법과 동사는 모든 split에서 공유한다. **새 명사에 같은 규칙을 옮기는 통제 실험**이며 새 문법·동사·한국어·자연스러운 실제 요청에 대한 검증은 아니다. 단일 작성자의 합성 템플릿이라는 한계가 크다.

학습 80요청/40문서, 검증 24요청/12문서, 최종 24요청/12문서. 최종 문서는 학습 음성으로도 사용하지 않았다. 이미 결과를 확인한 최종 세트는 다음 개선부터 개발용이다.

## 결과

| 모델 | 파일 | 첫 후보 정답 (24건, seed 1729 / 2718) | 평균 검증 (1729 / 2718) | 사전 기준 |
|---|---:|---:|---:|---|
| BM25 | 가중치 없음 | 12 / 12 | 1.500 / 1.500 | 비교군 |
| FP32 | 32,792B | 24 / 24 | 1.000 / 1.000 | 통과 |
| INT8 PTQ | 8,216B | 24 / 24 | 1.000 / 1.000 | 통과 |
| 3진 PTQ | 1,215B | 15 / 12 | 1.583 / 1.917 | 실패 |
| 3진 STE 학습 | 1,274 / 1,272B | 16 / 15 | 1.542 / 1.625 | 실패 |

기준은 recall@4 ≥95%, 평균 검증 ≤BM25의 80%였다. INT8은 이 작은 시험에서 BM25 대비 33.3% 적은 검증을 요구했다. 3진은 크기가 작아도 검증 비용이 증가해 채택하지 않았다. CLI의 BM25 교대 정책을 적용해도 INT8 평균은 1.0이고 3진 학습은 1.583/1.625로 개선 근거가 없다. 모든 검증 횟수는 정답 ID를 아는 oracle 기준이며 실제 LLM 호출 절감이 아니다.

평가 후 추가한 **비학습 진단 비교군**은 retain→keeps, discard→removes를 명시적으로 치환하고 단어·bigram 교집합을 센다. 18/24, 평균 1.25였다. 이 비교군은 선택에 사용하지 않았고 사전등록 결과도 아니다. 더 강한 문법별 규칙만으로 이 좁은 문제를 풀 여지도 있으므로 이 자료만으로 ‘모델이 반드시 필요하다’고 주장하지 않는다.

## 실행

```sh
mkdir -p .cache/semantic-learning
go run ./cmd/riido-hinttrain --out .cache/semantic-learning/local-train
go run ./cmd/riido-hinttrain --stage check \
  --selection .cache/semantic-learning/local-train/selection.json \
  --out .cache/semantic-learning/local-check
```

기존 출력 폴더는 덮어쓰지 않는다. `check`는 결과를 기록하는 연구 명령이며 종료 코드만으로 모델 승인 여부를 판단하면 안 된다. `results.json`의 모델별 `Pass`와 적용 범위를 확인해야 한다. CI 통과는 이 연구 모델의 실제 업무 성능을 보장하지 않는다.

실험 모델을 선택적으로 사용하려면:

```sh
go run ./cmd/riido-hints \
  --model .cache/semantic-learning/local-train/int8-1729.hbin \
  --sha256 7db5ca98a5fd4426cead07cce2bef897ccf164d3e60ecd731c8ef598f1ed65e8 \
  < examples/hints/request-model.json
```

위 해시는 이번 버전/자료로 생성한 seed1729 INT8 파일이다. 다른 버전의 학습 결과는 선택 기록의 해시와 먼저 비교한다. 예시는 이미 관찰한 최종 데이터의 한 요청이며 새 품질 시험이 아니다. 모델 없이 실행하면 BM25만 사용한다. 외부 `hints`와 `--model`을 동시에 지정할 수 없다. 모든 결과는 `unverified`이며 모델 해시를 포함한다. 모델 입력은 최대 256후보, 문서/요청별 4096B·64단어로 제한한다. 초과 시 자르지 않고 `bm25_model_input_out_of_scope`로 BM25에 복귀한다. SHA 불일치/손상 파일은 오류다.

## 구현·자원·증거

- 8192개 계수. unigram/bigram 요청-문서 조합을 부호 있는 해시 특징으로 변환하고 교집합 비율을 추가한다. float shadow, softmax cross-entropy, SGD+weight decay. 2seed×2학습률×FP32/3진 STE=8후보, 각60epoch. 검증 평균 순위, 동점이면 NLL로 epoch/config를 선택했다.
- 3진 forward는 전역 평균 절대 가중치 scale 및 0.7배 임계값, backward는 identity STE다. scale은 미분하지 않는다. 이번 실패는 이 방법의 결과이며 모든 3진 학습의 불가능성을 뜻하지 않는다. 다음에는 가중치 크기 차이와 전역 scale 손실을 분리해 볼 수 있다.
- 파일은 FP32/INT8 또는 presence bitmap+nonzero sign bits다. 런타임은 8192개 float64(계수만 64KiB)로 풀어 reference score를 계산한다. **압축 파일 크기는 실행 중 계수 메모리나 3진 전용 연산 성능이 아니다.** GPU/SIMD 가속 주장은 없다.
- 학습 전 plan/data SHA를 기록했고 선택 모델 8개를 저장한 뒤 final을 읽었다. final 파일 자체가 없는 디렉터리로 학습을 반복해 8개 모델 해시가 모두 재현됐다. 모델 decode 후 계수 동일성도 검사했다.
- Apple M4 Pro, Go1.27.1: 학습 전체 wall2.34초/userCPU1.18초, 최대RSS26,607,616B(25.4MiB), swap0. 단일 INT8 CLI 실행 최대RSS5,701,632B(5.44MiB), wall0.37초(프로세스 시작/로딩 포함), swap0. 이는 warm p95나 실제 에이전트 전체 메모리가 아니다. GPU를 사용하지 않았다.

코드와 원본 합성 데이터는 프로젝트 Apache-2.0 조건으로 공개한다. 외부 모델/데이터를 복사하지 않았다. 모델 파일은 GitHub에서 제외하고 현재 로컬에 보존한다. 새 HF 배포는 이 형식의 공개 패키지 검증기를 완성한 다음 진행한다. 현재 성공은 통제 실험에 한정되며 전체 목표는 계속 진행 중이다.
