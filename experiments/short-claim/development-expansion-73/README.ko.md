# 다음 실험은 새 동작 문제부터

두 학습 방식의 확인 비용은 기준27회 대비31/33회로 실패했다. 이번에는 학습 방식·feature·threshold를 다시 바꾸기 전에 다른 공개 코드의 실제 동작을 검증할 개발 자료를 늘린다.

[제안 목록과 우선순위](NEXT-DEVELOPMENT-SOURCES.proposal.v1.ko.md)는 새7저장소·서로 다른11동작 목표다. 새 부모·정답·실행 완료 계약은 아직0개다. 첫 후보는 datasize의 단위/오류 상태, query.Values의 태그/다중값, shlex.Split의 인용/EOF다. 일반 문장을 정답으로 삼지 않고 유한 입력의 값·오류·상태를 확인한다.

128→256→512는 더 많은 독립 의미 목표를 확보할 방향이며 현재11개를 리터럴 반복으로 채우지 않는다. 기존76개 회귀와 새 개발 자료, 별도 도메인별2,400개 최종 요청을 구분한다. 전체 원천·별칭·공유 helper·작성 흐름으로 묶어 역할 간 누출을 검사하고 미확정 정답은unknown으로 둔다.

[메모리 검토](../memory-scale-73/ANALYSIS.ko.md)에서 현재 구조는 이미CSR/SoA·uint16임을 확인했다. 먼저 숫자를 보존하는 작은 저장 형식·분할 입력을 검토한다. 모델 압축·새 objective·성능 승인을 이 준비와 합치지 않는다.

첫3원천은 이후 [실제 관찰](../source-observation-73/README.ko.md) 1회에서24유한 입력이 사전 Want와 일치했고 저장 결과 대조도 통과했다. 3목표와 미확정3부모 draft를 구분하며 새 정답·학습은 아직0이다. [저장 형식 실험](../compact-storage-73/README.ko.md)에서는 숫자를 그대로 보존하고 파일 크기 기준선을 비교했다. Runtime·loader나 입력 한도를 바꾸지는 않았다.

[실패 분석](../second-ranking-fit-72/ANALYSIS.ko.md) · [English](README.en.md)
