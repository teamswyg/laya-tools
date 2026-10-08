# 전체 inventory를 보존하면서 graph 만들기

`internal/inventoryprojection`은 전달된 원래 inventory에서 구조 그래프를 복사하는
작은 Go 준비 도구입니다. 원래 바이트·고정값·전체 타입 표현을 함께 보존합니다.
파일·네트워크 접근, 과거 기록 자동 변환, 수용 기록·검토 verdict·학습 권한 생성은
하지 않습니다. 테스트는 직접 만든 공개 소프트웨어 예시만 사용합니다.

```go
result, err := inventoryprojection.Project(inventoryprojection.Inputs{
    Config: configBytes, Full: fullBytes, Native: nativeContractBytes,
    Mapping: mappingContractBytes, Adapter: declaredAdapterSourceBytes,
    Source: nil,
})
```

각 입력은 별도로 예상한 상대 경로·SHA-256·정확한 크기가 있는 `PinnedBytes{File,
Bytes}`입니다. 닫힌 `Config`에는 `selected_full_inventory`, `adapter_source`,
`adapter_version`, `mapping_contract`, `native_contract`, `graph_output_path`를
넣고 실제 입력과 고정값을 맞춥니다. 지원 버전은 `riido-inventoryprojection-v1`입니다.
Adapter 바이트는 선언된 정체성의 고정값이며 실행 중인 바이너리가 그 코드로 만들어졌다는
증명은 아닙니다.

별도로 고정한 `NativeContract`에는 `native_inventory_schema`와 전체
`official_field_keys` 집합을 적습니다. 과거 schema·필드명은 추측하지 않습니다.
닫힌 `MappingContract`는 `graph_schema`, `copied_fields`, `full_only_fields`를
담으며 목록 순서까지 공개 `SupportedMapping()`과 정확히 같아야 합니다. 단순 schema
문자열이나 숨겨진 필드 설정으로 이 계약들을 대신할 수 없습니다.

`Result.Binding`에는 전체 inventory·graph·config·native·mapping·adapter의 고정값과
adapter 버전만 담습니다. Graph는 자신의 해시와 출력 경로를 갖습니다. Source·
observation·producer·review·수용·이전 버전·선택 이력의 연결은 생성하지 않습니다.
이들은 후속의 명시적인 과거 기록 adapter가 다뤄야 합니다. 현재 semanticframe V3에서
graph를 원래 inventory 대신 넣어 과거 수용 근거를 이어붙일 수는 없습니다.

별도 [FULL 검토 연결 인터페이스](accepted-full-graph-bridge.ko.md)는 새로 정의한
일반 FULL 검토 선언을 이 추출 결과와 작성 계획에 연결할 수 있습니다. 실제 과거
기록을 변환하는 어댑터는 아직 미완료입니다.

Graph 항목·설명·열린 polarity/kind/time 값·null 부모·근거 구간·배열 위치는 그대로
복사합니다. Graph에 없는 negative ID, 전체 부정 설명, 공식 필드 support, 미해결
질문, 출처·완전성 선언은 `Result.Full`과 정확한 `RetainedFullBytes`에 남깁니다.
공식 필드의 원래 객체 순서와 각 support의 원래 바이트도 보존합니다. `Full`을 Go에서
직렬화한 형태는 원래 객체 표현이 아니므로 원본을 참조할 때는 보존한 바이트를 씁니다.
반환값은 호출자 소유이므로 고정한 바이트를 수정하지 말고 보존해야 합니다. Support만
바꾸면 graph는 같아도 full/config 해시는 바뀔 수 있으며 정상입니다.

구조 검사는 graph와 전체 support의 명시적 끝점, negative ID 포함 여부, ID 중복,
부모 순환, 자기 의존성, 미개봉 참조를 포함한 같은 경로의 고정값 충돌을 확인합니다.
정확히 같은 고정값 반복은 허용합니다. Negative ID 포함 여부로 부정 의미를 증명할
수는 없습니다. 의존성의 등록·split·전체 코호트 연결은 후속 검사입니다. 완전성·검사
여부·status·value·미해결 필드는 의미 판단 없이 보존합니다. 누락된 관계·암묵적인
제한 의미는 전체 inventory의 충실도 검토가 확인해야 합니다.

`Source:nil`이면 선언된 Source 크기에 대한 숫자 범위만 확인하고
`UTF8SpansVerified`는 false입니다. 정확한 Source 바이트를 별도로 고정해 넣으면
전체 support를 포함한 모든 근거의 UTF-8 유효성·코드 포인트 경계를 확인합니다.
빈 근거 배열은 보존하며 완전성을 증명하지 않습니다. 구간 검사만으로 근거가 주장을
뒷받침하는지 알 수는 없습니다.

각 입력은 최대 128 KiB, Source는 16 KiB, 전체 입력은 8 MiB입니다. 배열·객체는
최대 256개, 깊이는 32, 키·ID는 128바이트, 디코딩 문자열은 65536바이트입니다.
ID와 공식 support 키는 ASCII 영문자·숫자·밑줄·하이픈을 사용합니다.
타입 목록을 할당하기 전에 항목·깊이 한도를 확인하고, graph를 직렬화하기 전에
HTML·제어 문자 이스케이프를 포함한 JSON 크기를 계산합니다. 출력 graph 한도는
128 KiB이며 넘으면 전체를 거부합니다. 필드 누락·중복·대소문자 별칭·알 수 없는 필드,
null 배열, 잘못된 UTF-8·surrogate 값·추가 JSON은 거부합니다. 오류에는 고정 코드만
나오며 입력 원문·경로는 없습니다. 배열·정렬된 조회 slice를 사용하며 전역 lock과
모델 실행 의존성은 없습니다.

`MeaningProven`, `TrainingEligible`은 항상 false입니다. 전체 Source QA, 권리,
인증된 독립성, 의미 수용, human Gold, frame 적격성, 모델 유용성을 증명하지 않습니다.
실제 projection·작성 전 전체 Source QA가 필요하며 새 승인 절차는 추가하지 않습니다.
