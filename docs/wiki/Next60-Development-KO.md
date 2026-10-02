# 다음 작은 모델 학습을 준비하는 방법

현재 새 개발 학습 자료는 **요청 2개·후보 라벨 5개**입니다. 긍정 2개·부정 3개에 가중치 1을 유지합니다. 기존 초안의 두 요청을 적격화한 것이며 새 모델 학습은 아직 0회입니다. 기존 79개 자료와 비활성 모델은 그대로입니다.

다음 단계는 한 번에 실행 절차를 재사용하면서 자료를 늘리는 것입니다. [배치 준비](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-source-batch-preparation)에서 네 대상을 골랐고 목표 열 개에는 여섯 개가 부족합니다. 보류는 13개입니다. 입력 19개·후보 설명 12개·호출 제안 57회는 준비 숫자이며 실제 실행·새 적격 요청 수가 아닙니다.

[반올림 제어기](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-ftoa-outside-preparation)는 Root의 합성 race 검사에서 28개 통과·1개 건너뜀을 기록했습니다. 원본·계산 함수·모델 호출은 0회입니다. 검사 부모 RSS 121,012,224바이트를 worker나 GPU 메모리로 해석하지 않습니다. 원래 작성자 봉인과 이후 Root 검사 기록을 함께 보존합니다.

자료가 30개·60개에 도달한 뒤 별도 학습 비교를 진행합니다. 주장 도메인별 보호 평가 2,400개 목표와 효용 기준을 유지하며 준비 입력을 독립 표본으로 부풀리지 않습니다. [PR111](https://github.com/teamswyg/laya-tools/pull/111)의 필수 CI 4개가 통과했고 자동 병합됐습니다.

[HF 개발 데이터](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-2-finite-v1)는 고정 태그 `next60-2-finite-v1`로 게시했습니다. 다시 내려받은 38개 파일이 원본과 같고 payload 해시 36개가 모두 일치했습니다. 현재 HF 미리보기 서버는 HTTP500으로 준비 중이라고 응답해 실제 미리보기 수용은 아직 확인되지 않았습니다. 파일 직접 다운로드는 확인됐습니다. [게시 기록](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/publication-proof-111/HF-PUBLICATION.v1.json)을 보세요.

[Go reader](https://github.com/teamswyg/laya-tools/tree/main/pkg/shortclaimdata)는 한 행을 모델용 문장·정답 배열·출처 정보로 나누어 읽습니다. 고정 배열과 값 복사를 사용해 공유 가변 상태나 lock을 두지 않았습니다. 로컬 race 검사 46건과 vet가 통과했으며, 이는 자료 형식과 정보 분리 검사입니다. 작은 힌트 모델의 성능이나 메모리 절감이 입증됐다는 뜻은 아닙니다.

[적격 자료 설명](Native2-Training-KO) · [실제 원본 관측](Native2-Observation-KO) · [English](Next60-Development-EN)
