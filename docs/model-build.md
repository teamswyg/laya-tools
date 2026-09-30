# Rebuild the native models

Normal users run `riidolaya setup`; Python is not a runtime dependency.

The published INT8 artifacts were exported with laya 0.3.21, torch 2.14.0,
transformers 5.17.0, onnx 1.23.1, onnxscript 0.7.2, onnxruntime 1.30.0.
Use an isolated Python environment for conversion. Download the exact Hugging Face
revisions listed in NOTICE (`model.safetensors`, `rl_agent_config.json`,
`tokenizer/*`, `encoder/*`) with huggingface_hub.snapshot_download and revision=SHA.
No remote custom Python model code is loaded.

```sh
python scripts/export_onnx.py --model /path/to/checkpoint --output /path/to/laya.onnx --quantize
```

The exporter uses dynamic batch, sequence and marker counts and opset 18. The
INT8 graph quantizes MatMul weights per output channel; activations remain float32.
Do not describe this as fully INT8 inference. FP32 export uses a separate
`laya.onnx.data` file. Keep that filename alongside its graph for developer tests.

For a Go model directory:

- copy `laya.int8.onnx` to `model.onnx`;
- copy `tokenizer/tokenizer.json` to `tokenizer.json`;
- copy `rl_agent_config.json` to `config.json`;
- include Apache-2.0 LICENSE and the pinned model's MODEL_CARD.md;
- preserve upstream NOTICE (including the laya-code copyright/attribution),
  add MODIFICATIONS.md and PROVENANCE.json, and mark the converted ONNX doc_string.

Run tokenizer reference parity and native positive/negative decision tests:

```sh
LAYA_MODEL_DIR=/path/to/model LAYA_RUNTIME=/path/to/native/library go test -count=1 -v ./internal/inference
```

`testdata/tokenizer-parity.json` contains exact Python SDK reference tokens for
original test inputs, including indentation, NFC normalization, Korean, emoji,
and special tokens. Runtime code implements the pinned tokenizer in Go. Only
this English NFC/ByteLevel/BPE format is accepted; arbitrary multilingual models
are not implicitly supported.

Current models-v2 archives contain those eight flat files.
`scripts/package_model_notices.py` adds missing attribution to the original v1
exports and verifies the serialized ONNX graph is unchanged; use the maintainer
export environment with onnx installed. It does not retrain or requantize. Remove host metadata from tar headers.
Compute the archive SHA-256 and every file's SHA-256, publish a new model version,
then update internal/assets/manifest.json via a CI-checked PR. Never overwrite
existing pinned model assets. CI verifies integrity on download; model license
and provenance are part of each archive. Re-exporting on another framework version
may change bytes and predictions: treat it as a new model artifact.
