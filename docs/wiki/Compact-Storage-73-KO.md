# 학습 수치를 그대로 두고 저장 줄이기

기존76개 자료를 저장하고 읽는 별도 실험에서 FP64 bit·인덱스·CSR·null 상태가 같았습니다. 파일 크기는 아래와 같습니다.

| 형식 | 원래 bytes | gzip bytes |
|---|---:|---:|
| Pretty JSON | 8,390,462 | 594,236 |
| 공백 제거 JSON | 4,697,570 | 400,256 |
| FP64 compact | 1,978,694 | 328,206 |

공백 제거만44.01%, 그 JSON 대비 compact는57.88% 감소했습니다. 원래 파일 대비76.42%에는 두 효과가 함께 들어 있습니다. gzip은 캐시·전송 크기 비교이며 풀린 JSON의 입력 한도를 바꾸지 않습니다.

파일과 RAM은 다릅니다. JSON import부터 compact 왕복·보고서까지 전체 worker peak RSS는약82.3MiB였습니다. 단독 decoder의 메모리나 추론 속도 개선을 입증한 결과는 아닙니다. 공유 loader·runtime·학습·정답·한도는 변경하지 않았습니다. 다음에는 수치와 메모리를 별도로 확인하고 단순 압축과 전용 codec 중 선택합니다.

후속74에서는 각 형식을 새 child에서 한 번씩 읽었습니다.24개 열의 수치 비트·길이·null·메타데이터가 모두 같았고 최대 RSS는 JSON18.23MiB/compact18.41MiB였습니다. **compact의 RAM 절감은 입증되지 않았습니다.** 누적 할당은44.84MB→21.61MB로 작았지만 현재 남은 Go heap도 compact가 조금 컸습니다. JSON 다음 compact의 고정 순서·미리 읽은 파일·한 번의 측정이므로 일반적인 가속률도 주장하지 않습니다.82.3MiB 전체 왕복과 직접 비교해서 메모리 개선이라고 계산하지 않습니다. 파일 캐시 preview로 남기며 기본 loader는 그대로입니다. [실제 단독 측정과 검토](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/standalone-reader-74).

[한영 상세·고정 계획·원래 결과·기준선](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/compact-storage-73) · [English](Compact-Storage-73-EN) · [새 자료 관찰](Source-Observation-73-KO)
