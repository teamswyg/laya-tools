# Source archive80 attribution / 원문 보존 고지

This is a maintainer-authored provenance note, not an upstream NOTICE file or a replacement license. 아래 설명은 관리자가 작성한 출처 고지이며, upstream NOTICE 또는 라이선스 전문을 대체하지 않는다. Sources and existing preparation documents are copied byte-for-byte; only inert archive filenames change. 원문과 기존 준비 문서는 바이트 그대로 복사하고, 원문 파일의 archive 이름만 실행되지 않는 확장자로 바꿨다.

## Google UUID — BSD-3-Clause

Origin: [google/uuid](https://github.com/google/uuid/tree/2d3c2a9cc518326daf99a383f07c4d3c44317e4d), revision `2d3c2a9cc518326daf99a383f07c4d3c44317e4d`.

The [complete LICENSE](upstream/google-uuid/LICENSE) is preserved unchanged: 1,480 bytes, SHA-256 `0a8d61ed3cbfd5312326e8126c31ce9c627a283adc99131b56896d29ada04b2d`. It contains `Copyright (c) 2009,2014 Google Inc. All rights reserved.` and all conditions and disclaimers.

The 15 selected runtime sources each retain the three-line copyright/BSD header. 파일마다 원문에 있던 저작권과 BSD 고지를 유지한다:

| File year / 연도 | Original files / 원래 파일 |
|---|---|
| 2016 | dce.go, doc.go, hash.go, marshal.go, node.go, sql.go, time.go, util.go, version1.go, version4.go |
| 2017 | node_net.go |
| 2018 | uuid.go |
| 2021 | null.go |
| 2023 | version6.go, version7.go |

Each header says `Copyright YEAR Google Inc.  All rights reserved.` (including the original two spaces after `Inc.`) and refers to the BSD-style license in LICENSE. `node_net.go` also retains its original `// +build !js` tag. `version4.go` retains its UUID Wikipedia uniqueness attribution; other hash/time derivation comments also remain unchanged. Inspection of these 15 files found no explicit additional copied-code license notice. This limited inspection is not an exhaustive ancestry, dependency, or authorship clearance.

각 header는 Google 저작권과 LICENSE의 BSD 조건을 가리킨다. `node_net.go`의 원래 build tag와 `version4.go`의 UUID Wikipedia 관련 주석도 그대로 있다. 이번 15개 원문에서는 추가 차용 코드의 별도 라이선스 고지를 발견하지 않았지만, 모든 원전·의존성·저자 독립성을 확인했다는 뜻은 아니다.

The retained BSD text requires source redistributions to retain its copyright, conditions, and disclaimer; binary redistributions to reproduce them in accompanying documentation/materials; and prohibits using Google/contributor names for endorsement without specific prior permission. The complete disclaimer remains with this archive. 보존 BSD 전문은 소스 재배포 시 저작권·조건·면책을 유지하고, binary 재배포 시 동반 문서 등에 이를 재현하며, 사전 허락 없이 Google 또는 기여자 이름을 제품 추천·홍보에 사용하지 못하도록 한다. 전문의 면책 조항도 함께 보존했다.

Scope: 15 retained Darwin non-JS runtime files plus original go.mod. JavaScript-only source and upstream tests are excluded. This source selection has not been compiled here and is not a hermetic toolchain/dependency proof. 범위는 Darwin non-JS 원문 후보이며 전체 cross-platform repository를 복제하거나 여기서 원문을 compile한 것이 아니다.

## Dustin go-humanize Ordinal slice — MIT

Origin: [dustin/go-humanize](https://github.com/dustin/go-humanize/tree/a1b4e66b9a6d890e9e15e7091cf16c8032367d6e), revision `a1b4e66b9a6d890e9e15e7091cf16c8032367d6e`.

The [complete LICENSE](upstream/humanize-ordinal-slice/LICENSE) is preserved unchanged: 1,136 bytes, SHA-256 `a973b4498c13eb74baa2a8e5c351426a6826f2fcdd909916dbe53ee2e755fd71`. Its copyright is `Copyright (c) 2005-2008  Dustin Sallings <dustin@spy.net>`. This public upstream copyright contact is retained as part of the original license.

Included: the complete, unchanged original `ordinals.go` file (371 bytes, SHA-256 `aac3d5ceefd8044baae1f3deb76613470c7eb94fc26af1ea51cec93f8eab075f`) and original go.mod. The selected file has no separate per-file copyright header; the root MIT notice accompanies it. Its only function is the original `Ordinal(int)`; no Ordinal64, upstream tests, other function bodies, or whole-package init profile are included. This is an explicit reduced source composition, not an import of the whole upstream library.

원래 ordinals.go 전체와 go.mod만 포함한다. 파일에 별도 저작권 header는 없으며, root MIT 전문을 함께 보존한다. 함수 추가·수정이나 전체 library/import/init 관찰을 의미하지 않는다. 작은 비음수 입력을 제안한 것은 음수 동작을 수정하거나 비음수 guard를 추가한 것이 아니다.

The retained MIT text requires its copyright and permission notice in copies or substantial portions; its complete warranty/liability disclaimer is preserved. MIT 전문의 저작권·허락 고지를 복사본 또는 상당 부분에 유지하는 조건과 전체 면책 조항을 보존했다.

Other humanize bodies were not acquired or bundled. The previously identified `number.go` gorhill gist/WTFPL lineage remains a known-missing concern only for any future whole-package composition. Its origin and license bodies are not included or newly acquired here; no license clearance for that broader composition is claimed. Other missing-file pointers remain unchanged in the historical proposal.

Humanize 전체의 `number.go` 차용 원전/WTFPL 조건은 선택 Ordinal slice 밖이다. 더 큰 구성을 선택할 때 확인할 known-missing 항목으로 남기고, 이번 준비에서는 원문 취득·복제를 하지 않았다. 이를 MIT 하나로 전체 package의 모든 차용분까지 허가됐다고 해석하지 않는다.

## Preparation and host repository / 준비 기록과 저장소

Six frozen preparation documents are copied unchanged under `preparation/`. UUID was initially called MIT in a parent instruction; the preserved preparation ledger records its correction to BSD-3-Clause before the original freeze. The repository's own LICENSE states Apache-2.0; these upstream files retain the BSD/MIT terms above, and this note grants no new license for them.

동결 문서 6개와 이전 실패 장부는 수정하지 않았다. 새 원문 pin·Want·API 결과·정답·role·weight를 만들지 않았고, 원문 compile/import/init/test/native·모델·Fit도 실행하지 않았다. 이 archive의 초기화 4곳은 정적 위치 수이며 실제 반환 수는 계측되지 않았다. AI-assisted, nonblind development preparation is not human-only authorship, independent held-out evidence, or training readiness. Original source copyright does not establish the independence of the later authored requests or captions.
