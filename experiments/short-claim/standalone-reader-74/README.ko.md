# 단독 입력 읽기: 파일 감소와 메모리 감소를 구분하기

JSON과 compact를 실제 새 프로세스에서 한 번씩 읽었고,24개 열의 수치 비트·길이·null·메타데이터가 같았다. JSON은 배열 할당 전에 크기를 확인하는2-pass 구현이고 compact는 기존 codec이다. 원래 학습·특징·정답은 다시 실행하거나 바꾸지 않았다.

최대 RSS는 JSON18.23MiB, compact18.41MiB로 **RAM 절감은 입증되지 않았다**. compact는 파일과 누적 할당이 작았지만 마지막 Go heap도 조금 컸다. 한 번의 warm-file·고정 순서·전체 child 측정으로 일반적인 속도 개선률을 계산하지 않는다. 이전82.3MiB 전체 왕복 측정과도 직접 비교하지 않는다.

사용 관점에서는 compact를 선택 가능한 파일 캐시 실험으로 남긴다. 기본 loader·입력 상한은 그대로다. 간단한 JSON 공백 제거와 gzip만으로도 저장·전송을 줄일 수 있으므로 전용 형식을 기본으로 채택할 근거는 아직 없다.128→512나2,400개 확장 시 전체 RSS가 작을 것이라는 보장도 없다.

[전체 수치와 한계](RESULTS-READER74.ko.md) · [실제 controller 결과](actual/controller.results.json) · [저장 기록 검토](independent-runtime/REVIEW.v1.ko.md) · [공개 사본 장부](SAFE-COPY-LEDGER.v1.json).

원본 plan의 host 경로는 별도 파생 공개 view로 대체했으며 view는 실행용이 아니다. 바이너리·원본 입력 blob·raw 로그·private helper는 Git에 없다. 소스는 `.go.txt` 참고 보관이며 public runtime에 연결하지 않았다. 이 README는 기존32파일 공개 stage에 root가 추가한 요약으로, stage 장부를 소급 변경하지 않는다.

[English](README.en.md)
