# 진행·완료보고·질문 표시 100개 독립 검토 기준

**범위:** 고정 Go v0.2 `ae6761dc81501b39ae17f29b48f0f0a8b35305fe11e297b9a787af0f9636df05`(32,960 B, 온도 0.9). V1 parent `c63b45571c866b779de90ebcfd1bf5be139064b578703ddfae23eb218553cf28`는 보존한다. Confidence 0.9·margin 0.05와 사전 품질 기준을 유지한다. 100개 결과로 가중치·온도·기준·정답·규칙을 바꾸거나 통과 기준을 낮추지 않는다. 실제 쓰기·알림은 실행하지 않는다.

## 1. 평가 전에 고정할 것

- 모델/소스/100개 원본/주석 기준/표시 매핑/지표·비용 가중치의 해시와 최초 실행 기록을 계획에 고정한다. 실행 후 수정은 새 버전이며 같은 100개를 미노출 평가로 재사용하지 않는다.
- 목표 정답은 `진행 / 완료보고 / 질문 / 표시없음`이다. 나머지 8의도 출력, 애매함·인용·가정·부정·복수 의도·문맥 부족도 제외하지 않고 `표시없음`의 오표시를 센다. 클래스·언어·대조 유형별 실제 분포를 공개한다.
- 학습/보정용 자료와 원문·사건·대화·번역/변형 수준의 겹침을 검사한다. 한 행마다 고유 group ID를 붙여도 통계적 독립은 증명되지 않는다. 작성자가 기존 실패를 알았는지와 주석자 간 이견을 기록한다.
- 별도의 자료로 온도를 맞추는 calibration label은 **작성자의 의도 주석**이다. 현업 완료·진행 사실이나 trusted event의 정답이 아니다. 이번 100개에는 보정을 다시 맞추지 않는다.

## 2. 세 축을 각각 관찰

현재 `Propose`는 저신뢰라도 version 누락/오류, reader 상태 계획 불가 또는 status 모호성을 `StateReason`에 먼저 기록한다. 완전한 metadata에서만 `Plan.Reason=below_threshold`가 보인다. **StateReason만으로 저신뢰 보류를 세면 누락된다.** Prediction이 남아 있으면 실제 confidence·margin·intent·source·guard로 아래 첫 축을 별도 계산한다. StateReason을 덮어 metadata 원인을 지우지 않는다.

| 기록 축 | 필요한 값 |
| --- | --- |
| 예측 자격 | `accepted / low_confidence / low_margin / unclear / untrained / guarded / not_predicted / predictor_error`; 복수 원인은 별도 flag로 보존 |
| 표시·매핑 | 요청된 3종의 정확한 label/emoji 매핑, 실제 새 제안, 이미 존재하여 no-op, 미연결/중복/inactive/group/허용목록 밖 제외 |
| 상태 metadata | 원래 StateReason·canonical version 유무·status 후보 모호성·reader 오류·snapshot 검증 실패; 동시 발생한 예측 보류와 독립 집계 |

Prediction 부재, reader/predictor 오류, missing work도 100개 모집단에 남긴다. 64개 요청 상한에 맞춰 나누되 batch 실패 사례를 성공 사례만 재실행해 덮지 않는다. 요청·반복 실행의 최초 결과와 실패 수를 보존한다.

## 3. 보고와 오표시 비용

- 4×4 혼동표와 **각 표시별** `그 표시가 맞은 제안/그 표시 전체 제안`(precision), `그 정답에서 맞게 표시한 수/그 정답 전체 사례`(정답 표시 coverage), 놓친 수를 적는다. 분모 0은 `null`이다. 전체 수락 coverage와 표시별 전체 100개 중 제안율도 별도로 보고한다.
- 예측 자격 coverage, 매핑 가능 coverage, 실제 새 표시 제안 coverage를 분리한다. 이미 같은 표시가 있는 no-op은 중복 억제 성공일 수 있으므로 새 표시 제안과 구분한다. 모호한 사례를 제거한 precision을 전체 품질로 제시하지 않는다.
- 잘못된 완료보고·질문 표시 비용을 진행 오표시와 별도 집계한다. 도메인이 사전 고정한 `w완료·w질문·w진행`으로 가중 오표시 비용을 계산하고 원래 오표시 건수도 함께 낸다. 값이 미정이면 임의 숫자로 통과 점수를 만들지 말고 미정으로 기록한다.
- 학습 모델의 NLL/Brier, confidence 구간별 정답/사례 수와 과확신 오답을 보고한다. 확률 공간을 고정한다: 4표시 평가의 `표시없음` 확률은 나머지 5의도 확률의 합이며 gate/no-op 결과를 확률 1로 만들지 않는다. 예측 부재의 수·계산 분모를 명시하고 확률을 꾸미지 않는다. 규칙 one-hot은 보정된 확률이 아니다. 상관된 100개로 제품 정밀도·독립 표본 신뢰구간을 보장하지 않는다.

## 4. 불변 조건과 공개 경계

모든 결과는 shadow, `MutationExecuted=false`, `StateChange=false`이며 임의 canonical version을 만들지 않아야 한다. Catalog 일관성은 `unqualified`; test double은 실제 조회 증거가 아니다. 개인정보·native ID·인증정보·실제 사용자 원문은 공개 결과에 저장하지 않고, 승인된 원본 허구 text 또는 최소 opaque 평가 ID와 점수·원인·집계만 남긴다. 완료보고 분류와 실제 완료 권한은 끝까지 별개다. 새 command/domain 구현 후 이 검토 기준과 실제 출력·오류 처리를 다시 대조한다.
