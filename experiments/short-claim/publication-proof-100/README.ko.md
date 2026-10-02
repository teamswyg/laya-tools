# 실제 게시 확인100

[PR100](https://github.com/teamswyg/laya-tools/pull/100)의 head `b8721814b743f1201e4b914fd016df2439af1710`는 Linux·macOS·민감정보·quality 검사를 모두 통과한 뒤 `172880c43c3f16dcf1efb119dda0c2d16cef1973`로 자동 squash 병합됐습니다. [실제 CI](https://github.com/teamswyg/laya-tools/actions/runs/36999950398)의 완료 응답과 head/merge 전체 tree 일치를 확인했습니다. 인간 리뷰 승인이나 보호 규칙 우회는 없었습니다.

Wiki는 필수 검사의 성공과 병합을 확인한 다음 6개 페이지·65,830바이트를 그대로 게시했습니다. 원격 commit `e2e260a5c6c9c682b160dde8537d2c03c3afa1fd`를 fetch한 뒤 모든 페이지의 바이트 일치를 확인했습니다. [준비 당시 pending 장부](riido100-wiki-copy-ledger.json)와 [실제 게시 장부](riido100-wiki-publication-ledger.json)는 각각 당시 상태를 유지합니다.

읽기 검증의 첫 시도는 root가 잘못된 전체 Wiki commit 인자를 전달해 무결성 검사에서 거절됐습니다. 원격 문서는 바뀌지 않았고, 실제 fetch한 commit을 사용한 두 번째 검증은 통과했습니다. [두 시도 기록](riido100-wiki-readback-attempts.v1.json)을 보존합니다. 모델·원 API·학습·벤치마크 재실행이나 HF 업로드는 없었습니다.

[CI 응답](riido100-ci-final.json) · [run 응답](riido100-run-success-proof.json) · [English](README.en.md)
