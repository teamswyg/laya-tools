"""Maintainer-only: add redistribution notices without changing the ONNX graph.

Usage: python scripts/package_model_notices.py --kind base --source DIR --output DIR
Requires onnx in the export environment. Existing published assets are immutable.
"""
import argparse
import hashlib
import json
import pathlib
import shutil
import tarfile

import onnx


def digest(path):
    with path.open("rb") as stream:
        return hashlib.file_digest(stream, "sha256").hexdigest()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--kind", choices=["base", "code"], required=True)
    parser.add_argument("--source", type=pathlib.Path, required=True)
    parser.add_argument("--output", type=pathlib.Path, required=True)
    args = parser.parse_args()
    root = pathlib.Path(__file__).resolve().parent.parent
    args.output.mkdir(parents=True, exist_ok=False)
    for name in ["tokenizer.json", "config.json", "LICENSE", "MODEL_CARD.md"]:
        shutil.copyfile(args.source / name, args.output / name)
    base = "Laya\nCopyright Convai Innovations\nApache License 2.0\nhttps://huggingface.co/convaiinnovations/laya\n\nModernBERT-large encoder and tokenizer\nApache License 2.0\nhttps://huggingface.co/answerdotai/ModernBERT-large\n"
    notice = (root / "licenses/laya-code.NOTICE").read_text() if args.kind == "code" else base
    modification = "Modified by teamswyg contributors: exported source checkpoint to ONNX opset 18 and dynamically quantized MatMul weights per channel to INT8; tokenizer and config copied unchanged. models-v2 adds this notice and ONNX doc_string attribution; graph and initializers are unchanged from models-v1. See MODEL_CARD.md and PROVENANCE.json.\n"
    (args.output / "NOTICE").write_text(notice + "\n" + modification)
    (args.output / "MODIFICATIONS.md").write_text(modification)
    model = onnx.load(str(args.source / "model.onnx"))
    graph_hash = hashlib.sha256(model.graph.SerializeToString()).hexdigest()
    model.doc_string += "\n" + modification
    onnx.save(model, str(args.output / "model.onnx"))
    del model
    check = onnx.load(str(args.output / "model.onnx"))
    assert hashlib.sha256(check.graph.SerializeToString()).hexdigest() == graph_hash
    del check
    revision = {"base": "55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851", "code": "25f97e5a2ec5f8cf7218a4f67504367d8832e1fe"}[args.kind]
    (args.output / "PROVENANCE.json").write_text(json.dumps({"source_revision": revision, "source_artifact": "models-v1", "source_onnx_sha256": digest(args.source / "model.onnx"), "unchanged_graph_sha256": graph_hash, "modification": modification.strip()}, indent=2)+"\n")
    archive = args.output.parent / (args.kind + "-int8-v2.tgz")
    with tarfile.open(archive, "w:gz", compresslevel=6) as tar:
        for path in sorted(args.output.iterdir()):
            info = tar.gettarinfo(str(path), arcname=path.name)
            info.uid = info.gid = info.mtime = 0
            info.uname = info.gname = ""
            info.mode = 0o644
            with path.open("rb") as stream:
                tar.addfile(info, stream)
    print(json.dumps({"url": "https://github.com/teamswyg/laya-tools/releases/download/models-v2/"+archive.name, "sha256": digest(archive), "files": {p.name: digest(p) for p in sorted(args.output.iterdir())}}, indent=2))


if __name__ == "__main__":
    main()
