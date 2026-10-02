# 공개 함수 동작 검증과 학습 자료 확장

UUID Parse·Scan·Ordinal의 24개 입력이 실행 전에 정한 기대값과 모두 일치했습니다. 재시도 없이 한 번 실행했고 API 24회와 오류 메시지 확인 5회, 총 29회가 기록됐습니다. [전체 결과와 근거](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/native-observation-80/ACTUAL-RESULTS.ko.md)를 보세요.

작은 주장 모델에 줄 근거를 준비하는 과정입니다. “오류가 나면 값은 모두 0”이라는 설명은 실제 UUID의 부분 반환값과 다릅니다. 이런 차이를 확인한 뒤 요청·후보 주장·학습 정답으로 연결해야 합니다. 현재 새 학습 정답·역할·부모 요청은 0이고 24개 입력을 독립 과제 24개로 세지 않습니다.

전체 프로그램 RSS17.265625 MiB·wall1.21초는 초기화·해시 검사·JSON·저장까지 포함합니다. 개별 함수나 모델 추론 비용이 아닙니다. Laya·GPU 추론과 Go heap은 측정하지 않았고 RAM 절감 비교도 없습니다.

[다음 원문 자료](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/source-observation-81/README.ko.md)는 네 목표와 입력 초안 32개를 담습니다. 원문·기대값·실제 동작 검증 후 학습 자료를 설계합니다. 기존 두 학습 모델은 기준선보다 확인 작업이 늘어 비활성 상태이며, 2,400개 독립 요청의 최종 평가도 미완료입니다.

원 BSD·MIT 고지를 보존하고 로컬 경로·바이너리·원시 로그는 제외합니다. [현재 도구 사용법](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/USAGE-56.ko.md) · [English](Source-Behavior-80-EN)
