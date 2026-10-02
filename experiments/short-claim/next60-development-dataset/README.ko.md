---
license: apache-2.0
language:
  - en
pretty_name: Riidolaya Next60 finite development subset
tags:
  - riidolaya
  - development-only
  - finite-code-behavior
configs:
  - config_name: development
    default: true
    data_files:
      - split: train
        path: data/train.jsonl
---

# Riidolaya Next60 유한 범위 개발 학습 자료

[English](README.md)

**개발 데이터 공개 후보**로, **요청2개·후보 라벨5개**(긍정2/부정3)를 담았습니다. 작은 주장 점수 모델이 코드 동작 설명 중 어느 것이 요청을 충족하는지 한정된 범위에서 배우는 데 쓸 자료입니다. 학습된 모델이나 독립 평가 benchmark는 아닙니다. 별도 dataset CI·최종 검사·Root 게시 계획을 마칠 때까지 게시 대기입니다.

실제 데이터 행은 **영어만** 있습니다. 한국어는 사람이 읽는 안내이며, 학습 데이터가 한·영 두 언어라는 뜻은 아닙니다.

| 요청 | 연결된 전체 소스 그룹 | 후보 라벨 | 근거 범위 |
|---|---:|---:|---|
| next60-pflag-native-ipnet | 77: 연결된 cobra/pflag | 3 | 고정 IPv4 입력5개·기존 관측15회 |
| next60-mapstructure-native-or | 78: mapstructure | 2 | 고정 정수 hook 구성4개·기존 관측8회 |

첫 요청은 CIDR을 마스킹한 네트워크, 오류 시 이전 상태 보존, bare address 거절을 다룹니다. 둘째는 hook을 같은 입력으로 재시도하고 nil 출력이라도 성공하면 멈추며 실패 메시지를 줄바꿈과 함께 모으는 동작입니다. **입력 사례9개·관측23회는 근거이며 추가 dataset 행이 아닙니다**. 후보 순서와 후보별 sample weight1을 유지하며 반복 관측으로 가중치를 늘리지 않습니다.

두 ID는 이미 Next60 초안에 있었습니다. 기존 catalog 대비 **새 globally unique 요청 ID0·새 source family0**입니다. 모두 development_train이며 연결 소스 그룹77/78을 통째로 유지합니다. 설명과 라벨은 AI 보조·소스/관측 참고·nonblind로 작성했습니다. 긍정은 고정 사례 전부 충족·unknown0, 부정은 한정 범위의 불일치를 뜻합니다. 모든 입력의 정확성·새 family 일반화·독립 held-out 성능을 입증하지 않습니다.

## 본문을 사용하고 관리 정보는 분리하기

`data/train.jsonl`은 1행당 요청1개입니다. 모델 feature로는 **request와 candidates[].text만** 사용하세요. 후보 Boolean 라벨과 unit weight는 학습 정답입니다. metadata_id·안정 ID·역할·그룹·source revision·범위·hash는 관리 정보이며 모델 feature에 넣지 않습니다. API 이름은 본문 밖의 metadata ID에만 있습니다. 전체 JSON 행 문자열을 encoder에 넣지 마세요.

확장 행은 명시적 adapter가 필요하고 shortclaim.LoadValidated에 바로 들어가지 않습니다. Root는 request/candidate 본문과 metadata ID로 원래 owned 입력2개를 복원해 고정 원본과 의미가 같은지 확인했고 **LoadValidated2 + Validate2, Fit0**으로 통과했습니다. 원래 [입력2개](evidence/INPUTS.v3.json)와 [정확한 검증 기록](evidence/INPUT-VALIDATION.v1.json)을 포함했습니다. 입력 표현 검증이며 모델 성능이나 이후 학습 reader 통합 완료를 뜻하지 않습니다.

Root가 고정 Go converter를 **1회 실행**했고 출력은 초안2행과 바이트 단위로 같았습니다. 2,597바이트, SHA256 `843e4776823fe44ad14f0c7a8b6bdf29827d92596ef3d8bf1dd31d498e87ca1c`입니다. [정확한 Root 검증 기록](evidence/ROOT-VALIDATION.v1.json)은 Root의 실제 작업을 기록하며 이 패키지 작성자의 실행은0입니다. [변환 소스](source/convert.go.txt)는 본문으로 보존했고 바이너리가 아닙니다.

viewer metadata는 data/train.jsonl만 development/train으로 선택하고 근거 JSON을 학습 행으로 읽지 않습니다. 형식은 공식 [dataset-card 안내](https://huggingface.co/docs/hub/datasets-cards)와 [repository-structure 안내](https://huggingface.co/docs/datasets/repository_structure)에 맞춰 확인했습니다. 실제 Hub/viewer 수용은 게시 후 확인할 조건으로 남았습니다.

## 출처·한계·라이선스

Root가 PR110의 필수 CI4개·merge e8382e83cb1ac2ec5af42d62c0936a548f53166d·동일 tree f331956e35acf0556818caa021f097f16e00b24c를 확인했습니다. wiki6개는 commit832f44ffea58dfd7afd2b2168cffa3cf1e7f8ae5에서 대조했고 [정확한 보존 기록](evidence/PR110-CI-WIKI-PUBLICATION.v1.json)에 있습니다. 데이터/소스 게시 근거이며 모델 성능 결과가 아닙니다. 과거 pending 카드는 preparation/에 바이트 그대로 남겼고 원래 preparation provenance도 과거 시점 기록으로 보존했습니다.

이 공개 후보에 대해 새 Fit·combined81 학습·30/60 checkpoint 전 학습·유료 모델 호출·보호 평가를 수행하지 않았습니다. 기존79 데이터와 과거 비활성 모델은 변경하지 않았습니다. LoRA·3진·MPS·GPU·production 준비·Codex 비용 절감·2,400개 수집 완료를 주장하지 않습니다. 다른 역할과 연결 소스 family가 충돌하면 이후 학습을 차단하며 역할 이동이나 seed 탐색으로 해결하지 않습니다.

YAML apache-2.0과 프로젝트 [LICENSE](LICENSE)는 저장소 라이선스에 따른 자체 작성 prose·유한 annotation에 적용됩니다. upstream 고지·전체 소스 계보·ancestor 모델·가중치를 Apache로 바꾸는 뜻은 아닙니다. [완전한 고지5개](evidence/NOTICE-MANIFEST.v1.json)는 pflag BSD-3-Clause·mapstructure MIT·별도 Go BSD 고지를 보존합니다. 원본 소스 본문·개인 경로·바이너리·모델 본체·raw trace·실사용자 prompt는 제외했습니다. 무시된 Go2022 join의 미해결 계보는 미해결/제외 상태이며 과거 [조건부 권리](evidence/rights-review/CONDITIONAL-ELIGIBILITY.v1.json)와 [재료 범위](evidence/MATERIAL-SCOPE.v1.json)의 한계를 유지합니다.

예정 저장소는 JooYoon/riidolaya-shortclaim-next60-development, 불변 tag는 next60-2-finite-v1입니다. [RootPublicationPlan](RootPublicationPlan.json)은 정확한 dataset CI111·최종 안전/권리 검사·Root 별도 승인이 끝날 때까지 게시를 차단합니다. 이 패키지 작성자의 Go/helper/모델/HF/auth/API/HTTP/upload/create/token 조회는0입니다.
