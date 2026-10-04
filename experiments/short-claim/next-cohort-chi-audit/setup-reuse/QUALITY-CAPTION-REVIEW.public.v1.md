# Caption review / 캡션 검토 v1

한국어: 결과가 노출된 개발 원문·술어의 읽기 검토입니다. 실행·모델 평가·보호된2400 접근·Fit=0입니다. 원본 요청·캡션·라벨·역할·결과를 보존합니다. `quality-caption-proposal-v2`는 적용·정답 지정 없이 별도로 제안한 미실행 영문입니다.

English: Nonblind saved-source/predicate review only. Execution, model calls, protected2400 access and Fit are zero. Original requests, captions, labels, roles and results remain unchanged. `quality-caption-proposal-v2` is a distinct unexecuted English proposal, not applied or assigned new truth.

Afero: 기존 음성 캡션은 Stat 기반 무제한 읽기와 접두부 잘라내기를 한 문장에 섞습니다. 실제 선택된 `TruncateSuccess`는 Stat을 보지 않고 limit까지만 읽으며 초과 검사 바이트를 읽지 않습니다. 읽기 오류는 보존하고 획득한 파일을 한 번 닫습니다. 기존 요청에는 음수 한도의 열기 전 거절과 한 번 닫기가 없지만 Wants에는 있습니다. `LimitReader`가 제한하는 전달 바이트 수는 `ReadAll`의 정확한 할당 용량이나 총 메모리 한도가 아닙니다.

The Afero negative caption conflates a Stat/unbounded branch with truncation. The selected `TruncateSuccess` ignores Stat, reads only through limit and has no overflow probe; it preserves read errors and closes an acquired file once. Wants additionally require negative-limit rejection before opening and single Close, which the original request omits. Bounded delivered bytes do not establish exact allocation capacity or a total memory bound.

For the existing five fixtures, propose these v2 strings. “Positive/negative” identifies existing candidate descriptions, not newly assigned labels:

- Request: “Add ReadFileLimit: reject negative limits before opening, close acquired files once, read through at most limit+1 bytes, accept exact limits, and distinguish overflow from read errors without truncated success.”
- Positive (`reference-design`): “Validate limit before opening; ignore Stat. Read at most limit+1 bytes, preserve read errors, close acquired files once, and return nil data with size_limit on overflow.”
- Negative (`wrong-seed-design`): “Reject negative limits before opening; ignore Stat. Read only through limit without an overflow probe; preserve read errors, close acquired files once, and silently accept an oversized file's prefix.”

한국어 의미: 음수·닫기 규칙과 limit+1 초과 검사를 명시하고 접두부 성공과 구분합니다. 큰 한도·Close 오류 정책은 입증하지 않습니다. 새 텍스트의 동결·검증이 필요하며 기존 결과를 수정 문구의 실행 결과로 옮기지 않습니다. Large limits/Close errors remain out of scope.

GJSON: 기존 `decoded_local_policy` 캡션은 객체별 decoded key 비교, 첫 중복에서 중단, 객체 경로 보존을 이미 설명합니다. `a`와 `\u0061`의 차이를 놓치는 raw spelling 후보와 구분되므로 필수 수정은 없습니다. Optional v2: “Walk valid JSON before publication; reject repeated decoded keys per object, including escaped aliases; stop at the first duplicate and preserve its object path.” This is also unexecuted; source review supports the existing caption, not a new model-quality result.

Evidence / 근거: [original Afero row](../../next60-development-thirty/data/train.jsonl), [read branch](../../next60-development-thirty/source/durable-worker/module/internal/read/candidates.go.txt), [saved Wants/Got predicates](../../next60-development-thirty/evidence/REMAINING-THREE-PREDICATES.v1.json); [GJSON row](../../next60-development-thirtyseven/data/train.jsonl), [per-object policy](../../next60-durable-evidence/source/worker/core/duplicate.go.txt), [saved comparison](../../next60-native-wire3-execution/NATIVE-CACHE-RESET-WANTED-COMPARISON.actual.public.v1.json). Pinned upstream: [Afero ioutil.go](https://raw.githubusercontent.com/spf13/afero/eb6a92826ea568e3f40ab91fcbda8d2b38b61d3e/ioutil.go), [GJSON gjson.go](https://raw.githubusercontent.com/tidwall/gjson/9378d3bb93e20854e1677e0e0248e1d71ae5712f/gjson.go).

Historical helper copying extent and ancestry remain unresolved. Existing roles are preserved; fresh role allocation and training admission are unresolved. 과거 도우미 복사 범위·계보와 새 역할·학습 편입은 이 검토로 확정하지 않습니다.
