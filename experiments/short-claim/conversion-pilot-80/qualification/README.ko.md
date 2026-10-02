# 유한 요청 정답과 학습 역할 확정

Root가 사전 Want72와 실제 후보 결과 전체 채널을 대조하고, 원 요청·후보 코드와 설명을 따로 검토했습니다. 실제 관찰의 72건 일치는 기대한 동작의 확인이고, 요청 정답은 원래 요청과 기준 동작을 만족하는 후보입니다.

| 요청 | 수용 후보 index | 다른 후보의 실제 반례 수 |
|---|---|---|
| 부분 UUID 결과와 오류 유지 | 1 | index0:2, index2:3 |
| 기존 상태와 스캔 오류 유지 | 0 | index1:6, index2:2 |
| 영어 서수의 teen 예외 유지 | 2 | index0:4, index1:3 |

아홉 설명은 각각 전체 후보 함수의 동작과 후처리에 충실합니다. 그래서 틀린 동작을 설명한 후보도 알려진 오답으로 사용할 수 있습니다. 세 요청의 label은 정답3·오답6, loss eligibility는9개, 표본 가중치는 각1입니다. 각 근거는 [정답·설명·채널·적격성 검토](ROOT-ELIGIBILITY.v1.json)에서 분리했습니다. AI 보조 비맹검 Root 검토이며 독립 사람의 맹검 판단은 아닙니다.

[학습 추가 manifest](QUALIFIED-TRAIN-APPEND.v1.json)는 원76 projection와 실제70 role SHA를 묶고 새 parent index76,77,78만 추가합니다. 기존 그룹의 최댓값74 뒤에 UUID caller그룹75와 서수 caller그룹76을 만들었습니다. 이 번호는 parent index나 기존 role 재배정 결과와 다릅니다. UUID Parse·Scan·공통 helper·오류와 모든 변형은 하나로 연결하며 두 새 그룹은 development_train 역할입니다. 기존76개 정답·후보·mask·weight·역할과 검증·보정 자료는 바꾸지 않았습니다. 전체 roleplan.Assign을 다시 실행하지 않았습니다.

표준 Go 입력 검증·79개 투영은 아직 실행하지 않았습니다. 이 manifest는 그 다음 실행의 고정 입력이며 모델의 qualification·training_ready·production_ready는 false입니다. 원문을 이미 본 source53/55 노출을 기록했고, 기존 선언 그래프에 새 두 가족이 없다는 것을 원천 독립성이나 heldout 증명으로 해석하지 않습니다. 유한8입력/요청과 고정 소스 버전의 정답이며 최종2,400개 검증에 넣지 않습니다.

[저장 결과 검토](ROOT-SAVED-REVIEW.v1.json), [실행 전 검사](ROOT-PREFLIGHT.v1.json), [기존 후보 준비](../preparation/PILOT.v1.ko.md), [다음 고정 FP32 학습 계획](../../data-effect-fit-79/plan/PLAN.v1.ko.md)을 함께 보세요. 관찰 당시 Want·Got/null 감독 정보는 그대로 보존하고, 이번 별도 manifest에만 새 정답을 기록했습니다.
