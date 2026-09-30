# 네이티브 모델 다시 만들기

[English](model-build.md) · 한국어

일반 사용자는 `riidolaya setup`을 실행합니다. Python은 실행 의존성이 아닙니다.

공개 INT8 모델은 laya 0.3.21, torch 2.14.0, transformers 5.17.0, onnx 1.23.1, onnxscript 0.7.2, onnxruntime 1.30.0으로 변환했습니다. 별도 Python 환경에서 변환합니다. NOTICE의 고정 Hugging Face revision을 `huggingface_hub.snapshot_download`의 revision=SHA로 지정해 model.safetensors, rl_agent_config.json, tokenizer/*, encoder/*를 받습니다. 원격 사용자 정의 모델 코드를 로딩하지 않습니다.

```sh
python scripts/export_onnx.py --model /path/to/checkpoint --output /path/to/laya.onnx --quantize
```

opset 18 및 동적 배치·시퀀스·선택지 수를 사용합니다. INT8 그래프는 MatMul 가중치를 채널별로 양자화하고 활성값은 float32로 남깁니다. 전체 INT8 추론이라고 표현하면 안 됩니다. FP32 export는 별도의 laya.onnx.data를 사용하므로 개발 검사에서는 그래프와 같은 폴더에 유지합니다.

Go 모델 폴더 구성:

- laya.int8.onnx를 model.onnx로 복사.
- tokenizer/tokenizer.json을 tokenizer.json으로 복사.
- rl_agent_config.json을 config.json으로 복사.
- Apache-2.0 LICENSE와 고정 모델의 MODEL_CARD.md 포함.
- 원본 NOTICE(laya-code 저작권·출처 포함)를 보존하고 MODIFICATIONS.md와 PROVENANCE.json 추가, 변환 ONNX doc_string에 고지 표시.

토크나이저 reference parity와 실제 positive/negative 판단 검사:

```sh
LAYA_MODEL_DIR=/path/to/model LAYA_RUNTIME=/path/to/native/library go test -count=1 -v ./internal/inference
```

`testdata/tokenizer-parity.json`은 들여쓰기·NFC·한국어·이모지·특수 토큰을 포함한 직접 작성 입력의 Python SDK 기준 토큰입니다. 고정 토크나이저를 Go로 구현하며, 이 영어 NFC/ByteLevel/BPE 형식만 지원합니다. 임의의 다국어 모델을 자동 지원하지 않습니다.

models-v2는 위 8개 파일을 평평한 구조로 포함합니다. `scripts/package_model_notices.py`는 원래 v1 export에 누락된 귀속 표시를 더하고 ONNX 직렬화 그래프가 같음을 확인합니다. onnx가 설치된 유지보수 환경에서 사용하며 재학습·재양자화하지 않습니다.

tar 헤더에서 호스트 메타데이터를 제거합니다. 아카이브와 각 파일의 SHA-256을 계산하고 새 모델 버전을 배포한 뒤, CI PR로 internal/assets/manifest.json을 갱신합니다. 기존 고정 파일을 덮어쓰지 않습니다. CI는 다운로드 무결성을 확인하며, 라이선스·출처는 아카이브에 포함합니다. 다른 프레임워크 버전으로 다시 변환하면 바이트와 예측이 달라질 수 있으므로 새 모델 파일로 취급합니다.
