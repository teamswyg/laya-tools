# Source81: 원문 동작 관찰과 저장 근거

Root가 4개 행동 목표/2개 원천 계열에서 32개 고정 입력을 한 번 관찰했습니다. 32개 모두 사전 Want와 일치했고 difference/panic/retry는 0입니다. 이는 함수의 유한 입력 관찰이며 32개의 독립 사용자 parent, 캡션 충실성, 학습 적격성, 모델 성능을 입증하지 않습니다. 신규 parent/label/role/weight는 0이고 qualification/training_ready/production_ready/protected_final은 false입니다.

결과 [final](actual/results.json)과 [마지막 partial](actual/results.partial.json)은 240,467바이트/SHA8f8a94dd…로 같고 외부 stdout도 동일했습니다. 고정 32개 입력·순서·Want/null과 Got 전체 채널을 보존합니다. 42개 pflag snapshot(nil slice2/nonnull empty3), time Date/Clock/Nanosecond/Zone4묶음, outer Error→Unwrap→cause Error6묶음이 사전 기대값과 일치합니다. 명시적 호출은 16종 API의 reserved/attempted/returned 각182회(원 public160+stdlib22)입니다. 동적 initializer 및 내부 registration/formatting/hook 계수는 미계측 null입니다.

전체 native child 측정은 real2.55초/user0.15초/sys0.18초, RSS19,611,648B(18.703125MiB), peak footprint16,679,416B이고 controller wall2.5557745초입니다. 초기화·소스 검증·JSON/fsync 포함 수치이며 함수별 비용이나 모델/GPU/Go heap 측정이 아닙니다. CPU1/Go softheap256MiB/바깥60초 guard를 사용했으나 RSS hard cap이나 엄격한 전체 controller 반환 시간을 보장하지 않습니다. 일반76 outside-controller를 재사용한 schema이며 이 실험의 binding은 [Root preflight](actual/ROOT-PREFLIGHT.v1.json), [Root ledger](actual/ROOT-ACTUAL-LEDGER.v1.json), [Root saved review](actual/ROOT-SAVED-REVIEW.v1.json)로 구분합니다.

관찰기 저자의 비맹검 [저장 기록 QA](observer-author-saved-QA/REVIEW.v1.ko.md)는 helper1회 성공, 원 API/worker/controller/tests/Features/Score/Fit 재실행0입니다. 별도 Want 작성자와 [원문/Want peer](literal-Want-peer/REVIEW.v1.ko.md)가 있지만 이 작성자 검사 자체를 독립 정답 승인이라 부르지 않습니다. 원문·Want·관찰기 [준비 장부](observer/PREPARATION-LEDGER.v1.json)와 Root metadata helper2회중실패1, saved-review1PASS 이력을 보존했습니다. presealed Want·원문·관찰기 binary는 수정하지 않았습니다.

source-first/와 want/·observer/의 역사 문서는 해당 시점의 실행0/미완성 설명을 그대로 보존합니다. 현재 실행 근거는 actual/입니다. 공개 [plan view](PUBLIC-PLAN-VIEW.v1.json)는 원 물리 plan의 observer_root를 제거한 별도 비실행 envelope이며 원 SHA를 유지합니다. 바이너리·물리 plan·private helpers·원시 OS 로그·실행 전 host 인자는 [copy ledger](SAFE-COPY-LEDGER.v1.json)의 SHA만 공개합니다. source/ 및 cached-stdlib/ 코드는 inert .go.txt/.mod.txt이고 자동 빌드·실행되지 않습니다. 공통 원문57개는 [고정 Source82 참조](COMMON-SOURCE82-REFERENCES.v1.json), 원문 라이선스는 [NOTICE](NOTICE.md)와 전문 파일로 연결합니다.

[PUBLIC-MANIFEST](PUBLIC-MANIFEST.v1.json)와 [SHA256SUMS](SHA256SUMS)는 선택 복사 및 별도 파생 파일을 구분합니다. 원본 Source81/82/Wants/실행 기록을 덮어쓰거나 새 학습 성공으로 해석하지 않습니다.
