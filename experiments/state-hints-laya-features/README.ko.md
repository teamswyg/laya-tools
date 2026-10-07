# 동결 학습 특징 추출

원본 train840가족1,680한영 행에서 별도의 원래 라벨 기반 Go지도 probe에 쓸
특징만 추출하는 오프라인 유지보수 도구입니다. Fit·loss·teacher가짜라벨·보정·최종
test·개인 태스크 읽기·모델 선택은 없습니다. 제품 런타임은Go이며 Python은
유지보수 예외입니다. 기존 동결 참고 실험과 원본 corpus바이트는 바꾸지 않습니다.

기본 모델·설정/tokenizer4개 핀·기존V1 QUESTION/순서는 동결 참고와 같습니다.
그 정의는V4범위 규칙과 여전히 다릅니다. sidecar의 expected_intent는 원래 라벨을
메타데이터로 복사할 뿐 tokenizer나 backbone입력에 넣지 않습니다. 모델에는
문장에서 만든ID와 고정QUESTION만 전달하며 라벨·단위·분할 배정은 전달하지 않습니다.

```sh
python experiments/state-hints-laya-features/extract_training.py \
  --base BASE_DIR --train TRAIN840.jsonl \
  --plan PROBE_PLAN.json --plan-sha256 PLAN_SHA \
  --reference-plan FROZEN_REFERENCE_PLAN.json --check
```

`--check`는 바이트 핀과840한영 가족/클래스당105가족을 확인하고 **backbone을
만들기 전에1,680행 전체를 로컬tokenizer로 확인**합니다. 고정512토큰/192head처리와
tokenizer만 사용하는 비절단 결과를 비교합니다. 입력이나 선택지가 하나라도
잘리면 backbone호출0으로 실패합니다. 행을 빼거나 줄이거나 다시 라벨링하거나
더 큰 모델 입력으로 재시도하지 않습니다. tokenizerimport가 라이브러리 정의를
읽더라도 모델을 만들거나 실행하지 않습니다. check는 특징·결과 파일을 쓰지 않습니다.

성공한 뒤에만 ROOT가
`--run --out .cache/statehint-laya-features/NEW-RUN`을 승인해 실행합니다.
동결 MPS/FP32작업자1개가 정확히1,680forward를 하여 행별8×1,024 pre-final특징을
받습니다. 미리 만든token ID를 재사용하고32KiB행 하나씩 파일에 씁니다. delta점수·
teacher예측라벨·전체backbone복사·55MiB특징buffer를 만들지 않습니다. MPS필수,
오프라인로딩/fallback0,프로세스최대RSS5GiB,MPS비율25%,시스템가용성20%를 고정합니다.
기존foundation·추출캐시와 새출력64MiB예약은 preflight전에512MiB태스크캐시 한도 안에
들어야 합니다. 다른캐시는 파일크기만 확인하고 내용을 열지 않습니다.

`features.f32le`는 정확히55,050,240바이트입니다. `sidecar.json`은
`riido-statehint-laya-feature-sidecar-v1`,base/corpus/instruction/feature해시,
고정의도순서,rows_count1680,shape[1680,8,1024],dtype`float32_little_endian`과
id·locale·text_sha256·expected_intent만 있는 순서별 행을 담습니다. Go가 원본
corpus의 라벨/ID와 대조하고 기존663fit/177dev전체그룹 분할을 적용합니다.
validation은 앞서 캡처한 파일을 별도로 재사용하며 이 도구에서는 열지 않습니다.

새 폴더0700/파일0600,전체출력64MiB상한입니다. 수치·추출 실패는 비공개 부분특징과
`FAILED.json`을 남기고 성공sidecar나 headFit은 없습니다. preflight실패는 결과/캐시를
만들기 전에 종료합니다. 자원표본은 순간GPU최고값이 아니며 통합RSS/MPS는 겹칩니다.
특징·행메타데이터를 업로드하거나Git에 넣지 않습니다. 별도Gohead도 새 문장에는
동결842MBbackbone이 필요하므로 독립 저메모리 의미모델이 아닙니다.

```sh
PYTHONDONTWRITEBYTECODE=1 python -m unittest discover \
  -s experiments/state-hints-laya-features -p 'test_*.py'
```

직접 만든tokenizer/forward stub으로 전체바이트·순서·라벨분리·전체preflight실패·
유한특징·원본쌍/핀·캐시/출력한도를 검사합니다. 실제 모델을 호출하지 않으므로
사전학습 정확도나MPS실행을 증명하지 않습니다. 실제tokenizer확인·추출은 소스검토와
자원확인 뒤 ROOT가 수행합니다. Apache-2.0소스/설정출처를 보존합니다.

## 실제 추출·지도 probe run01

ROOT가 tokenizer preflight121–312토큰/절단0을 확인한 뒤 동결 MPS forward1,680회를
완료했습니다. 특징은55,050,240바이트,producer전체56,300,062바이트로64MiB이하입니다.
원래 기대 라벨은 메타데이터에만 두었으며 teacher예측이나 MPS가중치 갱신은
사용하지 않았습니다. MPS프로세스를 종료한 뒤 별도의 순수Gohead를 학습했습니다.

| 단계 | 전체 프로세스 시간 | 프로세스 최대RSS | 범위 |
| --- | ---: | ---: | --- |
| 학습 특징 추출 |128.79초|2,976,038,912바이트 (2.77GiB)|MPS/FP32기본 모델1개,1,680행 |
| Go공유 head실행 |1.72초|24,182,784바이트 (23.06MiB)|기존 특징 캐시,backbone호출0 |

추출 작업자 시간125.865초는 로딩·자원확인·추출을 포함합니다. native전체 기본 모델
forward는69.980초,평균41.65ms였습니다. 추출 CPU user/system은19.17초/51.80초입니다.
MPS driver/할당 표본의 최대는2,174,435,328 /1,688,613,632바이트,시스템 가용성
최소는52%였습니다. 표본은 순간 최고값을 놓칠 수 있고 통합 메모리의RSS와 겹치므로
합산하지 않습니다. 로컬1회에서 서로 다른 단계를 잰 값이며 조건을 맞춘 속도 개선이나
새 문장의 추론 지연으로 해석하지 않습니다.

Gohead는663fit가족1,326행과 원래 AI검토 라벨만 사용했습니다. 새 영 가중치,
40epochs/batch32/rate.001/decay.01/seed1729/T1,1,680업데이트이며4,320바이트
파일의 학습 후 전체Prediction·재저장 바이트가 일치했습니다. 이미 노출된 두
진단 결과에서는 목표를 충족하지 못했습니다.

| 진단 |8분류 일반 정답 |Gated P/C/Q제안 |적격성 |
| --- | ---: | ---: | --- |
| 내부dev177가족 |118/354 (33.33%)|0|포괄률/지지 부족·완료 정밀도 미정으로 실패 |
| 노출된 validation120가족 |92/240 (38.33%)|0|같은 실패,모두 보류하는 비용 기준과도 같음 |

두 언어의 P/C/Q제안은0이므로 정밀도가 미정이며 숫자0자리값을0%나100%정밀도로
해석하지 않습니다. 선택·승격·calibration·최종test접근은0입니다. 유용한 완료 개선이나
저메모리 새 문장 적격성을 확보하지 못했습니다. 작은head도 새 문장에는
842,609,210바이트 동결backbone이 필요합니다. 특징·행메타데이터는 비공개로 유지하고
HF/Git자료에 넣지 않습니다. 기존 실험·코드·기준은 변경하지 않았습니다.
