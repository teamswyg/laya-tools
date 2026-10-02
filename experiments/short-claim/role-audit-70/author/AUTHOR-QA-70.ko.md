# 70 실제 역할 결과의 작성자 메타데이터 대조

저장된 한 번의 실행 결과와 원래 고정 입력을 읽는 stdlib Go 검사 1회가 성공했다. 검토자는 70 메타데이터와 실행기를 작성했으므로 이 보고서는 독립 역할 알고리즘 검증이나 독립 의미판단이 아니다. 역할·MembershipDigest·OrderDigest·Assign·Features·Validate·Project·학습·모델·원천 API는 재실행하지 않았다.

실제 결과는 29,864 B, SHA `24de7b745747966107b00c6748a61dc920d97ed9cc9f98a8f22fd367cd8b6b70`다. 고정 실행 계획 SHA `6bc94c340d3983f950349fb327453e5518b9c2980c8c09175ccb05886f2d94db`와 proposal SHA `4e9d36db9e151f807cc2c79da72d863308369d0708a86524b32f717a334f8c21`의 연결이 맞다. 계획·실행 파일·입력 14개·source 15개·결과·worker receipt/ledger·root ledger를 합친 35개 byte pin이 일치한다. 예전 17개 그룹의 전체 저장 객체가 그대로 남고 새 출처의 두 전체 가족이 추가된 19개 그룹이다. 부모 76개와 후보 위치 226개의 ID·순서·그룹·정답 상태·proposed 여부가 결과의 parent_roles와 일대일 대응한다.

| 관측된 저장 메타데이터 | train | validation | calibration |
|---|---:|---:|---:|
| 전체 그룹 | 12 | 4 | 3 |
| known 포함 그룹 | 11 | 4 | 3 |
| 요청 / 후보 위치 | 44 / 130 | 20 / 60 | 12 / 36 |
| answerable / no_answer / unknown 요청 | 22 / 9 / 13 | 10 / 5 / 5 | 6 / 3 / 3 |
| unknown null 위치 | 39 | 15 | 9 |
| mask 자격 P / N | 22 / 51 | 13 / 31 | 5 / 18 |
| mask 제외 known 위치 | 18 | 1 | 4 |
| mask 자격을 하나 이상 가진 그룹 | 10 | 4 | 3 |
| positive-bearing 그룹, 설명용 | 8 | 4 | 2 |

전체 unknown 요청 21개와 null 위치 63개가 유지된다. unknown-only 그룹 64는 train provenance이며 정답이나 가중치로 바뀌지 않았다. 모든 실제 학습 가중치는 아직 null이다. 원래 labels 46 P/107 N과 출처68의 proposed 5 P/5 N, 모든 acceptable 위치·mask가 유지된다. 저장 정답 coverage 6개가 원래 9/3/3 하한을 통과했다. mask 자격 그룹 10/4/3과 positive-bearing 8/4/2는 설명용이며 새 하한이 아니다.

출처68의 두 새 가족과 요청 4개는 이번 고정 seed의 저장 결과에서 train에 속한다. 따라서 이 역할 결과만으로 validation의 독립 upstream 전이나 저작 다양성을 증명할 수 없다. `training_ready=false`와 `source_authoring_diversity_cleared=false`는 그대로다. 원래 group 관계를 다시 계산하거나 유리한 seed를 찾거나 새로운 라벨을 만들지 않았다.

실행기의 저장 카운터는 compiled identity 1회, membership-set verification 1회, Assign 1회가 모두 완료됐다고 기록한다. module counter는 19 component 확인, 18 known order hash, coverage 6개 통과, 19 assignment 반환이다. root ledger는 실제 child 1회/exit 0/retry 0/timeout false를 기록한다. 이 QA는 저장 수치의 연결만 검사했으며 자원 측정을 추가하지 않았다.

고정 실행 계획의 limits[0]에는 draft의 `execution_authorized=false`와 아직 실행하지 않았다는 문장이 남아 있다. 원문을 바꾸지 않았다. 실제 상위 필드는 true/frozen이고 실행 receipt·결과·최종 ledger는 실행 완료 1회를 나타낸다. invocation receipt의 actual_assign_attempts 0은 호출 전에 만들어진 예약 상태이며 최종 결과의 완료 1과 모순되는 최종 카운터로 해석하지 않는다.

검사 원천은 `check.go`, 사전 계획은 `QA-PLAN-70.v1.json`, 결과는 `AUTHOR-QA-70.v1.json`(8,925 B, SHA `a0e401bc3a3baf1ba1b34315ad0faf0361df2734b7ef4f32b837e0adb66527dc`)이다. 성공 1회/실패 0회, 공유 파일 변경·새 역할·학습·원천 API 호출·모델·보호 final 열람·커밋·공개는 0이다. 관련 없는 private 임시 폴더의 권한 오류를 만난 넓은 파일 목록 읽기 1회는 별도 장부에 남겼다. AI가 준비와 읽기에 참여했으며 협업 비용은 측정하지 않았다.
