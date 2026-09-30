"""Maintainer-only oracle. Go runtime never imports this module."""
import argparse
import hashlib
import importlib.metadata
import json
from pathlib import Path

REFERENCE = "38c7bc9c93808ee15602265ba96a9b8fda58783f"
MODEL_SHA = "f95ffde02ad06f63ae38eb9d400038cd5ccaf8411ec3cb650c6025113f96cbb8"
VOCABULARY_SHA256 = "e67e803f624fb4d67dea1c730d06e1067e1b14d830e2c2202569e3ef0f70bb50"

def main():
    p = argparse.ArgumentParser()
    p.add_argument("--model-dir", required=True)
    p.add_argument("--out", required=True)
    a = p.parse_args()
    root = Path(a.model_dir)
    for name, digest in [("model.safetensors", MODEL_SHA), ("tokenizer.json", VOCABULARY_SHA256)]:
        if hashlib.sha256((root / name).read_bytes()).hexdigest() != digest:
            raise ValueError("source hash mismatch")
    dist = importlib.metadata.distribution("model2vec")
    provenance = json.loads(dist.read_text("direct_url.json") or "{}")
    if provenance.get("vcs_info", {}).get("commit_id") != REFERENCE:
        raise ValueError("install the pinned reference commit")
    from model2vec import StaticModel
    model = StaticModel.from_pretrained(root, normalize=True, max_length=None)
    texts = [
        "", "   ", "parse HTTP response headers", "def get_value(x): return x[0]",
        "parseHTTPResponse snake_case", "café naïve résumé", "CAFÉ", "cafe\u0301",
        "한국어 코드 검색", "中文測試", "日本語の関数", "emoji 😀🚀", "İstanbul I i ı",
        "Straße STRASSE", "ΑΒΓ αβγ Σ σ ς", "Привет мир", "١٢٣ 123",
        "[CLS] hello [SEP]", "[UNK]", "[PAD][MASK]", "x[CLS]y", "[cls] hello",
        "a\x00b\ufffdc\u200bd", "a\t b\n c\r d", "a\u00a0b\u2028c", "a+b=$5.00!",
        "can't won't isn't", "x" * 101, "x" * 100, "\U00020000中", "\ue000 hidden",
        "read only; do not delete", "delete only; do not read",
        "func (r *Reader) Read(p []byte) (int,error) { return 0,nil }",
    ]
    ids = model.tokenize(texts)
    vectors = model.encode(texts, normalize=True, max_length=None, use_multiprocessing=False)
    output = {
        "schema": "riido-static-reference-v1", "reference_commit": REFERENCE,
        "model_sha256": MODEL_SHA, "tokenizer_sha256": VOCABULARY_SHA256,
        "reference_version": dist.version,
        "tokenizers_version": importlib.metadata.version("tokenizers"),
        "numpy_version": importlib.metadata.version("numpy"),
        "cases": [{"text": text, "ids": tokens, "vector": vector.tolist()}
                  for text, tokens, vector in zip(texts, ids, vectors)],
    }
    Path(a.out).write_text(json.dumps(output, ensure_ascii=True, indent=2) + "\n")

if __name__ == "__main__":
    main()
