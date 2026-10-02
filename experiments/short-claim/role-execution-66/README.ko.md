# 원래 72개 요청의 실제 역할 배정66

[English](README.en.md) · [실제 결과](results-66.json) · [독립 검사](FINDINGS-RUNTIME-66.v1.ko.md)

CI를 통과한 Go 역할 모듈을 고정 seed1729로 한 번 실행했다. 연결된 코드·설명·요청 가족을 통째로 배정하며, 원래 72개 요청·216개 후보·정답·순서를 보존한다. 라벨을 가진 16개 그룹은 train10·validation3·calibration3이다. Unknown만 있는 별도 그룹은 train provenance에 포함되어 전체 소속은11·3·3이다. Unknown을 학습 음성으로 바꾸지 않는다.

| 역할 | 요청 | 후보 | unknown 요청 |
|---|---:|---:|---:|
| train |44|132|14|
| validation |16|48|4|
| calibration |12|36|3|

실행은 첫 공식 시도1회·실패0·재시도0이었다. OS 전체 자식 프로세스 관측은 real0.41초, peak RSS17,448,960bytes(16.64MiB)이다. Go heap·모델 추론·GPU 사용량 측정은 별도이며 이번 결과에는 없다. 결과의 `TrainingReady=false`는 이 역할 자료만으로 전체 학습 준비가 완료되지 않았다는 뜻이다. 권한을 다시 요청하는 조건은 아니다.

[root 예약](ROOT-INVOCATION-RECEIPT-66.v1.json), [실행 장부](ROOT-ACTUAL-EXECUTION-LEDGER-66.v1.json), [worker 장부](worker-attempt-ledger-66.json), [원본 복사 장부](PUBLIC-COPY-LEDGER-66.v1.json), [QA 복사 장부](QA-COPY-LEDGER-92.v1.json)를 함께 읽는다. 이전 [실행 전 준비](../role-execution-preparation/README.ko.md)는 당시의 미실행 상태를 보존한다. 독립 검사는 이전 역할 구현 작성자의 비블라인드 기계적 대조이며 새로운 데이터 작성자 독립성을 주장하지 않는다.
