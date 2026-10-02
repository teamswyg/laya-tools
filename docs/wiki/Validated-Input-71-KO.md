# 입력을 한 번 검증하는 Go 도구

사람과 에이전트의 명령어·JSONL 사용법은 유지합니다. 같은 입력을 읽기와 정렬에서 두 번 정규화하던 부분을 한 번 검증한 값으로 재사용합니다.

```sh
riido-shortclaim --stream --baseline lexical_ordered < requests.jsonl
```

Go 프로그램에서는 `shortclaim.LoadValidated(reader)` 또는 `shortclaim.ValidateInput(input)`으로 만든 값에 `.Rank(kind)`를 호출합니다. 입력 검증은 형태와 정규화 확인이며 정답이나 실행 권한의 승인이 아닙니다. 기존 `Rank(Prepared)`도 유지합니다.

공개8후보 예제에서 정렬 할당은 **81→0**, 요청 전체는 **291→210**으로 관측했습니다. 두 측정의 프로파일 설정이 달랐으므로 속도·전체 메모리의 개선 폭을 확정하지 않습니다. 전체 race/vet 검사와 원래 점수·순서·fallback 회귀 검사를 통과했습니다. 과거 측정은 원본 소스로 검증해 보존합니다.

[사용법·수치·한계](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/validated-input-runtime-71/README.ko.md) · [학습 결과](Second-Claim-Fit-72-KO) · [English](Validated-Input-71-EN)
