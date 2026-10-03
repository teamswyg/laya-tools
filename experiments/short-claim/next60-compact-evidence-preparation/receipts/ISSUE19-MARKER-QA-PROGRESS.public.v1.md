이번 진행: 실제 원본 연결 전, 관측 코드의 정확성과 파일 크기 한도를 점검했습니다.

- [x] [PR #129](https://github.com/teamswyg/laya-tools/pull/129): Linux·macOS·민감 정보 검사·quality 모두 성공, CI 봇이 자동 병합. 새 26개 검사는 자체 합성 검사입니다.
- [x] 원본 함수의 반환/실패와, 반환값을 복사·기록하다 생긴 실패를 별도로 표현하는 전송 형식: 19개 합성 검사 + 비작성자 2명의 소스 검토.
- [x] 실제 원본 연결용 관측 로직: 원본 호출 marker를 AFTER 기록 전에 확정하도록 수정. 11개 합성 검사(추가 하위 사례 5개), 일반 모드와 panic(nil) 호환 모드 race 검사, vet·포맷 검사 성공. 확인 응답 손실은 호출 효과가 없었다는 뜻으로 바꾸지 않습니다.
- [x] 컴파일 준비 코드의 자체 검사 5개와 vet 성공. 실제 실행기는 아직 활성화하지 않았습니다.
- [x] 초기 실패 원인과 수정 전 소스 보존: 테스트의 줄바꿈 전달 오류, 외부 구조체 필드 지정 누락, method 전용 marker가 상위 candidate 기록으로 전파된 문제.
- [ ] 결과 저장 구조 보완: 빈 행까지 포함한 39행 고정 배열의 자체 측정이 **179,946바이트**였습니다. 고정 결과 파일 예산 **65,536바이트**와 기존 소스의 131,072바이트 상한을 모두 초과합니다. 한도는 늘리지 않고 중복 데이터를 줄이는 두 구조를 검토합니다.
- [ ] Union 중간 읽기 결과와 JSON Map의 값 목록을 빠짐없이 연결하고, 작은 결과 파일과 검증 기록이 정확히 대응하는지 확인.
- [ ] 원본 소스·라이선스·컴파일 산출물 연결, 실제 프로세스 종료/회수·파일 동기화 확인 후 원본 실행.

이번 단계에서 새 원본 실행·모델 추론·학습·Golden 행 추가는 **0**입니다. 기존 35개 개발 사례와 제안한 20→60 Golden cohort는 서로 다릅니다. 보호된 도메인별 2,400개 평가와 전체 검증 작업량 5% 개선도 아직 미달성입니다. 테스트 통과를 모델 성능 개선으로 계산하지 않습니다.

English: PR #129 passed all four required CI checks and was merged by the bot. The new native-outcome channel passed 19 owned controls and two nonauthor source reviews. The owned binder passed 11 top-level controls plus five subcases in both panic modes; compiler preparation passed five controls. Initial failures and their source snapshots are retained. A padded 39-row owned result measured 179,946 bytes, exceeding the fixed 65,536-byte result reserve. We are comparing lossless compact output designs without enlarging the budget. Native source provenance, all intermediate observations, result/journal correspondence, process and durable-file lifecycle remain pending. New original executions, model inference, Fits and Golden rows: zero; protected 2,400/domain and 5% whole-work improvement remain unachieved.
