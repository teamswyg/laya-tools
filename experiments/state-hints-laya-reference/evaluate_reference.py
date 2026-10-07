# Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
"""Offline maintainer-only frozen Laya reference; no training or product runtime.

The original V1 question is deliberately preserved, not relabeled as the V4
scope ontology. Only the already exposed, exactly pinned validation120 is read.
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
DELTA_SHA = "e7037a4c92460dd0c77facb643a26a2f3ce99240e6a59bd78c15ad91df314fcc"
VALIDATION_SHA = "1cfe4911ff697f9c6c77f7be627a75eb14c87061b87207edf4b009c2f21e00cc"
RUBRIC_SHA = "b76a0e59bf046e82b0b868c56b9aba21a9d607c94bc42d5f56471a3947556215"
CONFIG_PINS = {
    "rl_agent_config.json": "ae287b56bbcf5f8c4f4541ae9dfd00c914c4c48b940b8398c3058af37ba92bbd",
    "encoder/config.json": "bf3ab80598fdccf414855a2ce80f22859e4492d06ca8a62ddd1cfb63972f8979",
    "tokenizer/tokenizer.json": "6c8aaa9a542084f2457eab775d4eeb51f92a70c0fd9de28d5edb0ddec3c08d30",  # gitleaks:allow public file SHA, not a credential
    "tokenizer/tokenizer_config.json": "50044de60daaa73df97d262e15a40d4faf0160e7d742df64b377877a1320dd12",  # gitleaks:allow public file SHA, not a credential
}
ROWS, OPTIONS, FEATURES = 240, 8, 1024
FEATURE_ROW_BYTES = OPTIONS * FEATURES * 4
FEATURE_CACHE_BYTES = ROWS * FEATURE_ROW_BYTES
OUTPUT_LIMIT = 16 * 1024 ** 2
INPUT_LIMIT = 8 * 1024 ** 2
BASE_BYTES, DELTA_BYTES = 842609210, 4252
ANCHOR = Path(".cache/statehint-laya-reference")


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
    if (plan.get("schema") != "riidolaya-frozen-laya-reference-plan-v1"
            or plan.get("status") != "FIXED_BEFORE_ANY_NEW_REFERENCE_MODEL_CALL"
            or plan.get("instructionAndOptions") != QUESTION
            or tuple(plan.get("intentOrder", ())) != INTENTS):
        raise ReferenceError("fixed instruction/plan mismatch")
    base, delta, data = plan["base"], plan["delta"], plan["input"]
    if (base["revision"] != BASE_REVISION or base["sha256"] != BASE_SHA
            or base["bytes"] != BASE_BYTES or base["configPins"] != CONFIG_PINS
            or delta["sha256"] != DELTA_SHA or delta["bytes"] != DELTA_BYTES
            or delta["modifiedNames"] != ["scorer.3.weight", "scorer.3.bias"]
            or data != {"families": 120, "rows": 240, "sha256": VALIDATION_SHA,
                        "partition": "validation", "previouslyExposed": True, "textOnly": True}):
        raise ReferenceError("fixed source/data pins mismatch")
    budget = plan["forwardBudget"]
    if budget != {"oneMPSWorker": True, "encoderCalls": 240,
                  "sourceBaseAndDeltaFromSameCapturedFeatures": True,
                  "maxPeakRSSBytes": 5 * 1024 ** 3, "maxMPSTotalMemoryFraction": .25,
                  "minSystemAvailableMemoryFraction": .2, "cpuFallbackAllowed": False,
                  "missingCacheDownloadAllowed": False}:
        raise ReferenceError("fixed resource plan mismatch")
    if (plan["temperatures"] != {"deltaLocked": .65,
            "baseHistoricalConfig": "choice:6-10 defaultcalibration recordedbeforecalls; no sweep",
            "raw1Diagnostic": False}
            or plan["gates"]["confidence"] != .9 or plan["gates"]["margin"] != .05
            or plan["outputs"]["cachedFeaturesBytes"] != FEATURE_CACHE_BYTES
            or plan["outputs"]["maxOutputBytes"] != OUTPUT_LIMIT):
        raise ReferenceError("fixed scoring/output plan mismatch")
    for key in ("calibrationAccess", "finalTestAccess", "privateTaskInputsAllowed",
                "modelSelected", "modelPromoted"):
        if plan[key] is not False:
            raise ReferenceError("reference scope mismatch")
    if plan["fitUpdates"] != 0 or plan["originalV1OntologyDiffersFromV4"] is not True:
        raise ReferenceError("reference history mismatch")


def load_validation(raw):
    rows = [json.loads(line) for line in raw.splitlines()]
    if len(rows) != ROWS:
        raise ReferenceError("validation row count")
    families, identifiers, texts = {}, set(), set()
    for row in rows:
        text, family, identifier = row["text"], row["family_id"], row["id"]
        if (row["partition"] != "validation" or row["locale"] not in ("ko", "en")
                or row["expected_intent"] not in INTENTS or row["role"] != "prose"
                or row["semantic_applicable"] is not True or row["wording_index"] != 1
                or row["ontology_freeze_sha256"] != RUBRIC_SHA or row["license"] != "Apache-2.0"
                or not isinstance(text, str) or not text.strip() or len(text.encode("utf-8")) > 4096
                or not isinstance(family, str) or not family or identifier != family + "-" + row["locale"]
                or identifier in identifiers or text in texts):
            raise ReferenceError("validation schema/identity")
        identifiers.add(identifier)
        texts.add(text)
        members = families.setdefault(family, [])
        members.append(row)
    counts = dict.fromkeys(INTENTS, 0)
    for members in families.values():
        if (len(members) != 2 or {r["locale"] for r in members} != {"ko", "en"}
                or len({r["expected_intent"] for r in members}) != 1
                or len({r["leakage_group_id"] for r in members}) != 1):
            raise ReferenceError("validation pair mismatch")
        counts[members[0]["expected_intent"]] += 1
    if len(families) != 120 or set(counts.values()) != {15}:
        raise ReferenceError("validation class quota")
    return rows


def historical_temperature(config):
    t = float(config.get("temperature_by_options", {}).get("choice:6-10", config["temperature"][0]))
    if not math.isfinite(t) or t <= 0:
        raise ReferenceError("historical temperature invalid")
    return t


def probabilities(logits, temperature):
    if len(logits) != OPTIONS or not all(math.isfinite(x) for x in logits) or not math.isfinite(temperature) or temperature <= 0:
        raise ReferenceError("nonfinite score")
    maximum = max(logits)
    values = [math.exp((x - maximum) / temperature) for x in logits]
    total = sum(values)
    return [x / total for x in values]


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


@dataclass
class Forward:
    features: bytes
    base_logits: list
    delta_logits: list
    native_base_logits: list
    tokens: int
    truncated: bool = False
    forward_seconds: float = 0.0


class LayaBackend:
    """Exactly one loaded backbone, one forward per row, two CPU final-linears."""
    def __init__(self, base, delta, config, budget):
        for key, value in {"HF_HUB_OFFLINE": "1", "TRANSFORMERS_OFFLINE": "1",
                           "PYTORCH_ENABLE_MPS_FALLBACK": "0", "TOKENIZERS_PARALLELISM": "false",
                           "USE_TF": "0", "OMP_NUM_THREADS": "1"}.items():
            os.environ[key] = value
        import torch
        from safetensors.torch import load_file
        from transformers import AutoTokenizer
        from laya.common import QTYPES, build_model, build_sequence, collate_items
        if (sys.platform != "darwin" or not torch.backends.mps.is_available()
                or torch.__version__.split("+")[0] != "2.14.0"
                or importlib.metadata.version("laya") != "0.3.21"):
            raise ReferenceError("pinned MPS maintainer runtime required")
        torch.set_num_threads(1)
        torch.set_num_interop_threads(1)
        torch.mps.set_per_process_memory_fraction(.25)
        budget.mps = torch.mps
        budget.check("runtime_ready", 0)
        self.torch, self.QTYPES = torch, QTYPES
        self.build_sequence, self.collate_items = build_sequence, collate_items
        self.tokenizer = AutoTokenizer.from_pretrained(base / "tokenizer", local_files_only=True)
        self.model = build_model(config, encoder_dir=str(base / "encoder"), pretrained=False)
        weights = load_file(base / "model.safetensors")
        self.model.load_state_dict(weights, strict=True, assign=True)
        del weights
        self.model = self.model.float().to("mps").eval()
        for parameter in self.model.parameters():
            parameter.requires_grad_(False)
        layer = self.model.scorer[-1]
        if not isinstance(layer, torch.nn.Linear) or layer.in_features != FEATURES or layer.out_features != 1:
            raise ReferenceError("pre-final scorer shape")
        self.original = {"weight": layer.weight.detach().cpu().contiguous().clone(),
                         "bias": layer.bias.detach().cpu().contiguous().clone()}
        d = load_file(delta)
        if set(d) != {"scorer.3.weight", "scorer.3.bias"} or tuple(d["scorer.3.weight"].shape) != (1, FEATURES) or tuple(d["scorer.3.bias"].shape) != (1,):
            raise ReferenceError("delta schema")
        self.delta = {"weight": d["scorer.3.weight"], "bias": d["scorer.3.bias"]}
        if any(not torch.isfinite(v).all() or v.dtype != torch.float32 for v in self.delta.values()):
            raise ReferenceError("delta numeric type")
        self.capture = []
        self.hook = layer.register_forward_pre_hook(lambda module, args: self.capture.append(args[0].detach().cpu().contiguous()))
        self.calls = 0
        budget.check("model_loaded", 0)

    def forward(self, text):
        torch = self.torch
        ids, markers, stats = self.build_sequence(self.tokenizer, text, QUESTION, 512, 192, return_stats=True)
        uncut, _ = self.build_sequence(self.tokenizer, text, QUESTION, 8192, 192)
        if ids != uncut or len(markers) != OPTIONS or stats["options_distinct"] != OPTIONS or stats["tokens_per_option"] is not None:
            raise ReferenceError("input/options truncation")
        item = {"ids": ids, "markers": markers, "qtype": self.QTYPES["choice"]}
        batch = self.collate_items([[item]], self.tokenizer.pad_token_id)
        self.capture.clear()
        self.calls += 1
        with torch.inference_mode():
            started = time.monotonic()
            native, _ = self.model(*(batch[k].to("mps") for k in ("input_ids", "attention_mask", "marker_pos", "marker_mask", "qtype")))
            torch.mps.synchronize()
            forward_seconds = time.monotonic() - started
            if len(self.capture) != 1 or tuple(self.capture[0].shape) != (1, OPTIONS, FEATURES):
                raise ReferenceError("captured feature shape")
            f = self.capture[0][0]
            if f.dtype != torch.float32 or not torch.isfinite(f).all() or not torch.isfinite(native).all():
                raise ReferenceError("nonfinite captured features")
            score = lambda state: torch.nn.functional.linear(f, state["weight"], state["bias"]).squeeze(-1).tolist()
            values = array("f", f.reshape(-1).tolist())
            if sys.byteorder != "little":
                values.byteswap()
            return Forward(values.tobytes(), score(self.original), score(self.delta), native.detach().cpu()[0].tolist(), len(ids), False, forward_seconds)

    def close(self):
        self.hook.remove()


def evaluate_rows(rows, backend, output, budget, base_temperature):
    """Backend sees text only; no expected labels or scope metadata. Stub-testable."""
    records, feature_hash, parity_max = [], hashlib.sha256(), 0.0
    native_times, token_counts = [], []
    feature_path = output.path / "features.f32le"
    with feature_path.open("xb") as cache:
        os.chmod(cache.name, 0o600)
        for i, row in enumerate(rows):
            budget.check("before_forward", i)
            if i >= ROWS:
                raise ReferenceError("encoder call budget")
            f = backend.forward(row["text"])
            values = array("f")
            values.frombytes(f.features)
            if sys.byteorder != "little":
                values.byteswap()
            if (len(f.features) != FEATURE_ROW_BYTES or f.truncated
                    or not 1 <= f.tokens <= 512 or len(f.native_base_logits) != OPTIONS or not all(math.isfinite(v) for v in values)
                    or not all(math.isfinite(v) for v in f.native_base_logits)):
                raise ReferenceError("feature/truncation budget")
            native_times.append(f.forward_seconds)
            token_counts.append(f.tokens)
            pb, pd = probabilities(f.base_logits, base_temperature), probabilities(f.delta_logits, .65)
            error = max(abs(a - b) for a, b in zip(f.base_logits, f.native_base_logits))
            parity_max = max(parity_max, error)
            if error > .001:
                raise ReferenceError("native/captured base logit parity")
            if output.used + FEATURE_ROW_BYTES > OUTPUT_LIMIT:
                raise ReferenceError("output budget")
            cache.write(f.features)
            output.used += FEATURE_ROW_BYTES
            feature_hash.update(f.features)
            records.append({"id": row["id"], "locale": row["locale"], "text_sha256": sha(row["text"].encode()),
                            "base_logits": f.base_logits, "delta_logits": f.delta_logits,
                            "base_probabilities": pb, "delta_probabilities": pd, "truncated": False})
            budget.check("after_forward", i + 1)
    if len(records) != ROWS or backend.calls != ROWS or feature_path.stat().st_size != FEATURE_CACHE_BYTES:
        raise ReferenceError("complete reference row/feature count")
    scores = {"schema": "riido-statehint-laya-reference-scores-v1", "status": "reference_only",
              "validation_sha256": VALIDATION_SHA, "intent_order": list(INTENTS),
              "base_sha256": BASE_SHA, "delta_sha256": DELTA_SHA,
              "base_temperature": base_temperature, "delta_temperature": .65,
              "training_steps_known": False, "training_steps": None,
              "base_origin": "base_pretrained",
              "delta_origin": "published_v1_delta",
              "instruction_sha256": sha(QUESTION["ins"].encode("utf-8")),
              "feature_cache_sha256": feature_hash.hexdigest(), "rows": records}
    output.write_json("scores.json", scores)
    return {"encoder_calls": backend.calls, "cached_features_bytes": FEATURE_CACHE_BYTES,
            "cached_features_sha256": feature_hash.hexdigest(),
            "cached_features_shape": [ROWS, OPTIONS, FEATURES], "cached_features_dtype": "float32_little_endian",
            "cached_features_order": "input_row_then_frozen_option_then_feature",
            "maximum_native_reconstruction_error": parity_max,
            "base_temperature": base_temperature, "delta_temperature": .65,
            "native_forward_seconds": sum(native_times),
            "mean_native_forward_seconds": sum(native_times) / ROWS,
            "minimum_native_forward_seconds": min(native_times),
            "maximum_native_forward_seconds": max(native_times),
            "minimum_tokens": min(token_counts), "maximum_tokens": max(token_counts)}


def run(args):
    plan_raw = read_pin(args.plan, args.plan_sha256, 1024 ** 2)
    plan = json.loads(plan_raw)
    validate_plan(plan)
    base, delta = local(args.base), local(args.delta)
    read_pin(base / "model.safetensors", BASE_SHA, BASE_BYTES, exact=BASE_BYTES, materialize=False)
    read_pin(delta, DELTA_SHA, DELTA_BYTES, exact=DELTA_BYTES, materialize=False)
    for name, pin in CONFIG_PINS.items():
        read_pin(base / name, pin, 16 * 1024 ** 2, materialize=False)
    config = json.loads(read_pin(base / "rl_agent_config.json", CONFIG_PINS["rl_agent_config.json"], 1024 ** 2))
    temperature = historical_temperature(config)  # recorded before any forward
    raw = read_pin(args.validation, VALIDATION_SHA, INPUT_LIMIT)
    rows = load_validation(raw)
    if args.check:
        return {"status": "checked; no model imports/calls or output", "rows": ROWS,
                "base_temperature": temperature, "delta_temperature": .65}
    budget = Budget()
    budget.check("preflight", 0)
    output = PrivateOutput(args.out)
    output.write("plan.source.json", plan_raw)
    output.write_json("preflight.json", {"base_temperature": temperature, "delta_temperature": .65,
        "base_sha256": BASE_SHA, "delta_sha256": DELTA_SHA, "validation_sha256": VALIDATION_SHA,
        "plan_sha256": sha(plan_raw), "source_sha256": sha(Path(__file__).read_bytes()),
        "base_revision": BASE_REVISION, "instruction_options_sha256": sha(canonical_json(QUESTION)),
        "device": "mps", "dtype": "float32", "cpu_fallback": False, "one_worker": True,
        "fit_updates": 0, "training_steps_known": False, "original_v1_ontology_differs_from_v4": True})
    backend, started = None, time.monotonic()
    try:
        backend = LayaBackend(base, delta, config, budget)
        extraction = evaluate_rows(rows, backend, output, budget, temperature)
        report = {"schema": "riido-statehint-laya-reference-resource-v1", "status": "reference_only",
                  "device": "mps", "dtype": "float32", "cpu_fallback": False,
                  "one_worker": True, "mps_allocator_fraction": .25, "fit_updates": 0,
                  "calibration_calls": 0, "final_test_calls": 0, "selection_performed": False,
                  "promotion_performed": False, "reference_qualification_performed": False,
                  "training_steps_known": False, "training_steps": None,
                  "original_v1_ontology_differs_from_v4": True,
                  "elapsed_seconds": time.monotonic() - started, "extraction": extraction,
                  "resource_samples": budget.samples,
                  "mps_samples_not_transient_peaks": True,
                  "unified_memory_rss_and_mps_overlap_not_added": True,
                  "runtime_versions": {name: importlib.metadata.version(name) for name in ("laya", "torch", "transformers", "safetensors")}}
        output.write_json("resources.json", report)
        return {"status": "reference_only; no selection/qualification/promotion", "encoder_calls": ROWS,
                "output_bytes": output.used}
    except Exception:
        output.write_json("FAILED.json", {"status": "failed; partial private features retained; no qualification",
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
    for flag in ("base", "delta", "validation", "plan", "plan-sha256"):
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
        print("frozen reference failed; inspect pinned inputs/resource limits; no automatic retry", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
