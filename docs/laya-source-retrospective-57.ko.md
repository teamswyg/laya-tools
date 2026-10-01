# Laya 원천을 다시 읽고 정리한 방향57

2026-10-02에 SDK commit [`4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c`](https://github.com/NandhaKishorM/laya/tree/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c)를 읽었다. 공개 파일19개의 Git blob과 내용을 대조했고 모델 실행이나 가중치 다운로드는 하지 않았다. 이번 검토는 성능 개선 실측이 아니다.

## 우리 목표와 원본의 차이

원본의 [Router](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/laya/router.py#L774)는 명시한 모델·작업·언어에 따라 **Laya 체크포인트를 선택**한다. Codex가 일을 끝낼 가능성이나 사용량을 학습한 예측기가 아니다. [typed decoder](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/laya/agent.py#L1179)는 choice의 argmax, 순서형 score의 확률 가중 평균, noul의 P(true)를 제공한다. 자유로운 설명 생성과는 다른 기능이며 score의 반올림이 argmax와 같다는 보장은 없다.

뤼이도의 우선 목표는 극히 작은 자원으로 반복해서 **검증할 후보에 대한 힌트**를 주는 것이다. 후보와 fallback을 모두 보존하고 실제 검증 도구가 정답과 승인을 맡는다. 새 모델 개수보다, 서로 다른 공개 동작에서 BM25보다 확인 작업을 줄이는지가 다음 판단 기준이다.

## 바로 참고할 세 가지

| 참고점 | 다음 범위 | 검증할 것 |
|---|---|---|
| 점수 의미와 보류 상태 | [confidence 구현](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/laya/confidence.py)을 참고해 확률/휴리스틱·보정 여부·passed/abstained/unevaluated를 작은 메타데이터로 구분 | 설정하지 않은 게이트를 통과로 표시하지 않음; NaN/missing 점수는 통과 불가 |
| 무거운 모델 로딩 전 지원 범위 | 기존 repo preview와 작업 라우터의 English checkpoint 언어 가드를 일관되게 검토 | 미검증 script에서는 loader 호출0; 명시 override 보존; Latin 문자만으로 영어 지원 인증하지 않음 |
| 반복 요청 캐시 | [원본 cache 예제](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/examples/hooks/cache.py)를 참고해 원문·후보 순서·구현/모델 revision·온도/예산을 key에 포함하는 작은 별도 실험 | 부정·`<`/`<=`·후보 순서·revision 변화에서 miss; mutation/byte cap/eviction/lookup 비용 |

세 항목은 제안이다. 현재 공개 결과에 소급하지 않으며 캐시·속도·메모리·요금 절감을 주장하지 않는다. 기존 정규화 feature SHA만으로 캐시하면 구두점과 대소문자 의미가 충돌할 수 있다. 현재 Go 힌트가 이미 수십 µs라 캐시 조회가 더 비쌀 가능성도 비교해야 한다.

`answer_confidence=max(p)`는 이전 SDK pin에도 있었다. choice/score의 entropy 기반 `confidence`와 구분하고, noul의 confidence는 max(p)다. 어느 값도 checkpoint·선택지 수·도메인의 독립 보정 없이 성공 확률이 되지는 않는다. 과거 “confidence는 entropy” 설명은 해당 choice/score 필드에 한정해 읽는다. 과거 snapshot 자체는 보존한다.

이미 Go에 있는 choice/noul·잘림 검사·상주 ORT·BPE 캐시·none/ambiguous·fallback·JSONL을 새 엔진으로 다시 만들 필요는 없다. Go `Predict`의 score 지원이 없으므로 모든 원본 primitive가 이식됐다고 하지 않는다. 가변 BPE 캐시와 Predict/Close 수명을 보호하는 lock을 원본 Router의 lock 위치만 보고 제거하면 안 된다. 원본 Agent에도 별도 추론 read lock이 있다.

원본 English421M/multilingual322M은 전체 모델 크기이며 backbone은395M/307M이다. 작은3진 head를 만들어도 남긴 backbone 메모리는 사라지지 않는다. [다중 질문](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/laya/agent.py#L1011)은 상태 토큰화를 재사용하지만 질문별 encoder 입력은 남는다. 질문에 의존하는 hidden state를 상태 문자열만으로 캐싱하지 않는다.

## 벤치마크와 라이선스

원본의 [모델 라우팅 예제](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/examples/29_presets_model_router.py)는 정책 예제다. [앱 벤치마크](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/research/scripts/bench_apps.py#L243)의399건 routing은 GSM8K/MBPP/AG News 도메인 분류이며 Codex 완료·비용 평가가 아니다. [원본 결과 문서](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/BENCHMARKS.md#L168)는 일부 fine-tuned 행의 committed result 부재와 낮은 base 결과도 명시한다. 다른 GPU에서 보고된 시간은 우리 Mac/Go 성능으로 옮기지 않는다.

[SDK LICENSE](https://github.com/NandhaKishorM/laya/blob/4aa6761be8173de4ce6d92c31b3e40b6eaf59a7c/LICENSE)는 Apache-2.0이다. 코드 복사/번역 시 LICENSE·관련 고지·변경 표시를 보존한다. 모델은 별도 카드 선언을 확인했다.

| 카드 | 확인 revision | 카드 선언 |
|---|---|---|
| [English](https://huggingface.co/convaiinnovations/laya/blob/55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851/README.md) | `55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851` | Apache-2.0 |
| [Multilingual](https://huggingface.co/convaiinnovations/laya-multilingual/blob/e4e9ddf21a7b1903b7acffd8814ad4307bf63a67/README.md) | `e4e9ddf21a7b1903b7acffd8814ad4307bf63a67` | Apache-2.0 |
| [Typed decisions](https://huggingface.co/convaiinnovations/laya-typed-decisions/blob/1a793eb568e6718f15941d08f85432581df534e3/README.md) | `1a793eb568e6718f15941d08f85432581df534e3` | Apache-2.0 |

이는 공개 선언 확인이며 학습 데이터 권리 전체를 인증한 결과는 아니다. 공개 teacher 라벨을 우리의 독립 정답으로 수입하지 않는다. 새 유료 모델 호출·fit·가중치·최종 평가0, 기존 최소15그룹/필요 효용5%/도메인별 고유 보호 최종2400개 기준을 유지한다.
