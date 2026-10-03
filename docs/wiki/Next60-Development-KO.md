# 주장·힌트 모델용 유한 개발 자료: 검증한 33개

목표는 요청에 도움이 될 코드 후보를 아주 적은 비용으로 자주 제안하는 작은 **주장·힌트 모델**입니다. 힌트 뒤에는 실제 검사와 반례가 필요합니다. 이 자료는 그 학습 근거를 쌓는 개발 자료입니다. 새 모델의 성능이나 Codex 비용 절감 결과를 담은 자료가 아닙니다.

**33개 요청·96개 라벨을 실제 Go로 생성하고 프로젝트 Reader의33회 호출·반환·값 일치를 확인했습니다.** [생성 기록](https://github.com/teamswyg/laya-tools/blob/511e8ac7c213f3b48568025eb9d83882d525bcdb/experiments/short-claim/next60-development-thirtythree/MATERIALIZATION.v1.json)과 [Reader 기록](https://github.com/teamswyg/laya-tools/blob/511e8ac7c213f3b48568025eb9d83882d525bcdb/experiments/short-claim/next60-development-thirtythree/READER.v1.json)을 제공합니다. 기존30행을 바이트 그대로 보존했습니다. 파일은45,390바이트, SHA-256은7b28ae6d119884c902b32ba781b3514f1d3774ca3a60cea16fe4ce11a3aad919입니다. PR124의 Linux·macOS 재현 검사와 필수 CI 네 작업이 통과했고 봇이 병합했습니다. 아래의 실제 게시 결과는 그 이후 검증했습니다.

| 단위 | 새 배치 | 검증한 전체33개 |
|---|---:|---:|
| 의미 요청 | 3 | 33 |
| 선택한 후보 라벨 | 8: 긍정3·부정5 | 96: 긍정33·부정63 |
| 고정 입력 | 14 | 160 |
| 전체 원관측 | 42 | 476 |
| 선택한 후보의 관측 | 37 | 467 |
| 이 자료 확장의 새 Fit·모델 추론 | 0 | 0 |

42번 실행은 3개 요청에 대한 입력 변형과 후보 비교입니다. 독립 작업 42개로 세지 않습니다. 실제 기존30행에는 후보88개가 있으며, 이를 임의로90개로 채우지 않습니다. 새 후보9개 중 미상 전용1개를 제외하여 라벨8개를 추가합니다.

## 새로 검사한 세 작업

| 요청 | 관측으로 드러난 구분 | 유한 적합 후보 위치 |
|---|---|---:|
| 지정 폭의 비트 보수 | 요청 폭64인데 저장 길이65인 입력에서 결과 길이까지64여야 함 | 첫째 |
| JSON 정수의 검사된 추출 | 큰 정수를 정확한 문자열로 다루고 소수·범위 초과를 오류로 구별해야 함 | 둘째 |
| 엄격한 이진 비트셋 읽기 | 잔여 바이트·짧은 입력·꼬리 비트를 거부하고 수신 객체를 보존해야 함 | 셋째 |

기존 API는 원래 계약에 따라 동작합니다. 여기서는 **새로 요청한 계약**을 만족하는지 비교합니다. 기존 동작이 추가 요구를 만족하지 않는다고 해서 upstream 버그라고 단정하지 않습니다. 저장소·출처는 추적 정보이며 모델 특징에는 요청과 후보 문장만 사용합니다.

참·거짓·미상을 구분합니다. 42개 후보-입력 결과는 **충족26·반례11·미상5**, 세부 조건318개는 **충족272·반례31·미상15**입니다. GJSON의 기존 Int에는 요구한 오류 반환 통로가 없어 5개 결과를 미상으로 보존하고 그 후보에 라벨을 붙이지 않습니다. 이진 읽기 두 결과에는 객체 변경이라는 명확한 반례와 오류 종류 미상이 함께 있습니다. 반례가 부정 라벨의 근거이며, 미상을 거짓으로 바꾸지는 않습니다.

새 배치의 미상15개 중 선택한 후보에는2개, 제외한 후보에는13개가 있습니다. 기존30행 생성 기록의 **23행 이후** 미상8개와 합친 범위는 각각10·23개입니다. 더 오래된 앞부분의 미상 기록은 기존 자료에 그대로 남아 있습니다. 이 수치를 전체 연구 이력의 미상 총계로 주장하지 않습니다.

## 사람과 에이전트의 사용법

데이터 파일은 data/train.jsonl이며 한 줄이 요청 하나입니다. 기존30행은 앞부분에 바이트 그대로 유지하고 새3행을 뒤에 붙입니다. 각 후보에는 명확한 라벨과 가중치1이 있습니다. 제외한 후보와 미상 조건은 qualification/ROOT-QUALIFICATION.v3.json 및 evidence/SAVED-COMPARISON.v1.json에 보존합니다.

저장소 루트에서 다음 검사를 실행할 수 있도록 Go 소스와 검증 명령을 함께 제공합니다.

```sh
bash scripts/verify-next60-thirtythree.sh
```

검사 범위는 저장된 조건 비교, 이전 데이터 보존, 데이터 생성, 실제 LoadDevelopmentRow 값 대조와 자체 오류 제어입니다. 원본 후보·Laya·주장 모델은 실행하지 않습니다. source/saved-comparer의 공개 재현 검사는 원래 결과와 모든42행·318조건을 비교합니다. 공개용 경로 없는 설명으로 바꾼 네 opaque 기록의 지문만 비교에서 제외합니다. 사적 실행 승인 원본을 재인증했다는 의미는 아닙니다.

Go Reader 사용 예시는 기존 [자료30 설명](https://github.com/teamswyg/laya-tools/blob/511e8ac7c213f3b48568025eb9d83882d525bcdb/experiments/short-claim/next60-development-thirty/README.ko.md)을 참고하세요. 최대16KiB 행과8개 후보의 고정 배열을 사용하며 lock이 없습니다. 새 원천은 비트셋 그룹82, GJSON·Match·Pretty 그룹83으로 묶어 development_train에 유지합니다. 저장 형식의 원천 이름은 bits-and-blooms-bitset과 tidwall-gjson이며 실제 저장소 이름은 별도 추적 필드에 보존합니다.

## 측정과 다음 학습

원관측은 한 번만 실행했습니다. 수집 프로세스의 OS 최대 RSS는 **17,612,800바이트(약16.8MiB)**, 전체 시간은 **8.052초**였습니다. 여기에는1,739번의 파일·디렉터리 저장 및 ACK가 포함됩니다. 모델 추론 메모리·GPU 실행·속도 개선을 뜻하지 않습니다. 저장 비교기 실행 파일은 디버그 정보를 제외해 **2,850,722바이트**였으며, 이것도 모델 본체 크기가 아닙니다. Go pprof와 OS RSS는 측정 범위가 다릅니다.

새 세 요청은 후보 위치를 나눴지만 전체 자료에서 항상 둘째를 고르는 기준은 여전히28/33입니다. 학습 전 위치·문체·어휘 기준과 의미 중복을 검사합니다. **33개는 수집 단계이며 자동 Fit 시작 조건이 아닙니다.** 약60개에서 출처 가족 단위의 준비도를 다시 판단하고, 도메인별 보호 요청2,400개 계획과 검증 작업량5% 감소 목표를 유지합니다. 기존 Fit3회·논리 모델3개·실패 모델 비활성은 유지합니다. 3진 저장 크기와 실제 CPU 효용은 별도로 검증합니다.

다음 수집 준비는 물리 JSON Lines의 완전한 값·오류 전 callback prefix와 잘못된 UTF-8 append 거부입니다. 기존 ForEachLine은 이미 callback의 중지를 존중하므로 이를 결함으로 재포장하지 않습니다. 새 입력의 기대값을 실제 결과 전에 고정합니다.

## 공개 범위와 기록

자체 코드·설명·입력은 Apache-2.0입니다. 선택한 upstream API의 BSD/MIT 전문 고지는 evidence/notices에 각각 유지합니다. 공개 자료에는 원본 코드 본문·모델 본체·실행 바이너리·원시 journal·사적 작업 입력·인증정보가 없습니다. 원본 source/go.mod/고지의 지문만 추적합니다. 모델 게시 권리는 별도 판단입니다.

초안의 미상 범위 오류·원천 이름 저장 형식 수정·최초 비교기 빌드의 크기 제한 실패와 후속 확인은 history/CORRECTIONS.v1.json에 기록합니다. 실패 파일과 이전 초안은 로컬에 보존했습니다. 모든 라벨은 고정 입력에 대한 유한 근거이며 새로운 입력의 일반 정답을 보장하지 않습니다.

[이슈19](https://github.com/teamswyg/laya-tools/issues/19) · [이전 HF30 고정 버전](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-30-finite-v1) · [English](https://github.com/teamswyg/laya-tools/blob/511e8ac7c213f3b48568025eb9d83882d525bcdb/experiments/short-claim/next60-development-thirtythree/README.en.md)

## 실제 게시와 간단한 다운로드

[HF33 고정 버전](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/2a1c2224e89f60eef9e381f98c0117206f60e0d2)에 공개했습니다. 새로 추가·변경한 **82개 파일 842,356바이트**와 관리 속성 파일을 고정 커밋에서 다시 받아 대조했습니다. 현재 목록은 소유638개·관리 속성을 포함639개이며 이전 파일 경로와 태그7개를 보존했습니다. 이전558파일 전체 검증은 HF30에서 수행한 기록입니다. 이번 커밋에서 옛 파일 전체를 다시 받았다는 의미는 아닙니다.

Viewer는 첫 요청에서 서버 준비 지연500을 반환했고, 두 번째 요청에서 **33행의 모든 필드와 순서 일치·partial=false·잘린 셀0**을 확인했습니다. 응답 커밋과 최상위 잘림 표시는 제공되지 않아 미확인으로 남깁니다. Viewer는 현재 main을 따르며 고정 커밋 파일 증명과 구분합니다.

아래는 **데이터 한 파일만** 다운로드합니다. 모델이나 학습 실행 없이 사람과 에이전트가 읽어 사용할 수 있습니다.

```sh
hf download JooYoon/riidolaya-shortclaim-next60-development \
  releases/next60-33-finite-v1/next60-development-thirtythree/data/train.jsonl \
  --type dataset --revision 2a1c2224e89f60eef9e381f98c0117206f60e0d2 \
  --local-dir ./riidolaya-next60-data
```

[Go Reader](https://github.com/teamswyg/laya-tools/blob/511e8ac7c213f3b48568025eb9d83882d525bcdb/pkg/shortclaimdata) 예시입니다. bytes와 github.com/teamswyg/laya-tools/pkg/shortclaimdata를 가져옵니다.

```go
example, err := shortclaimdata.LoadDevelopmentRow(bytes.NewReader(line))
if err != nil {
    return err
}
input := example.Input()
supervision := example.Supervision()
```

입력·정답·추적 정보는 구분됩니다. Reader는 점수나 학습을 실행하지 않습니다. 다음 입력 브리지는 1·7·33·60개 합성 자료와 듬성한 그룹 번호, 잘못된 계획·지문·행 연결을 검사하는 Go 테스트8개를 통과했습니다. 실제 새 자료를 학습에 넘기거나 Fit을 실행한 결과는 아닙니다. 후보 순서를 모두 바꾸는 편향 감사도 준비 중입니다.

<details>
<summary>이전 HF30 안내 원문 — 아래 “현재”와 수치는 당시 상태입니다</summary>

# 작은 주장·힌트 모델의 개발 자료

**30개 의미 요청·88개 후보 라벨을 실제 검증하고 HF에 공개했습니다.** [30행 자료와 설명](https://github.com/teamswyg/laya-tools/blob/560d5b2cfcdf6f1c23f8b624dbdca3a258ca462c/experiments/short-claim/next60-development-thirty/README.ko.md) 및 [HF30 고정 버전](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-30-finite-v1)을 제공합니다. 소유558파일과 표30행의 모든 필드·순서를 대조했고 이전 태그를 보존했습니다. [게시·실제 CI 근거](https://github.com/teamswyg/laya-tools/blob/560d5b2cfcdf6f1c23f8b624dbdca3a258ca462c/experiments/short-claim/publication-proof-122/README.ko.md)를 보세요. 이번 자료 확장의 새 주장 모델/Fit은0이며 Codex 절감 효과는 아직 입증하지 않았습니다.

목표는 아주 작은 CPU·메모리 비용으로 자주 호출하는 **주장·힌트 모델**입니다. “이 요청에 도움이 될 후보는 이것 같다”는 힌트를 주고, 최종 선택은 실제 테스트·검색·검증으로 뒷받침합니다. 문장 생성 능력보다 후보를 좁혀 전체 작업량을 줄이는지가 중요합니다.

| 단위 | 현재 검증·공개30행 | 이전 HF23행 |
|---|---:|---:|
| 서로 다른 의미 요청 | 30 | 23 |
| 후보 라벨 | 88: 긍정30·부정58 | 67: 긍정23·부정44 |
| 고정 입력 변형 | 146 | 113 |
| 전체 원관측 | 434 | 335 |
| 선택한 학습 후보의 관측 | 430 | 331 |
| 이번 확장의 새 학습·모델 추론 | 0 | 0 |

입력 변형이나 후보 실행 횟수를 독립 요청 수로 더하지 않습니다. 30행은 모두 development_train, 후보 가중치는1입니다. 기존23행과 별도 검증했던27행은 바이트 그대로 앞부분에 남겼습니다. 출처 가족을 묶은 그룹76–81을 사용하며 retryablehttp 가족의 다음 수집은 보류 상태입니다. 새 원천 일반화는 아직 검증하지 않았습니다.

## 사람과 에이전트가 사용하는 방법

[사용 설명](https://github.com/teamswyg/laya-tools/blob/560d5b2cfcdf6f1c23f8b624dbdca3a258ca462c/experiments/short-claim/next60-development-thirty/README.ko.md)을 읽고 data/train.jsonl을 한 줄씩 사용하세요. 파일은 **41,428바이트**, SHA-256은 **9cdfb758f03adc34a9fb5e00e3c1525921df26d0912f6b554e4ed4c890fd10a2**입니다. 요청·후보 문장, 라벨·가중치, 출처 추적 정보가 있습니다.

HF CLI가 있다면 아래처럼 **고정 commit의 데이터 한 파일만** 받을 수 있습니다. 학습이나 모델 다운로드 없이 읽기만 가능합니다. 전체 검증 명령은 이 저장소 checkout의 루트에서 실행하세요.

```sh
hf download JooYoon/riidolaya-shortclaim-next60-development \
  releases/next60-30-finite-v1/next60-development-thirty/data/train.jsonl \
  --type dataset --revision 89c0c7c9ddc3135e37e88e9da5d51ac55dff0e1d \
  --local-dir ./riidolaya-next60-data
```

[Go reader](https://github.com/teamswyg/laya-tools/tree/560d5b2cfcdf6f1c23f8b624dbdca3a258ca462c/pkg/shortclaimdata)는 최대 행16KiB·후보8개의 고정 배열과 불변 값을 사용하며 lock을 쓰지 않습니다. 아래 코드에서 bytes와 github.com/teamswyg/laya-tools/pkg/shortclaimdata를 가져옵니다.

```go
example, err := shortclaimdata.LoadDevelopmentRow(bytes.NewReader(line))
if err != nil {
    return err
}
input := example.Input()
supervision := example.Supervision()
```

모델 특징에는 **요청과 후보 문장만** 사용합니다. ID·출처·그룹·리비전·검사 결과는 추적 또는 정답 정보입니다. Reader 호출은 점수·학습·모델 추론을 실행하지 않습니다. 배열 구조를 사용한다는 사실만으로 속도 향상을 주장하지 않습니다.

다음 명령은 저장된 자료로 데이터 재생성·실제 프로젝트 Reader·판정 기록111개를 검사합니다. 로컬과 PR121의 Linux·macOS CI에서 실제 통과했습니다. 기존 Laya native 검사도 별도로 통과했으며 이 명령 자체는 모델이나 원본 함수를 실행하지 않습니다.

```sh
bash scripts/verify-next60-thirty.sh
```

## 실제로 확인한 범위

30행을 생성하고 실제 Reader의 **30호출·30반환·30일치**, 호출 전후 **60개 저장 checkpoint**를 확인했습니다. 입력·후보·라벨·가중치와 제한된 배열의 빈 칸까지 대조합니다. 판단 불가 조건 **8개**를 보존합니다. 미상을 거짓으로 바꾸지 않고, 부정 후보에는 별도의 알려진 반례가 있어야 합니다.

마지막 세 요청은 INI 전체 입력 바이트 예산, 제한된 파일 읽기, 배타적 파일 쓰기입니다. 실제 Go 원본에서 입력14개×후보3개의 **42관측**을 수집했습니다. 세부 조건111개는 참89·거짓20·판단 불가2였습니다. 참고 후보도 관측을 만족했기 때문에 채택했으며 그 이름 때문에 정답을 부여하지 않았습니다. 전체 parser 옵션·파일시스템 동시성·모든 입력을 보장하지 않습니다.

관측 수집 프로세스의 OS 최대 RSS는 약 **17.4MiB**, 시작·소스 확인·저장을 포함한 시간은 약 **2.52초**였습니다. Laya 또는 주장 모델의 추론 메모리·GPU 실행·Codex 절감 수치가 아닙니다. 원본은 한 번 실행했고, 저장 비교기의 최초 설정 형식 오류를 보존한 뒤 저장 파일 읽기만 다시 수행했습니다.

[다음 원천22개 제안](https://github.com/teamswyg/laya-tools/blob/560d5b2cfcdf6f1c23f8b624dbdca3a258ca462c/experiments/short-claim/next60-new-source-preview/README.ko.md)은 새 적격0개입니다. 기존 기능을 재포장하거나 현재 입력으로 차이를 구분 못하는 세 제안은 보류합니다. Go-list 의존성 목록 확인과 문구 중복0은 실행·의미 독립성·학습 자격을 보장하지 않습니다.

## 다음 학습의 판단 기준

현재 항상 두 번째 후보를 고르면 요청 기준 **27/30=90%**, 항상 부정을 고르면 후보 라벨 기준 **58/88≈65.9%**입니다. 분모가 다른 단순 기준이며 모델 정확도가 아닙니다. 의미 이해 없이 높은 점수가 가능해 후보 위치·설명 문체·어휘 대조군이 필요합니다. 기존 자료와 실패 기록을 고쳐 점수를 만들지 않습니다.

**30개는 자료 수집 checkpoint이며 자동 학습 시작 조건이 아닙니다.** 새로운 코드 가족의 의존성·공유 보조 코드·의미 중복을 확인하고, 전체 가족 단위의 학습·검증·보정 역할을 관측 전에 정합니다. 약60개 checkpoint에서 자료 준비도를 다시 확인합니다. 기존79개와 합친109개는 현재 단순 합계이며 계보·중복 검토 전에는 독립 학습량으로 주장하지 않습니다.

후속 모델은 단순 순서 선택·어휘·규칙 기준보다 검증에 필요한 전체 작업량을 최소 **5% 줄이는지** 평가합니다. 주장 도메인별 보호 요청 **2,400개** 계획을 유지합니다. 기존 Fit3회·논리 모델3개와 실패 모델 비활성은 그대로입니다. 3진 저장은 별도 실험이며 저장 크기 감소만으로 CPU 속도·효용 개선을 인정하지 않습니다.

[이슈19](https://github.com/teamswyg/laya-tools/issues/19)에서 실제 게시·CI 결과와 진행을 볼 수 있습니다. 자체 문서·소유 소스는 Apache-2.0, 원천 전문 고지는 해당 라이선스를 유지합니다. 원본 본문·모델 본체·사적 입력·인증정보·원시 journal은 공개 자료에 없습니다. 코드 라이선스 확인이 모델 계보 전체의 포괄적 권리 승인을 뜻하지 않습니다.

[첫 두 요청의 학습 자료](Native2-Training-KO) · [첫 원본 관측](Native2-Observation-KO) · [English](Next60-Development-EN)

</details>
