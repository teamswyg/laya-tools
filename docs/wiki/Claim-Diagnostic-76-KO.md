# 작은 힌트 모델: 설명을 보강하면 좋아질까?

[English](Claim-Diagnostic-76-EN) · [시작하기](Getting-Started-KO)

이번 검사는 “같은 코드 후보를 더 충실하게 설명하면 작은 모델이 확인 순서를 잘 고를까?”를 살펴봤습니다. 요청3개와 후보9개를 고정하고, 원래 주석·함수형 설명 A와 소스 동작을 설명한 B를 비교했습니다. 앞서 효용 검사를 통과하지 못한 두 모델을 그대로 사용했습니다. 재학습이나 기본 라우터 변경은 없습니다.

| 순서를 정하는 방식 | A의 모의 검사 수 | B의 모의 검사 수 |
|---|---:|---:|
| 첫 FP32 모델71 | 3 | 8 |
| 순위 손실을 더한 모델72 | 5 | 6 |
| 원래 순서 | 6 | 6 |
| BM25 | 8 | 3 |
| 단어·연속 두 단어 비교 | 8 | 3 |

낮을수록 미리 정한 수용 후보가 앞에 있습니다. 이 숫자는 **정답 기준으로 계산한 순위 위치의 합**입니다. 실제 검증기·Codex·LLM을 그 횟수만큼 실행한 결과가 아닙니다. 세 후보 중 Top3 성공은 항상3/3이라 성능 지표로 유용하지 않습니다.

이 작은 범위에서는 B가 단순 검색을 돕고 기존 모델은 악화시켰습니다. B는 길이와 문장 구성도 달라져 정보 보강만의 인과 효과를 입증하지 않습니다. 모델을 다시 학습하기 전에 반환과 panic, 조건부 상태 변경, 부정 표현처럼 단어가 비슷해도 동작이 다른 공개 사례를 더 확보하려 합니다. 비공개 최종 평가나 도메인별2,400개 목표를 완료한 것은 아닙니다.

실제 실행은 한 번이며 두 모델의 특징·점수 계산36회를 완료했습니다. 전체 자식 프로세스의 최대 RSS는10.0625MiB, 경과 시간은2.23초입니다. 시작·파일 검증·기준선·결과 저장을 포함하므로 한 번의 추론 시간이나 다른 실험 대비 RAM 절감으로 해석하지 않습니다. GPU·유료 모델 호출은 없습니다.

병렬 자료 작업에서는 wordwrap·godotenv의 두 행동 목표를 실제 Go로 관찰했고24개 유한 입력이 사전 기대값과 일치했습니다. 이24개는 독립 요청24개나 새 학습 정답이 아닙니다.

사용자는 현재 모델 다운로드 없이 [작은 동작 힌트 도구](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/USAGE-56.ko.md)를 사용해 모든 후보를 남긴 확인 순서를 받을 수 있습니다. 최종 코드 확인과 작업 실행은 별도로 합니다. 실패 모델을 기본값으로 활성화하지 않습니다.

[실제 비교·점수·검증·라이선스](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/caption-inference-76) · [함수 관찰75](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/native-observation-75) · [첫 모델의 불변 HF 보관본](https://huggingface.co/JooYoon/riidolaya-shortclaim-fp32-failed-71/tree/32b8f4579065247be0f71c83e1e7c143845e7b33) · [두 번째 모델의 불변 HF 보관본](https://huggingface.co/JooYoon/riidolaya-shortclaim-rank-bce-failed-72/tree/d090b00e9a5dab00d5372dfd6412c9aee0c60b7b)
