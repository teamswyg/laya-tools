# Laya 구현·변환 진단 18

**이번 12개 입력에서 Go와 Python의 INT8 출력은 일치했다. 원본 가중치 FP32 실행과 INT8 사이에는 최대 20.76%p의 확률 차이가 있었다.** 이는 품질 평가가 아닌 구현 진단이다. 독립 질문 2,400개 평가를 대체하지 않으며, 실사용 개선·보정된 확률·원본 모델의 우수성을 입증하지 않는다.

## 무엇을 비교했나

직접 작성한 코드·한국어·공백·유니코드·특수 토큰·긴 입력 6개를 두 선택지 순서로 실행했다. 이전 진단의 입력과 동일한지 확인했고, 결과를 보기 전에 [계획](plan-18.json)과 [자산 해시](assets-18.json)를 커밋했다. 정답 라벨은 없으며 CoSQA 등 원천 자료, 학습·검증·최종·reserve 파일을 열지 않는다.

1. Go와 설치된 `laya==0.3.21`의 `build_sequence`가 만든 토큰과 선택지 위치를 비교한다.
2. 같은 INT8 ONNX 파일을 Go와 Python ONNX Runtime 1.30.0 CPU에서 실행하고, 같은 온도 보정을 적용한 확률을 비교한다. 참조 스크립트는 upstream sequence builder와 직접 ORT 실행을 사용한다. Python `ONNXAgent`의 전체 API·훅·배치 동작을 시험한 것은 아니다.
3. `tindang/laya-code`의 고정 revision `25f97e5a2ec5f8cf7218a4f67504367d8832e1fe` 원본 F16 저장 가중치를 FP32 CPU로 실행하여 INT8과 비교한다. 설치된 Laya 모델 구현을 사용하고 strict state loading을 적용한다. 모델 저장소의 Python 코드를 다운로드하거나 실행하지 않는다.

## 결과

| 지표 | 결과 |
|---|---:|
| 토큰·선택지 위치 정확히 일치 | 12/12 |
| Go ↔ Python INT8 최대 확률 차이 | 1.73 × 10⁻¹⁸ |
| Go ↔ Python 최종 선택 불일치 | 0/12 |
| 원본 FP32 ↔ INT8 최대 확률 차이 | 0.207607, 약 20.76%p |
| 원본 FP32 ↔ INT8 평균 사례별 최대 차이 | 0.019952, 약 2.00%p |
| 원본 FP32 ↔ INT8 최종 선택 불일치 | 0/12 |

Go 재실행의 시간 외 모든 비교 수치는 동일했다. Python 참조의 전체 반복 실행은 하지 않았다. 참조 파일 해시는 시간 필드도 포함하므로 새로 생성하면 달라질 수 있다.

Go 이식 오류라는 설명은 이 입력 범위에서 지지되지 않는다. 하지만 FP32↔INT8 차이는 **PyTorch 실행·ONNX 변환·양자화·커널 차이가 합쳐진 값**이다. 동일 FP32 ONNX 대조군이 없으므로 순수한 양자화 손실이라고 단정하지 않는다. 12개에서 선택이 같았다는 사실도 실제 질문의 순위·AUC·기준값 통과 여부가 보존된다는 뜻은 아니다.

## 자원과 재현

Apple M4 Pro, macOS ARM64, CPU 4스레드, 최대 512토큰. 두 프로그램을 순서대로 실행했다. Python 참조는 5.69초 wall, 최대 RSS 3,143,892,992바이트(약 2.93GiB); Go 비교기는 2.31초 wall, 최대 RSS 1,600,077,824바이트(약 1.49GiB). Python은 INT8과 원본을 모두 실행하고 Go는 INT8만 실행하므로 이 시간·메모리 차이를 언어 성능 비교로 해석하지 않는다. 한 번의 외부 측정이며 pprof heap이나 GPU 메모리가 아니다.

호출 사이에 15분·8GiB RSS 제한을 확인한다. 호출 내부의 강제 제한은 아니다. maintainer 환경에 ONNX Runtime 1.30.0 wheel을 추가했고, 원본 선택 파일 8개 약 846MB를 고정 revision에서 내려받아 HF 해시 검증을 통과했다. 검증 경고의 누락 4개는 다운로드하지 않은 원격 파일, 추가 항목은 로컬 Hub 메타데이터다. 사용 파일은 별도 해시로 고정했다. 코드 모델의 Apache-2.0 LICENSE/NOTICE를 로컬에 보존하며 새 모델을 배포하지 않는다.

```sh
# 기존 maintainer 가상환경에 onnxruntime==1.30.0 필요; 사용자 Go 런타임 요구 사항 아님.
hf download tindang/laya-code README.md LICENSE NOTICE rl_agent_config.json encoder/config.json tokenizer/tokenizer.json tokenizer/tokenizer_config.json model.safetensors --revision 25f97e5a2ec5f8cf7218a4f67504367d8832e1fe --local-dir .cache/laya-code-source-18
python scripts/training/laya_parity_reference.py --model-dir .cache/models-v2/code --source-dir .cache/laya-code-source-18 --out .cache/parity18-reference.json
go run ./cmd/riido-parityprobe --reference .cache/parity18-reference.json --model-dir .cache/models-v2/code --runtime /path/to/pinned/libonnxruntime.1.30.0.dylib
```

실행 환경 버전·집계는 [결과 JSON](results-18.json)에 있다. 기존 참조 파일을 덮어쓰지 않으므로 재생성할 때 새 출력 경로를 사용한다. 비교기는 동일 자산 해시, 확률 유효성, 입력 일치와 확률 허용오차 10⁻⁵를 검사한다. 구현 일치 통과와 모델 품질 통과는 별개다.

다음은 **원본 FP32 모델을 실험 17b와 동일한 검증 질문·입력·범위로 평가**하는 것이다. 이것으로 점수 차이가 실제 관련성 구분에 영향을 주는지 확인한 뒤 학습·증류 방향을 정한다. 미사용 reserve 2,403개·2,401개는 계속 보존한다.
