# PR93·Wiki·Hugging Face 공개 확인

[PR93](https://github.com/teamswyg/laya-tools/pull/93)의 정확한 head `9c0eca98ac9e3f40b0eadafb21422a349fb732d8`에서 Linux/macOS·secrets·quality가 모두 성공했습니다. 2026-10-02 squash merge `9d204c2c700505658c108297d5fa769a835a60c3`을 확인했습니다. 이전 pending 기록은 보존하며 [최종 CI 상태](CI-PR93.v1.json)를 추가했습니다.

그 head의 한국어·영어 설명, Home, Sidebar 네 파일을 Wiki에 복사했고, push/fetch 후 원격 Git 내용이 일치하는지 확인했습니다. [Wiki 공개 장부](WIKI-PUBLICATION-93.v1.json)에 commit과 파일별 SHA가 있습니다.

[Hugging Face 데이터](https://huggingface.co/datasets/JooYoon/riidolaya-public-claim-preparation-69/tree/d80075c6160a53d5426019cf018016a3b02017bd)는 **원래 72개 요청의 배열 준비 스냅샷**입니다. 프로젝트 파일 20개·7,960,062 B를 정확한 commit에서 다시 내려받아 원본 바이트와 대조했습니다. [공개 장부](HF-PUBLICATION-69.v1.json)와 [컬렉션·태그 확인](HF-COLLECTION-69.v1.json)은 준비·게시·재다운로드를 구분합니다. 기존 컬렉션 14개 항목은 보존하고 새 데이터 항목 하나만 추가했습니다.

자료에는 학습된 모델 계수가 없고 fit0회입니다. 72행·216후보는 독립 과제 2,400개나 최종 골든셋이 아닙니다. 프로젝트 소유 공개 fixture와 파생 자료는 Apache-2.0이며 upstream 구현·가중치를 재배포하거나 재라이선스하지 않았습니다. Laya 인코더 실행·GPU 추론·LLM 절감을 증명하는 공개가 아닙니다. 후속 76개 요청/fit을 이 버전에 덮어쓰지 않습니다.

HF 단계의 v1 검토와 최신 카드 재독 delta를 [검토 폴더](hf-stage-qa/STAGE-QA-69.ko.md)에 보존했습니다. 이전 카드 원문이 없어 완전한 바이트 diff는 증명하지 않았습니다. 태그 검증 helper의 잘못된 target 가정 실패 1회도 보존했고, 태그로 조회한 dataset revision이 실제 upload commit과 같음을 확인했습니다.

[English](README.en.md)
