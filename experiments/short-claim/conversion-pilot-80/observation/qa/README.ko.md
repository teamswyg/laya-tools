# Conversion80 저장 결과 대조

저장된 후보 관찰 72개가 실행 전에 봉인한 72개 Want와 일치합니다. 이는 기존 3개 요청과 9개 코드 후보를 원 입력 24개에서 관찰한 한 번의 Go native 실행입니다. Laya, GPU, 새 모델 학습 또는 새 독립 요청 72개의 실험이 아닙니다.

검사기는 원 코드나 관찰기를 다시 실행하지 않았습니다. 각 행의 전체 literal, 11개 반환 채널, 입력의 nil/빈 값 구분, 후보·입력 순서, Scan wrapper argument receiver, panic 텍스트 미관측성을 대조했습니다. final·partial·stdout는 SHA `ee6b993dd5495e901ea9b516f96278eee76821bdb1e41ffaaac0ba0f2433ab8c`로 byte-exact 일치합니다.

후보 72개 dispatch 중 69개가 정상 반환하고, ParsePanicOnError 후보에서 예상된 3개 panic이 기록되었습니다. 반환 오류의 명시적 Error 관찰은 10개(URN 2개, 표준 errorString 8개)입니다. 따라서 측정한 callback은 82개입니다. 내부 원 API 56개와 startup 4개는 정적 경로 설명이며 동적 호출 수는 null로 유지됩니다. 오류 값 panic에 Error/String/Format을 추가 호출하지 않았습니다.

원본 20개를 포함한 폐쇄 파일 37개와 결과의 핀을 대조했습니다. 원 draft를 frozen false에서 true로 바꾼 단일 토큰 외 다른 bytes는 같습니다. 초기 예약은 static56, 실제 카운터 0 및 내부/init null입니다. 부모 시작은 1회, retry 0, exit0, Wait joined, timeout/overflow false입니다.

저장된 OS 기록의 whole-child RSS는 19,529,728 B(18.625 MiB), footprint는 16,728,544 B입니다. OS real/user/system은 2.84/0.34/0.18초이고 바깥 controller wall은 2.8445567499999997초입니다. startup·loader·동기 저장·후보를 모두 포함하므로 순수 함수 비용이 아닙니다. Go heap 256 MiB는 soft 설정이며 총 RSS hard cap을 뜻하지 않습니다.

stdlib metadata 검사 1회 성공, 실패 0입니다. receipt의 checks_completed 6,571은 decode·형태·핀 대조 등 기계적 검사 횟수이며 새 의미 정답 수가 아닙니다. 마지막 출력 fsync까지 포함한 terminal count는 6,575입니다. 본 검토자는 observer/controller 저자이며 결과를 미리 본 nonblind 상태입니다. 새 독립 source oracle, 독립 실행자, 블라인드 의미 검토 또는 학습 승인으로 주장하지 않습니다. 원 Wanted의 truth/role/weight/mask/label/Got null과 준비 flags는 그대로입니다. 부모의 후속 SAT/eligible 판단은 별도 기록입니다.

[결과 대조](NUMERIC.v1.json) · [영수증](RECEIPT.v1.json) · [시도 장부](LEDGER.v1.json)
