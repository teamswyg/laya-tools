# 16개 개발 학습 항목을 위한 입력 검증 준비

이 패키지는 앞으로 만들 16개 데이터의 연결과 형식을 확인하는 검증기 소스입니다. 기존 7개 학습 파일의 9,619바이트를 그대로 앞부분에 보존하고, 고정 Catalog10의 첫 9개 작업을 순서대로 추가하는지 확인합니다. 실제 16개 파일은 여기서 만들거나 실행하지 않았습니다. Root가 별도로 작성한 채택 기록은 SHA-256 `b2d5a732d7f5d9238883060a881fc0c8b21b08c5b26004a0c27ef374be0aa9d4`, 16,030바이트로 고정했습니다.

검증자는 후보·Wanted·observer·Root 채택 기록·학습 데이터의 작성자가 아닙니다. 기존 소스 검토와 저장 결과 비교를 알고 있으므로 눈가림 검토는 아닙니다. 이 검증기 자체의 작성자이므로 Root가 봉인된 소스를 독립적으로 읽은 다음 실제 데이터에 사용하도록 준비했습니다. 이것은 사용자 승인을 새로 요구하는 절차가 아니라 코드와 결과의 독립적 확인입니다.

`module/validate.go`는 표준 라이브러리만 사용하는 core입니다. `source/reader-main.go.txt`는 기존 `shortclaimdata.LoadDevelopmentRow`를 호출하는 CLI 연결 소스입니다. 두 파일을 같은 `package main` 디렉터리에 `.go` 파일로 복사하고, laya-tools 모듈 안에서 함께 실행하는 구성입니다. 순수 합성 검사에서는 연결 소스를 컴파일하거나 기존 reader를 실행하지 않았습니다.

예상 사용법은 다음과 같습니다. 파일 이름은 예시이며, DATA_SHA는 Root가 완성한 데이터의 정확한 SHA-256으로 바꿉니다. CLI는 정확히 7개 옵션을 `--이름 값` 형태로 한 번씩 받습니다. 중복 옵션, `--이름=값`, 위치 인자는 거절합니다.

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off CGO_ENABLED=0 \
go run source/validate.go source/reader-main.go \
  --data TRAIN_JSONL --data-sha256 DATA_SHA \
  --previous-data PREVIOUS_SEVEN_JSONL --catalog CATALOG10_V2_JSON \
  --adoption ROOT_QUALIFICATION_JSON \
  --adoption-sha256 b2d5a732d7f5d9238883060a881fc0c8b21b08c5b26004a0c27ef374be0aa9d4 \
  --finite-comparison FINITE_COMPARISON_JSON
```

준비 작업에서는 정확한 Go 1.27.1 darwin/arm64를 사용했습니다. 실제 사용도 그 버전을 명시적으로 선택합니다. 각 입력은 regular 파일, 최대 128 KiB이며, 데이터 한 행은 최대 16 KiB입니다. 경로나 원문을 진단에 포함하지 않습니다. stdout에는 고정된 구조의 JSON 요약을 출력하며, 성공은 종료 코드 0, 입력 실패는 1, 인자 실패는 2입니다.

검증은 파일 pin → 기존 7개 원문 prefix → Root 채택과 비교 기록 연결 → 16개 행의 내용과 메타데이터 → 기존 reader 호출 → 최종 집계 순서입니다. 각 행은 기존 reader와 같은 정확한 11개 필드를 갖습니다. 후보는 metadata_id/text/label/sample_weight 네 필드, finite_scope는 다섯 필드입니다. 필드 누락과 null, 명시적 false와 빈 배열을 구분합니다. 중복 또는 대소문자 별칭 필드, 잘못된 Unicode escape, 바뀐 request/caption, 비단위 가중치, 새로운 그룹·리비전·역할, 확대된 유한 보장을 거절합니다.

새 항목의 stable_id는 `next60-`와 Catalog ID를 붙인 문자열입니다. metadata_id는 Catalog의 original/authored_reference/authored_negative 순서입니다. 실제 라벨은 Root의 고정된 채택 기록에 있는 false/true/false와 일치해야 합니다. 참조 후보는 모든 고정 입력이 만족이고 unknown이 없어야 합니다. 음성 후보는 알려진 반례가 필요하며, 동시에 존재하는 unknown 위치는 저장 비교와 정확히 같아야 합니다. unknown 하나만으로 음성 라벨을 만들지 않습니다. source family는 기존 group77 pflag 연결군과 group78 mapstructure이며, 새 family나 평가 역할을 만들지 않습니다.

예상 성공 집계는 16개 request, 47개 후보 라벨(양성 16·음성 31), 80개 고정 입력, 236개 후보 관측입니다. 기존 reader는 16회 완료하고, feature·score·projection·Fit·모델·원본 후보 호출은 모두 0입니다. 검증기 자체의 labels_assigned와 qualified는 false입니다. 이것은 Root가 이미 한 유한 채택을 새 데이터에 충실히 옮겼는지 검사하는 것이며, 의미를 다시 판정하거나 학습 권한을 만드는 작업이 아닙니다.

표준 라이브러리 합성 검사 1회에서 상위 검사 9개와 하위 검사 31개가 통과했고, 실패·skip·stderr는 0입니다. vet 1회도 통과했습니다. 합성 데이터와 reader stub의 호출은 실제 데이터 검증, 기존 reader 호출 또는 모델 성능 증거가 아닙니다. 가짜 파일이 production pin을 통과하지 못하는 검사도 포함했습니다. 실제 16개 검증은 아직 0회입니다.

학습 해석에는 주의가 필요합니다. 새 9개 정답은 모두 후보 위치 1에 있고, 전체 16개 중 14개도 그 위치입니다. 따라서 위치 1만 고르는 기준도 이 고정 배열에서는 14/16, 즉 87.5%가 됩니다. 이는 실제 모델 점수가 아니라 산술 기준입니다. 캡션에도 요청을 그대로 되풀이하는 참조 문장과 `but`, `losing`, `exposes` 같은 오류 문장 스타일 차이가 있습니다. 이후에는 별도 고정 평가 자료에서 후보 순서를 바꾸고, 단순 순서·다수 라벨·어휘 기준과 비교해야 합니다. 원본 학습 데이터는 보존하며, 이 준비에서 그 검사나 학습을 실행하지 않았습니다.

[계획](PLAN.v1.json)과 [위치·문체 누출 분석](LEAKAGE-REVIEW.v1.json)에 범위와 제한을 기록했습니다. 소스 기반 확인은 모델 품질, 비용 절감, 외부 프로세스 증명, 라이선스 보증 또는 새 공개 CI 통과를 주장하지 않습니다. 공유 저장소·다른 봉인 폴더·모델·보호 자료·네트워크에는 쓰기나 실행을 하지 않았습니다.
