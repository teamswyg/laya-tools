# 다음 라운드의 개발 요청 7개

고정 입력을 실제 실행하고 따로 비교해 인정한 **요청 7개·후보 라벨 20개(긍정 7/부정 13)**입니다. 기존 3행을 바이트 그대로 보존하고 네 요청을 추가했습니다. 입력 34개와 후보 관측 98개는 요청 수와 구분합니다.

[데이터](data/train.jsonl)와 [Root의 적격화 근거](ROOT-QUALIFICATION.v1.json)를 확인할 수 있습니다. 네 새 요청의 텍스트·후보 설명·순서는 사전 고정 자료와 같습니다. 기존 source group 76/77/78을 재사용하며 모두 개발 학습용입니다. 새로운 독립 소스나 최종 평가 자료로 바꾸지 않았습니다.

[Go 입력 검사](INPUT-VALIDATION.v1.json)는 7행의 엄격한 reader 검사, 네 사전 입력과의 일치, 단위 가중치, 기존 3행 보존을 확인했습니다. 특징 계산·점수·투영·추가 Fit·원본 재실행은 0입니다. 검증기는 정답이나 적격 자격을 자동으로 부여하지 않습니다.

재현은 저장소에서 `bash scripts/verify-next60-development.sh data`를 실행하면 됩니다. Go 1.27.1이 필요하며 이 검사는 모델이나 Python을 실행하지 않습니다. 해당 PR의 CI 통과 여부는 GitHub 체크를 확인하세요.

현재 HF의 고정 공개 버전은 [3개 요청 태그](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-3-finite-v1)입니다. 이 7행 버전은 CI 확인 후 별도 태그로 게시할 계획이며, 이미 게시한 태그를 이동하지 않습니다. GitHub에는 모델 본체를 올리지 않습니다.

기존 79개 자료·실패 모델의 비활성 상태·누적 Fit 3회는 그대로입니다. 30개·60개 점검 전에는 새 corpus Fit을 하지 않습니다. [다음 10개 요청의 literal 보완](../next60-catalog10-literal-correction/TRANSITION.ko.md)은 관측·라벨 없는 제안이며, 보고 채널 계측이 필요한 여섯 입력은 계속 보류합니다.

[English](README.en.md)
