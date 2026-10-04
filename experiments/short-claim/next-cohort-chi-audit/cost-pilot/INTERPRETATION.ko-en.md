# 준비 비용 해석 정정 / Setup-cost interpretation correction

초기 SUMMARY의 `Preparation reuse reduces repeated setup`은 이번 관측보다 앞선 표현이었습니다. 각 warm 블록은 서로 다른 네 정책을 한 번씩 실행했고 모델 Rank는 한 번만 호출했습니다. warm 구간은 별도로 지불한 준비 비용을 제외합니다. 같은 모델을 반복 호출해 준비 비용을 분산하는 효과는 측정하지 않았습니다.

The initial SUMMARY sentence `Preparation reuse reduces repeated setup` exceeded this observation's scope. Each warm block ran four different policies once and called model Rank once. Warm intervals exclude separately paid setup. Repeated same-model setup amortization was not measured.

SUMMARY의 해석 문장과 해당 파일 체크섬만 정정했습니다. 모든 수치·관측·실행 계획·입력·기대값·점수·순서·소스 지문은 그대로입니다. 모델은 비활성입니다. [PR144 원본](https://github.com/teamswyg/laya-tools/tree/899fb15831946ada5716a83a1ffc994961b06fed/experiments/short-claim/next-cohort-chi-audit/cost-pilot)과 [초기 HF 자료](https://huggingface.co/JooYoon/riidolaya-shortclaim-data-effect-failed-79/tree/8b49f75205066d65d18017d1066d44af0d130ed4/followups/chi-cost144-899fb15)는 당시 상태를 보존합니다. 새 모델·원본 실행 결과가 아닙니다.

Only SUMMARY commentary and its checksum are corrected. All numeric fields, observations, sealed protocol, inputs, expectations, scores, order and source pins remain unchanged. The model stays inactive. The linked immutable PR144 and initial HF records preserve their publication-time state; this is not a new model or target execution result.

다음 Go 준비 실험은 아직 실행 전 제안입니다. 반복 횟수1·2·4·8에서 준비본 유지와 매번 재생성을 비교하고, BM25도 같은 입력 수명으로 맞춥니다. 준비와 매 호출의 실제 검증 비용을 모두 포함하며 점수 비트·동률·전체 후보·오류 동작을 대조합니다. 준비 할당은 모델 읽기·LoadValidated·View 생성·Prepare로 나눠 확인한 뒤 SoA 변경의 필요성을 판단합니다. pprof 원본은 공개하지 않고 누적 할당과 보유 메모리를 구분합니다. 이 사례의 순서 실패나 독립 학습 자료 자격은 이 최적화로 해소되지 않습니다.

The next Go experiment is a proposal, not an executed or sealed result: compare retained preparation with rebuilding at1,2,4,8 calls, using matched input lifetimes for BM25. Include setup and every actual verification; preserve score bits, ties, all candidates and errors. Attribute setup allocations to model reading, LoadValidated, View construction and Prepare before evaluating SoA. Keep raw profiles private and distinguish allocation activity from retained memory. Optimization does not resolve this case's ordering failure or independent training-data eligibility.
