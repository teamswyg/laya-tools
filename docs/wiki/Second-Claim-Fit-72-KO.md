# 두 번째 작은 주장 모델: 개선보다 먼저 확인한 실패

작은 모델은 확인할 코드 후보의 순서를 제안합니다. 최종 정답·실행·승인은 실제 검증 과정이 맡습니다. 기본 Go 정렬을 그대로 사용할 수 있으며 모델이나 Codex 연결은 필수가 아닙니다.

이번에는 후보별 정답 학습에 “같은 요청의 좋은 후보를 더 앞에 두기” 목표를 더했습니다. 같은 검증 요청에서 후보 확인은 **기본27회 → 첫 모델31회 → 두 번째33회**, Top1 정답은 **5→2→1**이었습니다. 효용에 실패해 기본 설정에 적용하지 않습니다.

학습에 실제 기여한 요청은16개뿐이었습니다. 훈련 손실은 줄었지만 검증에서는 관계를 거의 구별하지 못했습니다. 다음은 서로 다른 공개 코드의 동작·표현을 개발 자료에 추가하는 작업입니다. 기존76개는 회귀 자료로 유지하고, 별도 도메인별2,400개 최종 요청은 모델 선택에 쓰지 않도록 분리합니다. 작은 모델의 압축이나 학습 반복이 이 문제를 해결했다고 주장할 근거는 아직 없습니다.

CPU에서 한 번 학습했으며 계수 파일은32,792B, 전체 실행 peak RSS는약27.66MiB였습니다. 학습·검증·저장을 포함한 관측이며 Laya encoder·GPU 추론·Codex 토큰 절감 수치가 아닙니다. 모델 본체는 Git에 올리지 않습니다.

[결과·수치 검토](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/second-ranking-fit-72/README.ko.md) · [다음 실험의 이유](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/second-ranking-fit-72/ANALYSIS.ko.md) · [Go 입력 재사용](Validated-Input-71-KO) · [English](Second-Claim-Fit-72-EN)
