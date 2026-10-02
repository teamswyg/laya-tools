# Wrap/Marshal 75: 공개 관찰 기록

공개 Go 원천의 두 함수에 대해 사전에 고정한 24개 기대값을 실제 실행 결과와 비교했습니다. `wordwrap.WrapString` 12개와 `godotenv.Marshal` 12개 모두 기대값과 일치했습니다. 원 함수 반환은 24회, 차이와 panic은 0회이며, 자식 실행은 한 번이고 재시도는 없었습니다.

이번 실행은 Go 1.27.1의 네이티브 CPU 프로그램입니다. Laya 모델이나 GPU를 실행한 실험이 아닙니다. 유한 입력에서 원천 동작과 기대값을 연결하는 준비 자료이며, 24개 독립 요청이나 학습 레이블을 새로 확보했다는 뜻이 아닙니다. 모델의 일반화, 캡션 자격, 학습 준비 또는 사용량 절감을 증명하지 않습니다. 원 기록의 관련 자격 플래그는 모두 false로 유지했습니다.

| 관측값 | 결과 | 범위 |
|---|---:|---|
| 고정 기대값 일치 | 24 / 24 | Wrap 12개 + Marshal 12개 |
| 차이 / panic | 0 / 0 | 고정된 유한 입력만 |
| 실행 / 재시도 | 1 / 0 | 한 번의 자식 프로세스 |
| 최대 RSS | 18,677,760바이트 = 17.8125 MiB | 자식 전체의 최대값 |
| 최대 footprint | 16,089,616바이트 | RSS와 다른 OS 항목 |
| CPU user / system | 0.01 / 0.04초 | 자식 전체 |
| OS real | 0.87초 | 자식 전체 |
| 컨트롤러 경과 시간 | 0.879832625초 | 별도 바깥 실행 기록 |

OS 수치는 패키지 초기화, 사전 검증, 래퍼와 함수 호출, 결과 및 체크포인트 쓰기를 포함합니다. 순수 함수 지연 시간은 측정하지 않았습니다. CPU 1개 설정과 256 MiB Go soft heap 설정은 전체 RSS의 하드 상한이 아닙니다. 이 한 번의 관찰에서 성능 개선이나 일반적인 저자원 동작을 추론하면 안 됩니다.

읽는 순서는 [사전 기대값](finite-want.v1.json), [실제 결과](observations/results.json), [바깥 실행 장부](root/ROOT-ACTUAL-LEDGER.v1.json), [저장 결과 QA](saved-qa/SAVED-RECORD-QA.ko.md)가 좋습니다. nil 오류, nil 맵과 빈 맵, 입력 맵의 호출 전후 상태, 64자리 정수 문자열을 구분해 보존했습니다. 최종 결과와 마지막 부분 결과는 바이트까지 동일합니다.

[원천을 먼저 읽은 검토](source-peer/SOURCE-ORACLE.v1.ko.md)는 관찰기 작성자와 다른 AI 협업자가 수행했습니다. [컨트롤러 검토](controller-peer/FINDINGS.v1.ko.md)의 검토자는 관찰기 작성자입니다. [저장 결과 QA](saved-qa/SAVED-RECORD-QA.ko.md)는 컨트롤러 작성자가 수행했습니다. 이 역할 차이와 사전 노출을 독립적인 모델 평가나 블라인드 검증으로 바꾸어 설명하지 않습니다. 별도 모델 호출 0회는 협업 에이전트의 비용까지 0이라는 뜻이 아닙니다.

[공개 파일 장부](PUBLICATION-MANIFEST.v1.json)에는 exact copy, derived copy, omitted를 구분했습니다. exact copy는 원문 바이트와 SHA가 같습니다. 두 실행 계획의 파생본은 개인 경로 두 값만 공개용 표기로 교체한 비실행 자료입니다. 원 SHA를 함께 남겼으며 원 실행에 사용한 계획의 해시를 대신하지 않습니다. 바이너리, 개인 경로가 들어간 테스트·도구, 원시 로그는 공개하지 않았습니다. 이 묶음의 준비는 관찰기 재실행이나 원격 게시를 포함하지 않습니다.

원천과 MIT 고지는 PR99의 고정 커밋 [0a25491](https://github.com/teamswyg/laya-tools/tree/0a25491e83df23809c30d3497113e1d4969256b6/experiments/short-claim/source-audit-74/upstream)에 보존되어 있습니다. [go-wordwrap/LICENSE.md](https://github.com/teamswyg/laya-tools/blob/0a25491e83df23809c30d3497113e1d4969256b6/experiments/short-claim/source-audit-74/upstream/go-wordwrap/LICENSE.md)와 [godotenv/LICENCE](https://github.com/teamswyg/laya-tools/blob/0a25491e83df23809c30d3497113e1d4969256b6/experiments/short-claim/source-audit-74/upstream/godotenv/LICENCE)의 전체 고지를 유지합니다. [원천 참조와 파일 해시](SOURCE-REFERENCES.v1.json)는 실제 경로와 원문 SHA를 연결합니다. `.go.txt` 파일은 검토용 원문이며 제품 실행 인터페이스가 아닙니다.
