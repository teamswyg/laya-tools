# Chi 라우팅 순서와 작은 주장 모델 검증

[English](README.en.md) · [선택형 Go API](../../../pkg/hintprepared/README.ko.md) · [CI](https://github.com/teamswyg/laya-tools/actions/workflows/ci.yml)

이 실험은 설명이 비슷한 코드도 처리 순서 때문에 다른 결과를 낼 수 있다는 점을 검증합니다. 작은 모델은 후보의 검사 순서를 제안하고, 실제 계약 검사가 최종 판단을 담당하는 구성을 목표로 합니다. 점수는 확률이 아닙니다.

**이번 결과: Go API의 계산 호환성은 통과했지만, 기존 모델의 추천 품질은 부족했습니다.** 모델은 모든 검사를 통과한 후보를 5순위에 놓았고 BM25는 1순위에 놓았습니다. 기존 실패 모델은 계속 비활성입니다. 새 학습·모델 배포·기본 정책 변경은 없습니다.

최신 준비 최적화는 [생성자 배열 재사용 비교](constructor-scratch/README.ko.md)에 기록했습니다. 모든 값·순위를 유지하면서 Go 누적 할당 바이트 약37%를 줄였고, 쌍별 시간 중앙값은 약7.5–7.7% 늘었습니다. 할당 횟수와 최종 저장량도 구분해 공개합니다. 모델의 정답 후보 5위 실패는 그대로입니다.

## 어떤 문제인가요?

`GET /u/{id}`에서 `a%2Fb`를 이름으로 받는 상황을 생각하면 됩니다. 먼저 경로를 디코딩하면 `/u/a/b`가 되어 경로 구간이 늘어납니다. 라우트를 먼저 찾은 뒤 이름만 디코딩하면 하나의 이름 `a/b`를 유지할 수 있습니다.

계약은 다음과 같습니다. RawPath가 있으면 그 경로로 먼저 라우팅합니다. 사용자가 디코딩을 선택했고 원래 RawPath가 비어 있지 않을 때만, 매칭된 id에 PathUnescape를 한 번 적용합니다. 두 파라미터 API인 chi.URLParam과 Request.PathValue를 함께 갱신하고, 원래 URL의 Path와 RawPath를 보존합니다. RawPath가 없거나 선택하지 않았다면 기존 값을 유지합니다.

다섯 처리 방식을 함께 시험했습니다.

| 후보 | 처리 방식 | 9개 입력 중 계약 통과 |
|---|---|---:|
| post_path | 라우팅 후 PathUnescape 한 번 | 9 |
| post_query | 라우팅 후 QueryUnescape 한 번 | 8 |
| pre_path | 복제한 요청의 전체 경로를 먼저 디코딩 | 8 |
| twice_path | 라우팅 후 PathUnescape 두 번 | 8 |
| unchanged | 캡처된 이름을 디코딩하지 않음 | 6 |

QueryUnescape는 `+`를 공백으로 바꿉니다. 두 번 디코딩하면 `a%252Fb`에서 한 번 남겨야 할 `%2F`까지 풀립니다. 전체 경로를 먼저 풀면 인코딩된 슬래시가 라우팅 구분자가 됩니다. 오류나 패닉으로 실패한 것은 아닙니다. 서로 다른 정상 결과가 요구사항과 어긋났습니다.

9개 입력에는 인코딩된 슬래시, 중첩 인코딩, 더하기 기호, 공백, opt-out, RawPath 없는 일반 이름과 리터럴 `%2F`, 매칭되지 않는 경로가 들어갑니다. URL 필드를 직접 지정한 유한 실험입니다. 네트워크 요청 파싱·정규식·마운트·잘못된 escape·동시성·전체 요청의 숨은 상태는 범위에 포함되지 않습니다.

## 실제로 무엇을 실행했나요?

Go1.27.1에서 고정된 Chi 소스와 원본 구현 5종을 로컬 CPU로 실행했습니다. 새로운 요청과 라우터를 매번 만들었고, 라우트와 NotFound 핸들러 안에서 두 파라미터를 기록했습니다. 기대값은 실행 전에 문자 그대로 작성하고, 구현과 따로 소스 검토했습니다.

45회 실행이 모두 정상 반환했습니다. 명시적으로 감싼 API 호출 533회가 모두 반환했고 오류·패닉은 없었습니다. 이 호출 수는 숨은 라이브러리 호출이나 필드 쓰기를 포함하지 않습니다. 전체 관측 JSON 80,431B를 2,575B gzip으로 보존했고 CRC·끝까지 디코딩·전체 JSON·호출 수를 검증했습니다. [원본 관측](route/OBSERVATIONS.actual.public.v1.json.gz)과 [기대값 비교](route/COMPARISON.actual.public.v1.json)를 제공합니다.

별도로 기존 비활성 FP32 주장 모델을 한 번 읽어, 기존 계산과 새 [Prepared API](../../../pkg/hintprepared/README.ko.md)를 비교했습니다. 모든 후보의 점수 비트와 순서가 정확히 같았습니다. 제품 입력 검증도 실제 실행에서 통과했습니다. 모델은 Laya 인코더가 아니라 기존 해시 특징 기반 소형 주장 가중치입니다. 이 단계에서 Laya/MPS/GPU 실행이나 학습은 하지 않았습니다.

| 순서 제안 방식 | 첫 후보 | 9/9 후보의 순위 |
|---|---|---:|
| 기존 FP32 모델 | pre_path | 5 |
| 새 Prepared API + 같은 가중치 | pre_path | 5 |
| BM25 | post_path | 1 |
| 순서를 반영한 lexical control | post_path | 1 |
| 고정 입력 순서 | post_path | 1 |

고정 순서는 대조군입니다. narrow_rule은 일반 문장 요청을 지원하지 않아 `unsupported_rule_request`로 BM25에 fallback했습니다. 규칙 모델의 별도 성공으로 세지 않습니다. 순위 5와 1을 실제 LLM 호출 수나 작업 완료 시간으로 환산하지 않았습니다. [전체 점수·순서](preview/OBSERVATIONS.actual.public.v1.json)와 [fallback을 포함한 비교](preview/COMPARISON.actual.public.v2.json)를 참고하세요.

## 어떻게 재현하나요?

저장소 루트에서 Go1.27.1로 다음을 실행합니다.

```sh
go run ./experiments/short-claim/next-cohort-chi-audit/replay
```

다른 Go 실행 파일이나 작업 디렉터리를 사용한다면 `-go`와 `-root`를 지정할 수 있습니다. 검증기는 소스·고지·입력 지문을 확인하고 임시 디렉터리에서 컴파일합니다. 공개된 Chi 관측과 기대값을 다시 비교하고 임시 실행 파일을 정리합니다. 모델 다운로드나 인증은 필요하지 않습니다. 부모 1개·45회·명시 API533회·9/8/8/8/6 결과를 출력합니다.

로컬에서 이 명령을 실제 실행해 전체 관측이 일치했고, 종료 후 새 임시 디렉터리가 남지 않았습니다. [완료 기록](replay/EXECUTION.actual.public.v2.json)을 제공합니다. CI에도 같은 CPU 재현 단계를 추가했습니다. CI의 실행·통과 여부는 해당 PR의 검사 상태로 확인합니다. CI 통과는 코드 검사와 이 유한 계약의 재현을 뜻합니다. 모델의 일반화 성능, 학습 데이터 편입, 기본 모델 활성화는 각각 별도 기준입니다. 저장소의 기존 Laya 검사와 이 Chi 재현은 실행 대상을 구분합니다.

첫 재현 시도는 검증기의 임시 폴더 설정 때문에 Go가 모듈을 무시해 실패했습니다. 수정 검토 중에는 버퍼의 `ReadFrom`이 출력 제한을 우회할 수 있는 문제도 발견했습니다. 모듈과 컴파일 임시 폴더를 분리하고, 버퍼를 명명된 필드로 바꾸었습니다. 두 복사 경로의 정확한 한도·초과 한도를 실제 검사했습니다. [첫 소스](replay/FIRST-REPLAY-FAILURE.source.go.txt)와 [실패 기록](replay/EXECUTION.actual.public.v1.json)은 보존했습니다. Chi 입력·기대값·모델 비교를 수정하거나 성공으로 바꾸지 않았습니다.

## 비용과 메모리는 어떻게 읽나요?

실행 기록에 Darwin의 전체 프로세스 시간과 최대 resident-size 원시 보고값을 보존했습니다. 시작·파일 읽기·라우팅 또는 모델 계산·JSON/gzip 출력을 포함합니다. 단위 환산을 하지 않았고, 서로 다른 두 프로세스의 값을 모델 메모리 절감률로 비교하지 않습니다. 소수점 두 자리 CPU 시간이 0.00이어도 CPU 작업이 없었다는 뜻은 아닙니다. GPU·MPS·pprof 측정은 이번 실행에 없습니다.

GOMAXPROCS와 Go heap의 soft target을 제한했습니다. GOMEMLIMIT는 RSS나 전체 시스템 메모리의 상한이 아닙니다. 임시 컴파일 파일은 정리했고 보유 바이너리는 없습니다. 기록의 기존 512MiB 논리 파일 예산을 유지하면서, 종료된 공개 자료 사본은 원본 바이트를 복구할 수 있게 압축했습니다. 이 파일 예산은 RAM 한도와 다릅니다.

## 표본과 다음 학습 기준

**부모 사례는 1개입니다.** 9개 입력·5개 구현·45회 실행·추천 방식 6종을 독립 Golden 45개나 새 독립 부모로 세지 않습니다. 이미 소스와 결과를 본 개발용 감사 사례이며, 보호된 최종 평가셋으로 옮기지 않습니다.

이번 PDCA에서 확인한 개선 과제는 인코딩 순서와 guard를 문장 겹침보다 잘 구별하는 것입니다. 학습 전에는 코드와 설명을 대응시키고, 관계가 있는 부모·자식·helper를 같은 그룹으로 묶어 역할을 검토합니다. 다른 출처의 검토된 부모를 늘린 뒤 개발셋에서 튜닝합니다. 개선 기준에는 BM25 대비 올바른 후보를 찾기까지의 실제 검증 비용이 들어갑니다. 독립2400 평가와 전체 작업/LLM 비용 절감은 아직 입증되지 않았습니다.

설명의 정규화 단어 수 제한도 발견했습니다. 초안은 punctuation과 camel 이름 때문에 40/40/41/40/34단어였고 실제 제품 한도32를 넘었습니다. 실행 전에 별도 v2·v3를 만들고 guard·Path 대입·fallback을 다시 검토했습니다. 원본 입력은 보존했고 자르거나 학습 라벨을 바꾸지 않았습니다. 최종 요청31, 후보27/27/32/27/24단어로 실제 입력 검증을 통과했습니다.

## 출처와 라이선스

[go-chi/chi 고정 소스](https://github.com/go-chi/chi/tree/167e1e3bd039d060696b99c8da4e876ae04f42c1)는 MIT이며 전체 고지를 보존합니다. tree.go의 Armon Dadgar radix 출처 표시와 [Armon의 MIT 고지](source/ARMON-NOTICE.txt)도 함께 보존합니다. Armon revision은 고지를 확인한 비교 지점입니다. Chi의 실제 역사적 선행 revision과 복사 범위를 확정했다는 뜻은 아닙니다. 관측용 코드·직접 작성한 입력과 설명은 저장소 Apache-2.0을 따릅니다. 자세한 구분은 [NOTICE](NOTICE)에 있습니다.

[모델 본체의 고정 HF revision](https://huggingface.co/JooYoon/riidolaya-shortclaim-data-effect-failed-79/tree/5bef215895b69d3f2ef4b82bb3f1970279d67f46)을 참조하며 모델 본체·비공개 locator·프로파일·원시 실행 journal은 GitHub에 포함하지 않습니다. [PACKET](PACKET.public.v1.json), 두 실행 freeze와 SHA256SUMS가 공개 파일과 지문을 연결합니다. Freeze의 “실행 전/pending”은 당시 상태이고 실제 완료는 EXECUTION/COMPARISON에 기록합니다.

압축을 푼 증거의 비밀정보 검사에서는 공개 Go 파일명 `dir_plan9.go`가 오탐으로 걸렸습니다. 정확한 파일명과 고정 입력을 확인한 뒤 그 검사에만 한정한 임시 설정으로 다시 검사했습니다. 저장소의 검사 규칙은 그대로 유지했고 민감정보는 발견되지 않았습니다. [첫 검사](replay/SCAN-DECOMPRESSED-FIRST-FAILURE.actual.public.v1.json)와 [판정·재검사](replay/SCAN-DECOMPRESSED-DISPOSITION.actual.public.v1.json)를 보존합니다.

같은 입력의 [준비 재사용 비용과 별도 Go 프로파일](setup-reuse/README.ko.md)을 측정했습니다. N=8에서 준비 재사용은 재생성보다7.2~7.6배 빨랐지만 BM25가 더 빠르고 모델은 비활성입니다.
