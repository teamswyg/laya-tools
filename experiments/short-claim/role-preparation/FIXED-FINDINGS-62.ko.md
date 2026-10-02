# coverage ID 자원 경계 수정 확인

최종 source SHA는 `df8d53d0600e80c331c4912d90ff199aef753704f34c87e085fa738fe920ba0b`, 공개 tests SHA는 `c858a8138bde70629a51effaeb8259878419afffe7d2f5e0880f1b5bb79e9b92`다. 이전 `47af…`의 발견과 재현·장부는 바꾸지 않았다. 공개 작성자가 각 coverage 참조의 nonempty·UTF-8·512바이트 검사를 복사·정렬 전에 넣은 차이만 확인했다.

독립 합성 영향 검사 1회에서 named test/subtest 10개가 통과했다. 빈 ID·잘못된 UTF-8·512바이트 초과 ID는 ordering 전에 `coverage_component_id_invalid`로 거절되며, 정확히 512바이트인 ASCII와 UTF-8 ID는 허용된다. 작은 3:1:1 literal 비율, 입력·출력 소유권과 원래 requirement offset, unknown을 labeled floor에 넣지 않는 조건, 이전 coverage가 통과해도 뒤 실패에서 역할을 반환하지 않는 조건도 유지됐다. vet와 formatting이 통과했다.

이번 영향 검사의 합성 호출은 Assign 9회, AllocateCounts 5회다. 변화가 없는 membership·64MiB 관련 함수는 소스 대조와 이전 합성 증거를 유지하고 중복 실행하지 않았다. 시작 전 임시 작업 폴더가 아직 없어 도구가 한 번 실행을 거절한 이력도 별도 장부에 남겼으며, 그때 테스트나 API는 시작되지 않았다. 기존 합성 검사의 실패 0과 혼동하지 않는다.

잔여 mechanics blocker는 없다. 다만 이것은 준비 코드 검토의 통과다. opaque membership SHA·known flag·pointer/path/source 관계와 고정 seed·coverage의 실제 자격은 호출자 freeze에 의존한다. 정확한 membership 인코딩은 아직 제안이다. 이 검토자는 private prototype 작성자이고 이번 공개 port·수정 작성자는 아니지만 블라인드도 아니다. 실제 원본 graph·seed·roles·fit·모델·원천 API·새 라벨·리소스 실측·공유 코드 변경·게시 모두 0이며, training_ready는 false다. 별도 새 사람 승인 흐름을 추가하지 않는다.
