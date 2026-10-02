# 69 Hugging Face 로컬 준비 파일 독립 대조

현재 로컬 stage의 데이터·출처·권리·안전 대조가 통과했다. 검토자는 HF stage 생성기를 작성하지 않았으며, 새로운 의미 정답·학습 적격성·모델 품질을 판단하지 않았다. HF 명령·업로드·원본 API·특징 계산·투영·순위·역할·학습·모델 호출은 모두 0이다.

고정 manifest는 PR93의 정확한 head `9c0eca98ac9e3f40b0eadafb21422a349fb732d8` CI를 기다리는 staged 상태다. 7개 핵심 파일의 전체 7,927,992 B에서 byte 크기와 SHA가 일치했다. 원본 결과 7,699,810 B/SHA `348af320a8bdcbee433f3102cc72aba06404ddc568a9ae9ddf6ff8f8d2980a4f`는 그대로다. JSONL은 원본의 별도 보기이며 106,232 B/SHA `2c9f9834fea9e03132619423beecedba7597d6c3818dc20dfa2b0dd351dcde2c`다.

72개 부모와 216개 후보의 원래 index·Provenance·원문·저장 정규화·ID·순서·acceptable 순서·전체 그룹·역할·정답·loss mask가 모두 raw 결과와 같다. 역할 enum 0/1/2 및 truth enum 1/2/3의 문자열 변환이 정확하다. 정답은 known 34/no_answer 17/unknown 21이며 17개 전체 그룹을 유지한다. 역할별 부모는 44/16/12, 후보는 132/48/36이다. unknown 후보 63개의 label은 null이다. train/validation의 prepared fit weight 126개를 raw row refs와 하나씩 비교했고 0가중치 19개가 보존된다. calibration 후보 36개의 fit weight와 모든 unknown fit weight는 null이다. 이것은 준비 배열의 가중치이며 학습된 계수가 아니다.

현재 저장소 source 15개 pin, 공개 원문 복사 5개가 맞는다. 별도의 local Git byte 대조로 source commit `be05d073a12b7f2094e19470988e1ea034b1af5a`의 source 15개+LICENSE와 evidence head `9c0eca98ac9e3f40b0eadafb21422a349fb732d8`의 공개 evidence 4개를 확인했다. 합계 20개 blob 비교가 모두 일치하며 다운로드나 API를 실행하지 않았다.

LICENSE는 기존 프로젝트 Apache-2.0 원문과 같다. README/NOTICE는 자체 작성 공개 fixture·수치·파생 보기의 권리와 AI 보조 영어 작성, source/alias 그룹을 설명하며 upstream 구현·pretrained 가중치를 재라이선스하지 않는다. 이 stage에는 upstream 코드·학습 모델·바이너리·원시 profile·private 계획·로그·인증 값이 없다. 고정 파일 allowlist와 경로/토큰/개인키 표식 검사는 통과했다. 이 제한된 검사가 모든 비밀 탐지를 증명하는 것은 아니다.

카드의 현재 README SHA는 `bedcecb99ce5d4751f7c06b827eedd4000d1733a3cb5a0f4160759b5489b77c6`, NOTICE SHA는 `35456b351e29a369daf0e02cae91eb10bac9871649828fe09de0a4ba6a7e77a7`다. 원본 snapshot, 단일 preparation 관측, 계수 없음/fit 0, 미측정 AI 협업 비용, 미승인 다양성·일반화 범위와 별도 full76 버전을 정확히 구분한다. 영어 숫자 주변 공백을 보강하라는 가독성 메모만 부모에게 전달했고 의미·권리·안전에 남은 blocker는 없다. README가 변경되면 그 차이와 새 SHA를 별도 기록해야 한다.

독립 core 검사 1회와 Git pin helper 1회가 성공했다. core receipt는 `CORE-STAGE-QA-69.v1.json` SHA `beaa6b28518ad2ee4ac8c0b39c1bad83b0d2150bb594b7fe03c5a470b7892c7d`다. 새 원본 평가·라벨·역할·공유 수정·커밋·공개는 0이다. 학습 준비나 최종 2,400개/domain 검증을 새로 승인하지 않는다. 게시자는 정확한 CI 결과와 최종 README/NOTICE를 새 publication manifest에 넣어야 하며, 기존 7-file staged manifest는 역사 snapshot으로 보존한다. 이 보고서는 업로드 완료나 release 주장이 아니다.
