# Conversion80 저장 결과 대조

저장된 후보 관찰 72개가 실행 전에 봉인한 72개 Want와 일치합니다. 이는 기존 3개 요청과 9개 코드 후보를 원 입력 24개에서 관찰한 한 번의 Go native 실행입니다. Laya, GPU, 새 모델 학습 또는 새 독립 요청 72개의 실험이 아닙니다.

검사기는 원 코드나 관찰기를 다시 실행하지 않았습니다. 각 행의 전체 literal, 11개 반환 채널, 입력의 nil/빈 값 구분, 후보·입력 순서, Scan wrapper argument receiver, panic 텍스트 미관측성을 대조했습니다. final·partial·stdout는 SHA `ee6b993dd5495e901ea9b516f96278eee76821bdb1e41ffaaac0ba0f2433ab8c`로 byte-exact 일치합니다.

후보 72개 dispatch 중 69개가 정상 반환하고, ParsePanicOnError 후보에서 예상된 3개 panic이 기록되었습니다. 반환 오류의 명시적 Error 관찰은 10개(URN 2개, 표준 errorString 8개)입니다. 따라서 측정한 callback은 82개입니다. 내부 원 API 56개와 startup 4개는 정적 경로 설명이며 동적 호출 수는 null로 유지됩니다. 오류 값 panic에 Error/String/Format을 추가 호출하지 않았습니다.

원본 20개를 포함한 폐쇄 파일 37개와 결과의 핀을 대조했습니다. 원 draft를 frozen false에서 true로 바꾼 단일 토큰 외 다른 bytes는 같습니다. 초기 예약은 static56, 실제 카운터 0 및 내부/init null입니다. 부모 시작은 1회, retry 0, exit0, Wait joined, timeout/overflow false입니다.

저장된 OS 기록의 whole-child RSS는 19,529,728 B(18.625 MiB), footprint는 16,728,544 B입니다. OS real/user/system은 2.84/0.34/0.18초이고 바깥 controller wall은 2.8445567499999997초입니다. startup·loader·동기 저장·후보를 모두 포함하므로 순수 함수 비용이 아닙니다. Go heap 256 MiB는 soft 설정이며 총 RSS hard cap을 뜻하지 않습니다.

stdlib metadata 검사 1회 성공, 실패 0입니다. receipt의 checks_completed 6,571은 decode·형태·핀 대조 등 기계적 검사 횟수이며 새 의미 정답 수가 아닙니다. 마지막 출력 fsync까지 포함한 terminal count는 6,575입니다. 본 검토자는 observer/controller 저자이며 결과를 미리 본 nonblind 상태입니다. 새 독립 source oracle, 독립 실행자, 블라인드 의미 검토 또는 학습 승인으로 주장하지 않습니다. 원 Wanted의 truth/role/weight/mask/label/Got null과 준비 flags는 그대로입니다. 부모의 후속 SAT/eligible 판단은 별도 기록입니다.

[결과 대조](qa/NUMERIC.v1.json) · [영수증](qa/RECEIPT.v1.json) · [시도 장부](qa/LEDGER.v1.json)

이 폴더는 공개 준비 stage입니다. 공유 Git·외부 게시·원 코드 재실행을 하지 않았습니다. [실제 결과](actual/results.json)와 [봉인 Want](want/WANTS.v1.json), [observer 인계](observer/HANDOFF.v2.ko.md), [원 skeleton](skeleton/SKELETON.v1.ko.md)을 각각 보존합니다. 인계의 pending·초기 예약 카운터0·preflight 실행0은 실행 전 역사 기록입니다. [Root 실제 장부](actual/ROOT-ACTUAL-LEDGER.v1.json)는 완료를 별도로 기록하며 기존 generic76 schema를 재사용한 원문입니다. 이번 대상은 conversion80입니다.

[COPY ledger](COPY-LEDGER.v1.json)는 byte-exact 복사와 링크만 바꾼 파생 Markdown을 구분합니다. 파생 문서의 원 bytes는 `.md.txt`에도 있습니다. [MANIFEST](MANIFEST.v1.json), [SHA256SUMS](SHA256SUMS), [제외 목록](OMISSIONS.v1.json)은 원 SHA/공개 SHA와 제외 이유를 추적합니다. PRIVATE plan·binary·helper·raw OS/test log는 SHA-only입니다. Go/module은 `.go.txt`/`.mod.txt`로 비활성 보관하며 전체 MIT/BSD/Apache notices와 원문 범위는 [NOTICE](NOTICE.md)에 명시합니다. [English](README.en.md)

Root 후속 SAT·caption fidelity·eligible·corpus append 자료는 별도이며 observer snapshot을 소급 변경하지 않습니다. 실제 128/256개 독립 개발 문제 또는 도메인별 2,400개 보호 final 달성 주장도 없습니다. 준비 helper 작성의 문자열 syntax 실패 1회는 [원 장부](AUTHORING-FAILURE.v1.json)에 보존하며 native 실행 실패가 아닙니다.

원 exact-copy helper는 모든 원본 복사 후 아직 미생성인 MANIFEST 링크를 검사하여 metadata 단계에서 1회 중단했습니다. [실패 장부](COPY-FAILURE.v1.json)와 [별도 최종 장부](COPY-FINALIZATION.v2.json)에 보존합니다. 기존 COPY ledger의 실패0은 실패 전 작성된 역사 snapshot이며, 최종 상태는 새 장부가 설명합니다. 원 코드·관찰기 재실행과 저장 QA 재실행은 없습니다.
