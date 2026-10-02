# 다음 학습 자료: 실제 동작과 설명을 구분하기

작은 모델 두 방식이 확인 비용을 줄이지 못해 다음에는 자료 다양성을 늘립니다. datasize의 단위·오류 후 상태, querystring의 태그·다중값, shlex의 인용·불완전 입력이라는3개 동작 목표를 골랐습니다.

실행 전에24개 입력과 예상 결과를 고정하고 독립 원문·실행 경계 검토를 거쳐 실제 Go 원문을 한 번 실행했습니다.24개 모두 예상과 일치했으며, 큰 정수·오류8개·nil/빈 결과·부분 반환을 저장 파일 검사로 다시 확인했습니다. 전체 child peak RSS는약17.9MiB이며 모델 추론 비용은 아닙니다.

**24개 입력은 독립 문제24개가 아닙니다.** 다음3개 부모×3후보 설계는 아직 정답·역할·가중치가null입니다. 원문 주석이나 signature가 오류·상태 조건을 설명하지 못하므로, 관찰이 맞았다고 모델 입력에 충분한 정보가 있다고 결론내리지 않습니다. 한국어와 영어 표현도 같은 부모로 묶습니다. 실제 새 정답·학습·최종2,400 평가 완료는0입니다.

후속74에서는 같은9개 후보의 소스 근거 설명을 준비하고 독립 검토했습니다. 원문 설명 A와 새 설명 B는 같은 부모이며,8개 새 문구는 범위 내 근거가 있고 query 문구는 요청 조건을 더 좁혀야 합니다. 코드의 충족 여부와 모델이 볼 수 있는 정보는 별도입니다. 실제72는 같은 SampleWeights를 사용해 가중치0인 endpoint의 pair 손실·gradient도0으로 제외합니다. 감사 목록에 남긴 쌍이 모두 학습에 기여하는 것은 아닙니다. logfmt·dataurl·wordwrap·godotenv의 다음 범위와 전체 MIT·Go BSD 고지도 조사했습니다. [설명 검토·정정 기록](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/source-caption-74) · [다음 네 원천](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/source-audit-74). 새 정답·학습·보호 final 수는 늘리지 않았습니다.

[상세 결과·사용 범위·원문·검증 기록](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/source-observation-73) · [English](Source-Observation-73-EN) · [저장 형식 실험](Compact-Storage-73-KO)
