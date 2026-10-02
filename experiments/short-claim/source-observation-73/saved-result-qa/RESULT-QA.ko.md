# 첫 3원천: 저장 결과 대조

표준 라이브러리 Go 검사기를 1회 실행했고 성공 1회·실패 0회입니다. 원 API, 관찰기, controller를 다시 실행하지 않았습니다. 검사자는 controller 작성자이므로 이 결과는 **저장 파일 대조**이며 독립적인 실제 실행이나 광범위 의미 승인으로 설명하지 않습니다.

고정 Want에 있는 24개 Probe의 ID·순서·입력·초기 receiver·Want가 실제 Records에 그대로 남았고, Got 전체 및 Matches 플래그도 일치했습니다. datasize 10개, query 6개, shlex 8개이며 반환 24회, panic 0회, 불일치 0개입니다. `uint64` 최댓값 3개를 실수 변환 없이 비교했습니다. 오류 관측 8개도 원래 기대한 Kind·Text·Function·Input·Cause와 일치하므로 실패를 지운 결과가 아닙니다.

query의 nil map 1개·할당된 빈 map 2개·내용 있는 map 3개와 값 순서, shlex의 할당된 빈 slice 1개를 구분해 확인했습니다. 최종·중간 결과는 바이트가 같고, 동결 계획은 원래 draft의 Frozen 토큰 1개만 바뀌었습니다. 결과·계획·Want·controller 장부·원시 로그의 SHA 연결이 일치했습니다.

저장된 OS 기록도 로그와 일치합니다: child real 0.85초, user 0.01초, sys 0.04초, 최대 RSS 18,743,296 bytes, footprint 16,122,360 bytes, controller elapsed 0.854884초입니다. 새 측정이 아니며 초기화·사전 검증·관찰 wrapper·checkpoint I/O가 포함된 전체 child 기록입니다. 순수 API 또는 Laya 추론 성능, RSS 하드 제한으로 해석하지 않습니다.

24개는 행동 목표 3개의 유한 probe이며 새로운 독립 부모 24개가 아닙니다. 새 부모·라벨·역할·모델·학습·공유 파일 변경·외부 게시는 0입니다. caption/parent-label/broad-truth/training qualification은 모두 false를 유지합니다. `RECEIPT.v1.json`, `INVOCATION.v1.json`, `LEDGER.v1.json`이 exact 파일 핀과 시도를 보존합니다.
