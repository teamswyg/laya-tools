# 실험 48: 학습 구간 helper 여유 검사를 마치고 fitting 중단

고정한 새 helper 세 가지 모두 **20개 저장소의 적격 학습 요청 4,456개**에서 변경하지 않은 **5% 페이지 절감 필요조건**을 실패했습니다. 정답을 전부 알고 이득인 요청에서만 helper를 선택하는 oracle조차 5%에 도달하지 못했습니다. 이 helper들의 추가 claim fitting은 여기서 중단합니다. 새 validation 또는 보호된 final 결과를 채점하지 않았고, 모델 호출·계수 fitting·가중치 생성·정책 활성화도 하지 않았습니다. 실험 46의 실패 결론은 유지합니다.

[사전 커밋한 계획](plan-48.json)은 세 규칙과 학습 구간만 검사하는 범위를 고정합니다. [집계 결과](results-48.json)에 모든 정책과 저장소별 수치가 있고, [자원 관측 기록](resources-48.json)에 실제 오프라인 실행 두 번의 범위가 있습니다.

페이지는 **첫 번째 매핑된 정답 파일을 만날 때까지 후보 20개씩 읽는 페이지 수**로, 파일 검색 비용의 대리 지표입니다. 순절감은 `(baseline 페이지 − helper 페이지) / baseline 페이지`입니다. Oracle 여유는 `sum(max(0, baseline 페이지 − helper 페이지)) / baseline 페이지`로, 손해인 전환은 모두 피하고 정답 label을 아는 낙관적 진단입니다. 실행 가능한 라우팅 정책이 아니며 다른 품질·호출·저장소 제약도 무시합니다. **Baseline 43,914페이지**에서 5%를 충족하려면 최소 **2,196페이지 절감**이 필요합니다. 이 학습 검사를 통과했더라도 별도로 사전 계획한 다음 실험의 근거일 뿐, validation 유용성을 입증하지는 못합니다.

| 고정 정책 | 페이지 | 순절감 페이지 | 순절감률 | Oracle 절감 페이지 | Oracle 여유 | 승 / 패 / 동률 |
|---|---:|---:|---:|---:|---:|---:|
| Baseline | 43,914 | 0 | 0% | 0 | 0% | 0 / 0 / 4,456 |
| `normalized-positive64` | 42,816 | 1,098 | 2.5003% | 1,357 | 3.0901% | 125 / 255 / 4,076 |
| `joined-fields-rrf64` | 42,823 | 1,091 | 2.4844% | 1,615 | 3.6776% | 217 / 514 / 3,725 |
| `explicit-path64` | 43,365 | 549 | 1.2502% | 598 | 1.3618% | 86 / 37 / 4,333 |
| 기존 대조군 `legacy-normalized-full` | 41,894 | 2,020 | 4.5999% | 3,030 | 6.8998% | 184 / 550 / 3,722 |

기존 대조군의 **학습 oracle 6.8998%**는 [실험 46의 validation 실패](../path-cost-claim/RESULTS-46.ko.md)를 뒤집지 않습니다. 실제 학습 순절감은 4.5999%이고, oracle 상한만으로 쓸 만한 claim 선택기를 입증할 수 없습니다. 실험 46은 이미 validation oracle 상한 3.113%를 관측했고 고정된 유용성 관문을 실패했습니다. 이번 검사는 그 관문을 조정하거나 새 validation 승자를 고르지 않습니다.

세 규칙은 정답 label을 보지 않고 경로를 정렬하며, label은 이후 비용 집계에만 사용합니다. 문장을 생성하지 않고 파일 힌트를 작은 목록으로 제공합니다.

- `normalized-positive64`: 기존 식별자 분할과 전체 경로 BM25 정렬을 사용합니다. 점수가 양수인 경로를 최대 64개 남겨 baseline 후보를 먼저 두고 helper 후보와 교차 배치합니다.
- `joined-fields-rrf64`: 분할 식별자와 함께 소문자로 붙인 식별자도 유지합니다. 전체 경로와 파일 이름을 따로 정렬한 뒤, 각 필드의 양수 결과에 고정 점수 `1/(60 + 1부터 시작하는 순위)`를 더합니다. 결정적인 동률 처리로 합친 힌트를 최대 64개 남기고 같은 baseline 우선 교차 배치를 사용합니다.
- `explicit-path64`: slash 또는 dot이 들어간 완전한 query 토큰이 대소문자까지 정확히 catalog 경로나 slash 경계의 접미 경로와 맞는지 확인합니다. 맞는 경로 구간 수가 많을수록 우선하며, 다음은 토큰 길이와 catalog 순서입니다. 정확한 근거가 없으면 baseline을 유지합니다. 힌트는 최대 64개이며 BM25 index를 사용하지 않습니다.

품질 수치가 여유 검사 실패를 바꾸지는 않습니다. 저장소를 같은 비중으로 평균한 Hit10은 baseline **42.1142%**, positive64 **42.2723%**, joined RRF **41.4535%**, explicit path **44.2129%**입니다. Explicit path는 이 학습 지표를 높였지만 페이지 절감 여유가 부족하고, joined RRF는 지표가 낮아졌습니다. 개발 집계이며 작업 완료나 LLM 사용량 측정은 아닙니다.

**고정 학습 구성원 7,335개**를 coverage에 모두 유지했습니다. **적격 4,456개**, **source 대기 221개**, **사용 불가 2,630개**, **비용 무효 28개**입니다. 제외 구성원을 다른 요청으로 대체하거나 coverage 분모에서 숨기지 않았습니다. 평가 표본은 반복 실행이나 정책 변형 수가 아닌 요청 4,456개입니다. 새 validation 순위나 점수를 계산하지 않았고, **보호된 final 구성원 2,402개는 채점하지 않았습니다**. 봉인된 입력에 남아 있는 이전 validation 메타데이터는 새 validation 결과가 아닙니다.

새 helper는 각각 요청 4,456개 모두에서 시도했고 **fallback은 0건**입니다. Positive64는 힌트 274,050개, index build 3,568회, index search 4,456회입니다. Joined RRF는 힌트 274,332개, 필드 index build 7,136회, 필드 search 8,912회입니다. Explicit path는 힌트 6,088개와 anchor 비교 134,445,953회이며 index build/search는 없습니다. 공용 catalog cache는 **build 3,568회와 hit 888회**를 기록했습니다. 힌트 수가 적거나 Hit10이 높다는 사실만으로 helper 실행 회피나 후속 비용 절감을 측정한 것은 아닙니다.

독립 작성한 Go 구현은 호출자가 소유한 배열과 불변 index를 사용하며 query lock이 없습니다. 크기가 제한된 구성용 map은 남아 있으므로 전체 구현에서 map을 제거했다는 주장은 아닙니다. 완전한 UTF-8 query는 128 KiB, 고유한 정규 경로 catalog는 100,000개 경로 및 16 MiB로 제한합니다. 각 대응 text 필드는 16 MiB, 어휘 토큰 262,144개, vocabulary 131,072개가 상한입니다. Explicit matching은 비교 전에 고유 anchor 수 × 경로 수가 8,000,000 이하인지 확인합니다. Helper 한도나 오류에서는 이유를 기록하고 baseline으로 복귀하며, 입력을 잘라 쓰거나 요청을 버리지 않습니다. SIMD 성능 이득은 측정하거나 주장하지 않았습니다.

| 실제 전체 오프라인 관측 | 경과 시간(초) | 사용자 CPU 시간(초) | 시스템 CPU 시간(초) | 최대 RSS(bytes) | 최대 RSS(MiB) |
|---|---:|---:|---:|---:|---:|
| 첫 실행 | 82.84 | 89.67 | 2.64 | 111,902,720 | 106.71875 |
| 정확 재실행 | 82.30 | 90.25 | 2.46 | 112,492,544 | 107.28125 |

두 관측은 **오프라인 source/cache 검증, index 구성, 요청 4,456개 검사 전체**를 포함하며 **최대 RSS 256 MiB 목표**를 충족합니다. 전체 실행 두 번의 관측으로, 단일 요청 추론 지연·p95 추정·Go heap pprof 측정이 아닙니다. GPU는 사용하지 않았습니다. LLM 호출·token·retry·실제 helper 실행 생략·금전 절감은 측정하지 않았습니다.

실제 재실행의 **private case 출력과 공개 집계 출력 비교는 각각 exit 0**으로 끝나 두 비교에서 파일이 byte-identical함을 확인했습니다. 반복 실행은 재현성 검증이며 독립 표본을 추가하지 않습니다. Private case 본문, source 본문, 경로, query, 정답 label은 이 보고서에 포함하지 않았습니다.

| 증거 pin | 값 |
|---|---|
| Clean runner revision | `828d2c8dec00d144de7ddec96c27b6e20fe74311` |
| Runner binary SHA256 | `c211570fbf7aa64e58488105b6013848ba66eac158cdaf19fad861599d220da8` |
| Plan SHA256 | `f7e740f0ff0da3706592f79d30d240779d8e1261504ef4f4f4dec937904e61fd` |
| Input seal SHA256 | `3e8793cb0f3c5d23473fedfd3d4faed58c05dbd3949e77308ca1f8ee56f39187` |
| 집계 결과 SHA256 | `e23d6a76cdee999f119471882d31ea8170ea281c36fa60f566d847efbb7c7b4a` |
| Rankings SHA256 | `63f0f8ad72f3753736f34755054422d160b16eaf8205259c2534d799fa14ef5d` |

설계는 Apache-2.0 [laya-codex의 고정 revision `580bc73c2ed95fd319db93ef725f30bf35047428`](https://github.com/pilotspace/laya-codex/tree/580bc73c2ed95fd319db93ef725f30bf35047428)의 식별자·검색 개념과 [2009년 원본 reciprocal-rank-fusion 논문](https://plg.uwaterloo.ca/~gvcormac/cormacksigir09-rrf.pdf)을 참고했습니다. Go 구현은 독립 작성했으며 upstream 코드나 GPL Spiral 코드·표를 복사하지 않았습니다. 숫자 source 근거는 source 본문·학습 데이터·모델 공개 허가를 뜻하지 않습니다. Ternary 압축은 별도의 후속 방향이며 helper 유용성 검사 실패를 고칠 수 없습니다.
