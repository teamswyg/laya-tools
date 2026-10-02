# 80 관찰기: Want 봉인 전 스키마 준비

이 자료는 private schema 설계입니다. Adapter, original package import/compile, native init/API, worker, 모델/학습 실행은 아직 0입니다. source-first80의 세 목표/24 입력과 두 원천 가족을 유지하고 별도 작성자의 Want SHA를 받은 뒤에만 adapter를 작성합니다.

- UUID Parse 8개: 반환된 고정 [16]byte를 hex로 직접 복사합니다. UUID.String/MustParse/Validate를 관찰 helper로 호출하지 않습니다. 오류가 있어도 반환 배열의 부분 바이트를 보존합니다.
- UUID Scan 8개: 새 명시적 nonnil [16]byte receiver의 호출 전/후를 별도로 복사합니다. nil interface, typed nil bytes, nonnil empty bytes와 dynamic type을 구별합니다. byte 입력의 직접 16-byte copy와 내부 재귀 호출은 관찰기 직접 호출 수와 구별합니다.
- 원본 Ordinal 8개: int 입력과 반환 string을 그대로 기록합니다. 전체 humanize 패키지가 아니라 unchanged ordinals.go + 원 go.mod + 완전 MIT LICENSE의 명시적 slice입니다.

직접 entrypoint 예상 예산은 24입니다. 오류 5개에서 명시적 Error 메서드 관찰 5개(원 UUID URN 1개 + 표준 라이브러리 errorString 4개)가 예상됩니다. 내부 Parse/Scan 재귀와 fmt의 내부 formatting은 동적 계측하지 않습니다. 다른 실제 오류가 생겨도 Want를 바꾸거나 기록을 버리지 않고 mismatch로 보존하는 bounded 방법을 계획에 명시합니다.

Type 기록은 Error를 호출하지 않는 fmt %T를 사용합니다. 원 API panic과 관찰 Error 메서드 panic을 구별하며, 임의 panic 객체의 String/Error 메서드를 추가 호출하지 않는 serializer 정책을 봉인 전에 정합니다. 반환 채널이 없는 필드는 null이고 nil-error의 type/message도 null입니다.

원 UUID 전체 Darwin non-JS runtime source 15개 + 원 go.mod/BSD-3 LICENSE, Ordinal source 1개 + 원 go.mod/MIT LICENSE 총 20개 핀을 유지합니다. 원문 소스는 format하지 않습니다. UUID namespace 초기화 MustParse→Parse 4개는 정적 지점 예약이며 실제 init callback/return 수는 null/uninstrumented입니다. 바깥 controller의 예약이 child Start 전에 존재해야 합니다.

예정 worker는 fresh output directory와 fsync한 zero reservation/results file을 만든 뒤 callback별 durable reserve/return partial을 기록합니다. 전체 결과의 Want/Got/Match를 남기며 차이는 성공으로 고치지 않습니다. 외부 controller는 한 child, CPU 1, Go soft heap 256 MiB, 60초 process-group Kill+joined Wait, stdout 1 MiB 제안을 사용합니다. 이 설정은 측정된 RSS나 hard RSS cap이 아닙니다.

Pure schema/stub tests만 별도 package로 실행합니다. Original imports가 있는 main package 또는 `go test ./...`는 실행하지 않습니다. Want 후 native binary build는 가능하지만 init/API를 실행하는 binary/test invocation은 금지입니다. 논리 목표 3개와 finite fixture 24개를 독립 parent 24개나 새 labels로 세지 않습니다.
