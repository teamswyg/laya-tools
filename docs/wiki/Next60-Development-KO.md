# 작은 주장·힌트 모델의 개발 자료

**30개 의미 요청·88개 후보 라벨을 실제 검증하고 HF에 공개했습니다.** [30행 자료와 설명](../../experiments/short-claim/next60-development-thirty/README.ko.md) 및 [HF30 고정 버전](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-30-finite-v1)을 제공합니다. 소유558파일과 표30행의 모든 필드·순서를 대조했고 이전 태그를 보존했습니다. [게시·실제 CI 근거](../../experiments/short-claim/publication-proof-122/README.ko.md)를 보세요. 이번 자료 확장의 새 주장 모델/Fit은0이며 Codex 절감 효과는 아직 입증하지 않았습니다.

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

[사용 설명](../../experiments/short-claim/next60-development-thirty/README.ko.md)을 읽고 data/train.jsonl을 한 줄씩 사용하세요. 파일은 **41,428바이트**, SHA-256은 **9cdfb758f03adc34a9fb5e00e3c1525921df26d0912f6b554e4ed4c890fd10a2**입니다. 요청·후보 문장, 라벨·가중치, 출처 추적 정보가 있습니다.

HF CLI가 있다면 아래처럼 **고정 commit의 데이터 한 파일만** 받을 수 있습니다. 학습이나 모델 다운로드 없이 읽기만 가능합니다. 전체 검증 명령은 이 저장소 checkout의 루트에서 실행하세요.

```sh
hf download JooYoon/riidolaya-shortclaim-next60-development \
  releases/next60-30-finite-v1/next60-development-thirty/data/train.jsonl \
  --type dataset --revision 89c0c7c9ddc3135e37e88e9da5d51ac55dff0e1d \
  --local-dir ./riidolaya-next60-data
```

[Go reader](../../pkg/shortclaimdata)는 최대 행16KiB·후보8개의 고정 배열과 불변 값을 사용하며 lock을 쓰지 않습니다. 아래 코드에서 bytes와 github.com/teamswyg/laya-tools/pkg/shortclaimdata를 가져옵니다.

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

[다음 원천22개 제안](../../experiments/short-claim/next60-new-source-preview/README.ko.md)은 새 적격0개입니다. 기존 기능을 재포장하거나 현재 입력으로 차이를 구분 못하는 세 제안은 보류합니다. Go-list 의존성 목록 확인과 문구 중복0은 실행·의미 독립성·학습 자격을 보장하지 않습니다.

## 다음 학습의 판단 기준

현재 항상 두 번째 후보를 고르면 요청 기준 **27/30=90%**, 항상 부정을 고르면 후보 라벨 기준 **58/88≈65.9%**입니다. 분모가 다른 단순 기준이며 모델 정확도가 아닙니다. 의미 이해 없이 높은 점수가 가능해 후보 위치·설명 문체·어휘 대조군이 필요합니다. 기존 자료와 실패 기록을 고쳐 점수를 만들지 않습니다.

**30개는 자료 수집 checkpoint이며 자동 학습 시작 조건이 아닙니다.** 새로운 코드 가족의 의존성·공유 보조 코드·의미 중복을 확인하고, 전체 가족 단위의 학습·검증·보정 역할을 관측 전에 정합니다. 약60개 checkpoint에서 자료 준비도를 다시 확인합니다. 기존79개와 합친109개는 현재 단순 합계이며 계보·중복 검토 전에는 독립 학습량으로 주장하지 않습니다.

후속 모델은 단순 순서 선택·어휘·규칙 기준보다 검증에 필요한 전체 작업량을 최소 **5% 줄이는지** 평가합니다. 주장 도메인별 보호 요청 **2,400개** 계획을 유지합니다. 기존 Fit3회·논리 모델3개와 실패 모델 비활성은 그대로입니다. 3진 저장은 별도 실험이며 저장 크기 감소만으로 CPU 속도·효용 개선을 인정하지 않습니다.

[이슈19](https://github.com/teamswyg/laya-tools/issues/19)에서 실제 게시·CI 결과와 진행을 볼 수 있습니다. 자체 문서·소유 소스는 Apache-2.0, 원천 전문 고지는 해당 라이선스를 유지합니다. 원본 본문·모델 본체·사적 입력·인증정보·원시 journal은 공개 자료에 없습니다. 코드 라이선스 확인이 모델 계보 전체의 포괄적 권리 승인을 뜻하지 않습니다.

[첫 두 요청의 학습 자료](Native2-Training-KO) · [첫 원본 관측](Native2-Observation-KO) · [English](Next60-Development-EN)
