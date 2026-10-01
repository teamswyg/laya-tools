# 다음 확보56b: 숫자 증식보다 다른 실패 양상

[English](NEXT-56b.en.md) · [실제56a 결과](RESULTS-56a.ko.md) · [골든셋 규모](../../docs/golden-set-scale.ko.md)

**후보 검토이며 새 계약·독립 그룹·라벨·학습 적격은 아직0개 확정이다.** 56a의48요청·11그룹·학습0 실패와 기존 하한15를 유지한다. 다음 목표는15를 간신히 채우는 것이 아니라 소스와 관측 동작의 다양성을 확보하여 의미 주장 모델이 새로운 실패 양상을 구별할 수 있는지 시험하는 것이다.

## 먼저 필요한 typed truth

현재 `[]int → []int` 유한 검사는 부정·정수 경계·순서를 검사하는 데 충분했지만 상태·오류·소유권을 모두 표현하지 못한다. 새 표에 `(값, 오류 정체성, 이전/이후 상태, 호출 이벤트, 입력 변경)`을 구분한다. 후보 함수가 expected를 만들지 못하게 literal 정답과 실제 구현을 분리한다. 네트워크·실시간·회사 자료 없이 표준 Go 값과 가상 clock/event로 검사한다.

각 후보의 뜻을512바이트·정규화32단어 안에서 완전하게 전달할 수 있는지도 별도로 확인한다. 전달할 수 없다면 임의 요약·교사 생성·라벨 추측 대신 unknown/제외로 처리한다. 입력 예산을 바꾸려면 새로운 schema·계획·동일 조건 비교가 필요하다. 검사의 오류를 틀린 후보로 바꾸지 않는다.

## 자체 작성으로 준비할 여섯 후보

| 동작 | 독립 관측과 기존 원형과 다른 점 |
|---|---|
| 파싱 결과의 원자적 반영 | 정상은 반영, 마지막 잘못된 바이트는 오류·제로 결과·기존 대상 보존. 부분 commit을 구분 |
| 래핑된 오류의 정체성 | 직접/다중 래핑 원인 발견, 문자열만 같은 다른 오류는 거절. 문자열 유사도와 실제 원인을 구분 |
| 소유한 snapshot | 원본과 반환값을 각각 바꾼 뒤 반대편은 불변. 처음 출력이 같아도 뒤의 aliasing을 구분 |
| 취소와 자원 수명의 순서 | 사전 취소는 획득0, 획득 뒤 실패는 cleanup1, 성공/실패 이벤트 trace 고정 |
| 인용부호·escape 구분자 | quote 내부 구분자, 빈 항목, trailing escape의 literal token/error. 단순 split과 문맥 인식 구분 |
| 순환과 공유 DAG | ancestor back-edge만 거절하고 공유 child는 허용; 오류 경로와 원본 불변 검사 |

새 후보가 서로 같은 staged-commit·error-cause·ownership·lifecycle·lexical-state·graph core를 공유하는지도 감사한다. 기존 prefix-balance와 새 lexical-state가 코드 구조를 공유하면 함께 묶는다. 여섯 후보를 여섯 독립 그룹으로 미리 세지 않는다.

## 공개 원천 연결은 후속 단계

[역사적120개 조사53](../../docs/public-go-acquisition-53.ko.md)의 후보를 다시 읽어 다음을 제안한다. 지금 외부 파일이나 라이선스를 새로 복제하지 않았으며 실행 적격 계약도 아니다.

| 고정 원천 후보 | 실제 채택 전 확인 |
|---|---|
| [pflag snapshot](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/string_to_string.go) | 양방향 변경 격리와 기존 공유 getter 유지. 기존 std-only subset의 새 oracle/mutant·BSD 고지 검증. 자체 snapshot과 그룹 연결 가능 |
| [mapstructure hook 오류 원인](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/decode_hooks.go) | 실패 hook의 원인을 모두 보존하고 첫 성공 뒤 추가 호출0. MIT·closure·Go1.18 대비 errors.Join의 Go1.20+ 범위 확인. 자체 error core와 연결 |
| [INI quoted comment](https://github.com/go-ini/ini/blob/e2db55b0e088fa4ee0c128aa4ade263cdc1d7f08/parser.go) | quote 안 #/;는 값, 밖은 주석, 닫히지 않은 입력 오류. 이미 지원되는 옵션은 새 gap으로 세지 않음. Apache2와 복사 코드/외부 tests closure 재검증 |
| [retryablehttp wrapped TLS 오류](https://github.com/hashicorp/go-retryablehttp/blob/fd004584a46724fae09e2f21d7c382e15c893f42/cert_error_go120.go) | 직접/다중 wrapped 인증서 오류를 같은 영구 실패로 분류. MPL2 파일 고지·외부 의존성 때문에 후순위. 실제 TLS 대신 공개 오류 객체 사용 |

mapstructure typed-nil hook은 새 구현 gap으로 세지 않는다. 기존 composition이 typed nil을 보존하는 부분을 먼저 확인해야 하며 기존 동작을 검사하는 probe로만 사용할 수 있다. 원천 조사 기록의 status·원본 license bytes·고정 revision·module identity를 바꾸지 않는다.

## 실행 순서와 gate

1. 자체 원자적 반영·오류 정체성·snapshot의 typed truth 계약과 오답 대조군을 준비한다.
2. source/code/소유권·조건·문장 충실성을 독립 검토하고 모든 parent/negative/copy/core 관계를 전이적으로 합친다. 같은 작성자 한계도 공개한다.
3. 공개 pflag는 별도 권리·compile closure·요구·oracle·mutant pins를 먼저 봉인한다. hook/INI/TLS는 각각의 추가 문제를 해결한 뒤 준비한다.
4. 실제 그룹·그룹 크기·가족/언어/형태 편중을 보고하고 충분하지 않으면 계속 수집한다. 하한15 통과를 통계적 독립성이나 실사용 적격으로 해석하지 않는다.
5. 충분한 자료의 역할·모든 baseline·oracle상한·시간/RSS/저장 예산을 새 계획으로 봉인하고 CI 준비 후 기존 사용자 승인 범위에서 실행한다. 관측 후 primary나 gate를 바꾸지 않는다.

기존 학습 제안의 최대8 fits·최대16파생 artifact, FP32 주 후보와 삼진 sibling/INT8·PTQ child 관계를 유지한다. 새 학습 실행 계획은 아직 없다. 최종 평가는 도메인별 서로 다른 최소2,400요청을 별도로 확보하며 현재 protected final과 CoSQA reserve를 재사용하지 않는다. 이 문서가 HF 모델 게시나 production 활성화를 승인하는 품질 증거는 아니다.
