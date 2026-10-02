# 공개 Go 원천을 추가하는 다음 개발 준비안

현재 확인한 것은 **새 저장소 7개와 독립 의미의 행동 목표 제안 11개**입니다. 아직 새로운 부모 요청·정답·실행 완료 계약은 **0개**입니다. 원본 Go 함수 실행, 학습, 모델 호출, 보호된 final 자료 열람도 0회입니다. 기존 72/76 개발 자료, 정답, 역할, 마스크, provenance와 audit 신호는 바꾸지 않았습니다.

이번 자료는 [기계가 읽는 제안 목록](NEXT-DEVELOPMENT-SOURCES.proposal.v1.json)에 대응합니다. 원문 전체를 복제하지 않고 공식 GitHub의 고정 revision, 라이선스 원문 SHA, 최소 패키지 파일과 다음에 검증할 조건을 기록했습니다. **우리의 정답과 라벨은 아직 제안 상태**입니다.

## 어떤 원천을 확보했나

아래 파일 수는 원래 비테스트 Go 패키지 파일과 라이선스·모듈 자산을 합친 정적 후보입니다. 읽기만 했으며 빌드·오프라인 실행 성공을 뜻하지 않습니다. 각 closure는 현재 32파일 한도 안입니다. README와 읽은 test-import 근거는 closure와 분리했습니다.

| 원천 / 고정 revision | 라이선스 원문 | 비테스트 Go / closure 파일 | 서로 다른 목표 제안 |
|---|---|---:|---|
| [joho/godotenv](https://github.com/joho/godotenv/tree/97a2850142438b3c357c1d3439e325181afdcc1d) | [MIT, LICENCE](https://github.com/joho/godotenv/blob/97a2850142438b3c357c1d3439e325181afdcc1d/LICENCE) | 2 / 4 | 명시 값 파싱; 정렬·이스케이프 직렬화 |
| [google/go-querystring](https://github.com/google/go-querystring/tree/965d79f2113ea0ff039d29828a08a616a0223d48) | [BSD-3-Clause, LICENSE](https://github.com/google/go-querystring/blob/965d79f2113ea0ff039d29828a08a616a0223d48/LICENSE) | 1 / 4 | 태그가 있는 struct의 다중 URL 값 인코딩 |
| [google/shlex](https://github.com/google/shlex/tree/e7afc7fbc51079733e9468cdfd1efcd7d196cd1d) | [Apache-2.0, COPYING](https://github.com/google/shlex/blob/e7afc7fbc51079733e9468cdfd1efcd7d196cd1d/COPYING) | 1 / 3 | 인용·이스케이프·EOF를 포함한 단어 분리 |
| [go-logfmt/logfmt](https://github.com/go-logfmt/logfmt/tree/804e98fff868b206344991c57a8182172e5ba41e) | [MIT, LICENSE](https://github.com/go-logfmt/logfmt/blob/804e98fff868b206344991c57a8182172e5ba41e/LICENSE) | 4 / 7 | 레코드별 디코딩; 한 key/value의 인코더 상태 |
| [c2h5oh/datasize](https://github.com/c2h5oh/datasize/tree/aa82cc1e65004e2b59a6e44d26f774ca961b24d8) | [MIT, LICENSE](https://github.com/c2h5oh/datasize/blob/aa82cc1e65004e2b59a6e44d26f774ca961b24d8/LICENSE) | 1 / 3 | 단위 파싱과 오류 상태; 나누어떨어지는 단위 출력 |
| [mitchellh/go-wordwrap](https://github.com/mitchellh/go-wordwrap/tree/ecf0936a077a4bd73a1cc2ac5c370f2b55618d62) | [MIT, LICENSE.md](https://github.com/mitchellh/go-wordwrap/blob/ecf0936a077a4bd73a1cc2ac5c370f2b55618d62/LICENSE.md) | 1 / 3 | 공백·문자 폭을 고려한 줄바꿈 |
| [vincent-petithory/dataurl](https://github.com/vincent-petithory/dataurl/tree/d1553a71de50473073e188aa79cebf7f993f20fe) | [MIT, LICENSE](https://github.com/vincent-petithory/dataurl/blob/d1553a71de50473073e188aa79cebf7f993f20fe/LICENSE) | 4 / 5 | 미디어/바이트 디코딩; percent 바이트 변환 |

기존 historical acquisition의 8개 저장소와 humanize/UUID/semver/glob 원천은 이번 후보에 다시 넣지 않았습니다. 다만 저장소 이름이 다르다는 사실만으로 저자 흐름이나 공유 코어의 독립성이 입증되지는 않습니다.

## 우선 3개를 실제 계약으로 만든다

일반적인 설명을 그대로 정답으로 삼지 않습니다. 완전한 짧은 요청, 후보의 정확한 범위, 독립 작성한 유한 관측값을 먼저 묶고 다음 별도 실행 계획에서 원본 결과와 비교합니다.

| 우선 목표 / 직접 호출할 API | 독립 관측 계약 준비 | 어려운 오답 후보 |
|---|---|---|
| [datasize: UnmarshalText](https://github.com/c2h5oh/datasize/blob/aa82cc1e65004e2b59a6e44d26f774ca961b24d8/datasize.go#L116) | uint64 값, receiver의 실패 후 상태, NumError의 종류/원인. 큰 정수 oracle와 고정 리터럴을 독립 작성 | 1000 기반 단위, overflow wrap, 모든 오류에서 이전 값 보존, bits를 bytes로 허용 |
| [query: Values](https://github.com/google/go-querystring/blob/965d79f2113ea0ff039d29828a08a616a0223d48/query/encode.go#L125) | 고정 primitive struct의 키 부재/빈 값, 반복 값 순서, nested 이름. 임의 callback은 받지 않음 | 마지막 값만 남기기, 태그와 무관한 zero 생략, nested 이름 평탄화 |
| [shlex: Split](https://github.com/google/shlex/blob/e7afc7fbc51079733e9468cdfd1efcd7d196cd1d/shlex.go#L403) | 정확한 token 배열, 빈 quoted word, 뒤쪽 EOF 오류와 이미 완료한 앞 token | strings.Fields, 빈 token 제거, 오류 때 앞 결과까지 삭제 |

예를 들어 “크기 문자열을 파싱한다”는 말만으로 overflow의 성공/실패/포화 상태가 정해지지 않습니다. “셸처럼 나눈다”는 말도 실제 명령 실행이나 완전한 POSIX 문법을 포함하지 않습니다. 이 차이를 요청과 유한 truth에 모두 공개해야 합니다. 기존 512바이트/정규화 32단어 입력 계약은 이후 검증하며, 이번 초안이 이미 통과했다고 주장하지 않습니다.

**no_answer**는 지원되는 명확한 요청에 대해 제공한 후보 모두가 독립 계약을 위반하는 경우입니다. **unknown**은 문장이 모호하거나 원천/관측기가 요구를 지원하지 않아 판단할 수 없는 경우입니다. 둘을 합치지 않습니다. 오답 wrapper를 만들거나 숫자를 바꾸는 일은 새 독립 부모 요청을 만드는 일이 아닙니다.

## 초기 실행 범위와 권리의 한계

- godotenv의 보간은 입력에 없는 변수를 process environment에서 찾을 수 있습니다. 처음은 `$` 없는 명시 값으로 좁히고 `autoload`, 파일 접근과 `Exec`을 제외합니다. 구문 오류의 partial map도 보존하며 atomic-zero 정책을 추측하지 않습니다.
- query와 logfmt의 비테스트 소스 import는 표준 라이브러리지만 원래 모듈/테스트는 go-cmp를 요구합니다. 원래 go.mod와 언어 버전은 그대로 pin하고, 실행 전 별도 offline recipe를 확인해야 합니다. datasize는 원래 Go directive가 없고, dataurl은 원래 go.mod가 없습니다.
- logfmt는 다음 record에서 내부 slice가 바뀔 수 있어 관측기가 즉시 복사해야 합니다. 인코더의 validation 실패와 writer 실패를 구분합니다. wordwrap의 긴 단어는 폭을 넘을 수 있으므로 무조건 모든 줄의 최대 폭을 보장하는 요구로 바꾸지 않습니다.
- dataurl에는 Go `net/url` 및 `text/template/parser` 차용 주석이 있습니다. 추가 BSD 출처/고지 범위는 미해결이고 lexer goroutine의 오류 종료도 아직 실행 검증하지 않았습니다. root MIT 표기만으로 추출·학습 권리를 전부 승인하지 않습니다.
- 원본 라이선스와 파일의 저작권 표기를 보존합니다. 원래 없는 NOTICE를 upstream 파일로 만들지 않습니다. 코드의 재사용 조건, 우리 caption/literal의 작성 권리, 향후 dataset/weights의 라이선스는 각각 확인합니다.

## 개수와 분할은 어떻게 늘리나

한 저장소의 revision, 함수 별칭, wrapper, 공유 helper, nearvariant, 번역은 whole-family로 연결합니다. 다른 저장소의 복사된 semantic core도 분할 전에 전이적으로 합칩니다. 단순한 표준 라이브러리 관측 기반과 복사된 행동 helper는 구분해 근거를 남깁니다. 원본 개발자와 우리 AI-assisted English caption/oracle 작성 흐름도 따로 기록합니다. 한영 설명 문서는 한국어 모델 테스트가 아닙니다.

먼저 위 3개를 검증 가능한 개발 계약으로 전진시킨 뒤 나머지 8개 목표를 검토합니다. 이후 다른 실제 목표와 원천을 확보해 **약 128 → 256 → 512개의 독립 의미가 있는 development 문제**를 향합니다. 현재 11개 목표의 리터럴 반복만으로 128개를 채울 수 없습니다. 이 숫자를 모든 development fit의 새 최소 조건으로 만들지도 않습니다.

**도메인별 2,400개 이상의 새 protected final 요청**은 별도로 확보하고 보호할 목표입니다. 공개 개발 자료를 final로 다시 세거나, 7개 repo 그룹만으로 train/validation/final의 가족 다양성을 충족했다고 주장하지 않습니다. source/group/role/truth 정보는 provenance와 supervision에만 두고 scorer 입력에는 요청과 후보 문장만 투영합니다.

이번에 수행한 일은 공식 원천 읽기와 cached bytes의 SHA·경로/줄 범위 확인입니다. 원 API·benchmark·새 모델/학습·원격 게시·CI 실행은 하지 않았습니다. 유지보수용 읽기 자체와 이 AI 협업의 비용을 무료라고 측정한 것은 아닙니다.
