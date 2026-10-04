# 실제 모델의 packed 메모리 비용

기존 Go `hintweights` 읽기를 같은 공개 FP32 모델·요청·후보5개에 연결했습니다. 점수15개의 모든 float64 bit와 후보 순서가 기존 Decode 경로와 같습니다. **나쁜 legacy-first 추천도 그대로입니다.** 저장 방식 개선과 추천 품질은 별개이며 모델은 계속 비활성입니다.

| 측정 범위 | 기존 Decode/참고 계산 | packed 계수 직접 읽기 | packed Score + 변환 버퍼 |
| --- | ---: | ---: | ---: |
| 모델 준비 Go 할당/회 | 65,536B·1회 | 41,024B·2회 | 같은 View 생성 사용 |
| 준비된 후보5개 점수 Go 할당/회 | 0B·0회 | 0B·0회 | 0B·0회 |
| 전체 후보5개 Rank Go 할당/회 | 약130,616B·253회 | 약127,704B·239회 | 약127,704B·239회 |
| 논리 가중치 payload | 65,536B | 32,792B | 32,792B |
| 추가 재사용 변환 scratch | 없음 | 없음 | 20,976B |

View payload에는 모델 header24B가 들어갑니다. 원본 파일 버퍼·Go 구조체/slice header·할당 단위 여유·미리 준비한 feature 저장소·전체 heap/RSS·GPU는 위 논리 payload에 포함되지 않습니다. 두 표현을 같은 프로세스에 유지했으므로 전체 메모리가 줄었다고 주장하지 않습니다. 변환 scratch까지 더하면 참고 계수 대비 논리 차이는11,768B이며, scratch 없는 계수 경로는32,744B 차이입니다.

고정 입력1개를4회 교대 순서로32개 구간에서 측정했습니다. 준비64회, 준비된 점수1,024회, 전체 Rank128회씩입니다. 준비된 점수의 관측 중앙값은 참고4.24µs, 계수 경로18.23µs, 변환 경로6.78µs였습니다. 전체 Rank는 약448/456/445µs로 비슷했습니다. **순수 저장 방식 속도 비교가 아닙니다.** 계수 경로의 매-feature 계측 비용, 참고 Rank의 반복 검증, 고정 입력만 지원하는 packed adapter가 섞여 있습니다. 반복 횟수는 독립 사례 수가 아니며 production 지연·전체 작업·Codex 비용 개선으로 해석하지 않습니다.

## pprof로 본 다음 개선 대상

별도 계측 재실행에서 CPU·allocs 프로파일을 실제 생성하고 Go pprof로 파싱했습니다. 원래 비계측 측정은 보존했습니다. sampled alloc_space에서 Features가 flat74.72%, 하위 호출 포함85.07%를 차지했습니다. 이 값은 준비·게이트·32개 구간·프로파일 자체까지 포함한 누적 할당의 표본 추정이며 peak/live/RSS가 아닙니다. 짧은 CPU 프로파일은 runtime kevent/madvise 등 영향이 커 CPU·SIMD 병목을 확정하지 않습니다. [집계와 한계](PROFILE-READBACK.actual.public.v1.json)만 공개하며 원본 프로파일은 비공개입니다.

## 확인과 선택적 재실행

Go1.27.1로 저장 결과만 확인합니다. 모델 다운로드·모델 읽기·학습은 없습니다.

```sh
bash experiments/short-claim/next-cohort-xxhash-audit/packed-real-preview/verify.sh
```

정확한 기존 모델을 보유하면 실제 측정을 다시 수행할 수 있습니다. 시각·측정값은 달라질 수 있으므로 bit·순위·호출 일정·논리 payload를 대조합니다.

```sh
MODEL_FILE=/absolute/path/data-only8192-fp32-seed1729.hbin \
  bash experiments/short-claim/next-cohort-xxhash-audit/packed-real-preview/reproduce.sh
```

고정 모델·GitHub 상호 링크는 [HF 추가 기록](https://huggingface.co/JooYoon/riidolaya-shortclaim-data-effect-failed-79/tree/6ed4e42ea3bbd0549752451e1320a434717d4c8c/followups/xxhash135-136-134b76a)에 있습니다. 기존22개 파일의 metadata와 가중치 blob을 보존하고 새 공개 기록4개만 더해 전부 byte-exact readback했습니다. 새 모델 배포나 Laya/MPS/GPU 학습이 아닙니다. 라이선스 범위는 기존 scratch 모델·Apache-2.0 코드와 기존 출처 고지를 따르며 전체 계보의 적격화를 주장하지 않습니다.

## 다음 PDCA

동일 요청·후보를 반복할 때만 caller-owned 준비본을 재사용하는 실험을 분리합니다. 기존 Features의 토큰화·정렬·충돌 누적·값 bit를 유지하며, 먼저 길이를 맞춘 AoS를 비교하고 다음으로 uint16 index/float64 value SoA를 비교합니다. 특징당 논리16B→10B 가설이며 누적 offsets는uint32로 둡니다. 공유 map·lock 없이 각 호출자가 저장소를 소유합니다. 준비 비용과 반복1/2/8/64회 총 비용·실제 len/cap을 기록합니다. 한 번만 쓰는 요청에는 준비본 보유가 오히려 늘 수 있습니다.

부모1개·비맹검 사후 자료입니다. 새 역할·학습 라벨·Fit·보호2,400 평가·기본 활성화 모두0이며 기존37/HF37를 변경하지 않습니다. 실제 관측4557B와 [사전 고정](FREEZE.public.v2.json), [readback](READBACK.actual.public.v1.json)을 구분해서 읽어주세요.
