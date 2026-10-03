# 30행 공개와 실제 재검증

[HF30 고정 버전](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-30-finite-v1)을 실제 게시하고, 고정 commit **89c0c7c9ddc3135e37e88e9da5d51ac55dff0e1d**에서 소유 파일 **558개·4,559,126바이트**를 모두 내려받아 로컬 게시 파일과 대조했습니다. Hub가 관리하는 .gitattributes 1개·2,504바이트는 이전23 버전과 같았습니다. 이전2·3·7·16·21·23 태그도 이동하지 않았습니다.

[원격 파일·태그·표 검증 기록](HF30-PUBLICATION.actual.public.v3.json)과 [표 응답 검증 기록](HF30-VIEWER-ACCEPTANCE.actual.public.v1.json)을 제공합니다. HF 표의 **30행·모든 필드·순서**가 고정 데이터와 같고 num_rows_total=30, partial=false, 잘린 셀0입니다. 응답에는 commit과 최상위 truncated 필드가 없어 null로 남겼습니다. 표 검증과 고정 commit의 전체 파일 검증은 별도 증거입니다.

[PR121](https://github.com/teamswyg/laya-tools/pull/121)은 [실제 CI](https://github.com/teamswyg/laya-tools/actions/runs/37103825806)의 Linux·macOS·비밀정보·quality 네 검사를 통과해 CI 봇이 병합했습니다. [정확한 소스·병합 트리와 검사 기록](CI-MERGE121.actual.public.v1.json)을 보세요. 새30행 오프라인 재현 단계와 기존 Laya native CI 단계가 각각 성공했습니다. GPU 실행은 확인하지 않았습니다.

HF 표 조회는 처음3회 HTTP500이었고 네 번째200에서 값을 대조했습니다. 첫 로컬 검증의 존재하지 않는 최상위 truncated 필드 가정, 이어진 옛 응답 파일명 사용도 실패로 보존했습니다. 실제 스키마와 성공 응답 파일을 사용하는 새 검증에서 종료0을 확인했습니다. 실패를 성공으로 덮거나 원본·모델을 다시 실행하지 않았습니다.

자료는 **30요청·88라벨(긍정30/부정58)**이며 development_train입니다. [사용법](../next60-development-thirty/README.ko.md)과 [다음 원천 후보22개](../next60-new-source-preview/README.ko.md)를 보세요. 이번 자료 확장의 새 주장 모델/Fit은0이고 기존3 Fit·3 논리 모델 및 실패 모델 비활성 상태는 그대로입니다. 30은 수집 checkpoint이며 학습·일반화·Codex 절감이나 모델/GPU 성능의 증거가 아닙니다.

공개 파일은 자체 문서·판정 요약·체크섬입니다. 인증정보·개인 경로·사적 입력·원본 runtime 본문·모델 본체·원시 journal을 포함하지 않습니다. [English](README.en.md)
