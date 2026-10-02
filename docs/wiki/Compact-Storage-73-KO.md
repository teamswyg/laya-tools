# 학습 수치를 그대로 두고 저장 줄이기

기존76개 자료를 저장하고 읽는 별도 실험에서 FP64 bit·인덱스·CSR·null 상태가 같았습니다. 파일 크기는 아래와 같습니다.

| 형식 | 원래 bytes | gzip bytes |
|---|---:|---:|
| Pretty JSON | 8,390,462 | 594,236 |
| 공백 제거 JSON | 4,697,570 | 400,256 |
| FP64 compact | 1,978,694 | 328,206 |

공백 제거만44.01%, 그 JSON 대비 compact는57.88% 감소했습니다. 원래 파일 대비76.42%에는 두 효과가 함께 들어 있습니다. gzip은 캐시·전송 크기 비교이며 풀린 JSON의 입력 한도를 바꾸지 않습니다.

파일과 RAM은 다릅니다. JSON import부터 compact 왕복·보고서까지 전체 worker peak RSS는약82.3MiB였습니다. 단독 decoder의 메모리나 추론 속도 개선을 입증한 결과는 아닙니다. 공유 loader·runtime·학습·정답·한도는 변경하지 않았습니다. 다음에는 수치와 메모리를 별도로 확인하고 단순 압축과 전용 codec 중 선택합니다.

[한영 상세·고정 계획·원래 결과·기준선](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/compact-storage-73) · [English](Compact-Storage-73-EN) · [새 자료 관찰](Source-Observation-73-KO)
