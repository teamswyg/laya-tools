# 로컬 MPS 학습 준비

[English](mps-training.en.md) · [계획 이슈 #11](https://github.com/teamswyg/laya-tools/issues/11)

이 문서는 최초 준비 검사 기록입니다. 이후 완료된 실제 모델 학습은 [MPS pilot 결과](mps-pilot.ko.md)를 확인하세요.

일반 사용자는 Go 바이너리만 사용합니다. 이 문서의 Python은 유지보수자의 학습·변환 실험에만 필요합니다. `riidolaya`의 실행 의존성이나 기본 CI에 추가하지 않습니다.

## 이번에 확인한 것

2026-09-30 Apple Silicon, 24 GiB 통합 메모리 환경의 별도 가상환경에서 Python 3.14.7 / PyTorch 2.14.0으로 **MPS 합성 텐서 학습 3회**를 실행했습니다. CPU fallback을 끄고 실제 forward/backward, 유한한 기울기, 학습 헤드의 가중치 변경, 동결한 인코더의 불변을 확인했습니다. [경로를 포함하지 않는 결과 JSON](../scripts/training/mps-probe-20260930.json)을 보존합니다.

- 손실: 1.45736 → 1.41734 → 1.37812. 동일한 작은 합성 배치에 대한 값이며 정확도나 일반화 성능이 아닙니다.
- 동기화한 3회 구간: 약 1.36초. 기울기 검사와 메모리 샘플링을 포함하고 실행 환경 준비 시간은 제외합니다. 워밍업 없는 작은 probe이므로 학습 처리량 벤치마크가 아닙니다.
- 프로세스 최대 RSS: 약 430 MiB. MPS driver 단계 종료 샘플 최대: 약 18.7 MiB.
- 단계 종료 시점의 MPS 메모리 샘플은 순간 최대 GPU 메모리가 아닙니다. 통합 메모리이므로 RSS와 MPS 메모리를 합산하지 않습니다.

**Laya 모델을 로드하거나 학습한 결과가 아닙니다.** 원본 가중치는 저장소 캐시에서 찾지 못했으며 대규모 다운로드와 전체 학습을 시작하지 않았습니다. 작은 연산의 성공이 ModernBERT 전체 연산 지원이나 모델 학습 메모리를 보장하지 않습니다. NumPy 없는 최소 probe 환경에서는 PyTorch의 NumPy 초기화 경고가 나오지만 이 probe는 NumPy 변환을 사용하지 않습니다.

## 같은 준비 검사 실행

저장소 루트에서 실행합니다. 버전 파일은 이번 macOS arm64 probe 환경의 버전 고정 목록이며, 전체 Laya 학습 의존성 잠금 파일은 아닙니다.

```sh
python3 -m venv .cache/mps-training-venv
.cache/mps-training-venv/bin/python -m pip install -r scripts/training/requirements-probe-macos-arm64.txt
.cache/mps-training-venv/bin/python scripts/training/mps_probe.py --steps 3
.cache/mps-training-venv/bin/python -m unittest discover -s scripts/training -p 'test_*.py'
```

Probe는 최대 10회, FP32, 작은 고정 텐서만 사용하며 MPS 할당 한도를 권장 작업 메모리의 25%로 설정합니다. CPU fallback이 켜져 있거나 MPS가 없으면 실패합니다. 모델 다운로드, 네트워크 호출, 체크포인트 쓰기는 수행하지 않습니다. 합성 module 기반 단위 테스트 2개는 정책 계약을 검사하며 실제 Laya 통합 테스트가 아닙니다.

## 실제 Laya 학습 전 남은 단계

1. [고정 원본 revision](https://huggingface.co/convaiinnovations/laya/tree/55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851)의 가중치·encoder·tokenizer·config 파일을 확보하고 각각의 SHA-256, 출처, 라이선스를 기록합니다. INT8 ONNX 배포본을 학습 원본으로 사용하지 않습니다. [준비 메타데이터](../scripts/training/BASE_MODEL.json)에 검토 기준을 기록했습니다.
2. 전체 학습 의존성과 artifact hash를 잠급니다. 검토한 `laya==0.3.21` wheel은 실행 설치하지 않고 소스 구조만 읽었습니다. 공식 모델의 입력 포맷과 Go 입력/선택지 순서가 같은지 먼저 검사합니다.
3. `head_policy.py`로 `head`, `type_emb`, `scorer`만 학습합니다. **encoder 전체와 final norm, `act_head`는 동결**합니다. `act_head`는 답변/에스컬레이션을 다루는 별도 출력이므로 선택지 분류 라벨만으로 학습하지 않습니다. 호출 순서는 `model.train()` 이후 정책 적용이며, 동결 encoder는 `eval()`로 dropout도 끕니다. 이후 다시 `model.train()`을 호출했다면 정책을 재적용합니다.
4. 공개 합성 입력 한 건으로 실제 Laya MPS forward/backward 1회부터 확인합니다. FP32, 배치 1, 짧은 입력, compile off, CPU fallback off로 시작합니다. 유한한 loss/gradient, 지정 가중치 갱신과 동결 가중치 불변을 확인한 후 10/50회로 늘립니다. OOM 무한 재시도는 하지 않습니다.
5. head만으로 부족하다는 검증이 나온 뒤 encoder 상위 층의 점진적 해제, activation checkpointing, gradient accumulation을 비교합니다. encoder가 완전히 동결된 초기 단계에 필요한 최적화와 상위 층을 학습하는 단계를 구분합니다.
6. 학습·검증·보정·최종 테스트를 작업군별로 분리합니다. 기존 36개 개발 골든셋은 학습에 넣지 않습니다. 난이도와 분할 적합성은 별도 데이터와 지표로 평가합니다. Laya는 후보 선택/평가 모델이며, 분할 계획 생성은 상위 생성 에이전트가 담당합니다.
7. 별도 보정셋으로 확률을 보정하고 오래된 `temperature_by_options`가 새 보정을 덮지 않는지 검사합니다. CPU/MPS → FP32 ONNX → INT8 → Go의 선택·확률·보류 결과를 비교한 뒤 별도 버전 artifact를 만듭니다. 현행 모델을 덮어쓰지 않습니다.

원본 모델 전체 학습, resume 검증, 새 도메인 정확도, ONNX 재변환, Go parity 검사는 아직 수행하지 않았습니다. 이 환경 준비를 학습 성공이나 비용 절감의 증거로 취급하면 안 됩니다.

근거: [PyTorch MPS 문서](https://docs.pytorch.org/docs/2.14/notes/mps.html), [Laya SDK 0.3.21](https://pypi.org/project/laya/0.3.21/), [외부 MPS 실험 기록](https://github.com/pilotspace/laya-codex/blob/580bc73c2ed95fd319db93ef725f30bf35047428/finetune/runs/r1-setup.md). 외부의 전체 모델 학습 결과와 이 저장소의 합성 probe 결과는 서로 다른 증거입니다.
