# 입력을 한 번 검증하고 재사용하기

짧은 주장 도구는 요청을 읽으며 정규화한 뒤 정렬할 때 같은 입력을 다시 정규화했습니다. `ValidatedInput`은 기존 검증을 통과한 입력을 비공개 고정 배열에 보관하여 이 반복을 없앱니다. CLI 사용법과 결과는 유지합니다.

```sh
riido-shortclaim --stream --baseline lexical_ordered < requests.jsonl
```

Go에서 재사용하려면 `shortclaim.LoadValidated(reader)` 또는 `shortclaim.ValidateInput(input)`으로 만든 값에 `.Rank(kind)`를 호출합니다. `.Prepared()`는 독립 값 복사입니다. 영 값은 정렬할 수 없으며, 기존 `Rank(Prepared)`는 계속 원래 검증을 수행합니다. 이 계약은 입력 형태와 정규화만 확인하며 정답이나 실행 권한을 보증하지 않습니다.

공개 8후보 예제에서 정렬 할당/op는 **81→0**, 전체 요청은 **291→210**이었습니다. 이전에는 CPU 프로파일을 켰고 새 측정에서는 끈 단일 관측이므로 지연·RSS의 인과적 개선 폭은 확정하지 않습니다. 종료 Go heap은 증가한 사실도 기록했습니다. 새 map·lock·SIMD·모델·GPU·학습은 추가하지 않았으며, LLM 토큰 절감 증거로 사용하지 않습니다.

[측정과 한계](RESULTS-VALIDATED-INPUT-71.v1.ko.md) · [비교 수치](COMPARISON-VALIDATED-INPUT-71.v1.json) · [English](README.en.md)
