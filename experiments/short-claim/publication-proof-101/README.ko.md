# 압축 가중치 모듈 게시 확인

[PR101](https://github.com/teamswyg/laya-tools/pull/101)은 Linux·macOS 테스트, 민감정보 검사, quality 검사를 통과하고 자동 병합됐습니다. 테스트한 head와 실제 병합 commit의 전체 파일 내용도 같습니다. [실제 CI 응답](riido101-run-success-proof.json)과 [병합 응답](riido101-ci-final.json)을 보존합니다. 사람의 승인이나 보호 규칙 우회는 없었습니다.

이번 공개 모듈은 작은 힌트 모델의 가중치를 압축한 채 읽고 점수를 계산하는 Go 라이브러리입니다. **Laya 전체 모델을 이 크기로 줄였다는 뜻은 아닙니다.** 기존 형식과 계산 결과를 비교하는 테스트는 통과했지만, 공개 모듈 자체의 벤치마크 실행은 아직 0회입니다. 별도 비공개 프로토타입에서 측정한 속도와 구분합니다.

필수 CI 성공과 병합을 확인한 다음 한·영 설명과 Wiki 안내 4개 파일, 총 10,638바이트를 게시했습니다. 원격 Wiki commit `cac8e0d727eea36d5b7608977ea863904e7d474b`를 fetch해 파일별 바이트와 해시를 한 번 검증했고 모두 일치했습니다. [준비 당시 장부](riido101-wiki-copy-ledger.json)는 당시 pending 상태를 유지하고, [실제 게시 장부](riido101-wiki-publication-ledger.json)는 완료를 기록합니다.

[한국어 Wiki](https://github.com/teamswyg/laya-tools/wiki/Packed-Weights-78-KO) · [English Wiki](https://github.com/teamswyg/laya-tools/wiki/Packed-Weights-78-EN) · [정확한 복사 기록](PUBLIC-COPY-LEDGER.v1.json) · [English](README.en.md)

이 게시 확인에서는 모델·원 API·학습·벤치마크를 다시 실행하거나 Hugging Face에 업로드하지 않았습니다. 공개 응답 4개는 원본 14,452바이트 그대로이며, 로컬 경로·바이너리·비공개 로그와 게시용 helper는 복사하지 않았습니다.
