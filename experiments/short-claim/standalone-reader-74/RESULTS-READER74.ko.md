# 같은 저장 결과를 읽는 JSON·compact 비교

이번 한 번의 비교에서 compact reader의 누적 Go 할당과 실행 시간은 더 작게 관측됐지만, RAM 절감은 입증되지 않았습니다. 최대 RSS는 JSON 19,120,128 B, compact 19,300,352 B로 compact가 180,224 B 더 높았고, 종료 시 Go heap도 104,848 B 더 높았습니다. 모델·학습·원본 projection·features를 다시 실행하지 않은 저장 형식 읽기 실험입니다.

기존 76-parent projection의 JSON 8,390,462 B와 compact 1,978,694 B를 각각 새 프로세스 하나에서 읽었습니다. JSON reader는 숫자 배열을 먼저 토큰으로 훑어 길이와 소유 payload 상한을 확인한 뒤 두 번째 pass에서 배열을 할당합니다. compact reader는 기존 Decode를 그대로 썼습니다. 양쪽 모두 같은 배열 검증과 strict metadata 검사를 거쳤습니다. 입력 pin을 controller가 미리 읽었으므로 warm-file 상태이며 JSON 다음 compact의 고정 순서, 각 1회 표본입니다.

| 이번에 저장된 측정값 | JSON reader | compact reader |
| --- | ---: | ---: |
| OS real / user / sys, 초 | 0.49 / 0.08 / 0.01 | 0.05 / 0.01 / 0.00 |
| controller가 본 child wall, 초 | 0.50100075 | 0.058869916 |
| OS 최대 RSS, B | 19,120,128 | 19,300,352 |
| OS peak footprint, B | 16,269,840 | 15,745,528 |
| 종료 시 Go HeapAlloc, B | 4,660,664 | 4,765,512 |
| 누적 Go TotalAlloc, B | 44,840,120 | 21,608,800 |
| 소유 배열+메타데이터 payload, B | 1,978,406 | 1,978,406 |

RSS는 전체 프로세스의 관측이고 HeapAlloc은 마지막 Go heap 상태이며 TotalAlloc은 실행 중 누적 할당량입니다. 누적 할당량이 작다는 것은 peak RAM이 작다는 뜻이 아닙니다. peak footprint는 compact에서 더 작았지만 RSS와 마지막 heap은 더 컸으므로 포괄적인 RAM 이득으로 승인하지 않습니다. `0.00` sys는 시간 표시의 반올림이며 CPU 사용이 없다는 뜻이 아닙니다. Go 256MiB는 soft target이고 소유 payload 64MiB는 tokenizer/scratch/배열 headers/allocator/전체 RSS를 포함하는 강제 메모리 상한이 아닙니다.

실제 root 실행 기록에서 두 child는 각각 start 1회·retry 0회·exit0·timeout/overflow 없음·Wait 반환으로 완료했습니다. 종료코드만 보지 않고 `BothCompleted`와 `ArrayAndMetadataDigestsEqual`, 완료 상태를 함께 확인했습니다. 네 데이터셋의 여섯 열, 총 24열에서 count/null/원래 순서를 유지한 수치 비트 digest와 source JSON SHA·metadata SHA가 일치했습니다. 실제 24열은 모두 nonnil입니다. 행은 91/45/73/44이고 NNZ는 67,104/37,975/37,916/37,657입니다. 이는 FP64→FP32 변환, uint16 인덱스 축소, diagnostic 데이터 중복 제거 실험이 아닙니다.

별도 검토자는 저장된 16개 파일 pin과 184개 조건을 대조했습니다. 해당 runtime QA는 보고된 count/null/digest와 기록의 연결을 확인한 것이며 원본 배열을 다시 읽거나 비트 digest를 독립 재계산한 결과는 아닙니다. 이전 소스 검토와 root의 실제 배열 읽기·digest 계산이 근거입니다. 검토자는 reader/controller 작성자와 다르지만 이전 codec/preflight에 노출되었으므로 맹검은 아닙니다. 이 공개 복사 작업의 작성자도 원래 reader/controller 작성자이며 별도의 독립 성능 검증으로 표현하지 않습니다.

한 warm-file 표본의 전체 child 시간에는 소스·바이너리 pin 검사, 입력 읽기, 배열 검증, digest와 결과 저장이 포함됩니다. JSON은 안전한 cap 검사 때문에 두 pass를 사용합니다. 짧은 시간이 관측됐다는 사실을 형식만의 인과적 가속, cold I/O, 128→512 확장, 서빙 성능, LLM 토큰 절감이나 모델 효용으로 확대하지 않습니다. 기존 저장 roundtrip의 76.4% 크기 감소는 pretty formatting 제거와 binary 저장 효과가 합쳐진 비교이며 이 읽기 측정과 별개입니다. 보호된 2400 final 완료도 주장하지 않습니다.

원본 input snapshot/compact blob/binaries/raw stdout/time/private plans/private go.mod/helpers는 이 archive에 없습니다. 두 plan은 [draft 공개 view](PUBLIC-DRAFT-PLAN.v1.json)와 [frozen 공개 view](PUBLIC-FROZEN-PLAN.v1.json)로 별도 파생했고 원본 SHA와 제거한 네 host-path 필드를 명시했습니다. 공개 view는 실행 계획으로 사용할 수 없습니다. source는 원 SHA를 유지하는 `.go.txt` 참고 archive이며 독립 실행 모듈이 아닙니다. 원본과 공개본의 exact copy·파생 차이는 [SAFE-COPY-LEDGER](SAFE-COPY-LEDGER.v1.json)에 있습니다.

[root controller 결과](actual/controller.results.json), [JSON 결과](actual/json-reader.results.json), [compact 결과](actual/compact-reader.results.json), [독립 runtime QA](independent-runtime/RECEIPT.v1.json)를 함께 읽을 수 있습니다. 연구/reader/codec/원본 API·Features·Project·Fit·역할·라벨·모델·유료 실행·HF·공유 repo 게시의 새로운 실행은 0입니다. TrainingReady/PerformanceApproved/Final2400Complete는 계속 false입니다. 별도 모델 호출 0은 이 AI 보조 개발·검토 협업의 비용이 0이라는 뜻이 아닙니다.
