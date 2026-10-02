# 64 private pure projection 독립 검토

현재 공개 호출 형태 `Project`에서 다음 port를 막는 구체적인 코드 결함은 찾지 못했다. 이 결론은 새로 쓴 합성 입력과 제한된 코드/API 읽기에 대한 것이며 실제 corpus 적격성, 역할 계획 또는 학습 readiness 승인이 아니다. 원본 projection SHA `572e33a815b9a123c0e521d92567d6f254526250ab4d82e202487ddbe1d517e5`와 테스트 SHA `c312a1ba1aab690ff613510a572d4604b74cedc4e243b10c7a666d8ed6ef82c8`는 검토 전후 및 독립 private copy에서 모두 일치했다.

Known의 전체 acceptable set과 원래 순서를 보존하며, known의 나머지 후보 및 no_answer의 기존 negative는 label 0이다. mask는 원래 truth를 바꾸거나 행을 삭제하지 않고 같은 feature 행에 명시적 weight 0을 준다. Unknown은 nullable 감사 상태로 보존하고 fit 행을 만들지 않는다. Calibration 역시 전체 원문·truth·mask는 보존하며 loss와 학습 AUC 진단에 들어가지 않는다. Eight-candidate/all-acceptable/역순 acceptable, masked positive와 no_answer 및 unknown/calibration 조합도 독립 합성 검사에서 확인했다.

동일 `WholeGroup`이 다른 role에 있으면 unknown·all-masked·calibration을 포함해 거부한다. 세 role의 모든 서로 다른 조합을 검사했고 feature 호출 전에 실패했다. Positive-weight AUC view는 loss dataset과 다른 backing arrays/RowRefs를 소유하며 가중치 양수 행만 담는다. 이 view의 AUC는 기존 비가중 진단이며 full-candidate 순위·fallback·Top 또는 5% 효용 평가를 대신하지 않는다. Unknown-only/all-zero 그룹은 positive-weight group으로 세지 않는다.

64MiB는 출력 payload 한계로 구현되어 있다. 독립 테스트는 Projection 및 RuntimeParent 구조·slice headers, 보존 문자열 길이, 실제 sparse columns와 RowRef **capacity**를 별도로 합산해 `PayloadBytes`와 일치함을 확인했다. 합성 결과의 exact limit은 통과하고 1 byte 미만은 zero Projection으로 실패했다. 축소된 구현 한계로 부모 배열이 넘는 경우 malformed parent 검증·feature 계산보다 먼저 거부함도 확인했다. Empty projection의 네 offsets sentinel을 포함한 회계는 일치한다. 기존 preflight는 부모 수를 출력 RuntimeParent 크기로 제한하고 후보 수는 부모별 최대8로 제한하므로 feature 합계/row count는 오류 전에 임의로 int 범위를 넘지 않는다. 이 검토는 실제 64MiB 근처 workload나 RSS/CPU/GPU 측정이 아니다. Caller 입력, 정렬/row-plan scratch, feature 함수 임시 배열과 allocator 여유는 별도이며 전체 peak memory 64MiB를 주장할 수 없다.

두 feature pass는 길이·중복 index·dimension·유한성을 확인한다. 현재 `hintlearn.Dimension=8192`이고 index는 0..8191로 제한되어 uint16 변환이 안전하다. 최대 valid index는 보존되었으며 dimension,65535,65536,negative index와 두 번째 pass의 NaN/±Inf가 거부되었다. Metadata, truth, role, opaque group 및 mask는 feature 인자가 아니며 원래 request/candidate text만 전달된다. 원본 fixture의 ID/provenance는 감사 옆자료로만 남는다. 출력 후 caller의 acceptable/Prepared/group을 바꾸어도 반환 snapshot은 바뀌지 않았고 loss와 진단 배열도 서로 alias하지 않았다.

다음 한계는 그대로 남는다.

- `ValidatePrepared`는 값의 내부 일관성만 확인한다. 원래 source/text/truth SHA, 원래 binding, review 적격성 flag의 정당성을 증명하지 않는다.
- `WholeGroup int`는 caller의 opaque 식별자다. 같은 숫자의 cross-role 충돌은 잡지만, 원래 같은 component를 서로 다른 숫자로 속이거나 batch 밖 구성원을 빠뜨리는 경우는 이 함수가 증명할 수 없다. 전체 membership 및 원래 9/3/3 coverage는 동결된 caller 경로의 책임이다.
- `RoleEnumContract`는 현재 0/1/2 bridge만 설명한다. 최종 public integration은 실제 roleplan 소스/동결 입력을 따로 묶어야 한다.
- `Project`는 입력을 concurrently 수정하는 caller를 지원하지 않는다. 입력 문자열은 일반 Go의 불변 값으로 공유하고 반환 slice는 caller 소유의 mutable 결과다. Fit 전 동결/무결성은 별도 driver의 책임이다.
- Private injectable `project`는 두 pass의 **동일 길이인 다른 유한 값**까지 비교하지 않는다. 현재 exported `Project`는 callback을 받지 않고 읽은 순수 `hintlearn.Features`만 사용하므로 현재 호출 계약의 blocker는 아니다. 향후 추출기를 외부 주입형으로 바꾸면 이 가정을 재검토해야 한다.
- Empty/all-zero-weight projection 성공은 fitting 성공이나 readiness를 뜻하지 않는다. 별도 driver는 기존 Fit의 데이터·총 weight 검증을 그대로 사용해야 한다. Unknown 또는 loss0 그룹으로 old floors를 채우면 안 된다.

독립 private copy에 새 상위 테스트5개를 추가했고 원래7개와 함께 uncached race1회가 통과했다(상위12개, subtest 포함27개, package1.419s). Vet1회 및 formatting 검사도 통과했다. 실행 실패0, 원본/공유 파일 수정0이다. 출력 JSON은 도구 출력 한계로 일부 잘렸지만 최종 package pass·exit0은 별도 poll로 확인했다. 전체 원출력 파일이 있다고 주장하지 않는다. 실제 corpus/sourceAPI/roles/labels/seed/fit/AUC scoring/model/paid process/publication은 모두0이며 새 과학 gate를 만들지 않았다.
