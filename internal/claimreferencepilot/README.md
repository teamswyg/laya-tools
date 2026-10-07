# AI reference pilot guards / AI 참조 파일럿 검사

새 학습 자료를 만들기 전에, 한·영 개발 댓글의 판정 기준을 작은 별도
파일럿에서 확인하기 위한 Go 모듈입니다. 텍스트나 모델을 읽지 않고
검토 기록의 구조와 고정된 분모를 검사합니다.

This internal Go module checks the structural and numerical metadata contract of
a separate AI-reference definition pilot. It reads no comment bodies, loads no
model and emits aggregate counts only. It does not authenticate a provider,
establish semantic truth, grant execution authority or qualify a model.

## Fixed scope

- 48 families, 96 Korean/English comments, eight strata of six families.
- 16 distinct documented AI role contexts, including eight locale/panel raters.
- Two panels, two raters per locale: 384 records and 1,152 head judgments.
- First-panel per-rater support: at least two families per state in 36 cells.
- First-panel agreed true: at least eight families per locale/head in six cells.
- Each panel/locale/head: at least 39 exact agreements and 44 true-versus-nonpositive
  agreements, always with denominator 48.
- Both-locale naturalness: at least five of six pairs in each stratum; all 48
  source/fidelity/frame reviews remain separate requirements.

`false` and semantic `unknown` remain distinct. Missing or unable reviews do not
become votes; duplicate attempts remain in the submitted records and block a pass.
Confidence never supplies a semantic state, and a majority never overwrites the
first reference. Passing these checks is not human gold or operational usefulness.

## APIs

| API | Purpose |
|---|---|
| `ValidateBindings` | Check documented role/context and prerequisite receipts before words; S2 execution admission can still be absent. |
| `ValidatePlan` | Check the fixed pre-word inventory, input scope, budgets and zero content attempts. |
| `ValidateUsage` | Check inclusive invocation counts and byte usage during the workflow. |
| `Evaluate` | Join all completed records and check fixed numerical gates, lifecycle, retained dissent, usage and required S2 admission. |

Enums reserve zero for missing metadata. Fixed dimensions use arrays. The admitted
AI invocation ceiling is 49: 17 metadata starts, 25 content turns and seven
controller continuations, with separate per-role limits. Acquisition and review
limits are 8 MiB and 16 MiB; failed attempts and control files count too.

판정자가 실제로 읽은 문장, 출처의 충실성, 역할 실행 기록의 진위는 별도의
승인된 기록·검토 과정이 확인해야 합니다. 이 코드의 `PASS`를 학습 허가나
모델 정확도로 해석하면 안 됩니다. 96개 파일럿 문장은 최종 2,400개 자료를
대체하지 않으며 최종 학습·평가 자료에서 제외합니다.

Tests use newly authored abstract metadata only:

```sh
go test -race ./internal/claimreferencepilot
go vet ./internal/claimreferencepilot
```
