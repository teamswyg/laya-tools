# 33행 개발 자료의 공개 전 검사

[자료와 재현 방법](../next60-development-thirtythree/README.ko.md)을 공개하기 전에 오프라인 재현·전체 race 테스트·정적 검사·포맷·비밀정보 검사를 실제로 통과했습니다. [검사 기록](LOCAL-CHECKS.v1.json)을 확인하세요. 고정 관측42행·조건318개, 자료33행과 Reader 결과의 바이트 일치를 검사합니다.

이 파일은 로컬 검사 기록입니다. PR124의 Linux·macOS CI와 봇 병합, HF33 태그·고정 다운로드·viewer 검증은 관측한 뒤 별도 후속 기록으로 남깁니다. 기존30행은 그대로이며 새 학습은0회입니다. OS RSS와 Go pprof의 측정 범위를 구분하고 GPU 실행이나 Codex 절감을 주장하지 않습니다.

[English](README.en.md) · [이슈19](https://github.com/teamswyg/laya-tools/issues/19)
