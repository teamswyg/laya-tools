# Source에 근거한 family frame

`internal/semanticframe`은 원본 Source에서 만든 변경 불가 목록(inventory)과,
댓글에 무엇을 어떻게 표현할지 적은 계획을 연결하는 작은 Go 인터페이스입니다.
목록 버전을 잘못 섞거나, 명시된 관계를 빠뜨리거나, 남겨야 할 모호함을 임의로
해소하는 실수를 잡는 데 사용합니다. 글을 생성하거나 질문·진행·완료를 판정하거나
학습 자료를 적격으로 만드는 기능은 없습니다.

현재 지원하는 것은 **새로 정의한 일반 선언 형식뿐**입니다. 기존 비공개 coder,
수정본, reviewer 기록은 자동 변환하지 않습니다. 실제 작성 전에는 명시적으로
검토한 변환기, 전체 Source QA, 독립적인 frame·댓글 의미 검토가 필요합니다.
테스트와 문서에는 실제 코호트, 모델, 비공개 기록을 넣지 않았습니다.

## frame에 적는 내용

V3 frame에는 Source, family, 형제 번호(1–3)와 함께 선택한 Source·observation·
타입이 있는 inventory·producer 선언·표준 Source 검토·별도의 버전 수용 근거의
파일 고정값을 적습니다. definitions와 inventory schema source는 내용까지 여는
검증 대상이 아니라 방법을 가리키는 고정된 참조입니다.

표현 계획은 `/propositions/0`, `/referent_support/0`처럼 **목록 항목 전체**를
선택하고, 각 항목에 하나의 역할을 적습니다.

- `utterance_content`: 댓글에서 표현하려는 내용.
- `governing_constraint`: 표현을 제한하는 근거·문맥.
- `unavailable_source_context`: 댓글 독자가 알 수 없는 Source 문맥.

이 역할은 질문·진행·완료 라벨이 아닙니다. Source에서 참인 사실이라고 댓글에서
자동으로 주장하게 되는 것도 아닙니다. 계획에는 제한된 Source 근거 구간과 비공개
설명을 적습니다. 남길 모호함은 선택한 항목과 `preserve_without_resolution`으로
표시합니다. 세 형제에 서로 다른 세 주장 상태를 강제로 배정하지 않습니다.

## 근거를 전달하는 방법

호출자는 기존 읽기 영수증 절차로 바이트를 보존하고 `PinnedBytes{File, Bytes}`로
전달합니다. 예상 상대 경로·SHA-256·바이트 수는 별도로 보존한 고정값에서 가져와야
합니다. 이 패키지는 파일을 열거나 저장하거나 네트워크에 접근하거나 읽기 영수증을
만들지 않습니다.

```go
frame, err := semanticframe.DecodeFrame(frameBytes)
if err != nil {
    return err
}
accepted, err := semanticframe.NormalizeAccepted("original", suppliedInputs)
if err != nil {
    return err
}
summary, err := semanticframe.JoinFrame(frame, accepted, registeredSourceIDs)
```

`suppliedInputs`에는 version binding, Source, observation, inventory, producer,
표준 검토, 별도 inventory 수용 선언의 정확한 바이트를 넣습니다. 수정본(amended)은
이전 버전 binding도 보존해 전달합니다.

지원하는 일반 형식은 다음과 같습니다.

- Frame: `riido-semantic-family-frame-draft-v3`.
- Version: `riido-accepted-inventory-version-binding-draft-v1`.
- Inventory: `riido-inventory-graph-declaration-v1`.
- Producer: `riido-inventory-original-producer-declaration-v1` 또는
  `riido-inventory-amended-producer-declaration-v1`.
- 별도 inventory 수용: `riido-inventory-acceptance-declaration-v1`.
- 표준 Source 검토: 공개 `sourcecohort.Review` 형식.

공개 Go 타입에 닫힌 필수 필드를 정의했습니다. 원본과 수정본 producer는 서로 다른
schema를 사용합니다. `NormalizeAccepted`는 전달된 모든 파일의 해시·크기를 확인한
다음, 선택한 observation·inventory·producer·검토와 각각의 별도 inventory 수용
선언이 같은 Source·버전을 정확히 가리키는지 확인합니다. producer ID는 선언된
reviewer ID와 달라야 합니다. accepted라는 파일명, producer의 플래그, 표준 Source
검토만으로 별도 inventory 수용 근거를 대체할 수 없습니다. 원본에는 이전 버전이
없어야 하고, 수정본에는 같은 Source의 다른 inventory를 가리키는 이전 버전 binding이
하나 이상 있어야 합니다. 이전 binding 자체의 바이트·메타데이터는 확인하지만 그 안의
과거 참조 파일을 재귀적으로 열거나 다시 적격으로 만들지 않습니다. 과거에 보류된
버전은 호출자의 이력에서 그대로 보존해야 합니다.
여기서 `amended`는 inventory 버전의 수정을 뜻합니다. 변경되지 않은 원본 inventory를
새 reviewer가 검토한 경우에는 `original`을 선택하고 새 검토·별도 수용 근거를 고정합니다.
이를 inventory 수정본으로 분류하지 않습니다.

바이트 검증이 증명하는 것은 **전달된 선언과 그 연결**입니다. producer·reviewer의
신원을 인증하거나, 선언된 pass가 옳고 독립적으로 수행됐음을 증명하지 않습니다.
Observation은 UTF-8 바이트를 검증하지만 schema checker를 실행하지 않습니다.
검토의 읽기 영수증과 방법 파일도 열지 않습니다. 파일 고정값, 검토 UTC, 읽기 결과 경로,
전체 UTF-8 근거 경계, 선언된 의존성의 메타데이터 일관성은 요구하지만 완전한 Source
검토 검증은 아닙니다. `sourcecohort`의 전체 코호트 출처 검사를 재현하지 않습니다.

## 구조 검사와 한계

포인터 번호가 inventory 해시에 묶이므로 원래 목록 순서를 유지합니다. Polarity,
referent kind, time relation 문자열은 열린 값으로 두고 새 의미 enum을 만들지
않습니다. ID 중복, referent 부모 순환, proposition·referent·opposition·time·
constraint 관계의 빠진 끝점을 확인합니다. 계획에서 선택한 항목의 명시된 끝점도
선택해야 하지만 표현 역할은 달라도 됩니다. 전체 Source 의존성은 외부에서 전달한
등록 Source ID에 있어야 합니다. 선언된 정방향 관계만 검사하므로 암묵적 제한 의미,
누락된 관계, 같은 split에 속해야 하는 의존성 등은 독립 검토와 후속 절차가 확인해야
합니다. 범위 검사로 의미 충실도를 증명할 수는 없습니다.

Source 전체 경계는 `[0, Source.bytes)`입니다. 근거는 끝을 포함하지 않는 UTF-8
바이트 구간이며 비어 있거나 코드 포인트를 잘라서는 안 됩니다. 항목 전체 포인터는
하위 필드·별칭·앞자리 0·중복·역할 충돌을 거부합니다. 오류에는 고정 코드만 나오며
입력 원문·경로·ID는 포함하지 않습니다. 호출자의 slice는 변경하지 않습니다.

Source는 최대 16 KiB, frame·일반 inventory·근거 파일 각각은 최대 128 KiB,
한 번에 정규화하는 전체 입력은 최대 8 MiB입니다. 그래프 종류별 항목, 계획,
모호함 목록, 근거 목록은 각각 최대 256개이고 등록 Source ID는 별도로 최대 400개입니다.
초과하면 전체 입력을 거부합니다. 닫힌 JSON은 필드 누락·중복·대소문자 별칭·알 수
없는 필드·null 배열·잘못된 UTF-8/surrogate·뒤에 붙은 추가 값을 거부합니다.

성공해도 상태는 `STRUCTURAL_DECLARATIONS_JOINED_QA_PENDING`이며
`meaning_proven:false`, `training_eligible:false`입니다. 표현 의도·진실·양언어
충실도·자연스러움·권리·신원·human Gold·provider 독립성·전체 Source QA는 증명하지
못합니다. 뒤에서 평가할 여덟 QA 축은 평가 항목이지 여덟 frame 필드나 예상 head
상태가 아닙니다. 독립 frame·댓글 검토와 텍스트만 보는 블라인드 reference는 별도
후속 단계입니다. `familycohort`는 구조 검사만 하는 기존 상태를 유지합니다. 이
패키지에는 새 승인 절차, 실행 모델, 학습 작업을 넣지 않았습니다.
