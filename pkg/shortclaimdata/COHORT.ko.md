# 모르는 값을 보존하는 코호트 입력

새 `LoadCohortRow`는 후보 5개 또는 8개인 다음 개발 자료를 읽는 **선택형 Go API**입니다. 기존 `LoadDevelopmentRow`와 37개 자료의 형식·라벨·비교 결과는 그대로입니다. 아직 새 실제 요청이나 학습 결과를 추가한 단계는 아닙니다.

왜 필요한가요? “후보가 틀렸다”와 “아직 확인하지 못했다”를 같은 `false`로 학습시키면 모델이 근거 없이 후보를 배제할 수 있습니다. 요청 해석이 모호할 때도 이미 관측한 사실을 지워서는 안 됩니다. 이 API는 문장, 선언된 사실, 학습 사용 여부, 출처 기록을 따로 보존합니다.

| 입력 상태 | 감사 기록 | 이번 학습 변환 |
|---|---|---|
| T / `label:true` | 알려진 긍정 | 나머지 조건 충족 시 사용 |
| F / `label:false` | 알려진 부정 | 나머지 조건 충족 시 사용 |
| U / `label:null` | 미확인 | 부모 요청 전체 제외 |
| 모호한 해석 | T/F/null을 그대로 유지 | 부모 요청 전체 제외 |
| 알려진 라벨, 가중치 0 | 사실은 그대로 유지 | 선택된 부모에서 해당 행의 가중치만 0 |
| 모든 사용 마스크 0 또는 calibration | 모든 사실 유지 | 부모 요청 전체 제외 |

`Class()`의 `known_none`, `known_one`, `known_many`, `known_all`은 **주어진 라벨의 구조적 요약**입니다. 예컨대 F/F/U는 `unknown_containing`이며 “정답 없음”이 아닙니다. 전부 T인 사례도 보존합니다. opaque `evidence_sha256`는 증거 파일의 식별 지문일 뿐입니다. 이 reader는 그 파일의 개별 조건을 읽거나 사실·라이선스·전체 계보를 검증하지 않습니다. 따라서 `known_none`도 실제 의미 검증 완료를 뜻하지 않습니다.

## 사용 흐름

별도 검토한 원문·유한 관측·마스크 정책·전체 연결 가족의 역할 자료에서 `CohortBinding`을 먼저 고정합니다. 행 자체의 SHA-256과 예상 metadata도 포함합니다. JSON을 읽은 뒤 그 JSON에서 기대 지문을 만들어 붙이는 방식은 외부 binding 검증이 아닙니다.

```go
// frozen은 호출 전에 별도 고정한 shortclaimdata.CohortBinding입니다.
example, err := shortclaimdata.LoadCohortRow(rowReader, frozen)
if err != nil { return err }
input := example.Input()         // 요청·후보 문장만 정규화
audit := example.Supervision() // 후보 순서의 고정 배열; U는 LabelKnown=false
metadata := example.Metadata() // 문장 특징에 넣지 않는 역할·출처 기록
```

사용자가 명시적으로 호출해야 합니다. Codex 등록, 라우터 실행, 파일 경로 열기, 추론, 특징 추출, 학습은 수행하지 않습니다. 접근자는 값 복사를 반환하며 내부에 공유 가변 map·slice·lock이 없습니다. JSONL은 크기를 제한한 한 행씩 전달합니다.

## 새 wire 계약

객체 하나가 최대 16 KiB입니다. 요청과 각 후보 문장은 기존 512바이트·정규화 32단어 제한을 재사용합니다. 후보는 정확히 5개 또는 8개이며 순서를 보존합니다. 모든 필드는 필수이고 중복·대소문자 별칭·알 수 없는 필드·잘못된 Unicode·여분 JSON을 거절합니다. 원본 행 지문은 공백까지 포함한 바이트 전체를 묶습니다. 오류에는 입력이나 로컬 경로를 싣지 않습니다.

- 최상위 12개 필드: `schema`, `stable_id`, `request`, `candidates`, `role`, `whole_group`, `source_family`, `source_revision`, `text_revision`, `feature_policy`, `interpretation`, `bindings`.
- `schema`: `riido-shortclaim-cohort-row-v1`. `feature_policy`: `request_and_candidates_text_only`.
- `interpretation`: `unambiguous` 또는 `ambiguous`.
- 각 후보: `metadata_id`, `text`, `state`(T/F/U), `label`(true/false/null), `loss_weight`(정수 리터럴 0/1), `evaluation_eligible`(boolean).
- `bindings`: `source_sha256`, `evidence_sha256`, `mask_policy_sha256`, `roles_groups_sha256`. 각각 64자리 소문자 hex이며 0 지문은 거절합니다.
- 역할: `development_train`, `development_validation`, `development_calibration`. 전체 그룹은 1–65,535입니다. 출처 리비전은 40자리 소문자 hex입니다.

Train은 `loss_weight`로 경사 참여를 표시하고 evaluation 마스크는 false입니다. Validation은 loss가 0이며 `evaluation_eligible`로 검증 손실 참여를 표시합니다. Calibration은 둘 다 꺼집니다. U와 모호한 부모도 둘 다 꺼져야 하지만 알려진 T/F 라벨은 지우지 않습니다. 이 형식은 이전 Golden Reader·mask 소스 제안과 별도 버전이며, 이전 제안의 핀·파일을 바꾸지 않습니다.

## 기존 변환기와 연결

유지보수용 `internal/claimfit.ProjectCohort`는 최대 60개 입력을 받고, **선택 전에 전체 입력**의 중복 ID·그룹/가족 간 역할 충돌·공통 역할 계획과 마스크 정책 지문을 확인합니다. 감사 전용 요청 뒤에 역할 누수가 숨어 있어도 거절합니다. 알려진 관계만 검사하므로 실제 공유 계보는 별도 연결 그래프 검토가 필요합니다.

완전 known·모호하지 않음·calibration 아님·사용 행 존재인 부모만 기존 `claimfit.Project`로 전달합니다. T/F/U 중 U를 삭제해 남은 F만 전달하거나 U를 false로 바꾸지 않습니다. Train의 loss 마스크와 Validation의 evaluation 마스크를 기존 변환기의 역할별 행 가중치에 연결합니다. 선택된 부모에는 후보 전체와 가중치 0의 정확한 라벨도 유지합니다. 양성·음성 비교 쌍이 없는 전부 T/전부 F는 기존 BCE 행으로 표현되며, 비교 쌍 손실의 정의 여부는 기존 pairlearn 계약을 따릅니다.

전체 원본은 `Audit`에 남고, `SelectedAuditIndices`가 변환 후 부모 번호를 원 감사 번호로 대응합니다. `SelectedCount==0`은 정상적인 감사 전용 결과이며 학습 가능 상태가 아닙니다. Train/Validation의 양수 가중치 존재, 원천 중복·역할 분리·권리·정답 적격화는 별도로 확인해야 합니다. 이 API는 `Fit`이나 빈 자료의 `NLL`을 호출하지 않습니다.

다음 단계는 별도 20개 코호트에서 후보 수와 0/1/여러 정답·U·모호함을 균형 있게 수집하고, 원본 관측과 연결 가족 검토 뒤 확대하는 것입니다. 새 독립 요청 수·모델 품질·Codex 비용 절감·GPU 실행·2,400개 보호 평가 성과는 아직 증명하지 않았습니다.
