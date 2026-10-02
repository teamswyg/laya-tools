# 숫자를 유지하고 저장 크기 줄이기

다음 자료 확대에서 큰 JSON 파일이 먼저 입력 한도에 닿을 수 있어, **학습과 별도로 저장 방식만** 시험했다. 기존 76개 자료의 한 snapshot을 한 번 변환·읽기·대조했으며, [원래 FP64 bit·uint16·CSR·null 상태가 일치](ROUNDTRIP-HANDOFF.v1.json)했다. 네 CSR과 진단용 중복, metadata와 미승인 상태는 그대로다.

| 형식 | 파일 bytes | gzip bytes |
|---|---:|---:|
| 원래 pretty JSON | 8,390,462 | 594,236 |
| 공백 제거 JSON | 4,697,570 | 400,256 |
| FP64 compact | 1,978,694 | 328,206 |

[별도 크기 기준선](SAVED-BYTES-CONTROL.v1.json)에서 공백 제거만44.01%, 그 JSON 대비 compact는57.88% 감소했다. 원래 파일 대비76.42%는 두 효과를 합친 수치다. gzip은 동일 표준 설정으로 각각 한 번 생성했으며 캐시·전송 크기 비교다. 압축을 풀면 큰 JSON이 돌아오므로 decoded 입력64MiB 한도를 자동 해결하지 않는다.

**파일이 작아졌다고 RAM이 줄었다고 결론내리지 않는다.** [전체 worker OS 기록](ROOT-OS-MEASUREMENT.v1.json)의 peak RSS는86,343,680B다. JSON 읽기·변환·compact 쓰기·읽기·decode·보고서 생성이 함께 포함된 실행이며 단독 decoder가 아니다. 원래 projection payload1,928,154B와 새 array+metadata1,978,406B는 계산 범위도 다르다. 한 순차 warm 표본의 stage 시간을 서빙 속도 개선률로 사용하지 않는다.

[독립 소스 검토](independent/FINDINGS.ko.md)에서 고정 snapshot의 성공을 뒤집을 blocker나 배열 불일치의 새 근거는 없었다. 재실행 없이 구조를 읽은 검토다. Binary decoder는 네 dataset 모두의 count·실제 길이·64MiB 한도를 할당 전에 검사한다. 반면 JSON import는 typed allocation 뒤 payload 한도를 확인하고 중복 key가 map에 합쳐질 수 있으므로 일반 입력을 받는 runtime으로 바로 옮기지 않는다. 같은 codec의 왕복을 독립 source-array reader 비교로 부르지 않는다.

현재 공유 loader나 runtime을 변경하지 않았다. codec 순수 race/vet/build와 저장 roundtrip은 통과했지만, 더 큰 자료와 단독 reader의 메모리·할당·수치 확인은 후속이다. JSON 공백 제거와 gzip만으로 충분한지, 별도 codec의 유지 비용을 감수할 이점이 있는지 그 결과로 선택한다. 기존 파일·모델·실험 결과는 덮지 않았다. 원본 feature·projection·학습·API·역할·정답 변경·최종2,400 자료 읽기는0이다. 바이너리·compact 본문·원시 로그·호스트 경로는 공개하지 않는다.

[실험 범위](PLAN.ko.md) · [Roundtrip](ROUNDTRIP-HANDOFF.v1.ko.md) · [별도 기준선](SAVED-BYTES-CONTROL.v1.ko.md) · [English](README.en.md)
