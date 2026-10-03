# Hugging Face 개발37 게시 완료

[공개 자료](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-37-finite-v1)는 작은 주장·힌트 모델의 개발 근거이며 새 모델 배포가 아닙니다.

- PR132의 macOS·Linux·보안·quality 네 CI가 통과했고 동일 소스 트리가 병합됐습니다. [실제 기록](CI-MERGE132.actual.public.v1.json)
- 검증된 자체 자료50파일을 직접 게시하고 작은 메타데이터11파일을 추가했습니다. 최종 커밋은 `e9d4ba3308e7bafcf7460ccfbd569051cec59f5f`입니다.
- 기존721경로·9태그를 보존했습니다. 기존 경로 중 루트 한글·영문 안내만 갱신하고 옛 안내도 history에 보관했습니다. 최종780파일입니다.
- 새/변경61파일과 속성 파일1개를 고정 커밋에서 받아 크기·SHA를 모두 대조했습니다. 과거 전체 파일을 다시 검증했다고 주장하지 않습니다. [읽기 대조](HF37-FIXED-READBACK.actual.public.v1.json)
- `next60-37-finite-v1` 태그가 실제 커밋으로 해석됨을 확인했습니다. 과거9태그를 유지하며 총10개입니다. [태그](HF37-TAG-COMMIT.actual.public.v1.json)
- Viewer37행의 전체 값·순서와 잘림 없음이 일치했습니다. 응답은 커밋 revision을 주지 않아 null로 기록했습니다. 불변 파일 대조와 구분합니다. [Viewer](HF37-VIEWER.actual.public.v1.json)

게시 전 첫 검사는 주석이 있는 태그 객체 해시를 실제 커밋과 혼동해 중단됐습니다. 업로드0회였고 새 소스로 수정했습니다. 원래35태그 객체 `5722d80…`와 실제 커밋 `7621cd34…`는 다르게 표시되지만 동일 자료를 가리킵니다. 옛 기록·태그를 바꾸지 않았습니다. [보존 기준](HF37-TAG-PRESERVATION-BASIS.actual.public.v1.json) · [수정 설명](HF37-ANNOTATED-TAG-GUARD-CORRECTION.actual.public.v2.json)

데이터37행·108라벨, 새 Fit0, 새 독립 가족0입니다. 후보 위치 편향과 보호 평가2,400개, 실제 전체 검증 비용 개선은 여전히 별도 과제입니다. [학습 없는 비교](../next60-development37-bias-audit/README.ko.md)를 다음 모델의 기준으로 삼습니다. 원본 본문·모델 본체·바이너리·원시 journal·사적 입력·인증정보는 게시하지 않았습니다. 이전 전문 라이선스 고지를 보존하고 미상 계보를 자동으로 허용하지 않습니다.

[English](README.en.md)
