# 관찰80 실행 전 준비 자료

세 행동 목표를 24개 유한 입력에서 관찰하기 위해 준비한 **Go native CPU 관찰기**의 사전 기록입니다. UUID Parse 8개, UUID Scan 8개, 기존 Ordinal 8개를 대상으로 합니다. Laya·GPU·모델 추론이나 학습 자료 승격 결과가 아닙니다. 이 묶음은 준비 시점의 원 문서를 바이트 그대로 보존하며, Root의 이후 실제 실행 결과를 읽거나 보고하지 않습니다.

새 코드와 module 내용은 `.go.txt`·`.mod.txt`로 보관해 공개 저장소에서 실행 가능한 package가 되지 않게 했습니다. 원 저작권·주석·소스 본문은 변경하지 않았습니다. 과거 source 버전도 inert text로 보존했습니다. [원 Handoff](HANDOFF.v1.json)의 상대경로는 원 private workspace의 파일 이름을 뜻하며, 실제 공개 파일 이름은 [복사 장부](COPY-LEDGER.v1.json)의 `source_path`→`archive_path`에서 확인합니다. 원 Handoff를 공개 배치에 맞춘 원본인 척 다시 쓰지 않았습니다.

| 준비 범위 | 기록 |
|---|---|
| 논리 행동 목표 / 원천 가족 / 유한 fixture | 3 / 2 / 24 |
| 예상 callback | 직접 API 24 + 명시적 Error 5 = 29 |
| 보수적 callback 상한 | 최대 Error 24, tracked dispatch 48 |
| 원문 namespace 초기화 | 정적 site 4개, 실제 callback·return은 null |
| 준비 검증 | 순수 race 3회, vet 3회, compile-only build 2회 PASS |
| 준비에서 원 init·API·test·worker 실행 | 0 |
| 신규 labels·roles·weights·Features·Fit·모델·보호 final 열람 | 0 |

29회와 48회는 예상치와 상한입니다. 실제 관측 횟수가 아닙니다. 예상 오류 다섯 개가 달라져도 실제 동적 타입을 그대로 기록하고, 확인되지 않은 타입을 known 오류로 분류하지 않습니다. 호출 전 Got/Matches는 null이며, UUID 배열과 Scan receiver 전후, nil interface·typed nil bytes·nonnil empty bytes를 합치지 않습니다. Panic 값을 추가 Error/String 호출로 변환하지 않으며, 실패·불일치와 마지막 성공 checkpoint의 한계를 그대로 남깁니다. [준비 설명](HANDOFF.v1.ko.md)과 [장부](PREPARATION-LEDGER.v1.json)에 자세한 범위가 있습니다.

[UUID 원문과 전체 BSD-3-Clause](../upstream/google-uuid/LICENSE), [Ordinal slice와 전체 MIT](../upstream/humanize-ordinal-slice/LICENSE)는 기존 [원천 archive](../README.ko.md)에 있습니다. UUID의 고정 revision 원문 15개 및 원 go.mod, humanize의 원 ordinals.go 한 파일 및 원 go.mod를 선택했습니다. 원천 20개와 [source-first 자료](../preparation/SOURCE-FIRST-80.v1.json)는 SHA와 바이트 일치를 확인하고 재복사하지 않았습니다. 이 observer용 코드에는 저장소 본체의 Apache-2.0 조건이 적용되며, upstream BSD/MIT를 재허가하지 않습니다. 전체 humanize 또는 `number.go`의 다른 차용 원전까지 권리가 확인됐다는 뜻도 아닙니다. [기존 NOTICE](../NOTICE.md)를 함께 보존합니다.

Want는 별도 작성자의 봉인값을 [정확히 복사](WANTS.v1.json)했습니다. Want 작성자의 [관찰기 mechanics 검토](peer-static-review/RECEIPT.v1.json)는 비맹검 source reading이며, 새로운 독립 oracle 또는 인간 blind 검증이 아닙니다. 별도 source-only Want peer와 기존 원천 archive의 정확한 핀은 [외부 참조](EXTERNAL-REFERENCES.v1.json)에 있습니다. 반복 fixture를 독립 부모로 세거나 training/production qualification을 true로 바꾸지 않았습니다.

호스트 절대경로가 있는 draft/frozen plan, 바이너리, private helper와 원시 test/build metadata는 공개하지 않습니다. SHA와 제외 이유만 장부에 남깁니다. 실행 조건 CPU 1·Go soft heap 256 MiB·바깥 timeout 60초·재시도 0은 설정이며 실측 RSS hard limit이 아닙니다. 원 package 초기화는 main보다 먼저 일어납니다. 이 묶음은 Root의 실행 승인이나 일반화·성능·비용 절감 증거를 대신하지 않습니다.

[ARCHIVE-MANIFEST](ARCHIVE-MANIFEST.v1.json)와 [SHA256SUMS](SHA256SUMS)는 준비 파일의 바이트를 묶습니다. 공개 복사는 아직 완료하지 않았으며, 이번 작업은 private stage·metadata 검증만 수행했습니다. 실제 결과·worker·원 tests·모델은 재실행하지 않았습니다.
