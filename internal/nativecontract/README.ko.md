# 원본·수정본 검토 기록 연결

[English](README.md)

`nativecontract.Join`은 학습 자료의 검토 기록 네 개가 같은 자료와 검토자를 가리키는지 확인합니다. 자료를 그대로 재검토한 경우와 자료를 수정한 뒤 검토한 경우를 구분합니다. 검토자가 남긴 판정을 보존하며, 이 코드가 의미 판정을 새로 만들지는 않습니다.

호출자는 바이트 수와 SHA-256이 고정된 표준 검토 기록, 전체 목록 수락 기록, 선택 버전, 현재 상태를 전달합니다. `Expected`에는 원본·선택본·작성자·검토자·읽기·검사 기록의 예상 참조를 넣고, `RegisteredSourceIDs`에는 등록된 자료 ID를 넣습니다. 수정본이면 공식 항목 일곱 개의 키도 전달합니다. 원본 재검토 형식에는 이 키 필드가 없으므로 새 필드를 요구하거나 만들어 넣지 않습니다.

```go
result, err := nativecontract.Join(nativecontract.Inputs{
    Kind:                nativecontract.Amended,
    Review:              reviewBytes,
    Acceptance:          acceptanceBytes,
    Version:             versionBytes,
    Disposition:         dispositionBytes,
    Expected:            declaredAnchors,
    RegisteredSourceIDs: sourceIDs,
})
```

각 바이트 입력은 `PinnedBytes`입니다. `File`에는 상대 경로·바이트 수·SHA-256을, `Bytes`에는 실제 문서 바이트를 전달합니다. `Witnesses`를 추가하면 읽기 시작·읽기 결과·검사 연결·검사 결과·기존 보류 검토 기록의 공개 형식도 확인합니다. 참조된 경로를 자동으로 열지 않으므로, 필요한 바이트는 호출자가 별도로 준비해야 합니다.

중복·미등록·누락 키, 잘못된 타입이나 스키마, 변경된 바이트, 서로 충돌하는 참조, 검토자 불일치, 보류를 통과로 바꾼 조합은 오류로 반환합니다. 성공하면 원래 문서 바이트를 복사해 보존하며 기존 보류 참조도 유지합니다. 입력당 128 KiB, 합계 8 MiB, JSON 깊이 32, 컬렉션당 256개, 등록 ID 400개를 상한으로 둡니다. 현재 입력 문서는 최대 아홉 개이므로 실제 합계 상한은 개별 제한에 의해 더 작아집니다. 참조 비교는 정렬된 슬라이스를 사용하고 파일·네트워크·모델 호출이나 잠금은 사용하지 않습니다.

이 모듈은 두 가지 명시된 검토 형식만 지원합니다. 과거 모든 검토 형식의 변환이나 실제 자료의 의미·권리·검토자 신원 증명을 수행하지 않습니다. 성공해도 `MeaningProven`과 `TrainingEligible`은 항상 `false`입니다. 전체 자료 검증과 학습 허가는 별도 단계입니다. 테스트는 직접 만든 합성 메타데이터를 사용합니다.
