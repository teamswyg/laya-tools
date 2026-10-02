# Source80 변환 초안과 기존 그룹·역할 연결 검토

현재 선언된 76개 요청의 19개 whole-component에는 UUID 또는 humanize Ordinal 원천이 없습니다. 따라서 이번 세 초안의 기존 그룹·역할 연결은 **미정(null)**입니다. 기존 semver 그룹72와 doublestar 그룹74가 실제 train이라는 사실을 UUID·Ordinal에 이전하지 않습니다.

[RECEIPT.v1.json](RECEIPT.v1.json)에 원본 SHA와 전체 저장 역할을 기록했습니다. proposal70의 /parents와 /whole_groups, actual70의 /parent_roles를 original parent_index로 연결했고 76개 ID·그룹 및 그룹마다 하나의 역할이 일치했습니다. known38, no_answer17, unknown21; train44/validation20/calibration12개 요청이며 전체 그룹12/4/3, known 그룹11/4/3입니다. no_answer도 기존 알려진 진실이며 unknown과 합치지 않았습니다.

| 기존 그룹 | 원형/원천 | 저장된 실제 역할 | 요청 인덱스 |
|---|---|---|---|
| 0 | retain-active | development_calibration | 0, 1, 2, 3 |
| 4 | allow-owner | development_validation | 4, 5, 6, 7 |
| 8 | closed-window, clamp-window | development_validation | 8, 9, 10, 11, 36, 37, 38, 39 |
| 12 | quota-total | development_validation | 12, 13, 14, 15 |
| 16 | stable-odd | development_train | 16, 17, 18, 19 |
| 20 | remove-first | development_train | 20, 21, 22, 23 |
| 24 | compact-runs | development_train | 24, 25, 26, 27 |
| 28 | rotate-left | development_train | 28, 29, 30, 31 |
| 32 | nondecreasing | development_train | 32, 33, 34, 35 |
| 40 | prefix-balance | development_train | 40, 41, 42, 43 |
| 44 | transform-order | development_train | 44, 45, 46, 47 |
| 48 | atomic-commit | development_train | 48, 49, 50, 51 |
| 52 | error-identity | development_train | 52, 53, 54, 55 |
| 56 | owned-snapshot | development_calibration | 56, 57, 58, 59 |
| 60 | cancellation-lifecycle | development_validation | 60, 61, 62, 63 |
| 64 | quoted-delimiters | development_train | 64, 65, 66, 67 |
| 68 | ancestor-cycle | development_calibration | 68, 69, 70, 71 |
| 72 | whole_upstream_semver | development_train | 72, 73 |
| 74 | whole_upstream_doublestar | development_train | 74, 75 |

Parse·Scan의 여섯 후보는 같은 UUID family에 함께 묶어야 합니다. shared xtob, error 타입, Scan→Parse 내부 호출, UUID receiver, 패키지 원문·initializer, 모든 wrapper/postprocessor/alias/입력 변형을 포함합니다. Ordinal 세 후보는 원래 int Ordinal과 MIT 차용 last-digit 변형을 포함해 같은 Ordinal family로 묶는 제안입니다. 이것은 그룹 관계 제안이며 새 그룹번호·역할·감독을 생성하지 않습니다.

원천 이름·revision을 19개 선언 그룹에서 찾지 못했고 기존56/56b 원문 후보의 선택된 몸체 표지도 찾지 못했습니다. 이는 저장된 선언에 현재 연결이 없다는 근거일 뿐, 원천 독립성의 증명이 아닙니다. source80 /exposure에는 두 원천 모두 source53 inventory와 이전55 코딩 작업 노출이 기록되어 있습니다. 신규 독립 그룹·처음 보는 검증/최종 자료라고 부르지 않습니다. state/error 의미가 닮았다는 이유만으로 staged-commit48 또는 error-identity52와 합치지도 않습니다.

65의 기존 정책은 observer·표준 라이브러리 기반 코드를 명시적인 전역 provenance로 보존하되 자동 grouping edge를 만들지 않습니다. 같은 저자·Go Authors·stdlib 사용만으로 모든 그룹을 합치지 않습니다. 실제 원문 차용, helper/alias/fork, 후처리 공유가 발견되면 해당 관계를 기록해야 합니다. 기존 family로 연결되는 경우 그 whole-component 역할을 그대로 따릅니다. train/validation/calibration을 가로지르는 연결이 발견되면 영향을 받는 fit 계획을 멈추고 근거를 남기며, 역할 이동·component 분리·seed 탐색으로 해결하지 않습니다.

이 검토는 Want80·conversion80 작성자의 비맹검 메타데이터 검토입니다. 독립 원천 심사나 학습 승인으로 해석하지 않습니다. 기존 진실·mask·weight·role 변경0; 원 API/역할 배정/Features/Score/Fit/모델/외부 게시0입니다. 3개 요청 초안은 여전히 conversion pilot이며 후보 실제 검증, 감독 적격성, 전체 고정 control의 headroom 확인 후 별도 fit 계획을 준비해야 합니다. 기존5%·9/3/3·보호된 final2400 범위는 그대로이며 새 숫자 gate는 추가하지 않았습니다.
