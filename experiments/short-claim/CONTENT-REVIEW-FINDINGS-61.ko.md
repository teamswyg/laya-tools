# 61: 남은 설명문과 구현의 의미 검토

기존59의 남은 **615개 검토 항목을 모두 읽고 상태를 기록했습니다.** `pending`을 무조건 통과로 바꾸는 작업은 아닙니다. 이번 범위에서 446개는 제한된 근거와 일치하고,94개는 충돌하며,10개는 요구 범위가 빠졌고,65개는 미해결입니다. 이는 후보를 승인하거나 새 학습 정답을 만든 결과가 아닙니다.

범위는 기존15가족/60부모/180후보 위치/49소스 ID/60개 서로 다른 설명문 SHA/114 literal입니다. 부모마다2·3·4후보가 있는 수는11·38·11이고 원래 순서를 유지했습니다. 기존60의123개 overlay를 수정하거나 다시 생성하지 않았습니다. 원래59의738개 `pending` snapshot도 그대로이며,61의 새 overlay가615개 항목을 설명합니다.

| 기존59 항목 종류 | 근거와 일치 | 충돌 | 범위 누락 | 미해결 |
|---|---:|---:|---:|---:|
| 요청의 계약 표현60 |42|0|3|15|
| 후보 구현의 제한된 의존 관계 읽기180 |180|0|0|0|
| 설명문이 자기 구현에 충실한가180 |177|0|0|3|
| 설명문이 요청을 충족하는가180 |32|94|7|47|
| 계약 literal·관찰 필드15 |15|0|0|0|
| 합계615 |446|94|10|65|

관찰 필드와 부정·경계 조건은 별도360개 record로 남겼습니다. 기존 항목, 부모나 후보가 늘어난 것은 아닙니다. 관찰 축180개는 일치16/누락117/미해결47, 부정·경계 축180개는 일치121/누락10/미해결49입니다. 전체975개 record 모두 같은 기존4축과5개 상태 어휘를 사용합니다. `source_closure_only`는 기존 closure 항목을 기록하기 위해 `source_fidelity` 축을 참조하지만 `caption_fidelity_assessment=false`이고,177개의 실제 fidelity 일치 수에는 포함되지 않습니다.

가장 중요한 결과는 **틀린 구현을 정확히 설명하는 문장도 많다**는 점입니다. 예를 들어 관리자에게까지 검증을 요구하는 설명, 모든 중복을 지우는 설명, 실패 때 release를 생략하는 설명은 실제 코드에 충실할 수 있습니다. 그렇다고 관리자 bypass, 인접 중복만 제거, 획득 후 모든 terminal 결과에 release를 요구하는 요청을 충족하지는 않습니다.177개 fidelity 일치를177개 정답이나 승인 후보로 읽으면 안 됩니다.

원래 문장의 미해결 구간도 보존했습니다.

- `Reject only ...` 형태의 closed-window와 nondecreasing paraphrase는 완전한 거절 규칙인지, 거절의 필요조건만 말하는지 읽기가 갈립니다. 긍정·empty 조건을 소스에서 가져와 문장을 강화하지 않았습니다.
- rotate-left의 설명은 첫 값을 뒤로 옮기는 행동을 표현하지만 요청이 명시한 empty 처리까지 표현하지 않습니다. 코드의 empty guard는 fidelity 근거일 수 있어도 설명문의 coverage를 대신하지 않습니다.
- snapshot 설명들은 copy/alias/nil/empty/즉시 입력 변경을 표현하지만 요청의 `never panic`을 표현하지 않아 coverage와 observation 누락을 기록했습니다.
- quoted-delimiters의 짧은 요청·설명은 일반적인 error/clearing output을 표현해도 exact Syntax/Bounds/ASCII enum과 no-panic 관찰을 전부 표현하지 않습니다. 명확하게 말한 부분의 escape·partial-output 충돌은 기록하되, 기존 네 unknown 부모를 정답으로 승격하지 않았습니다.
- graph-mutates-input의 `erase root edges`는 논리적 `Count=0`인지 저장된 `Children[4]` 값까지 지우는지 불분명합니다. 실제 소스는 Count만0으로 하고 graph Count를 줄이며 Children 값은 지우지 않습니다. 이 세 위치의 fidelity는 미해결입니다. 어느 읽기든 입력을 변경한다는 사실은 unchanged-input 요청과 충돌하므로 coverage는 별도로 판단했습니다.

관찰할 수 있는 범위를 혼합하지 않았습니다. Legacy checker는 복제한 입력과 반환 시퀀스를 `slices.Equal`로 비교하며 panic이나 입력 변경을 unknown으로 처리합니다. Nil/empty, aliasing, capacity와 allocation을 검증하지 않습니다. Snapshot은 논리적 길이0..4에서 nil/empty, 복사 직후 입력, 결과에 쓴 뒤 입력, 입력에 쓴 뒤 결과를 구분합니다. Inspector가 수행하는 순차 쓰기와 후보의 입력 변경은 별개이며 spare capacity·whole-memory를 관찰한 것은 아닙니다.

Lifecycle은 실제 context/resource 대신 길이0..8 enum 이벤트와 반환 Events/Count/Status를 관찰합니다. Finish/fail before begin 같은 지원 제외 조건은 새 실패 label이 아닙니다. Quoted는 오류의 정확한 enum과 Count 및8개 Tokens를 따로 비교하며, 여러 오류 조건이 동시에 생기는 모든 문자열의 precedence를 새로 증명하지 않았습니다. Graph는 유효한1..4노드의 root0 탐색과 error enum, 전체 post-call graph 값을 관찰하며, 잘못된 count/index는 지원 실패로 남깁니다.

근거는 공통 [catalog](content-review-evidence-61.json),15개 family shard와 [manifest](content-review-manifest-61.json)에 있습니다. Original decoded request/caption의 SHA·UTF-8 길이·정확한 byte quote 범위와 official60 JSON pointer를 연결했습니다. 물리적인 raw source span98개를 공유해서 같은 enum 선언을 가리키는 여러 ID를 합치거나 코드 본문을 반복 복사하지 않았습니다. 논리적 declaration catalog link105개, observer/type line11개, literal source15개가 있고 기존49 root/172 component 관계를 유지합니다. `Input/Want`은 기존 authored literal이며 새 per-vector `Got`을 만들거나 후보를 실행하지 않았습니다.

실행 장부는 다음 작업을 구분합니다. 실제 source/text review는1회, record 직렬화 실행은1회 성공/실패0입니다. 직렬화 도구의 준비 빌드는2회 중 첫 map trailing-comma 문법 오류1회 후 성공1회였습니다. 작성자 byte verifier는2회 중 첫 자체 숫자 타입 가정 오류(`number_bytes`)1회 후 helper만 수정하여 성공했습니다. Assessment/catalog/shard/manifest를 다시 생성하거나 판단을 바꾸지 않았습니다. 이전 reference QA1회 성공 및 준비61 제작자의 metadata 생성1회 성공은 별도 기록입니다.

작성자 byte 검사에서는975 record digest/필수 필드,615+360 일대일 index,1860 text-reference pointer/SHA/길이 대조,984 quote 범위,98 raw source span을 확인했습니다. 별도 독립 byte 검토와 CI는 부모 작업이 수행합니다. 이 byte 무결성은 의미 판단의 객관적 정답이나 학습 자격을 증명하지 않습니다. [작성자 receipt](content-review-reader-byte-receipt-61.json), [실제 review/생성 장부](actual-content-review-attempt-1-ledger.json), [byte 검사 장부](reader-byte-QA-attempt-ledger.json), [빌드 장부](SERIALIZER-BUILD-LEDGER-61.json)를 구분해 읽어야 합니다.

검토자는 원래 caption/fixture/CLI 및 준비61 생성기의 저자와 독립이지만 protocol60 준비·제한 검토 이력과 metadata/outcome 노출이 있습니다. 따라서 blind 평가 또는 독립 데이터 작성 원천을 주장하지 않습니다. 이 검토는 **Codex 협업 안에서 AI가 수행한 source/text 읽기**입니다. `model_calls=0`·`paid_calls=0`는 별도 judge 추론/API/benchmark 프로세스를 실행하지 않았다는 뜻입니다. 이 대화의 AI 사용량이나 비용이0이라는 뜻이 아니며 일반 협업 비용은 측정하지 않았습니다.

기존 subset unknown18/전체21, 그룹17/known-containing16, original72/216과 closed-window/clamp-window 공유 그룹8을 그대로 유지했습니다. 새 label·roles·fit·weights·protected-final 읽기는0이고 `training_ready=false`입니다. 기존 synthetic_single_pipeline/no_roles_plan 및 데이터 원천 편향 문제를 이 문장 검토로 해소했다고 주장하지 않습니다. 새15저장소 gate나 모든 개발 fit에2400개를 요구하는 조건을 추가하지 않았고, 별도 protected-final의 도메인별 독립 요청 최소2400개 목표는 유지합니다.
