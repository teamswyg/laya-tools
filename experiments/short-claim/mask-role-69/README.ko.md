# 역할과 학습 제외 항목의 연결69

[English](README.en.md) · [상세 결과](MASK-ROLE-AUDIT.ko.md) · [Go 변환 연결](GO-BRIDGE-HANDOFF.ko.md)

실제 역할66과 설명 검토67을 원래72개 요청·216개 후보에 연결했다. Mask는 그 설명을 학습 손실에 반영할지의 표시다. 제외된 후보도 원래 후보 목록과 함수 정답에는 남으며, unknown의 정답은 null이다.

| 역할 | 양성 적격 | 음성 적격 | known 제외 | unknown | 적격 그룹 |
|---|---:|---:|---:|---:|---:|
| train |20|52|18|42|9|
| validation |10|25|1|12|3|
| calibration |5|18|4|9|3|

train의 그룹52는 known 후보 전부가 제외되어 적격 그룹 수가10에서9로 줄었다. 역할이나 mask를 바꾸지 않았으며 기존9·3·3 준비 하한을 유지한다. 양성 후보가 있는 그룹 수7·3·2는 설명용 수치이고 새로운 하한이 아니다. Calibration은 이후 보류 정책 검토용으로 남기며 학습에 넣지 않는다.

읽기 전용 metadata 집계1회로 총792개 연결 검사를 통과했다. 원래 역할·동작·특징·학습 함수를 다시 실행하지 않았다. [실제 결과](results-69.json), [예약](INVOCATION-69.json), [장부](ATTEMPT-LEDGER.json), [복사 장부](QA-COPY-LEDGER-92.v1.json)를 보존한다. 새로운 upstream4개 요청은 아직 이72개에 합치지 않았다.
