# 시점별 연구 기록

아래는 README 첫 부분에 있던 당시의 연구 기록입니다. 수치와 문장을 그대로 보존했으며, ‘최신’과 ‘다음’은 각 기록이 작성된 시점을 가리킵니다. 현재 사용법과 기본 동작은 [README](README.md)에서 확인하세요.

---

# laya-tools

[개선74](experiments/short-claim/feature-prefix-74/README.ko.md): 같은 Go 특징·점수 비트를 유지하고 공개 예제2개에서 특징 계산 시간 중앙값을약24%·20% 줄였습니다. 할당량은 그대로이며 [필수 CI·자동 병합](experiments/short-claim/publication-proof-97/README.ko.md)을 확인했습니다. [단독 reader](experiments/short-claim/standalone-reader-74/README.ko.md)는 JSON/compact의24개 열·수치·null·메타데이터가 같았지만 peak RSS18.2/18.4MiB로 RAM 절감은 보이지 않았습니다.

[다음 개발 자료73](experiments/short-claim/source-observation-73/README.ko.md): 서로 다른3개 동작 목표의24개 유한 입력을 실제 Go 원문에 한 번 적용해 사전 예상과 일치함을 확인했습니다. 아직 새 학습 정답은0이며, [모델 입력 설명의 정보74](experiments/short-claim/source-caption-74/README.ko.md)를 정답·학습 적격성과 분리해 보강합니다. [저장 형식73](experiments/short-claim/compact-storage-73/README.ko.md)은 FP64 bit를 유지하고8.39MB→1.98MB로 줄였으며 공백 제거·gzip 기준선도 따로 공개합니다. 파일 감소는 RAM·추론 성능 개선과 구분합니다.

[최신 개발 결과](experiments/short-claim/second-ranking-fit-72/README.ko.md): 작은 주장 모델 두 방식 모두 기존 정렬보다 많은 후보 확인이 필요했습니다(**27→31→33회**). 기본 정렬은 유지하며 [새 개발 데이터의 다양성](experiments/short-claim/second-ranking-fit-72/ANALYSIS.ko.md)을 먼저 늘립니다. [Go 입력 재사용](experiments/short-claim/validated-input-runtime-71/README.ko.md)은 정렬 할당81→0을 관측했고 전체 race/vet 검사를 통과했습니다. [첫 실패 모델의 HF 보관·검증](experiments/short-claim/publication-proof-96/README.ko.md)은 완료했으며 모델 본체는 Git에 올리지 않습니다. [English](experiments/short-claim/second-ranking-fit-72/README.en.md). 아래 항목은 각 당시의 기록입니다.

[첫 작은 주장 모델의 실제 결과](docs/wiki/First-Claim-Fit-71-KO.md): 전체76개 자료로 FP32 학습1회를 완료했지만 확인 횟수는 기존 규칙27회→모델31회로 악화해 효용 기준을 통과하지 못했습니다. 파일32,792B, 전체 학습 peak RSS약26.98MiB/0.805초이며 LLM 절감 수치는 아닙니다. 기존 규칙을 유지하고 같은 요청의 상대 순서 손실을 한 번 추가하는 후속을 진행합니다. [원래72개 준비 데이터](https://huggingface.co/datasets/JooYoon/riidolaya-public-claim-preparation-69/tree/d80075c6160a53d5426019cf018016a3b02017bd)는 공개·재다운로드 검증을 마쳤습니다. [English](docs/wiki/First-Claim-Fit-71-EN.md). 아래 항목은 각 당시의 기록입니다.

[실제 작은 힌트 준비69–70](docs/wiki/Claim-Projection-69-70-KO.md): 원래72개 요청의 첫 Go 배열 준비와 별도 검증을 완료했습니다. 반환 payload 약1.69MiB, 전체 준비 peak RSS 약53.95MiB이며 모델 추론 수치는 아닙니다. 새 사례를 합친76개는 전체로 다시 분할했고 첫 학습을 준비합니다. [English](docs/wiki/Claim-Projection-69-70-EN.md).

최근 [작은 주장 힌트의 연결66–69](docs/wiki/Claim-Checks-66-69-KO.md)에서는 원래72요청의 역할을 실제 배정하고, MIT upstream 설명으로 만든4요청·10후보의 유한 동작을 관측했습니다. 원래216후보의 학습 적격 양성35·음성95를 보존하며, 요청·설명 모두 기존 입력 한도 안에 있습니다. [같은 역할의 검사 비용](experiments/short-claim/stored-role-utility-69/STORED-ROLE-UTILITY.ko.md)은 검증 구간의 단일 최선 기준22회/oracle17회로 개선 여지를 확인했습니다. 이는 학습 성과나 Codex 절감이 아니며, 첫 실제 Go 투영을 별도 고정 계획으로 진행합니다. [English](docs/wiki/Claim-Checks-66-69-EN.md).

[설명 참조59](experiments/short-claim/RESULTS-REFERENCES-59.ko.md)는 Go 공식1회로 원문288곳과 정답표 표현식18곳을 연결했습니다. 다음 문장 검토에 사용할 근거이며, 설명 승인·새 정답·학습 성과는 아닙니다. [사용법](experiments/short-claim/USAGE-REFERENCES-59.ko.md), [공개 원천 전이 준비](experiments/short-claim/SOURCE-TRANSFER-59.ko.md), [그룹 전체의 역할 recipe](experiments/short-claim/role-recipe-59.json)를 공개합니다. unknown21·역할 미배정·fit0을 유지합니다.

[저장 정답 효용58](experiments/short-claim/RESULTS-STORED-58.ko.md)은 기존 72개 요청을 네 가지 비학습 방식으로 정렬했습니다. 가장 좋은 고정 기준의 91회 확인과 정답을 아는 순서의 73회 사이에 **약 19.8% 개선 여지**가 있었습니다. 이는 달성한 모델 성능이나 Codex 절감이 아니며, 판단 보류 21개와 학습 준비 미완성을 유지합니다. [사용법](experiments/short-claim/USAGE-STORED-58.ko.md)에서 Go 회귀 검사를 실행하고, [다음 준비](experiments/short-claim/NEXT-STORED-58.ko.md)에서 문구·원천 전이·역할 분할 계획을 확인할 수 있습니다.

[공개 동작 검증57](experiments/public-behavior/RESULTS-57.ko.md)은 버전·경로 패턴의 원본 Go API62건을 실제 관측했습니다.59건 일치·3건 차이를 그대로 보존하며, 큰 숫자의 비교 문제와 더 엄격한 패턴 검사 정책을 구분합니다. [사용법](experiments/public-behavior/USAGE-57.ko.md)에서 공개 결과를 재생할 수 있습니다. 두 원천 가족·여섯 유한 계약이며 모델 성능이나2400독립 작업 결과는 아닙니다. [Laya 원본 회고](docs/laya-source-retrospective-57.ko.md)는 체크포인트 선택과 Codex 비용 예측의 차이, 캐시·보류·라이선스의 다음 범위를 설명합니다.

별도 연구: **아주 작은 힌트로 비싼 탐색의 순서를 개선**하는 [의미 탐색 힌트 실험](experiments/semantic-hints/PLAN.ko.md)을 진행합니다. 모델 없이 직접 시험할 수 있는 [선택형 보조 검색](experiments/baseline-first/README.ko.md)은 `riido-hints --identifier-hints`로 실행합니다. 2,948개 문서 설명의 결과와 한계를 공개하며, 기존 난이도 라우터 및 기본 설정과 독립적입니다.

에이전트가 일부 후보부터 읽게 하려면 [`--limit 20` 페이지 출력](experiments/hint-pagination/README.ko.md)을 함께 사용하세요. 다음 cursor로 나머지를 복원할 수 있으며, 실제 토큰 절감은 별도 검증 중입니다.

여러 페이지를 읽을 때는 [`--session --limit 20`](experiments/hint-session/README.ko.md)으로 검색 결과를 재사용할 수 있습니다. 첫 요청 이후에는 원문 대신 cursor만 보내며, 입력을 닫으면 종료됩니다.

정적 임베딩 비교군: [Go 실행·압축·원본 일치 검증](experiments/static-embedding/README.ko.md). 코드 관련성 성능은 아직 부족하며 기본 모델로 적용하지 않습니다.

[작은 주장 모델 실험](experiments/path-cost-claim/RESULTS-46.ko.md)은 실제 학습 **4,456개·검증 2,599개**로 10개 후보를 실행했습니다. 규모 기준은 충족했지만 유용성 기준은 실패했습니다. 초기 24문제와 모델 설정 수를 구분하며, 최종 평가 2,402개는 아직 평가하지 않았습니다. 전체 준비·학습·정책 재생의 warm 실행은 최대 RSS 약 108.6MiB였으며, 단일 추론이나 Codex 절감 측정값은 아닙니다.

[골든셋 규모 안내](docs/golden-set-scale.ko.md)는 24개 예비 시험과 최소 2,400개 평가의 차이, 학습·검증·최종 평가 분리, 검색·모델 라우팅·저장소 선택·작업 분할별 정답을 설명합니다.

[외부 작업의 공정 비교](docs/fair-upstream-comparison.ko.md)는 전체 요구사항과 원본 Go 언어 조건을 고정하고, 하나의 공유 기록에서 네 실행 예약을 제한합니다. 기존 두 문제의 새 평가 버전은 고유 문제 수에 더하지 않습니다.

[실제 외부 작업 비교55](experiments/task-outcomes/RESULTS-55.ko.md)는 humanize·UUID 두 요청을 Sol6/Luna low에 각각 한 번 맡겼습니다. 서수는 둘 다 통과했고 UUID는 Sol6 요청만 통과했습니다. 네 실행의 사용량·종료·독립 검사·예산 영수증을 모두 보존했습니다. 실제 관측 누적은 **고유 요청7개·기록20개·다섯 코드 가족·세 저장소**이며, 새 학습·최종 평가·절감 증명은 없습니다. Laya는 두 전체 입력 모두 잘려 보류했고 cold CPU 최대 RSS는 약1.54/1.58GB였습니다. 초저자원 목표와의 차이도 공개합니다. [54의 파서 개선](experiments/task-outcomes/RESULTS-54.ko.md)과 [외부 계약 준비](docs/public-go-contracts-54.ko.md)는 역사적 단계로 보존합니다.

[짧은 주장56a](experiments/short-claim/RESULTS-56a.ko.md)는 모델 없이 실행되는 Go `riido-shortclaim`과 반복용 `--stream`을 구현했습니다. [사용법](experiments/short-claim/USAGE-56.ko.md)을 따라 후보 확인 순서만 제안받을 수 있습니다. 8후보 예제의 전체 호출 warm p95는0.063~0.079ms, 반복 측정 child peak RSS는10.25~10.56MiB였습니다. 48개 자체 작성 사례는11개 연결 그룹으로 하한15에 못 미쳐 학습0입니다. BM25→oracle의 가능한 검사 감소18.64%는 달성한 모델 성능·LLM 절감이 아닙니다. [준비56](experiments/short-claim/PLAN-56.ko.md)은 역사적 snapshot으로 보존하고, [다음 확보56b](experiments/short-claim/NEXT-56b.ko.md)는 오류·상태·소유권처럼 다른 동작을 먼저 검증합니다.

[동작 정답 감사56b](experiments/short-claim/RESULTS-56b.ko.md)는 오류·상태·소유권·수명·그래프를 검사하는 Go 유지보수 도구 `riido-typedaudit`을 추가했습니다. [사용법](experiments/short-claim/USAGE-56b.ko.md)에서 공개 입력을 재생할 수 있습니다. 새24개와 기존48개는 전체17그룹·라벨16그룹으로 두 최소15 기준을 통과했습니다. 524개는 원천 대조군 검사이지 독립 요청 수가 아닙니다. 불완전한 설명은 unknown으로 남겼고, 같은 합성 작성 흐름·역할 계획 미완성 때문에 학습·새 모델·가중치·최종 평가·성능 측정은0입니다.

[작은 속성 제안 준비56c](experiments/short-claim/RESULTS-56c.ko.md)는 Go `riido-scopeprep`으로 기존 부모4개에서 속성3개·제안12행을 준비했습니다. [사용법](experiments/short-claim/USAGE-56c.ko.md)을 따라 문구와 원천 연결을 생성할 수 있습니다. 설명36개는 모두 pending이며 독립 표본 증가·정답 승인·학습·성능 개선으로 세지 않습니다. [상주 자원56d 사전 계획](experiments/short-claim/PLAN-RESIDENT-56d.ko.md) 원문은 보존하고 실제 실행을 별도로 기록했습니다.

[상주 처리 비용56d](experiments/short-claim/RESULTS-56d.ko.md)는 고정한 실제 Go 바이너리로24행·25,080회 JSONL 요청을 완료했습니다. [사용법](experiments/short-claim/USAGE-56d.ko.md)에서 준비와 실행 경계를 확인할 수 있습니다. 이 Mac의 비학습 힌트 child peak RSS는9.22~10.52MiB, 행별 timed 왕복 p95는0.022~0.050ms이며 첫 응답577.13ms를 보존했습니다. 측정기33.11MiB peak는 별도입니다. 원본72개 중3후보48개를 반복한 것으로2400개 독립 최종 평가가 아니며, `narrow_rule`은 모두 BM25 fallback이었습니다. 모델 추론·학습·캐시·LLM 절감 증거와 구분합니다.

[유한 입력 주장 감사56e](experiments/short-claim/RESULTS-56e.ko.md)는 독립 소스 읽기로 문구 범위를 검토한 뒤36개 실제 함수 관측을12파생 행에 연결했습니다.11행은 지정 입력에서 일치 후보가 있고1행은 없습니다. [사용법](experiments/short-claim/USAGE-56e.ko.md)대로 Go에서 공개 관측을 검증할 수 있습니다. 기존 일반 문구의 pending/unknown은 보존하며 독립 요청 증가·모델 학습·유용성 증명은0입니다. [다른 공개 원천의 다음 계약](experiments/short-claim/SOURCE-NEXT-56e.ko.md)은 고정 버전의 semver/glob 구현·MIT 원문을 읽은 조사이며, 아직 새 데이터 확보 수로 세지 않습니다.

[사용량 기록과 독립 작업 검증](docs/task-outcomes.ko.md)을 준비했습니다. 기존 Codex 기록의 사용량과 공개 작업의 요구사항 충족 여부를 따로 확인합니다. 직접 작성한 예제로 도구를 검사했으며, 실제 모델 작업 결과나 비용 절감으로 세지 않습니다.

[공개 작업 실행기](docs/task-execution.ko.md)는 명시적으로 요청한 Codex 실행, 사용량, 결과 파일과 독립 검사를 연결하는 선택적 Go 개발 도구입니다. 실패·시간 초과도 기록하며, 현재 작은 개발 파일럿과 최종 골든셋 평가를 구분합니다.

[경량 검색 방식 세 가지](experiments/path-helper-headroom/RESULTS-48.ko.md)를 학습용 4,456개 전체에서 비교했습니다. 정답을 미리 아는 선택의 개선 여지도 3.09%·3.68%·1.36%로 기존 5% 기준에 못 미쳐 추가 학습을 중단했습니다. 전체 실행·재실행은 약 82초, 최대 RSS 106.7–107.3MiB였으며 단일 추론이나 Codex 절감 측정은 아닙니다.

