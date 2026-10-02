# 짧은 주장 감독을 Go 학습 배열로 옮기는 준비 패키지

[claimfit.Project](../../../internal/claimfit/projection.go)는 caller가 이미 검증하고 고정한 원문·truth·역할·mask를 기존 Go 학습 배열로 옮긴다. 자료를 읽는 loader, 의미 적격성 판정, 역할 배정 또는 학습기는 아니다. 이번 공개 port에서는 합성 입력으로 코드 연결을 검증했으며 실제 corpus 투영·학습·모델 추론이나 성능 측정은 하지 않았다. [English](README.en.md)

입력 `Parent`에는 전체 `shortclaim.Prepared`, 기존 `Truth`, 원래 순서의 `Acceptable []int`, `roleplan.Role`, `WholeGroup int`, 후보별 `LossEligible [8]bool`을 전달한다. 내부 driver에서 호출하는 예시다. `internal` 패키지이며 외부 프로젝트용 CLI/API는 이 port에 포함하지 않는다.

```go
projection, err := claimfit.Project([]claimfit.Parent{
    {
        Text: prepared,                 // 원본 전체, caller가 별도 binding을 확인
        Truth: claimfit.Known,
        Acceptable: []int{2, 0},         // 원래 허용 집합과 순서
        Role: roleplan.DevelopmentTrain,
        WholeGroup: frozenGroup,
        LossEligible: [8]bool{false, true, true},
    },
})
```

원래 입력·후보·순서·개수·전체 허용 집합은 `projection.Parents`에 남는다. known은 모든 acceptable 후보 label 1/나머지 0, no_answer는 모든 후보 0이다. mask된 known 행도 같은 라벨·features를 가진 행으로 남으며 weight만 0이다. unknown의 `NullableLabel.Known=false`는 **null 라벨**이다. `Positive=false`만 읽어서 negative로 해석하면 안 된다. unknown은 fit 행을 만들지 않는다. calibration은 원래 truth와 전체 후보를 보존하지만 fit/학습 AUC 진단에 들어가지 않는다.

development/validation의 sparse 배열은 기존 `pairlearn.Dataset` 형식이다. `Offsets`, `uint16 Indices`, `float64 Values`, `Labels`, `Groups`, 명시적인 `SampleWeights`, 원본 부모/후보 `RowRef`를 소유한다. feature는 동일한 `hintlearn.Features(original Request, original Candidate.Text)`다. ID/provenance/truth/역할/그룹/mask/review는 feature 인자가 아니다. class balance·부모별 재가중치·listwise loss·mask된 known 행 삭제를 추가하지 않는다. weight 0 행의 물리적 보존과 삭제가 같은 학습이라는 주장은 하지 않는다.

`DevelopmentAUC`/`ValidationAUC`에는 양수 weight 행만 복사하며 원래 fit 행·mask 행·unknown 후보·positive/negative 분모도 반환한다. 이 view에 기존 AUC를 적용하면 선택된 행의 **비가중 진단**이다. 전체 후보의 순위·fallback·Top·비용·기존 5% 효용 검증을 대신하지 않는다. 이번 합성 검사에서는 view와 분모를 확인했으며 AUC scoring이나 Fit은 실행하지 않았다. empty/all-weight-zero 투영 성공도 학습 성공이나 readiness 승인이 아니다.

같은 `WholeGroup`이 다른 role에 나타나면 unknown·all-masked·calibration까지 포함해 거부한다. 다만 caller의 숫자가 원래 component와 맞는지, batch 밖 멤버가 누락됐는지, 기존 9/3/3 coverage가 충족됐는지는 증명하지 않는다. `RoleEnumContract`는 현재 train=0/validation=1/calibration=2 bridge만 설명한다. source/text/truth/review SHA·전체 membership·감독/평가 적격성·roleplan 소스·실제 seed/coverage·Fit driver/config는 caller가 별도로 고정한다. 이 패키지는 새 역할·seed 또는 새 과학 gate를 만들지 않는다.

전체 prepared **출력 payload**를 64MiB로 제한한다. runtime parent/보존 문자열, slice headers, 두 fit 및 두 진단 view의 sparse arrays와 RowRefs를 함께 계산한다. 첫 feature pass에서 정확한 크기·유한성을 확인해 전체 결과 배열 할당 전에 overflow/상한을 검사하며 두 번째 pass가 동일 feature 경로로 값을 채운다. caller 입력·feature/정렬/row-plan scratch·allocator 여유를 포함한 peak RSS 64MiB 보장은 아니다. CPU/GPU/pprof 실측도 하지 않았다. 반환 배열은 caller 소유이며 수정 가능하므로 concurrent 입력 수정이나 수정 후 무결성은 별도 책임이다.

원형 및 이전 독립 QA의 공개 가능한 8개 보관본은 [archive guard](../../../internal/claimfit/archives_test.go)가 exact SHA/크기로 보호한다. 원본은 재작성하지 않았다. 원래 개인 경로가 있는 test module/README와 raw test 출력은 게시하지 않고 장부에 해시만 남겼다. 이전 [독립 결과](INDEPENDENT-FINDINGS-64.ko.md)와 [독립 장부](INDEPENDENT-AUDIT-LEDGER-64.json)는 원형의 검토이며 이 port의 새 독립 검토를 뜻하지 않는다. 이번 porter는 원형 저자이며 공개 변경은 다른 검토자가 별도로 읽어야 한다.

staging의 Go 1.27.1 race 최초 실행은 상위 14개/subtest 포함 29개 pass·실패 0, vet 1회도 성공했다. 원형 7개, 이전 독립 합성 5개, archive/enum 2개다. [PORT-LEDGER-64.v1.json](PORT-LEDGER-64.v1.json)에 실제 시도·원형/port SHA·보관 범위와 한계를 기록했다. 이는 합성 API 검증이며 실제 corpus 적격성·fit/readiness·성능을 승인하지 않는다. AI 보조 구현/검토가 사용됐고 일반 협업 비용은 측정하지 않았다. 별도 judge 모델/API·학습·가중치·protected-final 작업은 0이다.
