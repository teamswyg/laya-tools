# 본문에서 모델 입력을 추출하기

[English](README.en.md)

이미 인증된 앱이 읽은 본문에서 작은 Go 분류기에 넣을 문장을 추출합니다. 호출자가 `plain_text` 또는 `neutral_html_fragment` 형식을 명시해야 합니다. 형식이 없거나 지원하지 않으면 분류 입력을 만들지 않습니다. 실제 앱의 본문 형식·권한·배포 계약은 별도로 확인해야 합니다.

```go
var bodyWork statehintbody.Workspace
result, err := statehintbody.Extract(ctx, heldBody, statehintbody.NeutralHTML, &bodyWork)
if err != nil {
    // 이 본문은 보류합니다. 일부 문장을 잘라 예측하지 않습니다.
    return err
}
var modelWork statehint.Workspace
prediction, err := localModel.Predict(result.Text, &modelWork)
```

모델과 본문 형식은 명시적으로 제공하며 다운로드·본문 재조회·학습·상태 변경을 하지 않습니다. `prediction`은 문장의 의미에 대한 가설입니다. 현재 라벨 목록과 작업 버전·권한 확인을 대신하지 않습니다.

| 경계 | 값 |
|---|---:|
| 원본 본문 | 최대 16 KiB |
| 추출된 모델 입력 | 최대 4,096바이트; 삽입한 줄바꿈 포함 |
| 태그 중첩 | 최대 32 |
| 파서 토큰 | 최대 2,048 |
| 한 태그의 속성 | 최대 8 |

`plain_text`는 원본 UTF-8 바이트를 그대로 유지합니다. HTML 부분집합은 `p/div/br/span/strong/em/b/i/u/a`만 지원합니다. 블록 사이에 줄바꿈을 넣고 `<br>`를 각각 하나의 줄바꿈으로 유지합니다. `br` 외의 태그는 균형 잡힌 명시적 닫힘이 필요합니다. 인라인 안의 블록·중첩 링크처럼 브라우저가 구조를 바꿀 수 있는 조합은 거절합니다.

XML의 기본 entity 다섯 가지와 `&nbsp;`, 유효한 numeric entity만 한 번 해제합니다. 파싱 전에 정확한 `<br>`만 `<br/>`, `&nbsp;`만 numeric 표현으로 바꾸며, `&lt;br&gt;`나 `&amp;nbsp;`를 다시 해석하지 않습니다. 원본 해시는 이 치환 이전 바이트에서 계산합니다.

허용 속성은 `title`, 링크의 `href`, mention span의 `data-type="mention"`와 `data-id/data-label/data-mention-type`입니다. 표시 문장은 span 내부 텍스트에서 가져오며 ID·링크 주소·속성 라벨을 모델 입력에 덧붙이지 않습니다. class/style/hidden, 스크립트·이미지·iframe, 인용·코드·취소선, namespace, 중복 속성, CDATA·주석·선언·처리 명령과 다른 태그/속성은 본문 전체를 보류합니다. 이런 내용을 지우고 남은 문장을 완전한 본문처럼 분류하지 않습니다.

Go 표준 [`encoding/xml`](https://pkg.go.dev/encoding/xml#Decoder)을 strict 모드로 사용하고, namespace·허용 구조·중복 속성·numeric scalar·완전한 입력 소비를 별도로 확인합니다. 일반 HTML 파서나 브라우저 렌더러와 동일한 동작을 보장하는 기능은 아닙니다.

`RawBodySHA256`, `ModelTextSHA256`, `ExtractorVersion`은 서로 다른 표현의 출처를 추적하는 값입니다. 소유자 revision·권한·원자적 snapshot·실제 화면 일치를 증명하지 않습니다. `OwnerRevisionVerified`와 `RenderedVisibilityVerified`는 false입니다. 본문과 두 해시는 기본 JSON에서 제외되며 실제 업무의 기록은 비공개로 관리합니다.

Workspace는 호출자가 소유하는 고정 배열입니다. 한 Workspace를 동시에 공유하지 않습니다. 공용 가변 cache나 lock은 없고, 임시 문장 버퍼는 반환 전에 지웁니다. 반환된 `Text`의 수명과 비공개 처리는 호출자가 관리합니다. 오류는 원문을 되풀이하지 않으며 부분 결과를 반환하지 않습니다.

[가상 본문 149바이트의 로컬 비용](COSTS.development.json)은 추출·출처 해시만 포함합니다. 중복 파서 stack과 해시의 임시 할당을 줄여 호출당 할당은 2,248→1,544바이트, 43→30회였습니다. 같은 단일 조건의 시간 관측은 2.016→1.886µs였습니다. 분류기·JSON·조회·전체 RSS·GPU 비용이나 운영 속도 개선을 뜻하지 않습니다. raw pprof는 로컬에 보관합니다.

공개 테스트는 원본 가상 문장과 잘못된 입력으로 구조·예산·취소·비공개 출력 경계를 확인합니다. 실제 뤼이도 본문 지원률, 모델 정확도, 실제 서비스 권한 검증은 이 테스트의 결과가 아닙니다.
