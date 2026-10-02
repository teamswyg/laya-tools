# Source80 저장 결과 무결성 검사

Root가 한 번 실행한 관찰기의 저장 기록은 **동결된 기대값 24개와 일치**합니다. 원문 입력·순서·full Want·Got·부분 UUID·receiver·null·호출 수·파일 핀의 정합성을 확인했습니다. 검토자는 사전 Want 작성자 checkpoint_cli_45이며, 비맹검 작성자 측 사후 무결성 검사입니다. 독립 정답 승인이나 새 원천 의미 검증은 아닙니다. 원 API·관찰기·컨트롤러·Go tests/build·모델/Features/Fit를 재실행하지 않았습니다.

| 유한 목표 | primary 정상 반환 | nonnil 오류 | 기대값과 일치 |
|---|---:|---:|---:|
| UUID Parse | 8 | 3 | 8 |
| UUID.Scan | 8 | 2 | 8 |
| 원 int Ordinal | 8 | 오류 반환 채널 없음 | 8 |

전체 primary reserved/returned는24/24이고 explicit Error reserved/returned는5/5입니다. Error는 upstream URNPrefixError1·stdlib errorString4로 나뉘며 unclassified0, primary/Error panic0, differences0입니다. 성공 match 보고를 믿는 데서 그치지 않고 모든 Got 채널을 변경 없는 full Want와 대조했습니다. raw32 오류의 실패 슬롯 fe와 standard36의00, Scan 오류/빈 입력의 receiver 유지, typed nil과 nonnil empty bytes 구분, Ordinal 오류 채널 null도 그대로 보존됩니다. 입력24개는 목표3개·원천 계열2개이고, 독립 부모24개나 학습 라벨24개라는 뜻은 아닙니다.

worker final·partial·root stdout은 각67,635바이트, SHA256 `ce793a8633c75520e226bc894c2126caecac85cda7c809e56fd41792c9a877a6`이며 세 파일은 byte-exact입니다. frozen plan SHA `27186964abe2387f6dc8e5f33d3188bfb5dd5c4b56bf0cf7cdf4914a9b16f81a`는 기존 v2 draft의 frozen=false→true만 byte 수준으로 바꾼 것입니다. Want는 기존 `c93ac2b7…`, input은 `a6110ab8…`, binary는 `55f662cb…`이며 결과 source_pins33개가 frozen plan과 같고 실제 source 파일33개 SHA도 모두 일치했습니다. reservation은 direct 호출 수0·records null을 보존합니다.

Root 저장 장부는 Start1·joined Wait·exit0·retry0·timeout/overflow/start/wait failure0을 보고합니다. generic76 controller 장부 schema는 재사용한 원문 그대로이고 source80 worker/plan/binary SHA로 범위를 연결했습니다. 실행 전 root metadata helper는 compile 시도2 중 실패1이 있었지만 native 실행 전 실패이며 원 기록을 보존합니다. Root 실제 native 실행은 한 번이고 이 QA의 실행은 저장 자료를 읽는 jq checker1회 성공/실패0입니다.

원시 OS 로그에서 추출한 **전체 child** 값은 real1.21초, user0.02초, system0.07초, maximum RSS18,104,320바이트(17.265625MiB), peak footprint15,680,016바이트입니다. controller wall은1.217474375초입니다. 초기화·setup·24개 입력 관찰·checkpoint/fsync·저장을 모두 포함하므로 함수별 비용이나 모델 추론 비용으로 환산하지 않습니다. Go heap/GPU 개별 사용량은 측정하지 않았습니다. Go soft heap256MiB는 OS RSS hard cap이 아닙니다.

4개 namespace MustParse→Parse는 정적 startup site이고 실제 initializer callback/return 수는 null, 개별 동적 계측은 false입니다. 내부 Scan 재귀·Parse·fmt도 동적 계수하지 않습니다. labels/roles/weights/actual parents/model/Features/Fit0, training_ready·production_ready·qualification·protected_final=false를 유지합니다. 모든 입력의 안전성, RFC 검증, 원천 다양성, 2400개 데이터 성공이나 새 모델 품질을 승인하지 않습니다.

[MECHANICS.v1.json](MECHANICS.v1.json)과 [RECEIPT.v1.json](RECEIPT.v1.json)은 안전한 집계와 해시만 담습니다. private host 경로가 있는 frozen plan·실행 전 인자, 원시 OS 로그/stdout·native binary·검사 helper는 공개 복사하지 않습니다. 제외 원문의 SHA와 사유는 영수증에 남겼습니다. 원본 Want/소스/실행 결과는 수정하지 않았습니다.
