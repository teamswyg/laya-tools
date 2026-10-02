# Source80 후보 관찰기: 순수 골격 준비

이 인계는 원본 패키지를 실행하기 전의 관찰 스키마와 합성 callback 검증입니다. 기존 UUID·Ordinal 원본, 80의 관찰기, 실제 모델은 실행하지 않았습니다. 원본을 import하는 adapter와 실행용 main, 디스크 fsync 구현, 바이너리 및 최종 실행계획은 아직 없습니다.

새 제안은 기존 요청 3개와 후보 코드 9개의 조합입니다. 원래 입력 24개를 목표 → 후보 → fixture 순서로 배치하면 후보 관찰 72개가 됩니다. 이것이 72개 독립 부모 요청이나 학습 정답을 뜻하지 않습니다. 기존 원천 2가족과 원문 20파일을 그대로 연결할 예정입니다.

## 지금 준비한 인터페이스

`pure.Channels`는 기존 80의 JSON 필드 11개를 유지합니다. UUID·receiver·string·반환 오류·panic의 포인터 필드는 관찰되지 않은 null과 실제 빈 값 또는 false를 구분합니다. `Want_observed.nonstring_panic_text_unobserved`는 별도 메타데이터이며 Channels 필드를 늘리지 않습니다.

`pure.Literal`은 봉인된 Want의 72행 좌표, 원 입력 전체 객체, 네 가지 원문 참조, 원 입력 SHA, 소스 구간 및 null 감독 메타데이터를 보존하는 타입입니다. 후보 소스 참조의 `compiled`/`executed`와 `actual_candidate_run`/`actual_parent_created`는 Want 준비 당시 false 값입니다. 실제 관찰 카운터로 바꾸지 않습니다.

`pure.Check72Coordinates`는 닫힌 3×3×8 좌표와 ID·함수명을 검사합니다. `pure.Run`은 주입된 callback만 호출합니다. 실제 디스패치 전에 Save callback이 성공해야 하며, 반환 오류의 Error 메서드도 별도 예약·Save 뒤 최대 한 번 호출합니다. 디스크 내구성은 향후 Save adapter가 fsync를 구현해야 확보됩니다. 합성 테스트가 fsync를 증명하지는 않습니다.

원래 Scan 입력의 nil interface, typed-nil bytes, 비nil 빈 bytes를 `InputAudit`에 구분합니다. receiver-before는 wrapper에 전달한 원래 인자이며 ScanClearBefore 내부 zero receiver라는 뜻이 아닙니다. 후보 panic에서는 미반환 UUID·오류 채널을 null로 유지하고 panic payload의 Error/String/Format을 호출하지 않습니다. 예상 밖의 Error 메서드 panic도 primary panic과 구분하여 남깁니다.

## 예산과 기록

예상값은 후보 72회, 정상 반환 69회, 후보 panic 3회, 명시적 반환 Error 10회입니다. 예상 tracked 총합은 82회이고, 최대 예산은 후보 72회 + Error 48회 = 120회입니다. 후보 내부 원 API 56회는 소스에서 읽은 정적 매핑입니다. 동적 계측이 없으므로 actual 값은 null입니다. startup 원문 4구간 역시 실제 init 카운터가 아닙니다.

합성 race 테스트 2회와 vet 2회가 모두 통과했습니다. 첫 race는 테스트 10개와 하위 사례 5개, 봉인된 Want 메타데이터 보강 후 두 번째는 11개와 하위 사례 5개입니다. 실패는 0이며 두 번째 도구 표시가 잘린 사실 1건을 장부에 남겼습니다. 첫 골격의 Go 원문 3개는 inert history로 보존했습니다.

원본 compile/import/init/API, native observer 실행, 모델, Features, Project, Fit, 공유 저장소 수정 및 외부 쓰기는 모두 0입니다. 원 실행은 Root의 peer 검토와 다음 범위 확정 이후 별도 단계입니다. 학습 truth·역할·가중치·승격은 생성하지 않았습니다.

## 고정 참조와 남은 구현

- 후보 소스: SHA `a90b7d87f889319a3b9b6601eb0d654db8973a58ca0c46798620a9f8d3dedde2`, 1,620 B, inert 복사.
- literal Want72: SHA `798f0e8cbe99534599aac4c2c89a8bf25caaf97121d039600f0629d45afe76c2`, 244,161 B. 작성자는 별도 Source80 저자이며 이 observer 작성자는 아닙니다. 비맹검 source-derived 제안입니다.
- Want handoff: SHA `e45af5a37ba083e2d074bbbed5f44a31b00f9499777899c190a940f282ce4833`.
- [순수 타입](pure/types.go.txt), [callback 골격](pure/run.go.txt), [합성 검사](pure/run_test.go.txt), [실행 장부](PREPARATION-LEDGER.v1.json).

다음 구현은 exact-byte Want/source loader, 닫힌 후보 adapter, 기존 20원문과 라이선스 복사·핀 검증, fresh 출력 예약과 fsync checkpoint, compile-only 바이너리 및 최종 계획 봉인입니다. 원본 초기화는 main 전에 발생할 수 있으므로 실행파일의 도움말도 현재 단계에서 실행하지 않습니다. 향후 바깥 controller는 CPU 1, Go soft heap 256 MiB, 60초, 출력 1 MiB 미만, 자동 재시도 0을 별도로 결속해야 합니다. 이 설정은 실측이나 hard RSS 보장이 아닙니다.
