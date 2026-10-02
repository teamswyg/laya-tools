# 공개 원문 문구 자산63

[English](README.en.md) · [최신 상세 설명](WORDING-63.v2.ko.md) · [독립 검증](QA-63.ko.md)

기존 한 작성 흐름에서 만든 설명문에만 학습을 맞추지 않도록, 이전에 고정한 semver·doublestar 공개 원본에서 **완전한 문장·문법 항목25개**를 확보했습니다. 문장을 잘라32단어에 맞추거나 번역·재작성하지 않았습니다. 이 자산은 다음 공개 개발 입력의 재료이며, 현재 모델 입력·후보·학습 라벨을 추가한 결과는 아닙니다.

[최종 catalog](quote-catalog-63.json)는 원문 bytes, SHA, 정확한 파일·revision·줄·byte span을 보존합니다.25개 합계1818bytes, 최대26 normalized words이고 모두512bytes/32단어 이내입니다. 기존 retained 원본15파일99035bytes 및 **MIT 전문2개**도 확인했습니다. 원문과 저작권 고지를 함께 유지하며, 프로젝트의 새 Go 코드는 기존 Apache-2.0 범위입니다. 이것만으로 향후 모델 가중치의 학습·배포 자격을 승인하지 않습니다.

원문이 설명하는 의미와 기존57 관찰이 검증한 범위를 구분했습니다. StrictNewVersion 원문은 모든 nil·error 세부 필드와 반환 숫자를 명시하지 않습니다. `/**/` 문법 항목도 `src/**`가 `src`에 match하는 경계나 전체 pattern인 `**`의 모든 의미를 대신하지 않습니다. 기존 strict16행/slash9행의287개 필드 참조를 연결하면서 이런 관찰 빈틈을 남겼습니다. 새 동작·expected·정답 집합을 만들어 빈틈을 메우지 않았습니다.

MIT notice의 저작권 이름은 개별 문장의 저자나 human-only 작성 여부를 확인한 결과가 아닙니다. Pre-existing upstream wording과 프로젝트 AI-assisted input/Want 작성 원천은 다르지만, 아직 실제 후보 관계·감독 적격성·그룹 관계에 사용하지 않았습니다. 따라서 기존72요청의 `synthetic_single_pipeline`을 해소하거나 `authoring_diversity_cleared=true`로 바꾸지 않습니다.

첫 catalog는 row/topic 연결 오류를 자체 검토에서 발견해 보류했습니다. [첫 보관본](quote-catalog-63.attempt1.json)과 실패·수정을 [v2 장부](PREPARATION-LEDGER-63.v2.json)에 남겼으며, **학습 재료에는 최종 catalog만 사용해야 합니다.** 같은 제작자의 별도 검사 프로그램을 독립 검토로 부른 설명 오류도 v2에서 정정했습니다. v1 문서·장부는 당시 기록으로 보존하고, 별도 reader의 [독립 영수증](INDEPENDENT-QA-63.json)과 [장부](INDEPENDENT-QA-LEDGER-63.json)를 추가했습니다. 독립 검증은1회 성공·실패0이며 의미 전반·원천 저자·학습 자격 승인은 아닙니다.

다음 실제 입력에서는 원문이 표현하는 좁은 범위, 전체 후보 보존, 기존 기대값·unknown, 원천·파생 관계를 먼저 동결합니다. 이 단계의 새 label·candidate·roles·fit·weights·보호 최종 읽기는0이고 `training_ready=false`입니다. 보관 문서의 private/게시0 표현은 제작·검토 당시 범위입니다. 공개 통합과 CI는 별도 PR 기록으로 확인합니다.
