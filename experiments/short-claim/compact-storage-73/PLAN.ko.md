# 저장 형식만 비교하는 사전 범위

고정된 projection76 JSON 8,390,462B, SHA `f68bff5f48747a66f038c17262568c1bd91251a8c81b3f46f7ae5090c8fe4b99` 한 파일을 대상으로 저장→읽기 roundtrip을 최대 1회 수행합니다. 원본 파일·모델을 수정하지 않고 original Projection/Features/Prepare/Fit/role/label/behavior API는 0회입니다. 코드의 모든 import는 Go stdlib와 자체 codec뿐입니다. 공유 코드·Git/HF·공개 게시도 0입니다.

Envelope의 schema/state/counters/pins, 전체 parent metadata·문구·null label·membership/refs·diagnostic denominator·현재 false qualification을 남깁니다. 네 dataset의 Split/Excluded와 나머지 metadata를 compact JSON으로 보존합니다. CSR 네 개(Development, Validation, DevelopmentAUC.Data, ValidationAUC.Data)의 Offsets/Indices/Values/Labels/Groups/SampleWeights 여섯 배열은 별도 little-endian frame에 저장합니다. FP64는 bit 그대로, uint16 index는 현재 폭 그대로, native 64-bit int offset/group은 signed64 값 그대로입니다. Null/empty 배열은 flags로 구별합니다. 공백·object key 순서만 compact metadata 직렬화로 바뀌며 원본 JSON byte-identical 재출력을 주장하지 않습니다.

Magic/version/count·source SHA·정확한 전체 길이·metadata 길이와 4×6 배열 수를 기록합니다. 곱셈·합산과 file/owned payload를 **각각 기존 64MiB 이내**로 검사하고 decode allocation 전 전체 크기를 먼저 확인합니다. Truncation/trailing bytes, flags/reserved 값, CSR offset 범위·길이 불일치와 NaN/Inf를 거절합니다. -0와 인접 FP64 값은 bitwise toy test로 확인합니다. Optional SampleWeights null/empty도 보존하며 diagnostic 중복은 제거하지 않습니다.

순수 codec race tests/vet 후 한 번의 실제 저장 DTO 읽기→compact exclusive write→read→decode에서 metadata bytes·source SHA·모든 integer/FP64 bit·배열 null 상태를 비교합니다. File bytes·array+metadata owned payload·Go 할당 snapshot·전체 worker RSS는 별개이며 warm 순차 한 표본의 stage 시간은 인과 성능/서빙 속도 증거가 아닙니다. Source JSON·새 compact·binary·raw logs·host 경로는 private이고 공개 안전한 KOEN handoff에는 SHA/수량/집계만 넣습니다. 새 학습, 효용·성능 승인, 2,400 최종 성공·index 폭 축소·FP32 수치 변경은 주장하지 않습니다.
