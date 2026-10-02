# 준비 75의 pair 가중치 설명 정정

**현재72에서는 BCE sampleweight가 0이면 해당 후보와 연결된 pair의 학습 기여도 0입니다.** 이전 준비 v3와 한영 문서는 감사 목록에 남는 endpoint/pair와 실제 rank gradient·NLL 기여를 혼동했습니다. 원래 v1/v2/v3·문서·학습 사양·결과는 그대로 보존하고, 이 별도 v4와 정정문을 최신 설명으로 사용합니다. 기존 실험의 수치를 소급 변경하거나 무효화하지 않습니다.

고정 코드와 실제 실행 driver를 읽어 다음을 확인했습니다.

- [ranking.go 265–272](https://github.com/teamswyg/laya-tools/blob/d506f58ccf9629e2d2b6ca3cd1766ab20789bf64/internal/pairlearn/ranking.go#L265): pair 가중치는 긍정·부정 endpoint의 `SampleWeights` 곱입니다. 감사 배열에는 0 가중치 pair도 남습니다.
- [344–360](https://github.com/teamswyg/laya-tools/blob/d506f58ccf9629e2d2b6ca3cd1766ab20789bf64/internal/pairlearn/ranking.go#L344): 가중치 0이면 gradient를 계산하지 않습니다. [372–390](https://github.com/teamswyg/laya-tools/blob/d506f58ccf9629e2d2b6ca3cd1766ab20789bf64/internal/pairlearn/ranking.go#L372)도 pair NLL에서 0을 건너뜁니다.
- 저장된 driver `fit.go` 118은 projection의 Development/Validation을 그대로 전달합니다. 205–213은 LossEligible에 따른 원래 0/1 sampleweight와의 일치를 검사합니다. `main.go` 371–380도 같은 dataset을 `FitWithRankingTrace`에 전달합니다. Pair만을 위해 다시 unmask하지 않았습니다.

정확한 핀은 v4의 `factual_correction_of_preparation_v3.source_evidence`에 파일 SHA·바이트/줄 범위·span SHA로 있습니다. 고정 실제 계획 SHA는 `17afbf5c8d14fdd34f7c627d7b43b9582793984d010f52118f408b287bbc7daf`입니다. Reader의 독립 receipt SHA는 `c9c5df587cf8b9ac287692b1ac843ddd0fd6489dbbf23fcd4e6be088f0381f6e`입니다. Reader는 이번 request/mask 작성자로부터 독립이지만 caption74 작성자이고 nonblind입니다. 새 pair 수치나 학습을 재계산하지 않았습니다. 저장 audit의 0 가중치 pair 16개/train·2개/validation은 부모/peer의 기존 결과 확인을 참조합니다.

BCE 자격과 pair endpoint 자격을 개념적으로 별도 기록하는 것은 가능합니다. 하지만 **현72의 동일 0 가중치 제외를 보장하려고 새 runtime 구현을 추가할 필요는 없습니다.** 별도 endpoint 정책은 향후 다른 목적함수의 설계 가설입니다. v4에 남은 `enforcement_implemented: false` 류의 필드는 새75의 독립 endpoint-flag 정책을 구현하지 않았다는 뜻이며, 기존 sampleweight-product 처리가 없다는 뜻이 아닙니다.

요청 범위와 원천 코드 충족 제안은 그대로입니다. A는 충분한 설명이 있는 부정 4개·긍정 0개, B는 긍정 3개·부정 6개입니다. 제안 0/1 sampleweight를 현72 방식에 넣는다고 가정한 **nonzero pair 제안**은 A 0개·B 6개·공통 교집합 0개입니다. 실제 pair 생성이나 학습은 하지 않았습니다. 공통 교집합에 긍정이 없으므로 이 작은 묶음만으로 학습 가능한 paired 비교라는 주장은 여전히 성립하지 않습니다.

부모 3개·후보 9개·순서·A/B 원문·정답/역할/가중치/pair `null` 상태와 기존 평가 분모는 그대로입니다. v4는 고정 자료의 사실 정정과 인계용 metadata이며 원 API·Features·Project·Fit·모델·역할·공유/원격 수정은 0회입니다. 학습 자격이나 corpus 편입을 새로 만들지 않습니다.
