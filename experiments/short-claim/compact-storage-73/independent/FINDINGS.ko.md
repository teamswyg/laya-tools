# Compact 저장 prototype 독립 읽기 검토

현재 고정 파일의 성공한 왕복 결과를 뒤집을 코드 결함이나, 저장 배열이 원래 JSON 값과 다르다는 새 근거는 발견하지 못했습니다. 이는 source-reading 결론입니다. 이 검토자는 codec 작성자가 아니며 이전 projection/학습 준비에 참여한 비맹검 AI 보조 검토자입니다. 원본 데이터 왕복, decoder, 테스트와 helper를 재실행하지 않았습니다.

Binary Decode는 네 dataset의 전체 길이·개수·nil flags·reserved bytes·곱셈/합산·64MiB 한도를 먼저 확인한 다음 배열을 할당합니다. 각 열은 원래 순서대로 읽고, uint16 indices와 signed64 offsets/groups를 유지하며, FP64는 bits로 저장하고 복원합니다. nil과 빈 배열은 별도 flags로 남습니다. CSR 길이·offset 범위와 NaN/Inf 거절도 코드에서 확인했습니다. 저장 codec의 유한값·구조 검사는 원래 모델/학습 입력의 적격성 검사를 대신하지 않습니다.

JSON import는 여섯 배열만 typed DTO로 옮기고 Split/Excluded 및 나머지 envelope/parent/ref/null/false metadata는 RawMessage로 남깁니다. 일반 float64 map으로 metadata를 다시 계산하지 않는 점은 타당합니다. 공백·key 순서는 바뀌므로 원문 bytes 재현이 아닙니다. Source SHA는 원문 식별자이고 내부 checksum이 아니므로 이후 reader는 별도의 compact asset SHA와 source pin을 확인해야 합니다. 이 검토는 binary metadata를 새로 decode하지 않았습니다.

고정 관측을 보존하면서 후속 v2 또는 공개 loader 전에 구분할 제한은 세 가지입니다.

- `codec/json.go:77–83,113–114`: ImportJSON은 typed 배열을 할당한 뒤 총 owned payload를 검사합니다. 64MiB 파일 한도나 최종 payload 통과가 JSON import의 최대 할당을 64MiB로 제한하지 않습니다. 일반 입력을 받을 때는 배열 원소 수를 먼저 세는 bounded 단계가 필요할 수 있습니다. Binary Decode의 사전 검사와 구분해야 하며, 현재 고정 성공을 실패로 바꿀 이유는 아닙니다.
- `codec/json.go:20–33`: object를 map으로 읽어 중복 key를 마지막 값으로 합칩니다. 일반 JSON에 대해 모든 원문 속성을 보존하는 계약은 아닙니다. 현재 고정 JSON에 중복 key나 값 손실이 있다는 새 근거는 없습니다. 향후 일반 입력 계약에서 중복 key 거절 여부를 명시하면 됩니다.
- `cmd/roundtrip/main.go:87–105`: 기존 비교는 JSON에서 ImportJSON이 만든 Snapshot과 같은 codec으로 Decode한 Snapshot의 일치입니다. importer가 같은 방식으로 잘못 읽었다면 이 비교만으로 잡히지 않습니다. 기존 성공1/재시도0을 그대로 보존하고, 향후 별도 standalone reader가 source JSON의 각 경로/순서/정수/FP64 bits/null 상태와 metadata를 직접 대조하는 실험을 분리하는 것이 적절합니다. 이번에는 그 비교를 실행하지 않았습니다.

저장량 비교는 명확합니다. Pretty JSON 8,390,462B, minified JSON 4,697,570B, compact 1,978,694B입니다. Pretty 대비 76.4173%에는 공백 제거 효과가 함께 들어 있으며 minified 대비 감소는 57.8783%입니다. Gzip은 각각 594,236/400,256/328,206B입니다. Compact gzip은 minified JSON gzip보다 72,050B 작지만, JSON+gzip은 기존 도구와 schema를 활용하기 쉽고 전용 binary codec 유지 비용이 적습니다. Compact는 typed 배열을 바로 복원할 가능성이 있지만 그 이익은 독립 읽기·실제 loader·속도/메모리 비교 전에는 가설입니다. 이번 gzip 관측은 압축 파일 크기만 비교했고 해제·속도·RSS는 측정하지 않았습니다.

새 학습·모델·특징 계산·역할·label·최종 자료 읽기·공유 수정·게시 모두 0입니다. 기존 성공/실패 기록과 원본 파일을 수정하지 않았습니다. 대화에 사용된 AI 보조 검토의 비용은 측정하지 않았고, 별도 model/paid trial 0을 이 비용이 0이라는 뜻으로 사용하지 않습니다.
