# 작은 주장 모델을 위한 개발 자료

현재 공개된 고정 버전은 **23개 의미 요청·67개 후보 라벨**입니다. [Hugging Face `next60-23-finite-v1`](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-23-finite-v1)에서 내려받을 수 있습니다. 별도로 개발용 채택을 마친 풀은 **27개 요청·79개 라벨**이며, 뒤의 네 요청은 아직 이 23행 파일에 포함되지 않았습니다.

목표는 아주 작은 CPU·메모리 비용으로 자주 호출하는 **주장·힌트 모델**입니다. 요청에 도움이 될 후보를 좁히는 힌트를 주고, 최종 판단은 테스트와 검증 근거에 맡깁니다. 지금 늘리는 것은 검증된 학습 재료입니다. 이번 자료 추가에서 새 학습·모델 추론은 0회이며, LLM 가속이나 Codex 토큰 절감은 아직 입증하지 않았습니다.

| 단위 | 공개 고정 버전 | 별도 채택 풀 |
|---|---:|---:|
| 서로 다른 의미 요청 | 23 | 27 |
| 후보 라벨 | 67: 긍정23·부정44 | 79: 긍정27·부정52 |
| 고정 입력 변형 | 113 | 132 |
| 전체 원관측 | 335 | 392 |
| 선택한 학습 후보의 관측 | 331 | 388 |
| 새 corpus Fit | 0 | 0 |

입력 변형·후보 실행 횟수를 독립 요청 수로 더하지 않습니다. 공개 23행은 모두 `development_train`, 후보 가중치는 1입니다. 기존 humanize 그룹76, pflag/Cobra 그룹77, mapstructure 그룹78을 재사용합니다. 뒤의 네 요청에 필요한 Afero·retryablehttp·INI와 기존 보조 코드의 연결 그룹은 **다음 자료 생성 전에 확인할 항목**입니다. 새 번호를 임의로 부여하거나 새 원천 일반화를 주장하지 않습니다.

## 내려받아 읽기

HF에서 `development` 설정과 `train` split을 선택하세요. 실제 파일은 `releases/next60-23-finite-v1/next60-development-twentythree/data/train.jsonl`입니다. 루트의 `data/train.jsonl`은 최초 두 요청의 역사 자료입니다.

```sh
hf download JooYoon/riidolaya-shortclaim-next60-development \
  --type dataset \
  --revision 0d964a547708db596b13b0eae18d2c93dd3e3ac4 \
  --include "releases/next60-23-finite-v1/next60-development-twentythree/data/train.jsonl" \
  --local-dir ./next60-23
```

파일은 **31,493바이트**, SHA-256은 `39edb1bb60e88d56cb2fec271b511a5ce09a0ff8bd33ce21d0dda7defb3ce362`입니다. 이전 21행의 28,800바이트를 그대로 보존합니다. [자료 사용 설명과 실제 reader 검증](https://github.com/teamswyg/laya-tools/tree/67319e639b52d302292a0abdbbd83149c1fb0d99/experiments/short-claim/next60-development-twentythree)을 함께 읽을 수 있습니다.

[Go reader](https://github.com/teamswyg/laya-tools/tree/67319e639b52d302292a0abdbbd83149c1fb0d99/pkg/shortclaimdata)는 한 행씩 읽습니다. 최대 행16KiB·후보8개이며 고정 배열과 불변 문자열을 보유하고 lock을 쓰지 않습니다.

```go
example, err := shortclaimdata.LoadDevelopmentRow(bytes.NewReader(line))
if err != nil {
    return err
}
input := example.Input()
supervision := example.Supervision()
```

`bytes`와 `github.com/teamswyg/laya-tools/pkg/shortclaimdata`를 가져옵니다. 모델 특징에는 **요청과 후보의 문장만** 사용합니다. ID·출처·그룹·리비전·유한 범위는 추적 정보이고, 라벨·가중치는 정답 정보입니다. Reader 호출은 점수나 학습을 실행하지 않습니다. 배열 구조 자체를 측정된 속도 향상으로 해석하지 않습니다.

## 실제로 확인한 것

[PR119](https://github.com/teamswyg/laya-tools/pull/119)의 정확한 head `9d4b0b5e39628a8d7a2bb50a3eb9f9b018ccd0ef`에서 [필수 CI 네 개](https://github.com/teamswyg/laya-tools/actions/runs/37093356839)가 모두 통과했습니다. GitHub Actions 봇이 같은 소스 트리의 `67319e639b52d302292a0abdbbd83149c1fb0d99`로 병합했습니다. 두 추가 Go 검사는 저장된 유한 비교를 재현하고, 23행을 다시 생성해 실제 reader와 대조합니다. 기존 Laya native inference CI는 별도 단계입니다.

HF 커밋 [`0d964a547708db596b13b0eae18d2c93dd3e3ac4`](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/0d964a547708db596b13b0eae18d2c93dd3e3ac4)에서 소유 파일 **401개·2,989,766바이트**를 모두 다시 내려받아 확인했습니다. 현재 목록의 payload399개도 일치합니다. 원격402파일의 나머지 하나는 이전과 바이트가 같은 HF 관리 `.gitattributes`입니다. 루트 `FILE-MANIFEST.v6.json`과 `SHA256SUMS.v6`를 사용하며 과거 목록은 해당 과거 태그에서 검증합니다. 기존2·3·7·16·21 태그는 이동하지 않았습니다.

Viewer는 HTTP200, 관측23행·전체 필드·행 순서 일치, `truncated=false`, 잘린 셀0입니다. 이 응답에는 커밋과 전체 행수 필드가 없으므로 현재 화면 검증과 고정 커밋의 파일 검증을 구분합니다. 최초 조회의 HTTP500은 보존했으며 이후 실제200 응답으로 확인했습니다. 첫 실제 Go reader 검사도 23호출·23반환·23일치와 호출 전후46저장 checkpoint를 확인했습니다.

23개 버전에 새로 포함된 두 요청은 문자열 목록의 누적 항목 수 제한과 검증 오류 모두 수집입니다. 입력10개×후보3개를 실제 실행한30관측에서 만족23·불만족7·판단 불가0, 세부 조건은 일치125·불일치15·미상1입니다. 오류 코드 미상은 그대로 남깁니다. 이전 writer 원본의 미상4관측도 학습 후보 선택에서 제외한 상태입니다. 미상을 부정으로 바꾸지 않습니다.

별도 채택을 마친 두 요청은 `IOFS.Sub` 이름 검증과 요청 body의 byte-slice snapshot입니다. 최초27관측은 만족15·불만족12·판단 불가0, 세부 조건은 참57·거짓16·미상2입니다. Sub 오류 타입 미상2개는 알려진 반례와 함께 보존했습니다. 각 참고 후보는 고정 입력을 만족하고 다른 후보에는 알려진 반례가 있어 Root가 유한 개발 요청으로 채택했습니다. 이 결과는 모든 경로나 모든 body 타입을 보장하지 않습니다. 원본 worker의 내부·초기화 호출은 미계측 null입니다.

## 다음에 무엇이 달라지나

INI quoted value와 INI section 삭제 범위도 고정 입력 다섯 개씩의 실제 관측·저장 비교를 마쳐 별도 채택했습니다. 관측은 각각 만족9/불만족6, 만족10/불만족5입니다. 세부 조건 미상은 각각2개씩 유지합니다. 삭제에서 기존 구현의 getter panic 뒤 읽지 못한 상태는 미상으로 남겼고, 공개 발생 순서와 내부 인덱스를 구분합니다. 동시성·원자성이나 parser의 모든 옵션을 보장하지 않습니다. [상세 검증 기록](https://github.com/teamswyg/laya-tools/tree/research/next60-ini-two-qualified-and-hf23-120/experiments/short-claim/next60-ini-two-actual-observation)을 볼 수 있습니다.

남은 세 요청은 INI 전체 입력 바이트 예산, bounded file read, 배타적 file write입니다. 고정 입력과 정답을 먼저 검토한 뒤 원본 관측·저장 결과 비교·별도 채택을 거칩니다. 실행을 준비했다는 사실만으로 채택 수를 올리지 않습니다.

후보 위치 편향은 큽니다. 공개23행에서 항상 인덱스1을 고르는 기준은 요청 기준20/23≈87.0%, 항상 부정을 고르는 기준은 후보 라벨 기준44/67≈65.7%입니다. 별도27개 풀에서는24/27과52/79입니다. 분모가 다르며 모델 정확도가 아닙니다. 후보 순서 변경·어휘 대조군·연결 원천 분리 없이 높은 점수를 개선으로 채택하지 않습니다.

**적격 의미 요청30개 전에는 새 corpus Fit을 시작하지 않습니다.** 이후60개 checkpoint, 주장 도메인별 보호 요청2,400개, 단순 대조군 대비 효용5% 기준을 유지합니다. 기존79개 자료·역사적 Fit3회·논리 모델3개·실패 모델 비활성은 그대로입니다. 이용 가능한 힌트 모델인지 증거로 확인한 뒤 활성화합니다. [이슈19](https://github.com/teamswyg/laya-tools/issues/19)에서 진행을 확인할 수 있습니다.

자체 문서·주석·소유 소스는 Apache-2.0이며 원천의 전문 고지는 원래 라이선스를 유지합니다. 원본 본문·모델 본체·사적 입력·인증정보·원시 journal은 이 공개 자료에 없습니다. 미해결 과거 join 계보는 제외한 상태이며 모델 계보 전체의 포괄적 권리 승인을 뜻하지 않습니다.

[첫 두 요청의 학습 자료](Native2-Training-KO) · [첫 원본 관측](Native2-Observation-KO) · [English](Next60-Development-EN)
