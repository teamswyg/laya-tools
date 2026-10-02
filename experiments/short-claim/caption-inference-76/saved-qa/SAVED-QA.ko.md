# 76 저장 결과 대조

저장된 점수·순위·호출 장부·입력 핀의 대조는 통과했습니다. 새 모델 추론이나 원 API 실행 없이 Go 표준 라이브러리 검사기 한 번으로 계산했습니다. 모델 결과가 좋았다는 뜻은 아닙니다.

같은 세 요청에 각각 세 후보가 있으며, A/B는 후보 설명의 두 버전입니다. 요청 3개·후보 위치 9개·설명 18개를 유지했습니다. 모델 점수 36개는 36개의 독립 요청이 아닙니다. 원래 계획의 라벨·가중치·역할·그룹은 계속 `null`이며, 별도로 고정한 개발 진단용 rubric의 허용 후보 위치는 각 요청에 대해 `[0]`, `[1]`, `[2]`입니다. 이 검사는 그 rubric의 의미적 정확성을 새로 승인하지 않습니다.

`checks`는 허용 후보에 도달할 때까지 검사할 후보 수입니다. 작을수록 좋습니다. 아래 값은 저장 점수를 내림차순으로 정렬하고 동점에는 원래 후보 순서를 유지해 다시 계산했습니다. Top1의 분모는 각 arm의 같은 3개 answerable 요청입니다.

| 모델/기준선 | A checks | B checks | A Top1 | B Top1 |
|---|---:|---:|---:|---:|
| 실패 모델 71 | 3 | 8 | 3/3 | 0/3 |
| 실패 모델 72 | 5 | 6 | 1/3 | 1/3 |
| fixed_order | 6 | 6 | 1/3 | 1/3 |
| BM25 | 8 | 3 | 0/3 | 3/3 |
| lexical_ordered | 8 | 3 | 0/3 | 3/3 |
| narrow_rule | 8 | 3 | 0/3 | 3/3 |

모델 71의 B−A checks는 +5, 모델 72는 +1입니다. narrow_rule은 같은 6개 요청/arm 조합에서 전부 `unsupported_rule_request`로 BM25에 fallback했고, 저장 점수와 순위도 BM25와 동일했습니다. 각 arm의 전체 checks가 가장 적은 기준선을 고정 목록 순서로 고르면 A는 fixed_order, B는 BM25입니다. 이 비교로 새 모델이나 arm을 선택하지 않았습니다. Top3는 전부 3/3이지만 후보가 세 개뿐이어서 이 값은 구별력이 없습니다.

검사 범위는 모델 부모 순위 12개, 기준선 부모 순위 24개, 저장 scalar 점수 108개, 모델 점수 행 연결 36개입니다. 실제 동점 22쌍도 원래 순서로 유지됐습니다. 후보를 가중치/마스크로 삭제하지 않았으며, 각 모델·arm에서 모든 9개 후보를 점수화한 저장 기록을 유지했습니다. 실제 counts는 모델 읽기/Decode 2, Prepare/Baselines 6, Features/Score 36, Fit/Project/원 API/훈련 라벨 배정/유료 judge 0입니다. 이것은 저장 장부 대조이며 내부 API 호출을 새로 계측한 결과는 아닙니다.

stdout 결과, 최종 결과, partial 결과는 17,225 B로 정확히 같았습니다. 원 plan의 31,738 B 공개 사본도 바이트가 같고 18개 설명의 길이/SHA가 맞습니다. 원 plan의 `frozen_for_execution=false`는 원 기록 그대로 남아 있습니다. 실제 호출은 별도의 고정 rubric와 실행 전 root 인자/장부에 결합돼 있습니다. zero-call reservation과 완료 결과를 구분했고 재시도는 0입니다.

OS 기록은 전체 자식 프로세스의 maximum RSS 10,551,296 B = 10.0625 MiB, peak footprint 8,258,040 B, real 2.23 s, user 0.05 s, system 0.17 s입니다. controller wall은 2.24256375 s, worker wall snapshot은 최종 쓰기 전 1.853498208 s입니다. Go heap snapshot은 OS RSS와 다릅니다. 이 값들로 순수 추론·시작·pipe 시간을 빼서 추정하지 않았습니다.

검사자는 이번 worker의 작성자가 아니지만 이전 source/controller 작업과 결과 노출이 있어 블라인드 검토가 아닙니다. 모델 파일은 읽지 않았고 Decode·Prepare·Features·Score·Baselines·Project·Fit·worker/controller·원 API·테스트를 재실행하지 않았습니다. 별도 후보/API 실행, 새 학습 라벨·역할·가중치·모델은 0입니다. `qualification`, `training_ready`, `production_ready`, `protected_final`은 모두 false입니다. 독립 final 일반화나 비용 절감 증거가 추가되지 않았습니다.

검증 가능한 출력은 [수치 대조](NUMERIC-RESULT.v1.json), [영수증](RECEIPT.v1.json), [사전 실행 범위](INVOCATION.v1.json), [시도 장부](ATTEMPT-LEDGER.v1.json)입니다. private 검사기 소스·바이너리와 원 로그/호스트 경로는 공개 기록에 포함하지 않습니다.
