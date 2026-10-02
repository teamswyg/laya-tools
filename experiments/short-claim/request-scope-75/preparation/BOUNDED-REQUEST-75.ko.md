# 행동 범위를 드러내는 요청·설명 준비 75

이 자료는 같은 세 부모 요청과 같은 순서의 아홉 코드 후보에 대한 **새 요청 버전과 학습 자격 제안**입니다. 실제 정답·역할·가중치·pair 자격은 모두 `null`이며, 기존 코퍼스에 넣거나 학습하지 않았습니다. 원래 candidate v2, caption74, readiness 파일과 이번 준비의 v1/v2도 그대로 보존했습니다. 번역과 A/B 설명은 새로운 부모 요청으로 세지 않습니다.

## 모델에 보이는 요청

| 부모 | 영어 요청 | 바이트 / 자체 단어 계산 |
| --- | --- | --- |
| datasize | Find the nonnil receiver method parsing ASCII integers and binary byte units: return syntax or bits errors with receiver zero, and range errors with receiver uint64 maximum. | 173 / 27 |
| query | Find struct query function: exported nonignored default fields retain duplicate primitive slice order; omit empty slices, keep empty strings without omitempty; included named nonnil non-time structs use brackets; exclude custom encoders. | 237 / 32 |
| shlex | Find the string splitter that keeps empty quoted words and returns completed tokens before an unclosed quote or trailing escape error. | 134 / 21 |

앞의 두 요청은 별도 범위 버전이고 shlex 요청은 원문 그대로입니다. 512바이트·32단어 제한은 이 ASCII 자료에 대한 자체 식별자 분리 계산으로 점검했습니다. 원래 Normalize/Validate/Features는 호출하지 않았습니다.

query의 `included`는 중요한 조건입니다. 원본은 `omitempty`와 `IsZero`의 생략 판단을 **중첩 재귀보다 먼저** 합니다. 따라서 대괄호 주장은 실제 포함되는 이름 있는 nonnil·non-time struct에만 적용합니다. anonymous embedded struct의 평탄화, nil pointer, time 값의 특별 처리, custom Encoder 구현은 이 중첩 주장의 범위 밖입니다. 기본 primitive 필드·slice 표현만 다루며, delimiter나 다른 옵션의 일반적 동작까지 주장하지 않습니다.

## 코드의 동작과 설명의 정보는 다릅니다

고정 원본을 읽었을 때의 제안은 `UnmarshalText`, `Values`, `Split`이 각 요청을 충족하고 나머지 여섯 후보는 직접 선택된 API의 형태나 정책에서 충족하지 않는다는 것입니다. `Encoder`는 인터페이스 선언이며 실제 구현 함수가 아닙니다. 보이지 않는 구현이나 어댑터를 가정하지 않았습니다. `Parse`·`MustParse`의 위임, `Next`와 `Split`의 단일 반환/누적 목록 차이는 원본 wrapper 본문으로 확인 가능한 근거입니다. 추가 API 실행이나 모든 입력에 대한 형식적 증명은 아닙니다.

A는 기존 짧은 설명, B는 고정 caption74 설명입니다. 각 요청 조건에 `supported`, `contradicted`, `omitted`, `unknown` 근거를 별도로 둡니다. 코드가 실제로 맞는다는 근거를 모델에 보이지 않는 설명 정보로 채우지 않습니다.

| 제안 | A | B | A/B 공통 교집합 |
| --- | ---: | ---: | ---: |
| BCE 학습에 충분한 후보 설명 | 4 | 9 | 4 |
| 그중 긍정 후보 | 0 | 3 | 0 |
| 충분한 긍정·부정 endpoint 쌍 | 0 | 6 | 0 |

A의 네 후보는 설명에 결정적 불충족 근거가 있는 부정 후보뿐입니다. 예를 들어 receiver 없는 `Parse`, 오류 대신 panic하는 `MustParse`, 인터페이스라는 `Encoder`, 단어 하나를 반환한다는 `Lexer.Next`입니다. 긍정 후보의 서명이나 일반적인 한 줄 소개만으로는 요구한 오류 상태·중복 순서·빈 따옴표·완성 접두 목록까지 알 수 없습니다. B는 필요한 긍정 행동이나 결정적 불충족 정보를 드러낸다는 **검토 대기 가설**입니다.

## BCE 마스크만으로 pair 학습이 차단되지는 않습니다

기존 pair-ranking72는 known 부모에서 BCE 가중치가 0인 endpoint도 쌍에 사용할 수 있습니다. 이번 제안은 BCE 가중치와 pair endpoint 자격을 별도 필드로 기록합니다. 현재 코드에 이 자격의 강제 적용은 없습니다. 향후 사용 전 caller/pair plan이 충분하지 않은 endpoint를 명시적으로 제외하거나, 보수적으로 해당 부모의 pair 학습을 제외해야 합니다. 기존72 사양·수치·결과를 바꾸거나 소급 무효화하지 않습니다.

설명만의 효과를 비교하려면 A/B에 같은 고정 마스크와 가중치를 써야 합니다. 이 묶음의 공통 교집합에는 긍정 후보가 없어 **학습 가능한 paired 비교 자료라고 부를 수 없습니다**. A/B마다 별도 마스크를 쓰면 설명 정보와 학습 범위가 함께 바뀌는 다른 실험입니다. 원래 코드 정답·허용 후보·no_answer·unknown 평가 분모는 어느 경우에도 삭제하지 않습니다.

## 원천과 인계

원천은 [datasize](https://github.com/c2h5oh/datasize/blob/aa82cc1e65004e2b59a6e44d26f774ca961b24d8/datasize.go) ([MIT](https://github.com/c2h5oh/datasize/blob/aa82cc1e65004e2b59a6e44d26f774ca961b24d8/LICENSE)), [go-querystring](https://github.com/google/go-querystring/blob/965d79f2113ea0ff039d29828a08a616a0223d48/query/encode.go) ([BSD 3-Clause](https://github.com/google/go-querystring/blob/965d79f2113ea0ff039d29828a08a616a0223d48/LICENSE)), [shlex](https://github.com/google/shlex/blob/e7afc7fbc51079733e9468cdfd1efcd7d196cd1d/shlex.go) ([Apache 2.0](https://github.com/google/shlex/blob/e7afc7fbc51079733e9468cdfd1efcd7d196cd1d/COPYING))의 고정 revision입니다. 기존 원문 라이선스·저작권 고지 핀을 그대로 참조하며 원천 재라이선스를 주장하지 않습니다.

새 JSON은 파일 원문 SHA 3개, 증거 span 26개, 후보 선언 span 9개와 기존 ID·순서·설명 SHA를 대조한 stdlib metadata 도구로 준비했습니다. 실제 source AST·원 API·init·관찰기·역할·모델·학습·보호 final·원격 실행은 0회입니다. 도구의 생성 장부는 버전별 3회 성공·0회 실패이며 source scope 보완의 이력을 남깁니다.

작성자는 AI-assisted source 독자이며 이전 Want와 결과에 노출됐습니다. 비공개·미관측·독립 원천 작성 흐름 또는 의미 판단의 완전성을 주장하지 않습니다. 세 repo 가족과 alias/helper는 전체로 유지합니다. 2400/domain 최종 목표, 데이터 다양성 한계, 기존 학습 준비 조건은 변하지 않습니다. 후속 검토가 실제 코퍼스 편입 여부를 결정하며 이 파일 생성이나 CI 성공 자체가 학습 자격을 만들지는 않습니다.
