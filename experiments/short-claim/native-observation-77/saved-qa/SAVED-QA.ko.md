# 77 v2 저장 관찰 기록 대조

저장된 24개 관찰은 실행 전에 봉인한 기대값과 모두 일치했습니다. 이번 검사는 원본 함수나 관찰기를 다시 실행하지 않고, 저장 JSON·파일 해시·카운터·기록된 OS 수치만 대조했습니다. Go 1.27.1 표준 라이브러리 검사기를 한 번 실행해 통과했으며 실패와 재시도는 0입니다.

| 확인 항목 | 저장 기록과 대조 결과 |
|---|---|
| 유한 관찰 | logfmt 12개, Escape 4개, Unescape 8개; 24/24 기대값 일치 |
| 직접 추적 콜백 | 예약 217, 반환 217, panic 0; 원 패키지 공개 호출 209 + 표준 라이브러리 오류 문자열 호출 8 |
| EOF 이후 | 오류 없는 EOF 7건은 후속 반복 호출을 생략하고 결과를 null로 보존 |
| 오류 유지 | 오류가 있는 logfmt 5건의 반복 관찰과 오류 타입·문자열·위치를 보존 |
| 바이트 관측 | nil/빈 값, 길이와 hex, 성공한 key/value 쌍 15개, 순서·중복·부분 기록이 봉인한 값과 일치 |
| 파일 연결 | fixture 전체와 Want/Got 24건, source pins 26개, zero-call 예약 기록, 원 v1 Want 해시를 확인 |
| 실행 기록 | root Start 1, joined Wait, exit 0, 재시도·timeout·출력 초과 0 |

각 fixture의 모든 JSON 필드를 대조했습니다. 숫자는 부동소수점으로 바꾸지 않고 JSON 숫자 문자열로 보존해 비교했습니다. 78개 바이트 상태와 24개 오류 상태는 같은 오류가 여러 관측 위치에 나타난 경우를 포함하는 관측 횟수입니다. 서로 다른 78개 표본이나 24개의 오류 사례를 뜻하지 않습니다.

최종 `results.json`, `partial.json`, root가 저장한 표준출력은 43,604 bytes로 완전히 같으며 SHA256은 `e5305785e8aca293b8dcb5d9f684ca824e0aeb05b10bbe568e387357b941a99a`입니다. 실행 계획은 초안의 `frozen: false` 한 토큰만 `true`로 바뀌었습니다. root가 76용 일반 실행 장부 스키마를 재사용한 사실을 유지하고, 77의 실제 인자·최소 환경·바이너리·계획·기대값 해시로 연결했습니다. 장부 스키마 이름만으로 서로 다른 실험이 같다고 판단하지 않았습니다.

기록된 Darwin 전체 child 프로세스 수치는 최대 RSS **29,835,264 bytes = 28.453125 MiB**, peak footprint 22,807,104 bytes, OS real 5.20초·user 0.17초·system 0.46초입니다. controller wall은 5.210141667초이며 worker 내부의 완료 전 snapshot은 4.817438167초입니다. worker heap snapshot은 4,083,152 bytes, heap system은 24,707,072 bytes입니다. 누적 할당 51,355,616 bytes는 동시 상주 메모리가 아닙니다. 프로세스 시작·입출력·매 호출 전후 저장 비용이 포함되어 있으며 순수 API 시간이나 startup 시간을 차감하여 계산하지 않았습니다. soft Go heap 설정도 OS RSS의 하드 제한이 아닙니다.

이 실행은 **Go native CPU 관찰**입니다. Laya 모델이나 GPU 추론이 아닙니다. 추적한 217개는 지정한 공개 호출과 오류 문자열 호출이며, 내부 helper·표준 라이브러리 전체 호출·프로세스 초기화 비용을 전부 계수한 값은 아닙니다. 원 package 초기화는 main 이전에 일어나며 동적으로 계측되지 않았습니다.

검사자는 앞선 source-first/Want 검토 작성자이며 결과 요약을 이미 본 상태입니다. 관찰기와 root controller 작성자는 아니지만, 이 저장 기록 대조를 블라인드 의미 평가나 독립 runtime 재현으로 주장하지 않습니다. 원 v1의 EOF 사용 오류 분석과 해시는 그대로 남았고, 그 오류를 실제 panic으로 재현했다고 주장하지 않습니다.

두 행동 목표의 24개 유한 관찰은 24개 독립 부모 요청·학습 라벨·역할이 아닙니다. 새 모델 호출, 학습, Features, Project, 역할 배정, 원 API·관찰기·테스트 재실행, 공유 저장소·원격 수정은 모두 0입니다. `qualification`, `training_ready`, `production_ready`, `protected_final`은 계속 false입니다. 보호된 도메인별 최종 2,400개 목표의 일반화 증거를 이번 관찰로 대신하지 않습니다.

수치와 파일 연결은 [NUMERIC-RESULT.v1.json](NUMERIC-RESULT.v1.json), [RECEIPT.v1.json](RECEIPT.v1.json), 실행 이력은 [ATTEMPT-LEDGER.v1.json](ATTEMPT-LEDGER.v1.json)에 남았습니다. 검사기의 호스트 경로가 들어간 소스와 바이너리, 원 OS 로그·비공개 계획은 공개 묶음에서 제외합니다.
