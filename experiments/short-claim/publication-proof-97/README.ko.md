# 실제 병합·Wiki 게시 확인

[PR97](https://github.com/teamswyg/laya-tools/pull/97)의 정확한 head `ad61cd1a8b95434e2f2ed106aa154f3dcbe44621`에서 Linux·macOS·secrets·quality가 모두 통과했고 `2c51606ba6da7442fd4db3d9667daa8c573c2b83`로 자동 병합됐다. [CI](CI-97.v1.json)와 [실제 run](RUN-97.v1.json)을 보존한다.

병합 후 Source-Observation73·Compact-Storage73의 한영 설명과 Home·Sidebar 총6개 파일59,678B를 Wiki에 게시했다. 원격 Wiki commit `152ff088d6655bf0084aed07b9422db1c6ac81c3`을 새로 fetch하여 CI head의 문서와 바이트가 같음을 확인했다. [준비 장부](WIKI-COPY-97.v1.json)의 원격 대기 상태는 당시 기록이며, [실제 게시·읽기 확인](WIKI-PUBLICATION-97.v1.json)을 따로 연결한다. 이전 PR94/96 Wiki 기록도 보존한다.

[PR98](https://github.com/teamswyg/laya-tools/pull/98)의 Go FNV prefix 개선도 head `6e4a6d84008a9200fedf51b9bfd7d36834245c27`에서 필수4검사를 통과해 `7cea49090727555924bf22679ffeccc4f38c9865`로 자동 병합됐다. [CI](CI-98.v1.json)·[run](RUN-98.v1.json) 및 root의 전체 tree 일치 확인을 기록한다. 같은 특징·점수 비트를 보존하고 로컬 공개 예제의 계산 시간이 줄었으며 할당량은 그대로다. 모델 효용·RSS·GPU·실제 LLM 비용 개선을 입증한 것은 아니다.

이 게시 확인 단계에서 원본 행동 관찰·학습·benchmark·HF 게시를 반복하지 않았다. 모델 본체는 기존 HF 연구 보관소에만 있고 Git에는 없다.

[English](README.en.md)
