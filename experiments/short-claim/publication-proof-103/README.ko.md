# 준비 자료 게시 확인

[PR103](https://github.com/teamswyg/laya-tools/pull/103)은 head `6a68527e6d08ba5ee517faf5dd472c1392ec59a0`에서 필수 CI 네 가지가 통과한 뒤 자동 병합됐습니다. 실제 merge `01ce47dcca16c8f82022c97354d82105b6819875`와 head의 전체 Git tree는 일치합니다. [CI 실행](https://github.com/teamswyg/laya-tools/actions/runs/37015530233)은 Linux·macOS 테스트, secrets, quality 모두 성공했습니다.

그 뒤 Home·Sidebar·학습 자료 준비 한국어/영어 네 Wiki 파일 14,115바이트를 검토한 Git 원본에서 복사했습니다. 비밀정보 검사 후 게시하고 실제 Wiki commit `83cf5e1eb5500861b4cc4bde5d6826b0d02c7b4b`을 다시 fetch해 원격 네 파일의 바이트 일치를 확인했습니다.

- [병합 상태](CI-MERGE-OFFICIAL.v1.json), [네 실제 CI job](CI-JOBS-OFFICIAL.v1.json).
- [원격 게시 확인](WIKI-PUBLICATION-LEDGER.json), [로컬 복사 확인](WIKI-COPY-LEDGER.json).
- [당시 준비 인계](HANDOFF.v1.json)는 실제 게시 이전의 실행0·pending 상태를 보존합니다.

이 자료는 게시 검증입니다. 함수·후보 관찰, 새로운 학습이나 모델 성능 검사를 대신하지 않습니다. 호스트 경로를 가진 게시 helper·바이너리는 Git에 넣지 않았습니다.
