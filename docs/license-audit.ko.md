# 라이선스 검토와 배포 고지 보완

한국어 · [English](license-audit.en.md)

검토일: 2026-09-30. 범위는 현재 고정한 모델·이식 코드·CLI에 연결되는 Go 의존성·다운로드하는 네이티브 런타임입니다. **확인한 Apache-2.0/MIT/BSD-3-Clause 조건 아래에서 이 프로젝트를 수정·상업적으로 사용·재배포하는 방향은 가능합니다.** 다만 고지 보존 의무가 있으며, 모델 학습 데이터의 권리나 모든 특허까지 문제가 없다고 보증하는 분석은 아닙니다.

## 확인한 원문과 적용 범위

| 구성 요소 | 확인 근거 | 배포 시 처리 |
|---|---|---|
| Laya base | [고정 모델 카드](https://huggingface.co/convaiinnovations/laya/blob/55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851/README.md)의 `license: apache-2.0` | 해당 revision 루트에 LICENSE/NOTICE 파일은 없었음. 선언에 따른 Apache-2.0 원문과 출처·변환 고지를 모델 묶음에 포함 |
| laya-code | [고정 NOTICE](https://huggingface.co/tindang/laya-code/blob/25f97e5a2ec5f8cf7218a4f67504367d8832e1fe/NOTICE), 같은 revision의 LICENSE 및 모델 카드 | Apache-2.0. Tin Dang/Convai Innovations/ModernBERT 출처를 원문 그대로 보존 |
| ModernBERT-large | [모델 카드](https://huggingface.co/answerdotai/ModernBERT-large)의 Apache-2.0 선언, laya-code NOTICE | 인코더·토크나이저 출처를 유지. 선언만으로 학습 데이터 권리를 독립 검증한 것은 아님 |
| Laya SDK/exporter | [고정 SDK LICENSE](https://github.com/NandhaKishorM/laya/blob/6d942c92081fbc139e736bbd9ac0023223c29b7f/LICENSE) | Apache-2.0. 복사한 유지보수 exporter 및 시퀀스 구현 출처는 NOTICE에 기록 |
| system-one-router | [고정 LICENSE](https://github.com/mmornati/system-one-router/blob/a437d00bca33a4c10038b37a5efe04dc0e3d40bb/LICENSE) | Apache-2.0. 선택 구조의 출처·변경 내용을 소스와 문서에 표시 |
| pi-pignon | [고정 LICENSE](https://github.com/siiick/pi-pignon/blob/4d97d1a35134b813bac6a8e345a1bf9b0bbf0bdb/LICENSE) | MIT. Nicolas Chaintron 저작권 및 허가문을 보존. Go로 옮겨도 원본 고지를 없애지 않음 |
| regexp2 v1.11.5 | [LICENSE](https://github.com/dlclark/regexp2/blob/v1.11.5/LICENSE) | MIT, Doug Clark. 바이너리 배포에 원문 포함 |
| onnxruntime_go v1.36.0 | [LICENSE](https://github.com/yalue/onnxruntime_go/blob/v1.36.0/LICENSE) | MIT, Nathan Otterness. 원문 포함. 포함된 Microsoft C 헤더에도 별도 MIT 고지 보존 |
| golang.org/x/text v0.25.0 | [LICENSE](https://github.com/golang/text/blob/v0.25.0/LICENSE), PATENTS | BSD-3-Clause 및 추가 특허 허가문 포함 |
| Go 런타임 | 실제 Go 1.27.1 배포판의 LICENSE/PATENTS | BSD-3-Clause 및 특허 허가문 포함 |
| ONNX Runtime 1.30.0 | [공식 LICENSE](https://github.com/microsoft/onnxruntime/blob/v1.30.0/LICENSE) 및 공식 배포물의 ThirdPartyNotices.txt | MIT. setup은 이미 두 고지 파일을 해시 검증해 런타임과 함께 설치. CLI의 C 헤더 출처를 위해 MIT 원문도 바이너리 묶음에 포함 |

이식하지 않은 laya-codex/keel 코드는 해당 기능의 아이디어 비교 자료입니다. fast-laya-compaction은 라이선스를 확인하지 못해 이식하지 않았습니다. 웹사이트가 오픈소스로 소개한다는 사실만으로 복사 허가가 생기지는 않습니다.

## 우리가 지켜야 하는 조건

[Apache-2.0 원문](https://www.apache.org/licenses/LICENSE-2.0)의 §4는 재배포 시 라이선스 사본, 변경 파일의 변경 고지, 관련 원저작권·귀속 표시, 원본에 존재하는 NOTICE의 귀속 내용을 보존하도록 규정합니다. 자체 변경에 Apache-2.0을 적용할 수 있지만 원본 고지를 대체해서는 안 됩니다. §6은 일반적인 상표 사용권을 부여하지 않습니다. `riidolaya` 명령어와 출처 설명을 사용하고 공식 제휴·보증으로 표현하지 않습니다.

[MIT 원문](https://opensource.org/license/mit)은 수정·판매·재배포 등을 허용하면서 저작권과 허가문을 사본 또는 상당 부분에 포함하도록 요구합니다. 따라서 TypeScript를 Go로 번역했다고 pi-pignon의 고지를 삭제할 수 없습니다. 코드 파일은 소스 헤더에서 MIT 출처를 표시하고 배포물에 원문을 싣습니다.

BSD-3-Clause는 소스·바이너리 배포에서 고지와 면책을 보존하고 저자 이름으로 무단 보증·홍보하지 않는 조건입니다. 출처 URL만 제공하는 것으로 바이너리의 고지 의무를 대신하지 않도록 원문을 포함합니다.

이 조건들은 자체 작성 코드 전체를 동일한 강한 카피레프트로 전환하도록 요구하는 조건은 아닙니다. 현재 연결되는 구성 요소에서 GPL/AGPL 계열을 발견하지 않았다는 뜻이며, 앞으로 추가하는 모든 의존성에도 자동으로 같은 결론을 적용할 수 있다는 뜻은 아닙니다.

## 발견한 누락과 이번 수정

1. **models-v1의 code 묶음에 원본 NOTICE가 누락돼 있었습니다.** 새 `models-v2`에는 원본 NOTICE를 그대로 포함하고 base에도 출처를 넣었습니다. LICENSE와 모델 카드만으로 충분하다고 보지 않고 실제 배포물을 수정했습니다.
2. 모델 변환 사실을 분명히 하기 위해 ONNX 자체의 `doc_string`에 변환·수정 고지를 넣고 `MODIFICATIONS.md`와 `PROVENANCE.json`을 동봉했습니다. 원래 계산 그래프의 직렬화 해시가 같음을 확인했습니다. 기존 아카이브를 덮어쓰지 않고 새 URL/해시를 사용합니다.
3. **초기 CLI 배포는 일부 Go 의존성의 링크만 안내하고 원문 고지를 동봉하지 않았습니다.** 이번 바이너리에는 `licenses/` 전체를 포함합니다. 예전 릴리스에도 부속 고지 묶음을 제공하고 새 버전 사용을 안내합니다. 이것이 과거 배포 상태 자체를 소급해 없애는 것은 아닙니다.
4. `internal/compliance` 테스트는 실제 CLI의 Go 의존성 목록·버전과 검토 목록의 일치, 보존된 고지의 해시, Go 배포판 고지, 모델의 필수 고지 파일을 검사합니다. 새 의존성을 임의로 허용하거나 라이선스 호환성을 자동 추론하지 않습니다.

고지는 [NOTICE](../NOTICE), [라이선스 디렉터리](../licenses), [검토 목록](../licenses/manifest.json)에 있습니다. Go 런타임은 Go 배포물에서, Go 모듈 고지는 내려받아 검증된 해당 버전 모듈에서 복사했습니다. 모델 아카이브의 파일 목록과 해시는 [manifest](../internal/assets/manifest.json)에 고정합니다.

## 남아 있는 한계

**공개 라이선스 선언과 학습 데이터의 권리 검증은 별개입니다.** laya-code 모델 카드도 학습 데이터와 모델 가중치의 법적 지위에 관한 불확실성을 밝힙니다. base의 전체 학습 데이터 권리 사슬이나 권리자의 권한을 이 검토만으로 확인하지 못했습니다. 분류 모델이라 원문 생성 기능이 없다는 기술적 특징만으로 권리 문제가 모두 사라진다고 결론 내리지 않습니다.

자체 미세조정을 한다면 뤼이도의 비공개 코드·사용자 요청·고객 데이터에 대한 사용 권한, 학습 목적, 외부 공개 범위를 따로 정해야 합니다. 이번 preview는 직접 만든 가상 예제만 공개했고 실제 비공개 저장소나 고객 자료를 학습·재배포하지 않습니다.

현재 검토는 지정 버전의 기술적 배포 점검입니다. 라이선스 원문 밖의 계약, 상표 분쟁, 학습 데이터·특허에 대한 법률 보증은 포함하지 않습니다. 권리 보증이나 고객에 대한 배상 약정이 필요한 상용 계약 단계에서는 해당 범위를 별도로 법률 검토해야 합니다.
