# 실행 전 기대값 80: UUID Parse·Scan과 Ordinal

이 자료는 동결된 공개 원문과 입력을 읽어 작성한 **실행 전 예상값**입니다. 실제 반환값은 아직 없고, 모든 `Got`은 null입니다. 입력 24개는 Parse 8개, UUID.Scan 8개, Ordinal 8개로 나뉩니다. 행동 목표 3개·원천 계열 2개이며, 24개의 독립 부모나 학습 라벨을 만들었다는 뜻은 아닙니다.

입력 원본은 source-first80의 `INPUT-DRAFTS.v1.json` 8,265바이트, SHA256 `a6110ab85d52fa4b8c62af6aab26bd13a988f04a0a5e38ce1e24e26542fa165b`입니다. [WANTS.v1.json](WANTS.v1.json)은 각 입력의 JSON pointer, 원 객체, 순서와 ID를 보존합니다. source-first 작성자는 semantic_review60_prep, 기대값 작성자는 checkpoint_cli_45입니다. 서로 작성 담당은 다르지만 같은 프로젝트의 이전 자료에 노출된 AI 보조 개발이며, 블라인드 검증이나 독립적인 인간 원천을 주장하지 않습니다.

| 구간 | 원문에서 얻은 좁은 기대값 |
|---|---|
| raw32 정상 | `00112233445566778899aabbccddeeff`, nil 오류 |
| raw32 늦은 `ge` 오류 | `00112233445566778899aabbccddfe00`, `invalid UUID format` |
| standard36 정상 | 같은 정상 16바이트, nil 오류 |
| standard36 늦은 `ge` 오류 | `00112233445566778899aabbccdd0000`, 같은 오류 |
| mixed-case URN | EqualFold 경로를 거쳐 정상 16바이트 |
| 잘못된 URN prefix | 모두 0인 16바이트, `uuid.URNPrefixError`, `invalid urn prefix: "bad:uuid:"` |
| 38바이트 `{…}`와 `[…]` | Parse는 바깥 문자를 검사하지 않아 둘 다 정상 16바이트 예상 |
| Scan nil·빈 string·typed nil bytes·nonnil empty bytes | 새로 준 기존 receiver `ffeeddccbbaa99887766554433221100` 유지 |
| Scan 정상 text·raw16 bytes | 정상 16바이트로 receiver 변경 |
| Scan 늦은 잘못된 text·bytes | receiver 유지, `Scan: invalid UUID format` |
| Ordinal 0,1,2,3,11,12,13,112 | 각각 `0th`, `1st`, `2nd`, `3rd`, `11th`, `12th`, `13th`, `112th` |

raw32의 실패 슬롯은 값을 **대입한 뒤** 유효성을 검사합니다. `g`의 xvalues 값 255와 `e`의 값 14가 byte 연산에서 `(255<<4)|14 = 0xfe`가 되므로 그 슬롯은 fe이고 뒤 슬롯은 00입니다. standard36은 검사 후 대입하여 실패 슬롯과 그 뒤 슬롯이 00입니다. 이 부분 반환값을 공통된 '오류면 전부 zero'로 바꾸지 않습니다. Parse의 38바이트 처리와 별도 Validate 함수의 bracket 검사는 같지 않으며, Validate는 이번 관찰에서 호출하지 않습니다.

Scan은 각 입력마다 별도의 nonnil 16바이트 UUID literal을 받습니다. MustParse나 FromBytes로 receiver를 준비하지 않습니다. nil interface, typed nil []byte, nonnil empty []byte는 입력 객체에 각각 보존하며, 임의로 null이나 빈 배열 하나로 합치지 않습니다. 잘못된 text에서 지역 Parse 값은 오류 뒤 receiver에 commit되지 않습니다. non16 bytes는 Scan(string(src))로 재귀하는 원문 경로이며 직접 관찰 입력은 1개입니다.

오류 type은 반환된 오류에 `%T`를 적용한 문자열로 비교하고, 메시지는 nonnil 오류마다 `Error()`를 한 번만 호출해 얻도록 제안합니다. Go1.27.1의 errors.New 및 fmt.Errorf 원문도 읽었습니다. UUID format 오류 2개와 Scan 오류 2개는 `*errors.errorString`; URN prefix 오류 1개는 `uuid.URNPrefixError`를 예상합니다. Scan의 `%v`는 `%w` wrapping 약속이 아닙니다. Error identity·Unwrap·errors.Is는 기대값 null/범위 밖입니다. Ordinal은 오류 반환 채널이 없으므로 `error_nil`도 null입니다. 해당 API에 없는 UUID·receiver·출력 채널 역시 명시적인 null입니다.

직접 entrypoint 예산은 24회입니다. 위 다섯 오류의 직접 Error 관찰을 포함한 추적 dispatch 제안은 29회이며, 이 중 원 UUID Error 메서드는 1회, 표준 라이브러리 Error 메서드는 4회입니다. Scan 내부 Parse 3회·Scan 재귀 1회·fmt formatting 등은 소스 경로 해석이고 동적 계측이 아닙니다. UUID namespace 초기화의 MustParse→Parse 4개 정적 호출 사이트도 실제 callback 반환 수와 구분합니다. 실제 init 반환 수는 null, 개별 계측은 false입니다. 미래 native child는 외부 예약 뒤 시작해야 하며 전체 수명 OS 측정에는 초기화·관찰·저장이 포함됩니다.

원천은 google/uuid revision `2d3c2a9cc518326daf99a383f07c4d3c44317e4d`의 Darwin non-JS runtime 15파일+원 go.mod+BSD-3-Clause 전문과, dustin/go-humanize revision `a1b4e66b9a6d890e9e15e7091cf16c8032367d6e`의 **원 Ordinal 1파일 slice**+go.mod+MIT 전문입니다. 후자는 전체 humanize package가 아니며 Ordinal64를 추가하지 않습니다. 파일별 SHA와 라이선스 핀은 WANTS의 source_bindings에 있습니다. 공개 복사 시 원문 고지·전문 라이선스를 유지해야 합니다.

제안된 CPU 1, Go soft heap 256MiB, 외부 60초, 출력 1MiB 경계는 미래 실행 계획용입니다. soft heap은 OS RSS hard cap이 아닙니다. 원본 compile/import/init/API/test/adapter/observer/model/Fit/역할/라벨/공유 수정/게시 실제 실행은 모두 0입니다. [CHECKS.v1.json](CHECKS.v1.json)은 작성자 측 metadata 검사이며 원 함수 테스트나 독립 의미 검증이 아닙니다. 넓은 RFC 검증, 모든 입력의 no-panic, 다양성 충족, 학습 준비·production·protected-final 승인은 하지 않습니다. 결과가 다르면 v1을 그대로 보존하고 mismatch를 기록해야 합니다.
