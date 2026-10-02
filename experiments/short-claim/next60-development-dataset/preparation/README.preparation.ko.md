---
license: apache-2.0
language:
  - en
  - ko
pretty_name: Riidolaya Next60 finite development subset
tags:
  - riidolaya
  - development-only
  - finite-code-behavior
configs:
  - config_name: development
    data_files:
      - split: train
        path: data/train.jsonl
---

# Riidolaya Next60 유한 범위 개발 학습 자료

[English](README.md) · 예정 저장소: `JooYoon/riidolaya-shortclaim-next60-development`.

현재는 **미게시 준비 자료**이며, 독립 평가 benchmark나 학습된 모델이 아닙니다. `data/train.jsonl`은 정확히 **요청2행·후보 라벨5개**(긍정2/부정3)입니다. 이미 Next60 초안에 존재한 두 ID와 소스 영향을 받은 유한 개발 자격 판정을 옮겼습니다. 기존 catalog 대비 **새 globally unique 요청0·새 source family0**이며 독립 held-out coverage로 쓰면 안 됩니다.

| 안정된 요청 ID | 연결된 전체 그룹 | 후보 | 고정 입력 | 기존 후보 관측 |
|---|---:|---:|---:|---:|
| next60-pflag-native-ipnet | 77: cobra/pflag | 3 | IPv4 사례5개 | 15 |
| next60-mapstructure-native-or | 78: mapstructure | 2 | 정수 hook 사례4개 | 8 |

입력 사례9개와 기존 관측23회는 근거 범위이며 **dataset9행/23행이 아닙니다**. 후보별 sample weight1을 유지하고 반복 벡터로 가중치를 늘리지 않습니다. 모든 행은 development_train이고 그룹77/78은 이 프로젝트의 연결 소스 식별자이며 보편적인 분류가 아닙니다. 소스/fork/alias/template 관계는 같은 전체 그룹에 유지합니다. 다른 역할과 충돌하면 이후 Fit을 차단하며 역할 이동/seed 탐색으로 해결하지 않습니다.

Root의 자격 판정은 AI 보조·nonblind입니다. 모델 scoring/학습 전에 소스와 저장된 유한 결과를 보고 설명을 만들었습니다. 긍정은 고정 사례 전부 충족·unknown0, 부정은 완전한 한정 설명에서 확인한 불일치입니다. 모든 입력의 정확성·임의 hook/객체 불변성·독립 native 재현·새 family 일반화·모델 utility를 뜻하지 않습니다. 실제 작성/라벨 판정은 [qualification](evidence/QUALIFIED-TRAIN-SUBSET.v1.json)에 보존했고 과거 조건부 제안은 과거 제안 상태로 남겼습니다.

## 사용하는 방법

모델에 넣는 것은 request와 candidates[].text뿐입니다. metadata_id·안정 ID·라벨·가중치·역할·그룹·revision·범위·hash는 학습 정답/관리 정보이며 **모델 feature에서 제외**합니다. API 이름은 text 밖의 metadata ID에만 있습니다. 후보 순서를 유지하세요. Boolean 라벨은 요청 전체가 아닌 각 후보에 있고, loss 대상 후보5개에 unit BCE weight를 적용합니다. pair-loss 정책은 이 묶음에서 정하지 않았습니다.

확장된 HF 행은 **shortclaim.LoadValidated에 바로 호환되지 않습니다**. 추가 필드를 처리할 명시적 adapter가 필요합니다. 원래 두 입력은 [INPUTS.v3.json](evidence/INPUTS.v3.json)입니다. 비활성 [Go 변환 소스](source/convert.go.txt)는 bytes/hash로 묶인 입력과 판정을 읽으며 모델 호출/라벨 발명은 없습니다. 나중에 Root가 새 디렉터리에서 실행해 초안2행과 바이트 단위로 비교해야 합니다. 여기서는 컴파일/실행0입니다. JSON 파싱·변환 재현·학습 reader의 text projection·viewer YAML은 Root 검증이 남았습니다. viewer는 data/train.jsonl만 선택하고 근거 JSON을 학습 행에서 제외하는 제안이며 Hub 수용은 시험하지 않았습니다.

이 요청2건으로 새 학습을 수행하지 않았습니다. 준비 과정의 30/60 checkpoint 전 학습·combined81 projection/Fit·보호 validation/calibration/final 조회·LoRA/3진/MPS 학습·production 활성화·비용 절감 측정은0입니다. 기존79 corpus와 과거 모델 snapshot은 변경하지 않고 비활성 상태를 유지합니다. 모델 본체/바이너리/원본 소스 본문/복사한 upstream 설명/호스트 경로/raw trace/실사용자 prompt는 포함하지 않았습니다.

## 라이선스 범위와 출처

YAML apache-2.0과 정확한 프로젝트 [LICENSE](LICENSE)는 자체 작성 prose·유한 annotation의 저장소 라이선스를 가리킵니다. 포함된 upstream 고지·전체 소스 계보·ancestor 모델·모델 가중치를 Apache로 바꾸는 뜻이 아닙니다. 완전한 [고지5개](evidence/NOTICE-MANIFEST.v1.json)는 pflag BSD-3-Clause·mapstructure MIT·별도 Go BSD 고지를 보존합니다. 원문은 제외하고 무시된 Go2022 join의 미해결 계보도 미해결/제외 상태로 유지합니다. 과거 [조건부 권리 검토](evidence/rights-review/CONDITIONAL-ELIGIBILITY.v1.json)와 [재료 범위](evidence/MATERIAL-SCOPE.v1.json)의 한계는 그대로이며 포괄 법적 승인이 아닙니다.

원전 revision은 pflag c966cfef47379dcb01e7929504d66d94b540945b, mapstructure52aa5c6dc1d27226460807054ca2107b2d54fb2d입니다. 준비의 로컬 공개 소스 head는 acf514d3dd25db2e199e43da80ccf4477d8ff310, [PR110](https://github.com/teamswyg/laya-tools/pull/110)이며 **CI/merge는 준비 시점 pending**입니다. [PR109](https://github.com/teamswyg/laya-tools/pull/109)는 이전 native 관측 공개를 별도 보존 proof로 기록합니다. Root가 exact-head 필수 CI·커밋된 소스 핀·최종 공개 안전/라이선스 범위·JSONL/viewer metadata를 확인한 후 별도 게시 계획으로만 dataset을 생성/업로드할 수 있습니다. 준비의 HF repo/auth/API/HTTP/upload/create/token 조회는0입니다.
