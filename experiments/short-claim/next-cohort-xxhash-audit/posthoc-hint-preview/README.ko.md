# 기존 힌트 모델의 xxhash 사후 시험

작은 모델이 코드 변경 후보를 먼저 살펴볼 순서를 도울 수 있는지, 기존 공개 실패79 FP32 모델로 시험했습니다. 새 학습은 없습니다. 앞선15개 입력×5개 후보의 행동 검사는 그대로 보존합니다.

| 방식 | 추천 순서 | 1위 후보의 실제 계약 일치 |
| --- | --- | --- |
| 기존 학습 참고 모델 | legacy → roundtrip → reconstruct → digest → direct | 7/15 |
| BM25 | direct → digest → roundtrip → reconstruct → legacy | 15/15 |
| 어휘 순서 비교 | direct → reconstruct → digest → roundtrip → legacy | 15/15 |
| 고정 입력 순서 | direct → reconstruct → legacy → roundtrip → digest | 15/15 |

**이번에는 기존 모델이 더 나쁜 후보를 먼저 추천했습니다. 기본 활성화하지 않습니다.** 한 사례의 비맹검 사후 대조이며 일반 정확도나 독립 검증 성적이 아닙니다. 고정 순서의 좋은 결과도 입력 순서의 영향입니다. narrow_rule은 자연어 요청을 지원하지 않아 BM25로 대체되었으며, 별도의 규칙 성공으로 세지 않습니다. 점수는 확률이 아닙니다.

입력은 앞서 고정한 요청과5개 설명 그대로입니다. ID·Want·검사 결과·학습 역할은 모델 feature에 넣지 않았습니다. 학습 역할·새 라벨·Fit·보호2,400 평가 모두0입니다. 기존37 자료·HF37·기본 라우팅을 변경하지 않습니다.

## 저장 결과 확인

저장소 루트에서 실행합니다. Go1.27.1만 필요하며 모델을 읽거나 다운로드하지 않습니다.

```sh
bash experiments/short-claim/next-cohort-xxhash-audit/posthoc-hint-preview/verify.sh
```

전체 점수는 [관측 JSON](observations.actual.public.v1.json), source/입력/모델 pin은 [사전 고정](FREEZE.public.v1.json), 해석은 [readback](READBACK.actual.public.v1.json)에 있습니다. 원래 후보 index별 점수와 순서를 구분하고 float64 bit를 보존했습니다.

## 직접 재실행 — 선택 기능

정확한 모델을 이미 보유했다면 자신의 파일 경로로 실행합니다. 출력 전체를 저장 자료와 비교하며, 반복은 새 독립 사례가 아닙니다.

```sh
MODEL_FILE=/absolute/path/data-only8192-fp32-seed1729.hbin \
  bash experiments/short-claim/next-cohort-xxhash-audit/posthoc-hint-preview/reproduce.sh
```

모델은 [HF의 고정 버전](https://huggingface.co/JooYoon/riidolaya-shortclaim-data-effect-failed-79/tree/5bef215895b69d3f2ef4b82bb3f1970279d67f46)에 있습니다. 스크립트는 다운로드하지 않고 임시 Go 모듈에서 기존 파일을 가리키며 정확한32,792B와 SHA256을 검사합니다. 모델 본체와 개인 경로는 GitHub에 올리지 않습니다. 라이브러리는 사전 고정한 revision과 같아야 합니다.

## 메모리와 다음 개선

파일32,792B에 비해 현재 Decode는8,192개 float64 계수, 즉 계수 payload65,536B를 만듭니다. **전체 heap/RSS·GPU 메모리 측정은 아닙니다.** 현재 FP32/INT8/3진 압축은 디스크 형식이며 이 경로는 모두 float64로 풉니다. 기존 선택형 [packed 읽기](../../../../pkg/hintweights/README.ko.md)를 실제 입력에 연결하는 다음 실험은 복사본을 줄이되 기존 점수의 모든 bit와 순서를 유지하는지 먼저 검증합니다. 추천 품질 개선은 별도의 데이터·학습 실험으로 다룹니다.

실제 성공 Decode1회·Rank1회이며 내부5개 Features/Score는 고정 소스에서 유도한 수치입니다. 실패·중단까지의 완전한 per-call 계측은 아닙니다. 컴파일 포함 프로세스 시간은 추론 지연이나 속도 개선으로 해석하지 않습니다.

자체 코드 Apache-2.0, 기존 scratch 학습 자산 및 MIT/BSD 고지는 기존 출처 기록을 따릅니다. 미해결 전체 계보의 법적 인증이나 신규 재배포 적격화를 주장하지 않습니다. source의 “not a trained model” 주석은 이번 pilot에서 새 학습을 하지 않는다는 뜻이며, 참고 모델은 과거에 학습된 파일입니다.
