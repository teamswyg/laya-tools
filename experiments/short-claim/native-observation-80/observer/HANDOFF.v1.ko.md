# Source80 관찰기 준비 인계

준비 완료 상태입니다. 실제 native 관찰기, 원 패키지 초기화·API·원 테스트는 실행하지 않았습니다. 별도 작성자가 원문에서 작성하고 봉인한 Want를 받은 후에만 adapter, 원 패키지 import와 바이너리 빌드를 준비했습니다. 이 자료는 실행 결과나 학습 적격성을 증명하지 않습니다.

관찰 대상은 두 공개 원천 가족의 세 행동 목표입니다. UUID Parse 8개, UUID Scan 8개, 기존 humanize Ordinal 8개로 총 24개 유한 fixture입니다. 이를 독립 부모 24개나 새 정답 24개로 세지 않습니다. UUID는 15개 non-JS runtime Go 파일 전체, 원 go.mod 및 전체 BSD-3-Clause LICENSE를 보존했습니다. Ordinal은 원 ordinals.go 한 파일과 원 go.mod 및 전체 MIT LICENSE를 보존하는 명시적 source slice입니다. 원천·모듈·라이선스 20개와 Want는 byte-exact이며, Ordinal64나 전체 humanize 패키지는 포함하지 않았습니다.

실행 계획은 직접 API 24회와 명시적 Error 메서드 5회를 합쳐 **예상 29회**를 기록합니다. Error 관찰은 fixture당 최대 한 번만 수행합니다. 예상 밖 오류를 누락하지 않도록 보수적 최대 Error 24회, 최대 tracked dispatch 48회를 별도 기록합니다. 실제 오류 반환 채널이 있는 entrypoint는 Parse/Scan 16개이며, 48회는 느슨한 비용 상한이지 실행값이나 달성값이 아닙니다. 예상 Error 5회는 UUID URNPrefixError 1회와 표준 errorString 4회입니다. 예상과 다른 동적 오류 타입은 실제 타입과 unclassified 카운터를 보존합니다.

원 패키지의 네 namespace MustParse→Parse 초기화 위치는 원문에서 확인한 정적 site입니다. 실제 초기화 callback·return은 null이고, 내부 Scan 재귀·Parse·fmt formatting 호출도 동적 계측하지 않습니다. Go native CPU 관찰기이며 Laya·GPU 실행 준비가 아닙니다.

호출 전에는 새 디렉터리와 results.json을 배타적으로 예약·fsync하고, 각 dispatch 의도와 반환·panic을 동기 partial checkpoint로 남깁니다. 호출 전 Got/Matches는 null입니다. API 완료 뒤 UUID 16바이트 hex, Scan receiver 전후, nil interface/typed nil bytes/nonnil empty bytes, 입력 동적 타입, 문자열 결과와 nullable error 채널을 보존합니다. 오류 타입의 `%T` 관찰은 Error 메서드를 호출하지 않으며, 오류 메시지를 얻는 명시적 Error 호출은 별도 예산과 카운터를 사용합니다.

panic 값을 임의 Error/String 메서드로 변환하지 않습니다. 타입과 문자열 primitive만 보존하며, 그 밖의 panic text는 null과 미관측 flag로 남깁니다. panic이나 Want 불일치는 원 기록을 유지하고 Want를 바꾸거나 재시도하지 않습니다. 일반 불일치 상태는 기록을 마친 후 exit 0, 관측된 panic·저장 실패·불완료는 nonzero입니다. Final/partial/stdout은 같은 canonical JSON bytes를 사용합니다. 갑작스러운 종료에서는 마지막 성공한 sync snapshot만 증거이며, in-flight 반환과 storage 실패 후 모든 데이터 보존을 보장하지 않습니다.

계획은 Frozen=false입니다. Root가 별도 바깥 예약과 사전 핀 검증 후 false 토큰만 true로 바꾸어 최대 60초, CPU 1, Go heap soft limit 256 MiB, 자동 재시도 0으로 한 번 실행합니다. Soft heap 설정은 총 RSS hard cap 또는 측정값이 아닙니다. 실행 결과는 1 MiB보다 작게 제한합니다. 원 namespace init은 main보다 먼저 일어나므로 바깥 controller의 시작·timeout 범위와 worker의 직접 호출 예약 범위를 구분해야 합니다.

Go 1.27.1, darwin/arm64, CGO=0, trimpath, buildvcs=false로 두 번 컴파일했습니다. 원 모듈은 UUID의 directive 부재 및 humanize Go 1.21을 유지합니다. Go package-selection metadata로 UUID 15개와 Ordinal slice 1개가 선택됨을 확인했으며, 이는 실행이 아닙니다. 순수 패키지 race 검사 3회, vet 3회가 통과했습니다. 최종 실행은 top-level test 7개와 subtest 13개입니다. 이전 top 6개 검사와 중간 경계 검사는 실패로 재분류하거나 숨기지 않았습니다. 고정 코드 적용 patch의 context 실패 1회는 source change 0으로 장부에 보존했습니다.

전체 모듈 또는 native main에 대한 go test는 수행하지 않았습니다. 그런 테스트는 원 패키지 초기화를 실행할 수 있습니다. 코드 작성자 자신의 synthetic 검증이며 인간 blind 또는 독립 원천 정확성 증명은 주장하지 않습니다. 별도 Want 작성자·source-only peer review의 범위는 참조 핀으로 구분합니다. 신규 labels, roles, weights, 모델, Features, Fit, 보호 final 열람, 공유 코드 변경 및 외부 게시는 모두 0입니다. training_ready/production_ready/qualification은 false입니다.

정확한 파일·바이트·SHA는 [HANDOFF.v1.json](HANDOFF.v1.json) 및 [PREPARATION-LEDGER.v1.json](PREPARATION-LEDGER.v1.json)에 있습니다. 실행에 필요한 private plan, 원시 테스트·빌드 metadata, helper와 바이너리는 공개 복사 대상에서 제외하고 SHA와 이유만 남깁니다. 초기 설계, 첫 draft·바이너리 및 변경 전 소스는 별도 history로 보존했습니다. 이전 실험의 결과·Want·컨트롤러는 변경하지 않았습니다.
