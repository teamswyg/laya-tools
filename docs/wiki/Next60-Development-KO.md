# 다음 작은 주장 모델 학습을 준비하는 방법

현재 새 개발 학습 자료는 **요청 7개·후보 라벨 20개**입니다. 긍정 7개·부정 13개에 가중치 1을 유지합니다. 기존 초안 일곱 개를 유한 범위에서 적격화했으며 새 모델 학습은 아직 0회입니다. 기존 실제 79개 자료와 비활성 모델은 그대로입니다.

이번 [네 작업의 실제 실행](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-four-selector-actual-observation)은 쉼표 문법, 키 충돌, 맵(키·값 자료구조) 스냅샷, 소수 바이트 처리입니다. 입력 19개 × 후보 3개 = 관측 57개를 한 번 실행했습니다. 독립 검토에서 일치 40개·알려진 불일치 17개·판정 불가 0개를 확인했고, 각 작업의 후보 1이 모든 고정 입력을 만족했습니다. 입력 19개와 관측 57개를 독립 요청 수로 세지 않습니다.

원본 검사 프로세스의 OS 최대 RSS는 약 7.89 MiB, 시작·입력 검증·저장 과정을 포함한 시간은 약 1.078초입니다. 명시적 후보 호출 57회와 오류 메서드 호출 14회가 반환됐지만 원본 내부 API·초기화 호출 수는 측정하지 않았습니다. 모델 호출은 0회이며 GPU 성능이나 추론 비용 절감 결과가 아닙니다.

[현재 일곱 줄 자료와 사용법](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-development-seven)을 보세요. Go reader 7회와 새 입력 검사 4회가 통과했고, 이전 세 줄은 바이트까지 그대로 보존했습니다. 요청·후보 문장만 모델 특징에 들어가며 정답·출처는 별도로 읽습니다. 새 네 작업도 기존 humanize·pflag·mapstructure 학습 그룹에 속하므로 독립 소스 가족이나 보호 평가 자료가 늘어난 것은 아닙니다.

[직접 반올림 작업](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-ftoa-actual-observation)을 실제로 한 번 실행했습니다. 입력 6개 × 후보 3개 = 관측 18건이며 독립 요청은 하나입니다. 개선 후보는 6/6 입력을 만족하고, 원본과 오류 후보는 확인된 반례가 있습니다. 원본의 오류 반환 채널 미확인은 그대로 남습니다. 같은 humanize 소스의 기존 Ordinal 학습 그룹 76에 연결하며, 새 독립 소스나 평가 자료로 세지 않습니다.

작업 프로세스의 전체 시간은 약 0.378초, macOS 종료 시 보고한 직접 자식 최대 RSS는 약 5.55 MiB입니다. Go heap 207,056 B는 한 시점의 값입니다. 이 수치는 원본 코드의 정답 확인 작업이며, 작은 모델의 추론·GPU·전체 머신 메모리나 비용 절감을 입증하지 않습니다.

[이전 세 줄 자료](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-development-three)의 실제 reader 3회와 입력 검사 1회 통과 기록도 보존합니다. [PR112](https://github.com/teamswyg/laya-tools/pull/112)의 필수 CI 4개가 통과해 Go reader가 자동 병합됐습니다.

[HF 3개 요청 고정 버전](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-3-finite-v1)을 게시했고 71개 파일을 다시 내려받아 확인했습니다. Viewer는 HTTP200으로 세 행을 반환하며 모든 필드가 고정 자료와 일치했습니다. [게시 기록](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/publication-proof-114/HF-PUBLICATION.v2.json)을 보세요. 이전 2개 요청 태그와 당시 viewer HTTP500·나중 성공 기록도 보존합니다. 모델 학습·성능 검증과는 별도의 자료 게시입니다.

HF의 현재 고정 버전은 위의 3개 요청 자료입니다. 7개 버전은 CI로 검증한 뒤 별도 태그로 게시해 이전 버전을 보존합니다.

다음 [의미 작업 10개 제안](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-catalog10-literal-correction)은 구체적 입력 46개와 관측 방법이 더 필요한 보류 입력 6개입니다. 아직 실제 관측·정답 라벨은 아닙니다. 기존 초안 20개를 모두 적격화해도 새 라운드의 30개 체크포인트에는 별도 의미 작업 10개가 더 필요합니다. 문장 변형을 새 작업으로 세지 않습니다.

자료가 30개·60개에 도달한 뒤 별도 학습 비교를 진행합니다. 주장 도메인별 보호 평가 2,400개 목표와 5% 효용 개선 기준을 유지합니다. 새 모델은 실제 비교를 통과하기 전 활성화하지 않습니다.

[Go reader 사용법](https://github.com/teamswyg/laya-tools/tree/main/pkg/shortclaimdata) · [첫 두 요청의 학습 자료](Native2-Training-KO) · [첫 원본 관측](Native2-Observation-KO) · [English](Next60-Development-EN)
