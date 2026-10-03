# 7개 요청 버전의 실제 게시 확인

[PR115](https://github.com/teamswyg/laya-tools/pull/115)는
[필수 CI 네 개](https://github.com/teamswyg/laya-tools/actions/runs/37079477122)가
통과한 뒤 GitHub Actions가 자동 병합했습니다.
[Hugging Face의 7개 요청 버전](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-7-finite-v1)은
별도로 게시하고 고정 커밋에서 다시 검증했습니다.

소유 파일 112개를 모두 내려받아 바이트와 파일 목록을 비교했고,
manifest 자체와 체크섬 목록을 제외한 payload 110개의 체크섬을 확인했습니다.
Hub가 관리하는 `.gitattributes` 한 개는 소유 payload에 넣지 않았으며 기존과 동일합니다.
viewer는 첫 요청에서 HTTP 500을 반환했습니다. 이후 HTTP 200 응답의
7행 전체 필드가 원본 JSONL과 일치했습니다. 이전 2행·3행 태그는 그대로입니다.

- `CI-MERGE115.v1.json`: CI·자동 병합과 head/merge 트리 비교 기록.
- `HF-PUBLICATION.v3.json`: 실제 업로드·고정 다운로드·새 태그·viewer 확인 기록.
- `HF-FILE-MANIFEST.v3.json`, `HF-SHA256SUMS.v3`: 원격 게시 버전의 파일 목록과 체크섬.

자료는 7개 요청·20라벨·34개 고정 입력·98개 후보 관측입니다.
추가 모델 학습은 없었습니다. 이 기록은 자료 게시 확인이며,
모델의 정확도 개선, GPU 실행 또는 Codex 사용량 절감을 입증하지 않습니다.
PR115의 CI는 그 소스를 검증합니다. 이후 만든 Hub 카드와 게시 metadata의
검증은 별도 기록으로 구분합니다.
