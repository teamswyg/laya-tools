# Development data verification / 개발 자료 검증

This Go command checks the frozen 30-row development release with the actual
project Reader. It can also check the 42 saved records and 111 predicates for the
last three requests. It runs no original candidate, model, feature encoding,
ranking or training. It assigns no labels or source groups.

이 Go 명령은 고정된 30행 자료를 실제 프로젝트 Reader로 확인합니다. 마지막
세 요청의 저장 기록42개·판정111개도 선택적으로 확인합니다. 원본 후보나 모델을
실행하거나 새 라벨·소스 그룹을 부여하지 않습니다.

From the repository root / 저장소 루트에서:

```sh
go run ./cmd/riido-developmentverify \
  --data experiments/short-claim/next60-development-thirty/data/train.jsonl \
  --metadata experiments/short-claim/next60-development-thirty/MATERIALIZATION.v1.json \
  --finite experiments/short-claim/next60-development-thirty/evidence/REMAINING-THREE-PREDICATES.v1.json
```

Exit0 means exact hashes, bounded row values, 30 Reader returns and eight retained
unknown metadata entries match. With `--finite`, all saved literal Wants, typed
predicate values, ordering and counts match their frozen golden values too.
Exit1 rejects input/output; exit2 rejects arguments. Diagnostics are bounded JSON
without caller text or host paths. Omitting `--finite` leaves
`finite_predicate_golden_verified=false`.

종료0은 해시·행 값·Reader30반환·미상 메타데이터8개가 일치한다는 뜻입니다.
`--finite`를 넣으면 저장된 기대값·타입별 관측값·순서·개수도 고정 근거와
일치해야 합니다. 종료1은 입력/출력 거절, 종료2는 인자 거절입니다. 진단은
개인 경로나 입력 문장을 담지 않는 크기 제한 JSON입니다. 옵션을 생략하면
111판정 검증 여부는 false로 남습니다.

The finite check verifies saved bookkeeping. Its
`original_worker_reexecution_verified` and `semantic_predicates_recomputed`
remain false. Unknown is distinct from known false, including rows that have
both. The data's 27/30 second-position bias is reported as a simple control, not
model accuracy. This command is intentionally pinned to this development release;
the general row-loading API is `pkg/shortclaimdata`.

판정 검사는 저장 기록의 일관성 검사입니다. 원본 재실행·의미 판정 재계산
플래그는 false입니다. 알려진 거짓과 미상을 구분하고 둘이 함께 있어도
보존합니다. 두 번째 후보가 정답인27/30의 편향을 모델 정확도로 해석하지
않습니다. 이 명령은 해당 고정 버전용이며 일반 행 읽기 API는
`pkg/shortclaimdata`입니다.

To reproduce data and run owned synthetic failure controls offline / 자료
재생성과 소유한 가상 실패 검사를 함께 실행하려면:

```sh
bash scripts/verify-next60-thirty.sh
```

See [한국어 설명](../../docs/wiki/Next60-Development-KO.md) and
[English guide](../../docs/wiki/Next60-Development-EN.md).
