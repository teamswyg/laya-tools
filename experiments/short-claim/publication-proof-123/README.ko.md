# 실제 게시·저장 공간 검증 기록

[PR122](https://github.com/teamswyg/laya-tools/pull/122)은 [필수 CI4검사](https://github.com/teamswyg/laya-tools/actions/runs/37106112785)를 통과해 봇이 병합했습니다. 소스와 병합 Git tree가 같습니다. 새30행 오프라인 검사와 기존 native Laya 검사는 별개이며 GPU 실행 확인은 아닙니다. [실제 기록](CI-MERGE122.actual.public.v1.json).

[한영 Wiki와 탐색 메뉴](https://github.com/teamswyg/laya-tools/wiki/Next60-Development-KO) 4페이지를 실제 게시하고 원격 Git blob 전체 바이트를 확인했습니다. [Wiki 기록](WIKI30-PUBLICATION.actual.public.v1.json). [이슈19 체크리스트](https://github.com/teamswyg/laya-tools/issues/19#issuecomment-5963281223)도 이전 내용을 보존해 갱신·전체 재읽기했습니다. [이슈 기록](ISSUE19-SYNC.actual.public.v14.json).

완료된 임시 사본4개는 개별 압축·실제 파일 복원·전체SHA·mode0600·저장 확인 후 정리했습니다. 33,561,848 B → 2,072,888 B, 기록 제외 **31,488,960 B**를 확보했습니다. 별도 검토자가 전체 gzip EOF/CRC·SHA·현재 부재·기존16개 보관 기록을 확인했습니다. [독립 검토](storage-review/REVIEW.v1.json)와 [범위 설명](storage-review/README.ko.md)을 보세요. 디스크 저장량 수치이며 모델 RAM·GPU·속도 개선 수치가 아닙니다. 저장 census는 비원자적 과거 측정으로 실행 전 재확인이 필요하고, 기존512MiB 범위를 유지합니다. 향후 원래 위치 복원은 inode/mtime가 달라 새 metadata 동결이 필요합니다.

현재 검증 자료30/88과 HF30은 그대로입니다. [다음 세 과제](../next60-native-three-preparation/README.ko.md)는 입력·기대값·설명·순서만 사전에 고정했으며 새 실행/적격/Fit/모델은0입니다. 모델 본체·사적 경로·원본문·raw journal을 이 기록에 넣지 않았습니다.

[English](README.en.md)
