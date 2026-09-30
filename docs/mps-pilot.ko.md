# 실제 MPS 학습 pilot — 2026-09-30

[English](mps-pilot.en.md) · [이슈 #11](https://github.com/teamswyg/laya-tools/issues/11)

이번에는 합성 텐서 probe를 넘어 **원본 Laya를 실제로 MPS에서 학습**했습니다. encoder와 별도 action head는 동결하고 head/type_emb/scorer 26,248,193개 파라미터를 학습 대상으로 삼았습니다. 두 과제는 같은 원본에서 따로 출발했습니다. 사용자 런타임은 Go이며 아래 Python은 유지보수 학습·평가 전용입니다.

## 실험 구성과 결과

자체 작성 영어 합성 개발 사례 88개(난이도 48, 분할 40)를 사용했습니다. [데이터 설명](../benchmarks/training/README.md)과 [모든 실측 JSON](../benchmarks/results/mps-pilot-20260930)을 공개합니다. 기존 36개 골든셋은 학습에 사용하지 않았습니다. 새 사례가 계획한 독립 400그룹 골든셋을 대신하지 않습니다.

각 과제에서 학습률 1e-5/5e-5를 비교했습니다. 먼저 최대 2 epoch를 수행한 뒤, 최대 5 epoch 실행으로 확장했습니다. 모두 원본에서 출발하고 validation NLL로 체크포인트를 골랐습니다. 이후 별도 개발 calibration partition에서 temperature를 0.5~5.0 범위로 보정했습니다. 기존 0.9 수용 기준을 낮추지 않았습니다.

| 과제 | 원본 개발 평가 | 선택한 학습 모델 | 0.9 이상 수용 | 선택 |
|---|---:|---:|---:|---|
| 난이도 | 6/9 | 6/9 | 0/9 | 5e-5, epoch 5 |
| 분할 방식 | 5/8 | 4/8 | 0/8 | 5e-5, epoch 2 |

분할은 더 오래 학습한 checkpoint의 validation 손실이 나빠져 epoch 2가 유지됐습니다. **정답 수 개선과 유효한 추천 증가를 입증하지 못했습니다.** 분할 정답 수 악화도 그대로 기록했습니다. 원본 확률은 upstream 보정, 학습 모델 확률은 새 pilot 보정을 쓰므로 확률 지표 개선이 가중치 학습만의 효과라고 주장하지 않습니다.

같은 개념 패턴이 분할 사이에 있고, 첫 test 결과를 본 뒤 실험 길이를 늘렸으므로 전부 개발 평가입니다. 독립 저장소 일반화·실제 코딩 성공·비용 절감의 증거가 아닙니다. 사람이 예상한 수준을 자동 실행의 정답으로 취급하지 않습니다.

## 메모리와 저장 공간

24GiB Mac에서 한 모델씩 실행했습니다. 확장 run의 프로세스 최고 RSS는 약 2.75GiB, 단계 종료 MPS driver 샘플 최대는 약 3.04GiB입니다. 통합 메모리라 서로 겹치며 합산할 수 없습니다. 순간 GPU peak를 지속 측정한 수치도 아닙니다. 난이도 약 42초, 분할 약 32초는 로딩·평가·해시·저장을 포함하는 이번 작은 run 시간이며 처리량 벤치마크가 아닙니다.

각 run은 프로세스 RSS 5GiB, 시스템 메모리 여유 20%, 디스크 여유 30GiB 및 최대 20분을 검사하며 MPS allocator 한도도 제한했습니다. 학습 중 다른 앱을 종료하거나 전역 설정을 바꾸지 않았습니다. 실제 사용 중인 Mac의 메모리 상황에 따라 같은 설정도 중단될 수 있습니다.

원본 가중치는 약 804MiB를 한 번 저장합니다. 공개 모델은 약 100MiB의 FP32 head 교체 가중치만 포함합니다. 원본 전체를 복제하지 않습니다. optimizer/RNG checkpoint는 로컬에만 보존하며 자동 업로드하지 않습니다. 여러 버전의 캐시를 무제한 쌓지 말고 새 실험 전 디스크 여유를 확인합니다. 현재 도구는 사용자 파일이나 이전 checkpoint를 자동 삭제하지 않습니다.

## 재현

Python 3.14.7 / PyTorch 2.14.0 / Laya SDK 0.3.21 환경을 사용했고 [전체 실험 의존성](../scripts/training/requirements-pilot-macos-arm64.txt)을 고정했습니다. 이 목록은 패키지 버전 목록이며 모든 wheel의 해시를 고정한 공급망 lock은 아닙니다. 원본 파일 해시는 배포 provenance에 기록합니다.

```sh
python3 -m venv .cache/mps-training-venv
.cache/mps-training-venv/bin/python -m pip install -r scripts/training/requirements-pilot-macos-arm64.txt
.cache/mps-training-venv/bin/python scripts/training/train_pilot.py \
  --base .cache/training/base --task difficulty \
  --out .cache/training/new-difficulty-version --epochs 5 --minutes 20
```

`--base`에는 revision `55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851`의 원본 model.safetensors, rl_agent_config.json, encoder/config.json과 tokenizer 디렉터리가 필요합니다. 원본 가중치 SHA-256은 코드에서 검증합니다. 기존 출력 디렉터리는 거절합니다. 다른 과제는 `--task decomposition`을 사용합니다.

학습 실행은 CPU fallback 없이 finite loss/gradient, 실제 head 갱신, 동결 파라미터 전체 해시 보존, 저장 head의 재로딩 추론 일치를 검사했습니다. optimizer 상태를 저장했지만 **학습 재개 동등성은 아직 검증하지 않았습니다**. ONNX·INT8·Go 변환도 미완료입니다.

## Hugging Face 배포와 다음 버전

배포 후보는 `JooYoon/riidolaya-difficulty-pilot-v0.1`, `JooYoon/riidolaya-decomposition-pilot-v0.1`입니다. 실제 게시 여부·revision은 이슈 #11의 결과 댓글을 확인합니다. 이름은 배포 버전이며 로컬 두 번째 튜닝 run의 `v0.2`와 별개입니다.

`package_pilot.py`는 라이선스·원본 모델 카드·NOTICE·변경 사항·학습 데이터·실측·의존성·해시를 묶습니다. `publish_pilot.py`는 명시적 파일 목록만 허용하고 optimizer, 추가 파일, symlink, 변조된 가중치를 거절합니다. 게시 시 정확한 소스 commit의 성공한 CI와 JooYoon 인증을 확인하며 이미 존재하는 버전 repo를 덮어쓰지 않습니다. 토큰은 환경/공식 Hub 로그인 저장소에서만 읽고 코드·로그·모델에 넣지 않습니다.

Laya와 ModernBERT의 공개 모델 카드가 Apache-2.0을 선언함을 확인했고 SDK 라이선스도 확인했습니다. 새 데이터는 자체 작성이며 비상업/동일조건변경허락 외부 데이터는 추가하지 않았습니다. 선언된 라이선스 기준으로 출처와 전문을 보존합니다. upstream의 모든 사전학습 데이터 권리를 독립적으로 입증했다는 뜻은 아닙니다.

이것은 head 교체 파라미터이며 LoRA나 완성된 독립 모델이 아닙니다. 정확한 원본과 FP32로 결합해야 합니다. `evaluate_head.py`가 유지보수자용 CPU 참조 로딩을 제공합니다. 현재 Go 런타임의 기본 모델을 변경하거나 자동 작업 실행을 켜지 않습니다.

다음 튜닝은 epoch를 무한히 늘리기보다 공개 실제 작업의 근거와 레이블을 개선하고 작업군을 분리하는 데 우선순위를 둡니다. 새 릴리스는 별도 버전으로 같은 검증·CI·라이선스 절차를 거칩니다. 이 도구가 백그라운드에서 자동 학습·게시하도록 예약된 것은 아닙니다.

[후속 PDCA 튜닝과 다음 평가 준비](pdca-tuning.ko.md)
