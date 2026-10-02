# 원천 관찰80: 공개 원문과 사전 입력 제안

UUID의 파싱과 receiver 변경, 작은 비음수의 서수 표현을 나중에 구별해서 관찰하기 위한 **원문 보존 자료**다. [동결된 설명](preparation/SOURCE-FIRST-80.ko.md)과 [입력 초안](preparation/INPUT-DRAFTS.v1.json)은 그대로 복사했다. 이 폴더 준비에서는 Want·adapter·API 호출·학습을 만들거나 실행하지 않았다. 별도 관찰 작업의 이후 상태를 대신 보고하는 자료도 아니다.

| 원천 | 보존한 범위 | 전체 라이선스 |
|---|---|---|
| [google/uuid](https://github.com/google/uuid/tree/2d3c2a9cc518326daf99a383f07c4d3c44317e4d) | 고정 revision의 Darwin non-JS runtime 원문 15개와 원래 go.mod | [BSD-3-Clause 전문](upstream/google-uuid/LICENSE) |
| [dustin/go-humanize](https://github.com/dustin/go-humanize/tree/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e) | 원래 ordinals.go 하나와 원래 go.mod. 전체 upstream package를 가져온 구성이 아님 | [MIT 전문](upstream/humanize-ordinal-slice/LICENSE) |

원문 Go 파일은 `.go.txt`, module 파일은 `.mod.txt` 확장자로만 보존한다. 본문·저작권·주석·build tag·module 내용은 변경하지 않았다. 실행 가능한 library port나 compiler closure 증명이 아니며, JavaScript용 UUID 원문·upstream tests·추가 Ordinal64·관찰기 helper·binary는 포함하지 않는다. Humanize의 다른 파일과 `number.go`의 별도 WTFPL 원전은 이 Ordinal slice에 포함하지 않았고 새로 취득하지 않았다.

UUID 15개 원문의 파일별 Google 저작권과 BSD 고지, 두 원천의 전체 LICENSE를 함께 배포하도록 준비했다. 세부 출처·파일별 고지·각 SHA-256은 [NOTICE](NOTICE.md)와 [manifest](ARCHIVE-MANIFEST.v1.json)에 있다. 저장소 본체의 Apache-2.0 표기가 이 upstream 원문을 재허가한다는 뜻은 아니다. 재배포 조건은 해당 전문을 함께 확인해야 한다.

제안은 UUID Parse·Scan과 Ordinal의 **3개 행동 목표, 2개 원천 가족, 24개 입력 슬롯**이다. 같은 UUID package의 두 목표와 반복 fixture를 독립 source 가족으로 부풀리지 않는다. 모든 기존 truth·role·weight의 null 상태와 qualification·training·production·protected-final false를 그대로 보존한다.

UUID `hash.go`의 namespace 초기화에는 정적으로 읽은 MustParse→Parse 호출 위치 4곳이 있다. 원문 init 안을 계측하지 않았으므로 실제 init callback·반환 수 4를 관측했다는 뜻이 아니다. 이후 실행은 child 시작 전 외부에서 이 시작 경계를 예약하고, 실제 CPU·RSS는 초기화 등을 포함한 child 전체 수명으로 따로 측정해야 한다. 이 archive 준비의 원문 import/init·API·tests·compile·native·모델·Fit은 모두 0이다.

[SHA256SUMS](SHA256SUMS)로 원문과 사전 문서의 복사 바이트를 확인할 수 있다. [복사 장부](COPY-LEDGER.v1.json)는 이번 metadata 준비만 기록한다. 원래 준비 중 발생한 helper 컴파일 실패도 [이전 장부](preparation/LEDGER.v1.json)에 그대로 남아 있다. 사전 Want를 별도 저자가 작성하고 나중에 실제 관찰과 후보 만족 조건을 검토하는 단계가 필요하며, 이 자료 자체는 정답·학습 적격성·일반화 승인이나 2,400개 미노출 final 검증이 아니다.
