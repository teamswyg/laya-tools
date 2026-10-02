# 다음 bounded Go loader의 구체 연결

이 문서는 구현 handoff입니다. loader·Project·Features·순위·fit을 실행한 결과가 아닙니다. 기존 입력 경계는 부모의 별도 결과 `53b492d0629f794785e592c90a96f3d86223dbf81891393d763a89e4a9d92d55`에서 전체 72/216 supported로 확인됐습니다. 새 과학적 조건을 추가하지 않습니다.

## 첫 범위: 기존 72개 그대로

loader의 중간 DTO는 배열·슬라이스로 충분합니다. `BoundParent`에는 `OriginalIndex int`, `ParentID string`, `RawRequest string`, `Candidates []BoundCandidate`, `Truth claimfit.TruthState`, `Acceptable []int`, `WholeGroup int`, `Role roleplan.Role`을 둡니다. 후보 DTO에는 원래 `ID/Text`, 원래 `NullableLabel *int`, `LossEligible bool`을 둡니다. SourceID·CodeSHA·group·role은 side metadata로 보존하고 scorer text에 넣지 않습니다.

고정된 두 probes의 raw 입력 순서를 legacy 48개 다음 typed 24개로 연결합니다. 각 원문과 저장 정답/mask/role을 이번 감사 결과의 `/parents/{i}`와 대조합니다. 후보 배열 길이·순서·ID, nullable label, acceptable의 순서도 유지합니다. shortclaim 입력은 `Schema: shortclaim.Schema`, raw 요청, 후보 ID/Text, 고정 provenance 문자열로만 구성합니다.

```go
prepared, err := shortclaim.Validate(shortclaim.Input{
    Schema: shortclaim.Schema,
    Request: original.Request,
    Candidates: original.Candidates, // 원래 ID/Text와 원래 순서
    Provenance: "frozen-original-56-56b",
})
if err != nil { return zero, fixedInputError }
p := claimfit.Parent{
    Text: prepared,
    Truth: boundTruth,             // known=1, no_answer=2, unknown=3
    Acceptable: ownedAcceptable,   // 원래 허용집합, unknown/no_answer는 빈 slice
    Role: boundRole,               // train=0, validation=1, calibration=2
    WholeGroup: originalGroupID,
    LossEligible: boundMaskArray,  // unused slots=false; 기존 mask만 복사
}
```

Prepared의 내부 배열을 임의 생성하지 않습니다. `claimfit.Project`는 전체 입력을 검증하고 feature를 생성하는 최초 단계이므로, 그 호출 전에 새로운 source/input/binary/plan 핀과 1회 실행 장부를 동결해야 합니다. 역할 배정은 다시 실행하지 않고 기존 66 결과의 role을 읽습니다. whole-group의 역할 불일치·duplicate index·원문 SHA 불일치·unknown 정답 치환은 고정 오류로 중단합니다.

`Project`의 train/validation row에는 모든 원래 known 후보가 남고 mask가 false이면 원래 label과 함께 weight 0이 됩니다. unknown과 calibration은 audit/runtime 부모로 보존하며 fit row를 만들지 않습니다. 이번 감사의 eligible 후보 수를 실제 Project 출력으로 선채우지 말고, 원래 분모와 출력 row references를 함께 대조해야 합니다. 모든 no_answer와 음성 후보를 보존합니다.

Primary A 효용은 기존 51 known/no_answer 요청의 원래 acceptable과 모든 후보입니다. mask는 check cost를 공짜로 만들지 않습니다. B complete-caption parent 진단은 따로 보고하고 A/58을 대체하지 않습니다. BM25와 oracle을 계산할 때 둘 모두 같은 A 분모·단위 검사 비용을 사용하고 unknown 21의 개수를 따로 보존합니다. 최초 Project와 같은 범위의 효용 실행은 별도 동결 후 부모가 진행합니다.

## 이후 범위: 68 두 family의 고정 결합

68 계약 `818eed54f6b14b1f512025804cf7782d4a59d94ed30eeeaa46c56a270afbd0eb` 및 독립 QA `ecc9de25fea14872257fc50af2ddb312db644d0e0b5e9ed07ffdc9f3fb119083`을 같은 bytes에서 해시 확인 후 decode합니다. 원래 네 부모 순서는 prefix-reject, prefix-allow, nested-allow, nested-reject이며 후보 수는 2/2/3/3입니다. 제안 label은 계약의 Want를 새로 실행해서 만들지 않고 QA의 고정 proposed metadata와 acceptable을 연결합니다. 실제 가중치 null 및 proposed-only 상태를 정식 결합 동결 전에는 유지합니다.

결합 DTO는 기존 65의 모든 `canonical_members`와 `relationships`를 온전히 보존한 `OriginalWholeMembership json.RawMessage`와 두 `ProposedWholeFamily` record를 별도 소유합니다. 새 record에는 family ID, 원래 68 parent indices, 전체 원천·helper·module·license closure 핀을 보존합니다. 후보 하나만 골라 closure를 줄이거나, 새 source family의 일부만 기존 role에 덧붙이지 않습니다.

실제 combined 역할에는 **새 전체 membership 매핑과 동일 seed ASCII1729의 새 계획**이 필요합니다. 기존 66 역할을 새 코퍼스에 자동 복사하지 않습니다. 새 family 간 또는 기존 corpus와의 relation이 기록으로 확정되지 않으면 누락을 추측해 채우지 말고 정확한 gap을 반환합니다. opaque 해시만으로 외부 edge 누락을 증명하지 않습니다. 기존 하한·unknown·mask를 보존하며 추가 표본·역할·학습 승인을 이미 얻었다고 주장하지 않습니다.

공유 레포 코드 수정, 새 Assign/Project/Features/Baselines/API/fit/model 실행은 이 handoff에서 모두 0입니다.
