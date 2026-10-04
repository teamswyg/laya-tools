# 작은 힌트가 도움이 될 구간을 먼저 찾기

새 모델을 반복 학습하기 전에, 단순 검색보다 줄일 수 있는 검증 작업이 있는지 확인했다. `google/btree`의 원본 API 여덟 개를 대상으로 경계값 포함·제외, 순회 방향, 콜백 조기 종료, 빈 입력을 시험했다. 요청·설명·정답·역할·소스·자원 예산을 먼저 고정하고, 합성 실패 시험을 통과한 뒤 첫 원본 실행을 보존했다.

**12개 요청, 한 연결된 소스 가족,34개 입력,272개 후보 관측**이다. 개발 학습 group84로 사전 배정했다. 기존37 데이터와 역할은 그대로이며,49개 독립 Golden 샘플이나 새로운 validation을 확보했다는 뜻은 아니다. upstream README가 밝힌 GoLLRB API 계보도 같은 가족으로 묶었다.

## 실제 결과

| 순서 기준 | 정답까지 확인 횟수 | Top1 / 답 있는10개 | Top3 / 답 있는10개 |
|---|---:|---:|---:|
| 고정 순서 | 54 | 3 | 5 |
| BM25 | 59 | 2 | 4 |
| lexical | 58 | 2 | 4 |
| narrow rule | 59 | 2 | 4 |

narrow rule은12건 모두 BM25 fallback이다. 정답을 미리 안다고 가정하는 최소 확인 횟수는26회다. 최선의 확인54회 대비 **51.85%의 개선 여지**가 있지만, 아직 학습 모델이 달성한 개선은 아니다.

후보96개의 유한 판정은 T18/F78/U0이었다. 요청은 정답 하나8개, 복수 정답1개, 모두 정답1개, 정답 없음2개다. 정답 없는 요청은 후보 여덟 개를 끝까지 확인했다. 모든 Got·정답·콜백·순서·판정을 보존하고 별도 비교기가 재계산했다. 정상 종료272회, 실제 콜백1,088회다. 후보별 트리 생성·삽입·순회·콜백·literal 정답 비교를 포함한 단일 시간 표본이다.

이 작은 Go 검증에서는 낙관적인 추가 힌트 계산 여유도 요청당 약 **8.62µs**다.5% 감소까지 요구하면 같은 부가 비용이라는 가정 아래 약7.85µs가 된다. 준비+네 기준의 묶음 시간854,045ns는 별도이며, 개별 기준의 latency로 나눌 수 없다. 고정 실행 순서의 첫 표본이므로 실제 재정렬 속도 향상·CPU 사용량·LLM 토큰 절감을 입증하지 않는다.

## 다음 튜닝에 쓰는 방법

별도 [cohort.train.v1.jsonl](cohort.train.v1.jsonl)은 원본 문서 설명과96개의 유한 라벨을 보존한다. 각 후보의 loss mask는1, 독립 평가 eligibility는false다. 실제 Go `LoadCohortRow`12회와 `ProjectCohort`1회가 통과했다. **새 Fit·가중치·Laya encoder·GPU/MPS 실행은0**이다. 이 자료를 구조적으로 읽을 수 있다는 것과 모델 학습·활성화 준비가 됐다는 것은 구별한다.

현재 no-answer 두 요청만 `with`, 다른 경계 요청은 `keeping`을 쓴다. 문체를 외우는 지름길이 가능하므로 원문을 사후 수정하지 않고 다음 수집에서 문체와 truth를 독립적으로 다양화한다. 다음 표현 비교와 검증 계획은 [한국어](NEXT-ABLATION-PLAN.ko.md) / [English](NEXT-ABLATION-PLAN.en.md)에 기록했다. 독립2400 평가와 실제 에이전트 절감은 여전히 미충족이다.

## 재현

저장 기록·합성 실패 시험·실제 Reader/Project를 확인하려면 저장소 루트에서 실행한다. Go1.27.1과 Node가 필요하다. Node는 연구용 소스 준비와 저장 기록 비교에 사용하며, 관측기와 학습 입력 인터페이스는 Go다.

```sh
bash scripts/verify-btree-utility.sh
```

원본 B-tree API도 다시 실행해272개 Got와 기준 순위를 비교하려면 다음을 사용한다. 새 시간값은 원래 실험 수치를 교체하지 않는다. CI의 Linux/macOS도 이 모드를 사용하며 모델을 내려받거나 실행하지 않는다.

```sh
bash scripts/verify-btree-utility.sh replay
```

## 소스·고지와 기록

[google/btree 고정 소스](https://github.com/google/btree/tree/aeba20f7a1e1315badec4eca4fdc9f754f5f880a)의 Apache-2.0 소스와 문서 주석을 사용했다. [전체 upstream 라이선스](LICENSE.upstream.txt)와 정확한5파일의 `upstream-source.json.gz`를 보존한다. Go1.27.1은 `btree_generic.go`를 선택하며 legacy `btree.go`는 제외된다. 모든 조상·미선택 코드·미래 모델의 권리를 포괄 인증한다는 주장은 하지 않는다.

`FREEZE`는 원본 실행 전, `READER-FREEZE`는 실제 데이터 Reader/Project 전의 봉인이다. `OBSERVATIONS`는 전체 첫 기록, `INDEPENDENT-READBACK`은 별도 계산 결과다. `QUALIFICATION`, `MASK-POLICY`, `ROW-BINDINGS`가 유한 라벨·역할·해석을 연결한다. `RESOURCE`의 로컬 보관 파일 이름은 자원 감사용이며, 공개 패키지에는 동일한 완전한 소스·입력·기록을 개별 파일로 제공한다. 모델 본체나 raw profile은 Git에 넣지 않았다.

CI 재현에서는 Got와 유한 판정을 정확히 비교한다. 점수는 기존 플랫폼 정책과 같은 절대·상대 오차 각 `1e-12`를 허용하되, 각 플랫폼 실제 점수에서 순위·확인 횟수·비용을 다시 계산한다. 충분히 떨어진 점수의 순서는 유지해야 한다. 작은 동점 주변의 순위 변화는 보고하고 원래 Mac 기록을 교체하지 않는다. 첫 stdout·stderr·상태를 판정 전에 보존하고 실패해도 CI artifact에 남긴다. [정책](REPLAY-POLICY.public.v1.json)과 19개 합성 플랫폼 대조를 함께 검사한다.
