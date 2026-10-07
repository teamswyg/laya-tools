# Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
"""Offline maintainer-only frozen training-feature extraction; no Fit or product runtime.

The original V1 question is deliberately preserved, not relabeled as the V4
scope ontology. Only the exactly pinned original train840 is read.
Actual model calls require --run; tests use explicitly owned stub forwards.
"""
from array import array
import argparse
from dataclasses import dataclass
import hashlib
import importlib.metadata
import json
import math
import os
from pathlib import Path
import re
import resource
import stat
import subprocess
import sys
import time

INTENTS = ("question", "blocker", "reference", "progress", "completion_report",
           "cancel_request", "planned", "unclear")
QUESTION = {
    "t": "choice",
    "ins": "Choose the single intent expressed by this message. Quoted or hypothetical claims are not actual reports or requests.",
    "crit": dict(zip(INTENTS, (
        "asks for an answer or explanation",
        "current work cannot proceed because something is missing",
        "shares background information or reference material",
        "reports work actively being carried out now",
        "reports finished work; completion is not independently verified",
        "explicitly requests cancellation or withdrawal",
        "describes future work that has not started",
        "intent is absent, contradictory, quoted, or hypothetical",
    ))),
}
BASE_REVISION = "55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851"
BASE_SHA = "891102d372688fc2a094dac56a384bc537b87c63f21f9f3dac0be2b7cbc8d86c"
CORPUS_SHA = "586f1862241bf0e43734494911d503b6aaedac979215f9fdd802f43e5c7a660e"
REFERENCE_PLAN_SHA = "cfd153d8a80c011e1ac82f8a55086f0db8a891230b68155dd0ad340d6fdb2003"
RUBRIC_SHA = "b76a0e59bf046e82b0b868c56b9aba21a9d607c94bc42d5f56471a3947556215"
CONFIG_PINS = {
    "rl_agent_config.json": "ae287b56bbcf5f8c4f4541ae9dfd00c914c4c48b940b8398c3058af37ba92bbd",
    "encoder/config.json": "bf3ab80598fdccf414855a2ce80f22859e4492d06ca8a62ddd1cfb63972f8979",
    "tokenizer/tokenizer.json": "6c8aaa9a542084f2457eab775d4eeb51f92a70c0fd9de28d5edb0ddec3c08d30",  # gitleaks:allow public file SHA, not a credential
    "tokenizer/tokenizer_config.json": "50044de60daaa73df97d262e15a40d4faf0160e7d742df64b377877a1320dd12",  # gitleaks:allow public file SHA, not a credential
}
ROWS, OPTIONS, FEATURES = 1680, 8, 1024
FEATURE_ROW_BYTES = OPTIONS * FEATURES * 4
FEATURE_CACHE_BYTES = ROWS * FEATURE_ROW_BYTES
OUTPUT_LIMIT = 64 * 1024 ** 2
INPUT_LIMIT = 8 * 1024 ** 2
BASE_BYTES = 842609210
ANCHOR = Path(".cache/statehint-laya-features")


class ReferenceError(ValueError):
    pass


def sha(data):
    return hashlib.sha256(data).hexdigest()


def canonical_json(value):
    return json.dumps(value, ensure_ascii=False, sort_keys=True,
                      separators=(",", ":"), allow_nan=False).encode("utf-8")


def local(path):
    p = Path(path)
    if p.is_absolute() or not p.parts or any(x in (".", "..") for x in p.parts) or "\\" in str(path):
        raise ReferenceError("invalid local path")
    # Refuse symlink traversal, including model/config directories.
    current = Path()
    for part in p.parts:
        current /= part
        if current.is_symlink():
            raise ReferenceError("symlink input or output")
    return p


def read_pin(path, expected, limit, *, exact=None, materialize=True):
    p = local(path)
    if not re.fullmatch(r"[0-9a-f]{64}", expected):
        raise ReferenceError("invalid SHA pin")
    before = p.stat(follow_symlinks=False)
    if not stat.S_ISREG(before.st_mode) or not 0 < before.st_size <= limit or (exact is not None and before.st_size != exact):
        raise ReferenceError("input size or type")
    h, chunks, count = hashlib.sha256(), [], 0
    with os.fdopen(os.open(p, os.O_RDONLY | os.O_NOFOLLOW), "rb") as stream:
        opened = os.fstat(stream.fileno())
        if (opened.st_dev, opened.st_ino) != (before.st_dev, before.st_ino):
            raise ReferenceError("changed input")
        while block := stream.read(min(1 << 20, limit + 1 - count)):
            count += len(block)
            if count > limit:
                raise ReferenceError("input budget")
            h.update(block)
            if materialize:
                chunks.append(block)
    if h.hexdigest() != expected or count != before.st_size:
        raise ReferenceError("input pin mismatch")
    return b"".join(chunks) if materialize else None


def validate_plan(plan):
    if (plan.get("schema") != "riidolaya-laya-frozen-features-shared-go-probe-plan-v1"
            or plan.get("status") != "FIXED_BEFORE_EXTRACTION_OR_HEAD_FIT"
            or plan["baseSHA"] != BASE_SHA or plan["baseRevision"] != BASE_REVISION
            or plan["frozenReferencePlanSHA"] != REFERENCE_PLAN_SHA):
        raise ReferenceError("fixed extraction source plan")
    expected = {"input": "train840 original1,680rowsONLY pinned586f...660e",
        "encoderCalls": ROWS, "oneMPSWorker": True, "frozenBackbone": True,
        "requiresFullNoTruncation": True, "featureBytes": FEATURE_CACHE_BYTES,
        "maxTotalOutputBytes": OUTPUT_LIMIT, "maxRSSBytes": 5 * 1024 ** 3,
        "maxMPSFraction": .25, "minAvailableFraction": .2,
        "noDownloadsOrCPUFallback": True}
    if plan["extraction"] != expected:
        raise ReferenceError("fixed extraction recipe/budget")
    for key in ("calibrationAccess", "finalTestAccess", "promotionAllowed", "fitOnMPSAllowed"):
        if plan[key] is not False:
            raise ReferenceError("extract-only scope")


def directory_bytes(path):
    # Cache-budget metadata only; never open other corpus/feature contents.
    if not path.exists():
        return 0
    total = 0
    for root, dirs, files in os.walk(path, followlinks=False):
        dirs[:] = [x for x in dirs if not (Path(root) / x).is_symlink()]
        for name in files:
            st = (Path(root) / name).stat(follow_symlinks=False)
            if stat.S_ISREG(st.st_mode):
                total += st.st_size
    return total


def check_cache_budget():
    used = directory_bytes(Path(".cache/statehint/v4-foundation")) + directory_bytes(ANCHOR)
    if used + OUTPUT_LIMIT > 512 * 1024 ** 2:
        raise ReferenceError("512 MiB task-cache budget has insufficient reserved headroom")
    return {"existing_task_cache_bytes": used, "reserved_output_bytes": OUTPUT_LIMIT,
            "task_cache_cap_bytes": 512 * 1024 ** 2}


def load_training(raw):
    rows = [json.loads(line) for line in raw.splitlines()]
    if len(rows) != ROWS:
        raise ReferenceError("training row count")
    families, identifiers, texts = {}, set(), set()
    for row in rows:
        text, family, identifier = row["text"], row["family_id"], row["id"]
        if (row["partition"] != "train" or row["locale"] not in ("ko", "en")
                or row["expected_intent"] not in INTENTS or row["role"] != "prose"
                or row["semantic_applicable"] is not True or row["wording_index"] != 1
                or row["ontology_freeze_sha256"] != RUBRIC_SHA or row["license"] != "Apache-2.0"
                or not isinstance(text, str) or not text.strip() or len(text.encode("utf-8")) > 4096
                or not isinstance(family, str) or not family or (not isinstance(identifier, str) or not identifier)
                or identifier in identifiers or text in texts):
            raise ReferenceError("training schema/identity")
        identifiers.add(identifier)
        texts.add(text)
        members = families.setdefault(family, [])
        members.append(row)
    counts = dict.fromkeys(INTENTS, 0)
    for members in families.values():
        if (len(members) != 2 or {r["locale"] for r in members} != {"ko", "en"}
                or len({r["expected_intent"] for r in members}) != 1
                or len({r["leakage_group_id"] for r in members}) != 1):
            raise ReferenceError("training pair mismatch")
        counts[members[0]["expected_intent"]] += 1
    if len(families) != 840 or set(counts.values()) != {105}:
        raise ReferenceError("training class quota")
    return rows


class PrivateOutput:
    def __init__(self, path):
        self.path = local(path)
        if self.path.parent != ANCHOR:
            raise ReferenceError("output anchor")
        Path(".cache").mkdir(exist_ok=True)
        ANCHOR.mkdir(mode=0o700, exist_ok=True)
        if ANCHOR.stat().st_mode & 0o777 != 0o700:
            raise ReferenceError("output directory mode")
        self.path.mkdir(mode=0o700)
        self.used = 0

    def write(self, name, data):
        if "/" in name or "\\" in name or self.used + len(data) > OUTPUT_LIMIT:
            raise ReferenceError("output budget")
        with (self.path / name).open("xb") as stream:
            os.chmod(stream.name, 0o600)
            stream.write(data)
        self.used += len(data)

    def write_json(self, name, obj):
        self.write(name, json.dumps(obj, ensure_ascii=False, indent=2, allow_nan=False).encode() + b"\n")


class Budget:
    def __init__(self):
        self.samples = []
        self.mps = None

    def check(self, phase, completed):
        if sys.platform != "darwin":
            raise ReferenceError("MPS required; no CPU fallback")
        peak = resource.getrusage(resource.RUSAGE_SELF).ru_maxrss  # macOS bytes
        result = subprocess.run(["memory_pressure", "-Q"], capture_output=True, text=True, check=True)
        match = re.search(r"free percentage: (\d+)%", result.stdout)
        if not match or int(match.group(1)) < 20 or peak > 5 * 1024 ** 3:
            raise ReferenceError("RSS/system-memory budget")
        sample = {"phase": phase, "completed_rows": completed, "peak_rss_bytes": peak,
                  "system_available_percent": int(match.group(1))}
        if self.mps is not None:
            sample["mps_allocated_bytes"] = self.mps.current_allocated_memory()
            sample["mps_driver_bytes"] = self.mps.driver_allocated_memory()
            sample["mps_recommended_max_bytes"] = self.mps.recommended_max_memory()
            if sample["mps_driver_bytes"] > .25 * sample["mps_recommended_max_bytes"]:
                raise ReferenceError("MPS driver sample exceeds memory fraction")
        self.samples.append(sample)
        return sample


def offline_environment():
    for key, value in {"HF_HUB_OFFLINE": "1", "TRANSFORMERS_OFFLINE": "1",
                       "PYTORCH_ENABLE_MPS_FALLBACK": "0", "TOKENIZERS_PARALLELISM": "false",
                       "USE_TF": "0", "OMP_NUM_THREADS": "1"}.items():
        os.environ[key] = value


class LocalTokenizer:
    def __init__(self, base):
        offline_environment()
        from transformers import AutoTokenizer
        from laya.common import QTYPES, build_sequence
        self.tokenizer = AutoTokenizer.from_pretrained(base / "tokenizer", local_files_only=True)
        self.build_sequence, self.choice = build_sequence, QTYPES["choice"]

    def prepare(self, text):
        ids, markers, stats = self.build_sequence(self.tokenizer, text, QUESTION, 512, 192, return_stats=True)
        uncut, _ = self.build_sequence(self.tokenizer, text, QUESTION, 8192, 192)
        if (ids != uncut or len(markers) != OPTIONS or stats["options_distinct"] != OPTIONS
                or stats["tokens_per_option"] is not None):
            raise ReferenceError("training input/options truncation; no backbone calls")
        return {"ids": ids, "markers": markers, "qtype": self.choice}


def preflight_rows(rows, tokenizer):
    # Store bounded token IDs, not feature vectors or expected-label inputs.
    items = []
    for row in rows:
        items.append(tokenizer.prepare(row["text"]))
    if len(items) != ROWS:
        raise ReferenceError("complete tokenizer preflight required")
    return items


@dataclass
class Forward:
    features: bytes
    forward_seconds: float = 0.0


class LayaBackend:
    def __init__(self, base, config, prepared_tokenizer, budget):
        offline_environment()
        import torch
        from safetensors.torch import load_file
        from laya.common import build_model, collate_items
        if (sys.platform != "darwin" or not torch.backends.mps.is_available()
                or torch.__version__.split("+")[0] != "2.14.0"
                or importlib.metadata.version("laya") != "0.3.21"):
            raise ReferenceError("pinned MPS runtime required; no fallback")
        torch.set_num_threads(1)
        torch.set_num_interop_threads(1)
        torch.mps.set_per_process_memory_fraction(.25)
        budget.mps = torch.mps
        budget.check("runtime_ready", 0)
        self.torch, self.collate = torch, collate_items
        self.pad = prepared_tokenizer.tokenizer.pad_token_id
        self.model = build_model(config, encoder_dir=str(base / "encoder"), pretrained=False)
        weights = load_file(base / "model.safetensors")
        self.model.load_state_dict(weights, strict=True, assign=True)
        del weights
        self.model = self.model.float().to("mps").eval()
        for parameter in self.model.parameters():
            parameter.requires_grad_(False)
        layer = self.model.scorer[-1]
        if not isinstance(layer, torch.nn.Linear) or layer.in_features != FEATURES or layer.out_features != 1:
            raise ReferenceError("last-linear input shape")
        self.capture = []
        self.hook = layer.register_forward_pre_hook(lambda module, args: self.capture.append(args[0].detach().cpu().contiguous()))
        self.calls = 0
        budget.check("model_loaded", 0)

    def forward(self, item):
        torch = self.torch
        batch = self.collate([[item]], self.pad)
        self.capture.clear()
        self.calls += 1
        with torch.inference_mode():
            started = time.monotonic()
            self.model(*(batch[k].to("mps") for k in ("input_ids", "attention_mask", "marker_pos", "marker_mask", "qtype")))
            torch.mps.synchronize()
            elapsed = time.monotonic() - started
            if len(self.capture) != 1 or tuple(self.capture[0].shape) != (1, OPTIONS, FEATURES):
                raise ReferenceError("captured shape")
            f = self.capture[0][0]
            if f.dtype != torch.float32 or not torch.isfinite(f).all():
                raise ReferenceError("nonfinite feature capture")
            values = array("f", f.reshape(-1).tolist())
            if sys.byteorder != "little":
                values.byteswap()
            return Forward(values.tobytes(), elapsed)

    def close(self):
        self.hook.remove()


def extract_rows(rows, items, backend, output, budget):
    if len(rows) != ROWS or len(items) != ROWS:
        raise ReferenceError("preflight must cover full train840")
    metadata, feature_hash, times = [], hashlib.sha256(), []
    feature_path = output.path / "features.f32le"
    with feature_path.open("xb") as stream:
        os.chmod(stream.name, 0o600)
        for i, (row, item) in enumerate(zip(rows, items)):
            budget.check("before_forward", i)
            f = backend.forward(item)
            values = array("f")
            values.frombytes(f.features)
            if sys.byteorder != "little":
                values.byteswap()
            if len(f.features) != FEATURE_ROW_BYTES or not all(math.isfinite(x) for x in values):
                raise ReferenceError("feature size/finiteness")
            if output.used + FEATURE_ROW_BYTES > OUTPUT_LIMIT:
                raise ReferenceError("output budget")
            stream.write(f.features)
            output.used += FEATURE_ROW_BYTES
            feature_hash.update(f.features)
            times.append(f.forward_seconds)
            metadata.append({"id": row["id"], "locale": row["locale"],
                "text_sha256": sha(row["text"].encode("utf-8")),
                "expected_intent": row["expected_intent"]})
            budget.check("after_forward", i + 1)
    if backend.calls != ROWS or feature_path.stat().st_size != FEATURE_CACHE_BYTES:
        raise ReferenceError("full feature-call accounting")
    sidecar = {"schema": "riido-statehint-laya-feature-sidecar-v1", "base_sha256": BASE_SHA,
        "instruction_sha256": sha(QUESTION["ins"].encode("utf-8")), "intent_order": list(INTENTS),
        "corpus_sha256": CORPUS_SHA, "feature_sha256": feature_hash.hexdigest(),
        "rows_count": ROWS, "shape": [ROWS, OPTIONS, FEATURES], "dtype": "float32_little_endian",
        "rows": metadata}
    output.write_json("sidecar.json", sidecar)
    return {"encoder_calls": backend.calls, "feature_bytes": FEATURE_CACHE_BYTES,
        "feature_sha256": feature_hash.hexdigest(), "native_forward_seconds": sum(times),
        "mean_native_forward_seconds": sum(times) / ROWS}


def run(args):
    plan_raw = read_pin(args.plan, args.plan_sha256, 1024 ** 2)
    plan = json.loads(plan_raw)
    validate_plan(plan)
    reference_raw = read_pin(args.reference_plan, REFERENCE_PLAN_SHA, 1024 ** 2)
    reference = json.loads(reference_raw)
    if (reference["instructionAndOptions"] != QUESTION or tuple(reference["intentOrder"]) != INTENTS
            or reference["base"]["configPins"] != CONFIG_PINS):
        raise ReferenceError("frozen reference preprocessing mismatch")
    base = local(args.base)
    read_pin(base / "model.safetensors", BASE_SHA, BASE_BYTES, exact=BASE_BYTES, materialize=False)
    for name, pin in CONFIG_PINS.items():
        read_pin(base / name, pin, 16 * 1024 ** 2, materialize=False)
    config = json.loads(read_pin(base / "rl_agent_config.json", CONFIG_PINS["rl_agent_config.json"], 1024 ** 2))
    rows = load_training(read_pin(args.train, CORPUS_SHA, INPUT_LIMIT))
    cache = check_cache_budget()
    budget = Budget()
    budget.check("before_tokenizer", 0)
    tokenizer = LocalTokenizer(base)
    items = preflight_rows(rows, tokenizer)  # all rows, before any backbone creation
    budget.check("tokenizer_preflight_complete", 0)
    token_report = {"rows": ROWS, "minimum_tokens": min(len(x["ids"]) for x in items),
        "maximum_tokens": max(len(x["ids"]) for x in items), "truncated_rows": 0,
        "backbone_created": False, "encoder_calls": 0}
    if args.check:
        return {"status": "checked; tokenizer only, no backbone/model forwards or output",
                "preflight": token_report, "cache_budget": cache}
    output = PrivateOutput(args.out)
    output.write("plan.source.json", plan_raw)
    output.write_json("preflight.json", {"tokenizer": token_report, "cache_budget": cache,
        "plan_sha256": sha(plan_raw), "source_sha256": sha(Path(__file__).read_bytes()),
        "reference_plan_sha256": REFERENCE_PLAN_SHA, "instruction_options_sha256": sha(canonical_json(QUESTION)),
        "base_revision": BASE_REVISION, "label_source": "original_pinned_train840_expected_intent",
        "expected_labels_passed_to_backbone": False, "fit_updates": 0})
    backend, started = None, time.monotonic()
    try:
        backend = LayaBackend(base, config, tokenizer, budget)
        extraction = extract_rows(rows, items, backend, output, budget)
        output.write_json("resources.json", {"schema": "riido-statehint-laya-feature-resource-v1",
            "status": "feature_extraction_only", "device": "mps", "dtype": "float32",
            "one_worker": True, "cpu_fallback": False, "mps_allocator_fraction": .25,
            "fit_updates": 0, "calibration_calls": 0, "final_test_calls": 0,
            "label_source": "original_pinned_train840_expected_intent",
            "expected_labels_passed_to_backbone": False, "teacher_pseudo_labels": False,
            "elapsed_seconds": time.monotonic() - started, "extraction": extraction,
            "resource_samples": budget.samples, "mps_samples_not_transient_peaks": True,
            "unified_memory_rss_and_mps_overlap_not_added": True})
        return {"status": "complete feature extraction only; no head Fit", "encoder_calls": ROWS,
                "feature_bytes": FEATURE_CACHE_BYTES, "output_bytes": output.used}
    except Exception:
        output.write_json("FAILED.json", {"status": "failed; partial features retained; no head Fit",
            "encoder_calls": 0 if backend is None else backend.calls})
        raise
    finally:
        if backend is not None:
            backend.close()


class Parser(argparse.ArgumentParser):
    def error(self, message):
        raise ReferenceError("invalid reference arguments")


def main():
    p = Parser(description=__doc__)
    for flag in ("base", "train", "plan", "plan-sha256", "reference-plan"):
        p.add_argument("--" + flag, required=True)
    mode = p.add_mutually_exclusive_group(required=True)
    mode.add_argument("--check", action="store_true")
    mode.add_argument("--run", action="store_true")
    p.add_argument("--out")
    try:
        args = p.parse_args()
        if args.check and args.out or args.run and not args.out:
            raise ReferenceError("output/mode mismatch")
        print(json.dumps(run(args), allow_nan=False))
    except Exception:
        print("training extraction failed; inspect pins/tokenizer truncation/resource limits; no automatic retry", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
