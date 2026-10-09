# Decoder allocation follow-up / 디코더 할당 개선

2026-10-09 · Apple M4 Pro · Go 1.27.1

선택형 Go CPU/heap 프로파일링을 실제 별도 로컬 서버에서 검증했습니다. 32·256·1,024·4,096바이트의 합성 ASCII 입력을 순환해 8초 동안 총 **85,043회** 직렬 요청했고 모델 SHA를 확인했습니다. 판정값은 저장하지 않았습니다. 종료 후 0600 CPU/heap 파일이 정상 gzip으로 읽혔습니다. 프로파일 원본은 로컬에만 보관합니다.

CPU 프로파일의 전체 기록 구간은 서버 준비·대기·종료를 포함해 약 155.73초였습니다. 따라서 이 실행을 8초 동안만 기록한 CPU 프로파일이나 기본 상태의 처리량으로 해석하지 않습니다. heap의 `alloc_space`에서는 JSON 스트리밍 버퍼의 fetch가 누적 할당의 약 35.65%, `io.ReadAll`이 약 23.44%로 관측됐습니다. 이는 프로파일의 추정 누적 할당 비중이며 현재 사용량이나 최대 메모리가 아닙니다. heap 표본에서 작은 모델의 전체 메모리를 추정하지 않습니다.

측정 근거에 따라 JSON 입력을 다시 스트리밍 디코더에 넣는 대신 이미 제한된 바이트에서 직접 읽도록 변경했습니다. 단일 `text` 문자열 규칙, 정확한 키 이름, 중복·미지 필드·null·후속 JSON 거부와 UTF-8 처리는 보존했습니다. 원래 디코더와 비교하는 fuzz 검사가 **286,999회** 실행되어 통과했습니다. 별도 공용 코드 검토에서도 필수 수정점이 없었습니다. 모델·특징·판정 기준은 변경하지 않았습니다.

[측정 원문](decoder-benchmark.txt)은 디코더만 비교합니다. 4KB 입력에서 20,896 → 4,120 B/op, 18 → 3 allocs/op, 5,178 → 1,819 ns/op였습니다. 이는 약 **80.3% 적은 할당 바이트와 2.8배 빠른 디코딩**이며 서버 전체 상주 메모리나 모델 추론 속도의 개선 수치가 아닙니다. 입력은 같은 ASCII 문자열이고 다른 검토·fuzz 작업도 진행 중이었으므로 현실적인 문장이나 고립된 CPU 조건의 수치가 아닙니다. 특히 반복된 작은 문자열의 결과를 다양한 입력 전체에 일반화하지 않습니다.

재현 명령:

```sh
go test ./internal/claimsdemo -run '^$' -bench BenchmarkDecodeText -benchmem -benchtime=100ms
go test ./internal/claimsdemo -run '^$' -fuzz FuzzDecodeTextCompatibility -fuzztime=10s
```

English: Optional local CPU/heap profiling was exercised with 85,043 serial synthetic ASCII requests across four lengths over eight seconds, validating the response model SHA and discarding predictions. The complete CPU recording spanned about 155.73 seconds including setup/idle/shutdown; it is not an isolated eight-second baseline. Raw profiles remain private. Sampled cumulative allocation attributed about 35.65% to JSON streaming-buffer fetch and 23.44% to `io.ReadAll`; these shares are not resident or peak memory.

Direct decoding of the already-bounded request bytes preserves the strict single-string schema and legacy UTF-8/escaped-surrogate behavior. Differential fuzzing passed 286,999 executions and independent public-source review found no required fixes. At 4KB, the decoder-only benchmark changed from 20,896 to 4,120 B/op, 18 to three allocations/op and 5,178 to 1,819 ns/op: about 80.3% fewer allocation bytes and 2.8× faster decoding. Model features and decision thresholds are unchanged. This does not establish total server memory reduction, inference speedup or semantic quality; repeated ASCII inputs and concurrent work limit generalization.
