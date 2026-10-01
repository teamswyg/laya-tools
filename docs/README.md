# Documentation / 문서 안내

Choose a language; both versions cover the same features, measurements, and limits.
같은 기능·측정·한계를 두 언어로 제공합니다.

**Start with the [user Wiki](https://github.com/teamswyg/laya-tools/wiki) / 처음에는 [사용자 Wiki](https://github.com/teamswyg/laya-tools/wiki)를 보세요.**

Latest owned task evidence: experiment 52 recorded all six attempts on three development requests. Six candidates passed independent closure checks; five attempts exited zero with complete core usage, while one accepted candidate timed out with unknown usage. Including preserved experiment 51, there are eight records over three requests in two code families and one repository. These remain development evidence, not population capability labels or demonstrated savings. The target remains at least 2,400 distinct final requests per evaluation domain, with related cases grouped.

최근 실험 52는 개발 요청 3개에서 계획한 여섯 시도를 모두 기록했습니다. 후보 6개는 독립 범위 검사를 통과했고, 다섯 시도는 정상 종료와 완전한 기본 사용량을 확인했습니다. 나머지는 후보가 수용됐지만 시간 초과로 사용량이 미확정입니다. 보존한 실험 51까지 합치면 두 코드 가족·한 저장소의 요청 3개에 기록 8건입니다. 모집단의 능력 라벨이나 절감 근거로 쓰지 않습니다. 영역별 구분되는 최종 요청 최소 2,400건을 목표로 하며 관련 사례는 같은 그룹으로 관리합니다.

| Topic / 주제 | English | 한국어 |
|---|---|---|
| Project overview / 프로젝트 개요 | [README](../README.en.md) | [README](../README.md) |
| Router design / 라우터 설계 | [Design](design.en.md) | [설계](design.ko.md) |
| Decomposition decision plan / 작업 분할 판단 계획 | [Preview](decomposition-preview.en.md) | [Preview](decomposition-preview.ko.md) |
| Maintainer MPS preparation / 유지보수 MPS 준비 | [Readiness and limits](mps-training.en.md) | [준비 검사와 한계](mps-training.ko.md) |
| Real MPS training pilot / 실제 MPS 학습 실험 | [Results and publishing](mps-pilot.en.md) | [결과와 배포](mps-pilot.ko.md) |
| Upstream adaptations / 생태계 이식 | [Ecosystem](ecosystem.en.md) | [생태계](ecosystem.ko.md) |
| Repository selection preview / 저장소 선택 | [Preview and evaluation](repository-routing-preview.en.md) | [Preview와 평가](repository-routing-preview.ko.md) |
| License audit / 라이선스 검토 | [Audit](license-audit.en.md) | [검토](license-audit.ko.md) |
| Initial measurements / 초기 측정 | [Measurements](measurements.md) | [측정](measurements.ko.md) |
| Golden-set scale and domain boundaries / 골든셋 규모와 도메인 구분 | [24 versus 2,400](golden-set-scale.en.md) | [24개와 2,400개의 의미](golden-set-scale.ko.md) |
| Golden-set acquisition, grouped splits and budgets / 골든셋 확보·그룹 분할·실행 예산 | [Staged plan](golden-set-acquisition.en.md) | [단계별 확보 계획](golden-set-acquisition.ko.md) |
| Usage recording and independent task verification / 사용량 기록·독립 작업 검증 | [Use and evidence limits](task-outcomes.en.md) | [사용법과 증거 범위](task-outcomes.ko.md) |
| Explicit public task execution / 명시적 공개 작업 실행 | [Owned process and evidence](task-execution.en.md) | [직접 실행과 근거 연결](task-execution.ko.md) |
| Owned coding attempts 52 / 실제 코딩 시도 52 | [Six accepted closures, five complete executions](../experiments/task-outcomes/RESULTS-52.en.md) | [여섯 후보 수용·다섯 실행 완료](../experiments/task-outcomes/RESULTS-52.ko.md) |
| Owned coding attempts 51 / 실제 코딩 시도 51 | [Two attempts, no profile comparison](../experiments/task-outcomes/RESULTS-51.en.md) | [두 시도와 비교 중단](../experiments/task-outcomes/RESULTS-51.ko.md) |
| Stopped packaging pilot 50 / 패키징 문제로 중단한 파일럿 50 | [Prelaunch refusal](../experiments/task-outcomes/RESULTS-50.en.md) | [모델 실행 전 거절](../experiments/task-outcomes/RESULTS-50.ko.md) |
| Training-only search-helper headroom / 학습용 검색 방식의 개선 여지 | [4,456 requests, stop decision and resources](../experiments/path-helper-headroom/RESULTS-48.en.md) | [4,456개 결과·중단 판단·자원](../experiments/path-helper-headroom/RESULTS-48.ko.md) |
| Search-claim training 4,456 / validation 2,599 / 검색 주장 학습·검증 | [Results, failure and memory](../experiments/path-cost-claim/RESULTS-46.en.md) | [결과·실패·메모리](../experiments/path-cost-claim/RESULTS-46.ko.md) |
| Historical source review 47 / 과거 출처 검토 47 | [Source scope](../experiments/path-cost-data/SOURCE-REVIEW-47.en.md) | [출처 범위](../experiments/path-cost-data/SOURCE-REVIEW-47.ko.md) |
| Maintainer model export / 유지보수 모델 변환 | [Model build](model-build.md) | [모델 빌드](model-build.ko.md) |

The Wiki focuses on installation, choosing a workflow, interpreting results, integrating agents/Go, and troubleshooting. These documents retain technical evidence and provenance. Raw data, fixture JSON, and license originals are shared across languages rather than translated.

Wiki는 설치·기능 선택·결과 해석·에이전트/Go 연결·문제 해결을 설명합니다. 이 문서들은 기술적 근거와 출처를 보존합니다. 원본 측정 JSON·예제 설정·라이선스 원문은 번역하지 않고 공통으로 사용합니다.

## Maintaining the Wiki / Wiki 유지보수

Edit [wiki sources](wiki) in both languages through a repository PR. After `quality` passes and the PR merges, copy those Markdown files into the separate `laya-tools.wiki.git` checkout and commit/push the reviewed content. Preserve unrelated Wiki pages. GitHub's first Wiki page must already exist before the Wiki Git remote can be cloned. Publication uses existing authorized Git access; no new token or scheduler is configured.

[Wiki 원본](wiki)의 두 언어를 저장소 PR로 수정합니다. `quality` 통과·병합 후 검토한 Markdown을 별도 `laya-tools.wiki.git` checkout에 복사하고 커밋·푸시합니다. 관계없는 Wiki 페이지는 보존합니다. 최초 페이지가 있어야 Wiki Git 원격 저장소를 clone할 수 있습니다. 기존에 허용된 Git 접근으로 게시하며 새 토큰이나 스케줄러를 만들지 않습니다.

Wiki pages use absolute links so they work both in the source repository and on GitHub Wiki. Update both languages and check code examples, measured numbers, feature status, and links together. Direct Wiki edits do not inherit the main repository's branch protection.

Wiki 페이지는 원본 저장소와 Wiki 모두에서 동작하도록 절대 URL을 사용합니다. 두 언어의 명령 예제·측정 숫자·기능 상태·링크를 함께 확인하세요. Wiki 직접 수정에는 메인 저장소의 브랜치 보호가 적용되지 않습니다.
