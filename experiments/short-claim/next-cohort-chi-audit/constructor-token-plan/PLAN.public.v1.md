# 원문 토큰 준비 공유 / Shared raw token preparation

이번 가설은 생성자 안에서 같은 원문을 반복 토큰화하는 비용을 줄이는 것입니다. 요청1개와 후보최대8개의 원문 단어·순서 있는 terms를 한 번 준비하고, 요청의 FNV(term+NUL) prefix를 공유합니다. 임시 준비 자료는 최종 Prepared 객체에 남기지 않습니다. 새로운 pool·전역 cache·lock은 없습니다.

This experiment shares constructor-local raw token plans and ordered query hash prefixes between cardinality counting and feature generation. Both cross-hash passes and the final copy remain paid work. The final owner retains no token/prefix plans. This is an execution-cost experiment, not training or a model-quality claim.

비교 기준은 PR147로 병합된 scratch Prepare 전체 소스5,971B, SHA256 `7776f6a6a4e16a20009e65c59b64604a2222980eb2fd8f787431aac24b319bdf`입니다. 기준 함수 본문은 그대로 두고 이름만 바꿉니다. 기존 Features·Rank·텍스트 clone·offsets·tight 저장·오류 우선순위를 보존합니다. normalized 문장을 원문 대신 사용하지 않습니다. 기존의 불안정 정렬과 FP64 누적 순서, 상쇄된+0 특징도 그대로 유지합니다.

The baseline is the exact merged scratch source above. Only its comparison function name changes. The original Features implementation remains an independent bit oracle; Rank and final ownership contracts remain unchanged. Ordered and repeated raw terms must survive overlap-word sorting/compaction. A successfully prepared empty token plan is valid; an uninitialized plan is invalid.

별도 FREEZE는 변환 레시피·생성 코드·기준 본문·새 코드·입력·모델·예산의 전체 크기/지문을 고정합니다. 합성 대조군이 먼저 통과한 뒤 첫 실제 결과를 보존합니다. 이전 constructor-scratch의16파일과 관측은 수정하지 않습니다. 이전 결과를 새 기준의 결과로 다시 해석하지 않습니다.

실행 일정은 N=1/8 각각6쌍이며 AB/BA 각3쌍입니다. 정상 완료 시 측정24구간은 Prepare24/Rank108, 공통 준비와 anchor를 포함한 명시 래퍼는140회·Prepare26/Rank110입니다. 모든 준비·토큰화·prefix·cross scan·최종 복사 비용을 포함합니다. GC/MemStats 경계와 결과 변환 위치는 양쪽에서 같고, 전체 특징·점수 비트·순위·소유권 증거를 남깁니다. 후속 모델 없는 Features 재계산은 당시 명시 호출 수에 더하지 않습니다.

Separate-freeze adoption criteria / 별도 동결의 채택 기준:

- 모든 쌍에서 새 누적 Go 할당 바이트가 기준을 넘지 않습니다.
- 모든 쌍에서 기준보다 Mallocs가 최소144회 적어야 합니다.
- N별 쌍별 새/기준 **경과 시간** 비율 중앙값이1.00 이하여야 합니다.
- 모든 특징/점수 비트, 순위, 후보, offsets, tight capacity, 텍스트 소유권 및 오류 계약이 같아야 합니다.

These are paired elapsed-time criteria, not CPU utilization or GPU measurements. Cumulative Go allocation is not peak memory, retained heap or RSS. A valid saved record may describe a failed cost gate; recording it does not authorize adopting the candidate. Do not rerun the same frozen experiment merely to obtain a passing number.

노출된 개발 부모1개와 비활성 failed79 RIIDOH01 FP32 해시 특징 점수기를 사용합니다. 기존 모델 파일32,792B의 SHA256은 `dff05140098845943ece87c3ab8a31a15564a4b170a59ab017ee393501158004`입니다. 모델 가중치는 GitHub에 넣지 않습니다. 정답5위 실패는 비용 최적화로 해소되지 않습니다. 독립2400평가·source-family 편입·새 역할/라벨·Fit·모델 활성화·실제 에이전트 비용 개선은 별도 미완료 과제입니다.

One exposed development parent, no new independent Golden cases, labels, roles, training or activation. Existing failed79 weights remain inactive. This experiment neither executes a Laya encoder nor proves useful routing, GPU/MPS performance or LLM-cost savings.
