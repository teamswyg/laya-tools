# 다음 4개 공개 Go 원천: 실행 전 조사

이번에 확인한 것은 godotenv, logfmt, wordwrap, dataurl의 고정 원문·라이선스와 다음 행동 계약 후보입니다. 원 API 호출, 패키지 초기화·빌드·테스트, 모델 호출, 학습, 새 부모 요청·정답 확정은 모두 0입니다. 별도 stdlib Go 도구가 저장된 바이트만 한 번 검사하고 사본을 만들었습니다. **해시 확인은 의미 정답이나 학습 준비 완료의 증명이 아닙니다.**

[검사 영수증](SOURCE-READ-RECEIPT.v1.json)은 기존 inventory SHA `7c4e1029e8edf704b02d1f5bb7f67be855b82ab00653d60e817f68c4447b857f`를 묶습니다. 4개 폐쇄 원천에는 라이선스·모듈을 포함해 **19개 파일 / 11개 non-test Go 파일**이 있으며, 기존 inventory의 **7개 행동 목표**를 다룹니다. 원래 7개 repo의 11개 목표는 제안입니다. 리터럴·입력값·별칭·번역을 늘려 128개 과제라고 세지 않습니다. development 128→256→512에는 실제로 다른 행동 목표가 더 필요하며, 도메인별 보호 final ≥2,400개는 별도이고 읽지 않았습니다.

## 원본과 권리

| 원천 | 정확한 revision | root Go / 원본 closure 파일 | 모듈 및 라이선스 |
|---|---|---:|---|
| godotenv | `97a2850142438b3c357c1d3439e325181afdcc1d` | 2 / 4 | 원본 go1.13, MIT [LICENCE](https://github.com/joho/godotenv/blob/97a2850142438b3c357c1d3439e325181afdcc1d/LICENCE) |
| logfmt | `804e98fff868b206344991c57a8182172e5ba41e` | 4 / 7 | 원본 go1.21, MIT [LICENSE](https://github.com/go-logfmt/logfmt/blob/804e98fff868b206344991c57a8182172e5ba41e/LICENSE) + 아래 Go BSD 고지 |
| wordwrap | `ecf0936a077a4bd73a1cc2ac5c370f2b55618d62` | 1 / 3 | 원본 go1.14, MIT [LICENSE.md](https://github.com/mitchellh/go-wordwrap/blob/ecf0936a077a4bd73a1cc2ac5c370f2b55618d62/LICENSE.md) |
| dataurl | `d1553a71de50473073e188aa79cebf7f993f20fe` | 4 / 5 | 실제 go.mod/go.sum 없음, MIT [LICENSE](https://github.com/vincent-petithory/dataurl/blob/d1553a71de50473073e188aa79cebf7f993f20fe/LICENSE) + 아래 Go BSD 고지 |

고정 Git tree는 모두 비절단 상태였고, 19개 closure 파일의 크기·SHA256·Git blob SHA1과 root non-test Go 목록을 대조했습니다. README 4개도 기존 해시 그대로 보존했습니다. 원본 root NOTICE는 4개 모두 없습니다. godotenv/wordwrap에는 go.sum도 없습니다. logfmt의 go-cmp require는 원본 테스트용이며 이번에는 테스트를 실행하거나 내려받지 않았습니다. root 소스 import는 표준 라이브러리로 제한되어 있지만 실제 offline 빌드는 아직 하지 않았습니다.

MIT 고지를 그대로 보존하면 godotenv/wordwrap 원문을 재사용하는 경로가 있습니다. logfmt/dataurl도 root MIT만 보고 제외할 필요는 없지만, **빌리거나 수정한 Go 코드에는 별도 BSD 조건을 함께 지켜야 합니다.** [보충 출처 고지](attribution/NOTICE-Go.txt)와 [Go BSD 전문](attribution/Go-BSD-LICENSE)을 별도 보존했습니다. 이것은 우리가 작성한 보충 고지이며 upstream에 없던 NOTICE가 있었다고 주장하지 않습니다.

- logfmt의 [jsonstring.go](https://github.com/go-logfmt/logfmt/blob/804e98fff868b206344991c57a8182172e5ba41e/jsonstring.go#L13-L17)는 Go encoding/json을 수정했다고 명시하며 Go Authors 2010 헤더를 갖습니다. 최초 retained path commit `8b0437fa…`에도 헤더가 있습니다. `getu4` 원문은 고정 [Go JSON source](https://github.com/golang/go/blob/c303df658d43b9f3e98e56e646f8e84a83495991/src/encoding/json/decode.go#L953-L962)와 같습니다.
- dataurl의 [unhex](https://github.com/vincent-petithory/dataurl/blob/d1553a71de50473073e188aa79cebf7f993f20fe/rfc2396.go#L64-L75)는 net/url에서 빌렸다고 명시합니다. 고정 [Go URL source](https://github.com/golang/go/blob/402d3590b54e4a0df9fb51ed14b2999e85ce0b76/src/pkg/net/url/url.go#L38-L48)의 Go Authors 2009 헤더 및 동일 원문을 확인했습니다.
- dataurl의 [lexer](https://github.com/vincent-petithory/dataurl/blob/d1553a71de50473073e188aa79cebf7f993f20fe/lex.go#L140-L179)는 text/template 계보를 명시합니다. 고정 [Go lexer source](https://github.com/golang/go/blob/402d3590b54e4a0df9fb51ed14b2999e85ce0b76/src/pkg/text/template/parse/lex.go#L124-L137)는 Go Authors 2011 헤더를 갖고 `backup`/`ignore` 원문이 같습니다.

공식 Go 비교 revision은 `402d3590b54e4a0df9fb51ed14b2999e85ce0b76` 및 `c303df658d43b9f3e98e56e646f8e84a83495991`입니다. 양쪽 [LICENSE](https://raw.githubusercontent.com/golang/go/c303df658d43b9f3e98e56e646f8e84a83495991/LICENSE)는 1,479 B, SHA256 `dd26a7abddd02e2d0aba97805b31f248ef7835d9e10da289b22e3b8ab78b324d`, Copyright 2012 The Go Authors로 같습니다. **이는 공식 라이선스·비교 코드의 고정 출처이지 실제 복사 당시 release를 확인한 것은 아닙니다.** 수정 전체의 완전한 계보나 미확인 기여자의 권리를 추측하지 않습니다. 원본 MIT·관련 Go 저작권·BSD 조건·면책을 함께 배포하고 비추천 조건을 지키는 조건부 소스 재사용 경로이며, 데이터셋·학습된 가중치의 라이선스가 자동으로 확정되지는 않습니다.

## 다음에 만들 수 있는 계약: 모두 제안

아래는 새 정답표나 실행된 후보가 아닙니다. 각 목표의 완전한 짧은 문장, 독립 literal Want, observer, 오류·panic·nil/empty 관측, source/recipe를 먼저 고정해야 합니다. source reading에서 발견한 정책과 일반 라이브러리 설명을 섞어 정답으로 만들지 않습니다.

| 행동 목표 | 정확한 원천 / 필요한 의미 helper | 관찰할 계약과 어려운 오답 |
|---|---|---|
| `dotenv-explicit-values` | [godotenv.go:113](https://github.com/joho/godotenv/blob/97a2850142438b3c357c1d3439e325181afdcc1d/godotenv.go#L113), parser.go의 parseBytes/getStatementStart/locateKeyName/extractVarValue/expandEscapes/expandVariables, quote·space helpers, regex·sentinel | dollar 없는 입력의 CRLF·인용된 hash·single quote·중복 키·오류와 부분 map. 모든 hash 제거, single quote도 unescape, 오류 은폐를 오답 후보로 제안. 오류 시 전체 map zero/변경 없음 정책을 추가하지 않음. |
| `dotenv-canonical-marshal` | [godotenv.go:183](https://github.com/joho/godotenv/blob/97a2850142438b3c357c1d3439e325181afdcc1d/godotenv.go#L183), isInt/doubleQuoteEscape/doubleQuoteSpecialChars | 렌더링한 전체 행 정렬·정수 문자열 그대로·escape·마지막 LF 없음. int64 변환으로 넓은 수/leading zero 손실, map 순서 의존, escape 생략. 같은 라이브러리 roundtrip만으로 정답 정의 금지. |
| `logfmt-record-decoding` | [decode.go:54/71](https://github.com/go-logfmt/logfmt/blob/804e98fff868b206344991c57a8182172e5ba41e/decode.go#L54), Decoder/SyntaxError/Err/syntaxError/unexpectedByte, jsonstring.go의 unquoteBytes/getu4 | record/keyval 순서·bare key/빈 값·quoted escape·첫 syntax 오류 위치. key/value는 다음 record 전에 복사하여 nil/empty를 그대로 기록. 중복 map 덮어쓰기·escape 미처리·오류 무시를 후보로 제안. |
| `logfmt-single-keyval-state` | [encode.go:48](https://github.com/go-logfmt/logfmt/blob/804e98fff868b206344991c57a8182172e5ba41e/encode.go#L48), Encoder/needSep/EndRecord, writeKey/value helpers·sentinels·quote buffer | bytes.Buffer에서 primitive 입력의 출력/정확 오류/다음 성공 호출 separator. key의 invalid rune는 제거되고 제거 후 빈 key가 오류임. 잘못된 rune가 있으면 항상 거절, validation 실패에도 separator 변경, nil과 문자열 null 혼동을 후보로 제안. |
| `wordwrap-whitespace-layout` | [wordwrap.go:16](https://github.com/mitchellh/go-wordwrap/blob/ecf0936a077a4bd73a1cc2ac5c370f2b55618d62/wordwrap.go#L16), 전체 WrapString/nbsp | rune 수 기준·LF·NBSP·긴 단어·공백 위치의 정확 문자열. byte 폭, 모든 긴 단어 자르기, NBSP 분리를 후보로 제안. 보편적 최대 줄 폭·모든 공백 보존은 source 계약 아님. |
| `dataurl-percent-bytes` | [rfc2396.go:15/79](https://github.com/vincent-petithory/dataurl/blob/d1553a71de50473073e188aa79cebf7f993f20fe/rfc2396.go#L15), isUnreserved/isHex/unhex | byte별 uppercase percent 출력·plus는 literal·잘린/invalid hex·raw non-ASCII 오류·nil/empty. query decoding으로 plus→space, 부정확한 escape 집합을 후보로 제안. 독립 byte/hex literals를 사용. |
| `dataurl-decoded-media-bytes` | [dataurl.go:256](https://github.com/vincent-petithory/dataurl/blob/d1553a71de50473073e188aa79cebf7f993f20fe/dataurl.go#L256), DataURL/MediaType/defaultMediaType/parser.parse, readers, lexer item/state/constants, Unescape helpers | default media/charset·명시 media·parameter·base64/percent payload·nil/error. base64만 처리하거나 parameter 삭제하는 후보. 첫 실행 전에 lexer lifecycle·전체 4개 Go closure·새 모듈 recipe를 별도 고정. |

`no_answer`는 지원된 요청에서 제공 후보가 모두 독립 계약을 어긴 것이 관찰되었을 때의 빈 허용집합입니다. 환경 의존, 모호한 문장, 미관찰 panic/오류/lifecycle 또는 더 강한 임의 정책을 후보 실패로 바꾸지 않고 `unknown`으로 남깁니다.

특히 godotenv는 `expandVariables`가 환경을 읽으므로 첫 범위는 dollar-free로 제한합니다. `extractVarValue`의 comment-start 경로에는 `line[i-1]`가 있어 no-panic을 추측할 수 없습니다. Load/Read/Overload/Write/Exec/autoload는 제외합니다. logfmt는 bytes.Buffer와 primitive 입력만 먼저 다루며 failing writer·custom callback의 보편적 atomicity는 주장하지 않습니다. dataurl DecodeString은 unbuffered lexer goroutine을 시작하고 parser가 먼저 오류 반환할 수 있으므로, 조기 오류에서 goroutine 종료가 보장되는지는 아직 미확인입니다. percent helper가 더 작은 첫 실행 대상입니다. dataurl의 없는 go.mod를 upstream 원문으로 만들지 말고 별도 유지보수 recipe로 명시해야 합니다.

## 실제 다음 순서와 그룹

1. wordwrap 및 godotenv Marshal의 작은 계약·독립 Want·bounded observer를 준비합니다.
2. logfmt는 보충 Go BSD 고지를 포함한 폐쇄 소스와 ordered state/error 관측을 준비합니다.
3. dataurl은 percent helper부터, DecodeString은 lexer 종료 관측·새 모듈 recipe를 준비한 뒤 별도 범위로 다룹니다.

전체 repo 계보·revision·alias·wrapper·동일 helper·nearvariant·번역은 함께 묶습니다. 다른 repo라도 실제 복사된 의미 core를 공유하면 합쳐야 합니다. 빌린 Go helper는 행동의 일부라 일반 stdlib observer 인프라라는 이유로 제거하지 않습니다. 같은 Go Authors 명칭 또는 단순 표준 라이브러리 사용만으로 모든 원천을 자동 합치지도 않습니다. godotenv의 Ruby dotenv 포트 설명, wordwrap와 기존 mapstructure의 Mitchell 작성자 계보, logfmt/dataurl의 Go 원형 계보를 별도 보존합니다. 원 소스 작성 흐름과 우리의 AI 도움을 받은 영어 caption/oracle 작성 흐름은 다릅니다. 한국어 설명은 한국어 성능 시험이 아니며 4개 repo로 다양성이 해결됐다고 주장하지 않습니다.

조사 장부에는 공식 GitHub HTTP 18건 성공/0건 실패, browser 1회·2개 page open, 오래된 worktree의 없는 inventory 경로 조회 1건을 구분합니다. metadata helper는 1회 성공/0회 실패이며 원문 19파일·README 4파일·4개 동일 raw helper span만 검사했습니다. 새로운 성능/자원 수치, 새 학습, 공유 파일 변경, 게시, 보호 final 읽기는 0입니다. 별도 model 호출 0은 이 Codex 협업 읽기 자체의 비용이 0이라는 뜻이 아니며 그 비용은 측정하지 않았습니다. 앞선 첫 3원천의 24개 관찰은 재실행하지 않았습니다.
