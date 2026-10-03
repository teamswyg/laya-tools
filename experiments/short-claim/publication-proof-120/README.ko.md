# HF23 배포와 현재 개발 자료

이 프로젝트는 아주 작은 모델이 짧은 힌트를 자주 제시해 LLM 작업을 돕도록 만드는 실험입니다. 지금은 요청 문구와 후보 설명을 고정된 실제 관측에 연결하는 자료를 늘리고 있습니다. **공개 고정 자료는 23개 요청·67개 후보 라벨(양성 23/음성 44)**입니다. 별도로 Root가 채택한 개발 풀은 **25개 요청·73개 라벨(양성 25/음성 48)**이며, 마지막 두 요청의 원천·별칭·보조 코드 연결을 확인한 뒤 다음 자료를 생성합니다. 25행 자료는 아직 생성·배포하지 않았습니다. 새 corpus Fit·모델 추론은 0회이고, 적격 30개 이전 새 학습 금지를 유지합니다.

[HF23 고정 버전](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-23-finite-v1)의 커밋은 `0d964a547708db596b13b0eae18d2c93dd3e3ac4`입니다. 고정 다운로드에서 소유 401파일·2,989,766B와 399개 payload 체크섬이 일치했습니다. 원격 402파일 중 나머지 하나는 이전과 같은 HF 관리 `.gitattributes`(2,504B)입니다. Viewer는 HTTP200·관측 23행의 순서/모든 필드 일치·잘린 셀 0을 확인했습니다. 응답에 커밋·전체 행수 필드가 없으므로, 고정 커밋 파일 검증과 현재 Viewer 대응을 구분합니다. 최초 HTTP500 기록과 기존 2/3/7/16/21 태그도 보존했습니다.

재현할 때는 태그를 지정하고 `releases/next60-23-finite-v1/next60-development-twentythree/data/train.jsonl`을 선택합니다. 파일은 31,493B·SHA256 `39edb1bb60e88d56cb2fec271b511a5ce09a0ff8bd33ce21d0dda7defb3ce362`입니다. 이전 21행·28,800B prefix가 그대로이며 실제 프로젝트 reader 23회 대응과 46개 저장 checkpoint를 확인했습니다. 고정 입력 113개·원관측 335개·선택 관측 331개는 요청 수 23과 다른 분모입니다.

[PR119](https://github.com/teamswyg/laya-tools/pull/119)의 head `9d4b0b5e39628a8d7a2bb50a3eb9f9b018ccd0ef`에서 [필수 CI 네 개](https://github.com/teamswyg/laya-tools/actions/runs/37093356839)가 통과했고 봇이 `67319e639b52d302292a0abdbbd83149c1fb0d99`로 병합했습니다. 두 트리는 같습니다. 추가 Go CI는 저장된 비교와 23행 생성/reader 대응을 재현합니다. 기존 Laya native inference는 별도 CI 단계입니다. 이번 자료 추가의 모델·GPU·Codex 비용 절감 효과는 아직 측정하지 않았습니다.

공개 자료에서 항상 후보 인덱스 1을 고르면 20/23, 항상 음성으로 보면 44/67입니다. 별도 25개 풀은 22/25와 48/73입니다. 서로 다른 분모의 산술 대조군이며 모델 정확도가 아닙니다. 후보 위치·어휘·연결 원천 대조가 필요합니다. 알려진 반례와 함께 있는 미상 조건도 유지합니다. 추가 두 요청의 미상 2개와 기존 B_code 미상 1개를 거짓으로 바꾸지 않습니다. 기존 79자료·역사 Fit3/논리모델3·실패 모델 비활성, 60 checkpoint·주장 도메인별 보호 2,400요청·효용 5% 기준은 그대로입니다.

[한국어 Wiki](https://github.com/teamswyg/laya-tools/wiki/Next60-Development-KO)와 [English Wiki](https://github.com/teamswyg/laya-tools/wiki/Next60-Development-EN)의 master는 `5d7a5090cabd7c9e44153c80ca0fb0254013b83d`입니다. [Issue19 댓글](https://github.com/teamswyg/laya-tools/issues/19#issuecomment-5963281223)의 v7 원문과 저장 readback 바이트도 같습니다. 이 묶음은 이미 저장된 Root 근거에서 작성한 공개 기록입니다. 작성자가 원본·모델·배포를 새로 실행하거나 정답·역할·권리를 재승인하지 않았습니다. 자체 문서/metadata는 Apache2이며 원천 라이선스를 바꾸지 않습니다. Git120 게시·CI와 이후 적격화 상태는 Root가 별도 근거로 확정합니다.
