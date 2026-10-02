# 저장 형식 roundtrip 관측

고정 JSON 8390462B를 private compact 1978694B로 저장했고, metadata 163758B와 네 CSR의 모든 FP64 bit·integer·null/empty 상태가 일치했습니다. 저장 크기 감소 76.4173%, 원래 uint16 폭·diagnostic 중복은 그대로입니다.

정확한 집계·pin은 ROUNDTRIP-HANDOFF.v1.json에 있습니다. Owned array+metadata payload 1978406B, 원본 published payload1,928,154B는 회계 범위가 다릅니다. Go heap/전체 RSS/파일 크기는 별개이며 한 warm 순차 표본의 시간은 인과 속도 개선이 아닙니다. SourceJSON 공백 재현은 주장하지 않습니다.

순수 codec race test1·vet1·build1 통과, 저장 roundtrip1·retry0. 원본 Projection/Features/Prepare/Fit/role/label/behavior/model API·HF·공유 수정 모두0. Raw JSON·compact·binary·host 경로·로그는 private이고 새 학습/성능 승인/2,400 성공은 false입니다. 128/512 확대에서 I/O·텍스트 parsing 병목을 줄일 가능성은 있으나 확대 실험이나 loader 변경은 아직 없습니다.
