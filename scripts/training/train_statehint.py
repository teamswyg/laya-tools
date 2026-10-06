"""Bounded, offline Laya final-linear fine-tuning for authored state hints.

Maintainer experiment only. The pretrained English backbone remains frozen.
Synthetic template variants do not establish independent product accuracy.
This script makes no downloads, uploads, API calls, or runtime dependency changes.
"""

import argparse
from collections import Counter, defaultdict
import gc
import importlib.metadata
import json
import math
import os
from pathlib import Path
import random
import resource
import sys
import time

os.environ["PYTORCH_ENABLE_MPS_FALLBACK"] = "0"
os.environ["USE_TF"] = "0"
os.environ["TOKENIZERS_PARALLELISM"] = "false"
os.environ["HF_HUB_OFFLINE"] = "1"
os.environ["TRANSFORMERS_OFFLINE"] = "1"

import torch
import torch.nn.functional as functional
from safetensors.torch import load_file, save_file
from transformers import AutoTokenizer
from laya.common import QTYPES, build_model, build_sequence, collate_items
from train_pilot import BASE_SHA, file_hash, frozen_hash, guard, metrics


BASE_REVISION = "55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851"
ORIGIN = "original_authored_synthetic_development"
INTENTS = (
    "question", "blocker", "reference", "progress", "completion_report",
    "cancel_request", "planned", "unclear",
)
SPLITS = ("train", "validation", "calibration", "test")
QUESTION = {
    "t": "choice",
    "ins": "Choose the single intent expressed by this message. Quoted or hypothetical claims are not actual reports or requests.",
    "crit": {
        "question": "asks for an answer or explanation",
        "blocker": "current work cannot proceed because something is missing",
        "reference": "shares background information or reference material",
        "progress": "reports work actively being carried out now",
        "completion_report": "reports finished work; completion is not independently verified",
        "cancel_request": "explicitly requests cancellation or withdrawal",
        "planned": "describes future work that has not started",
        "unclear": "intent is absent, contradictory, quoted, or hypothetical",
    },
}
PINNED_BASE_FILES = (
    "rl_agent_config.json", "encoder/config.json", "tokenizer/tokenizer.json",
    "tokenizer/tokenizer_config.json",
)
TRAINABLE_NAMES = ("scorer.3.weight", "scorer.3.bias")


def load_templates(path):
    obj = json.loads(path.read_text())
    if (obj.get("schema") != "riido-statehint-templates-v1"
            or obj.get("origin") != ORIGIN
            or obj.get("license") != "Apache-2.0"
            or tuple(obj.get("intents", ())) != INTENTS
            or obj.get("variants_per_family") != 25):
        raise ValueError("Unsupported template provenance/schema")
    families = obj.get("templates", [])
    if len(families) != 96:
        raise ValueError("Expected exactly 96 authored template families")
    by_id = {}
    counts = Counter()
    for family in families:
        group = family.get("family_id", "")
        locale, intent, split = (family.get(k) for k in ("locale", "intent", "split"))
        template = family.get("template", "")
        if (not isinstance(group, str) or not 1 <= len(group) <= 12 or group in by_id
                or locale not in ("ko", "en") or intent not in INTENTS or split not in SPLITS
                or not isinstance(template, str)
                or "{project}" not in template or "{item}" not in template):
            raise ValueError("Invalid authored template family")
        try:
            template.format(project="FictionalProject", item="FictionalItem")
        except (KeyError, IndexError, ValueError) as error:
            raise ValueError("Unsupported template placeholder") from error
        by_id[group] = family
        counts[locale, intent, split] += 1
    for locale in ("ko", "en"):
        for intent in INTENTS:
            for split, expected in zip(SPLITS, (3, 1, 1, 1)):
                if counts[locale, intent, split] != expected:
                    raise ValueError("Expected 3/1/1/1 families per locale and intent")
    return by_id


def load_corpus(path, families):
    if path.stat().st_size > 2 * 1024 ** 2:
        raise ValueError("Authored corpus exceeds 2 MiB budget")
    rows = []
    ids, texts = set(), set()
    groups = defaultdict(set)
    split_counts = Counter()
    with path.open() as stream:
        for line in stream:
            row = json.loads(line)
            identifier, group = row.get("id"), row.get("group_id")
            locale, intent, split = (row.get(k) for k in ("locale", "intent", "split"))
            text = row.get("text")
            family = families.get(group)
            if (not isinstance(identifier, str) or identifier in ids or family is None
                    or not isinstance(text, str) or not text or len(text.encode()) > 4096
                    or text in texts or row.get("origin") != ORIGIN
                    or any(row.get(k) != family[k] for k in ("locale", "intent", "split"))):
                raise ValueError("Invalid, duplicated, or family-leaking corpus row")
            prefix = group + "-"
            variant = identifier[len(prefix):] if identifier.startswith(prefix) else ""
            if len(variant) != 2 or not variant.isascii() or not variant.isdigit():
                raise ValueError("Invalid variant identity")
            number = int(variant)
            if not 0 <= number < 25 or number in groups[group]:
                raise ValueError("Duplicate/out-of-range template variant")
            project = ("가상프로젝트-" if locale == "ko" else "FictionalProject-") + variant
            item = ("가상항목-" if locale == "ko" else "FictionalItem-") + variant
            if text != family["template"].format(project=project, item=item):
                raise ValueError("Text differs from the pinned original fixture")
            ids.add(identifier)
            texts.add(text)
            groups[group].add(number)
            split_counts[locale, intent, split] += 1
            rows.append(row)
    if len(rows) != 2400 or set(groups) != set(families):
        raise ValueError("Expected 2400 rows and all 96 template families")
    if any(variants != set(range(25)) for variants in groups.values()):
        raise ValueError("Each family must contain exactly 25 unique fictional variants")
    for locale in ("ko", "en"):
        for intent in INTENTS:
            for split, expected in zip(SPLITS, (75, 25, 25, 25)):
                if split_counts[locale, intent, split] != expected:
                    raise ValueError("Wrong per-locale/intent partition size")
    return {split: [row for row in rows if row["split"] == split] for split in SPLITS}


def verify_plan(path, data, templates, base):
    plan = json.loads(path.read_text())
    fixed = {
        "schema": "riido-statehint-training-plan-v1",
        "source_sha256": file_hash(Path(__file__)),
        "reference_sha256": file_hash(Path(__file__).with_name("train_pilot.py")),
        "data_sha256": file_hash(data),
        "templates_sha256": file_hash(templates),
        "base_sha256": BASE_SHA,
        "base_revision": BASE_REVISION,
        "learning_rates": [0.001, 0.003],
        "epochs": 20,
        "seed": 1729,
        "batch_size": 32,
        "confidence_threshold": 0.9,
        "extraction_minutes": 10,
        "training_minutes": 10,
        "overall_minutes": 20,
    }
    for key, expected in fixed.items():
        if plan.get(key) != expected:
            raise ValueError("Pinned training plan mismatch: " + key)
    if file_hash(base / "model.safetensors") != BASE_SHA:
        raise ValueError("Original pretrained weight SHA-256 mismatch")
    pins = plan.get("base_files", {})
    if set(pins) != set(PINNED_BASE_FILES):
        raise ValueError("Pin original configuration and tokenizer files")
    for name in PINNED_BASE_FILES:
        if pins[name] != file_hash(base / name):
            raise ValueError("Original configuration/tokenizer hash mismatch: " + name)
    return plan


def configure_last_linear(model):
    if (not isinstance(model.scorer[-1], torch.nn.Linear)
            or model.scorer[-1].in_features != 1024
            or model.scorer[-1].out_features != 1):
        raise ValueError("Review the pinned Laya scorer structure")
    model.eval()
    for parameter in model.parameters():
        parameter.requires_grad_(False)
    model.scorer[-1].requires_grad_(True)
    names = tuple(name for name, parameter in model.named_parameters() if parameter.requires_grad)
    if names != TRAINABLE_NAMES:
        raise ValueError("Only the pretrained final scoring linear layer may train")
    if sum(parameter.numel() for parameter in model.parameters() if parameter.requires_grad) != 1025:
        raise ValueError("Unexpected trainable parameter count")
    return model.scorer[-1]


def linear_state(layer):
    return {"weight": layer.weight.detach().cpu().contiguous().clone(),
            "bias": layer.bias.detach().cpu().contiguous().clone()}


def restore_linear(layer, state):
    with torch.no_grad():
        layer.weight.copy_(state["weight"].to(layer.weight.device))
        layer.bias.copy_(state["bias"].to(layer.bias.device))


def logits_for(features, state):
    return functional.linear(features, state["weight"], state["bias"]).squeeze(-1)


def labels_for(rows):
    return [INTENTS.index(row["intent"]) for row in rows]


def summarize(logits, rows, temperature):
    output = {}
    for locale in ("all", "ko", "en"):
        indices = [i for i, row in enumerate(rows) if locale == "all" or row["locale"] == locale]
        subset = logits[indices]
        labels = [INTENTS.index(rows[i]["intent"]) for i in indices]
        result = metrics(subset, labels, temperature)
        probabilities = torch.softmax(subset / temperature, dim=-1)
        predictions = probabilities.argmax(-1)
        confusion = [[0] * len(INTENTS) for _ in INTENTS]
        for expected, predicted in zip(labels, predictions.tolist()):
            confusion[expected][predicted] += 1
        eligible = probabilities.max(-1).values.ge(0.9) & predictions.ne(INTENTS.index("unclear"))
        correct = predictions.eq(torch.tensor(labels))
        result.update({
            "temperature": temperature,
            "confusion_labels": list(INTENTS),
            "confusion_true_rows_predicted_columns": confusion,
            "non_unclear_accepted": int(eligible.sum()),
            "non_unclear_accepted_correct": int((eligible & correct).sum()),
            "non_unclear_coverage": float(eligible.float().mean()),
            "non_unclear_accepted_precision": float(correct[eligible].float().mean()) if eligible.any() else None,
        })
        output[locale] = result
    return output


def upstream_temperature(cfg):
    # Eight choices fall in the upstream 6-10 bucket. Do not silently ignore it.
    value = cfg.get("temperature_by_options", {}).get("choice:6-10", cfg["temperature"][0])
    value = float(value)
    if not math.isfinite(value) or value <= 0:
        raise ValueError("Invalid upstream calibration")
    return value


def make_item(tokenizer, row):
    ids, markers, stats = build_sequence(tokenizer, row["text"], QUESTION, 512, 192, return_stats=True)
    uncut, _ = build_sequence(tokenizer, row["text"], QUESTION, 8192, 192)
    if (ids != uncut or len(markers) != len(INTENTS)
            or stats["options_distinct"] != len(INTENTS)
            or stats["tokens_per_option"] is not None):
        raise ValueError("Input/options truncated: " + row["id"])
    return {"ids": ids, "markers": markers, "qtype": QTYPES["choice"]}


def extract_features(model, tokenizer, rows, budget, original):
    """One complete frozen pretrained forward per row; never rerun for epochs."""
    features, costs, wall_costs = [], [], []
    capture = []

    def hook(module, arguments):
        capture.append(arguments[0].detach().cpu().contiguous())

    handle = model.scorer[-1].register_forward_pre_hook(hook)
    maximum_parity_error = 0.0
    try:
        with torch.no_grad():
            for index, row in enumerate(rows):
                row_started = time.monotonic()
                budget.check("extract")
                item = make_item(tokenizer, row)
                batch = collate_items([[item]], tokenizer.pad_token_id)
                capture.clear()
                started = time.monotonic()
                logits, _ = model(*(batch[key].to("mps") for key in
                                    ("input_ids", "attention_mask", "marker_pos", "marker_mask", "qtype")))
                torch.mps.synchronize()
                elapsed = time.monotonic() - started
                if len(capture) != 1 or tuple(capture[0].shape) != (1, 8, 1024):
                    raise RuntimeError("Unexpected final-linear feature shape")
                feature = capture[0][0]
                if not torch.isfinite(feature).all() or not torch.isfinite(logits).all():
                    raise RuntimeError("Nonfinite pretrained output/features")
                error = float((logits_for(feature, original) - logits.detach().cpu()[0]).abs().max())
                maximum_parity_error = max(maximum_parity_error, error)
                if error > 0.001:
                    raise RuntimeError("Cached final-linear reconstruction parity failed")
                features.append(feature)
                costs.append(elapsed)
                if index % 100 == 0 or index + 1 == len(rows):
                    budget.sample("extract", index + 1)
                    print(json.dumps({"phase": "extract", "split": row["split"],
                                      "rows": index + 1, "total_rows": len(rows)}), flush=True)
                row_wall = time.monotonic() - row_started
                wall_costs.append(row_wall)
                budget.extraction_seconds += row_wall
                budget.check("extract")
    finally:
        handle.remove()
    return torch.stack(features), {
        "rows": len(rows), "forward_seconds": sum(costs),
        "wall_seconds": sum(wall_costs),
        "mean_forward_seconds": sum(costs) / len(costs),
        "minimum_forward_seconds": min(costs), "maximum_forward_seconds": max(costs),
        "maximum_cached_logit_absolute_error": maximum_parity_error,
    }


class Budget:
    def __init__(self, root, plan, started):
        self.root = root
        self.plan = plan
        self.started = started
        self.training_started = None
        self.extraction_seconds = 0.0
        self.samples = []

    def check(self, phase):
        guard(self.started, self.plan["overall_minutes"], self.root)
        if phase == "extract" and self.extraction_seconds > self.plan["extraction_minutes"] * 60:
            raise RuntimeError("Feature extraction exceeds 10 minutes; no automatic retry")
        if (phase == "train" and self.training_started is not None
                and time.monotonic() - self.training_started > self.plan["training_minutes"] * 60):
            raise RuntimeError("Cached-feature training exceeds 10 minutes; no automatic retry")

    def sample(self, phase, completed):
        torch.mps.synchronize()
        self.samples.append({
            "phase": phase, "completed": completed,
            "process_peak_rss_bytes": int(resource.getrusage(resource.RUSAGE_SELF).ru_maxrss),
            "mps_allocated_bytes": torch.mps.current_allocated_memory(),
            "mps_driver_bytes": torch.mps.driver_allocated_memory(),
        })


def fit(layer, original, features, rows, budget, plan):
    train_features = features["train"]
    train_targets = torch.tensor(labels_for(rows["train"]), dtype=torch.long)
    validation_targets = torch.tensor(labels_for(rows["validation"]), dtype=torch.long)
    trace, best_state, selection = [], None, None
    best_nll = math.inf
    updates = 0
    budget.training_started = time.monotonic()
    for learning_rate in plan["learning_rates"]:
        restore_linear(layer, original)
        optimizer = torch.optim.AdamW(layer.parameters(), lr=learning_rate, weight_decay=0.01)
        for epoch in range(1, plan["epochs"] + 1):
            order = list(range(len(train_features)))
            random.Random(plan["seed"] + epoch).shuffle(order)
            total_loss = 0.0
            for offset in range(0, len(order), plan["batch_size"]):
                budget.check("train")
                positions = order[offset:offset + plan["batch_size"]]
                inputs = train_features[positions].to("mps")
                targets = train_targets[positions].to("mps")
                optimizer.zero_grad(set_to_none=True)
                logits = layer(inputs).squeeze(-1)
                loss = functional.cross_entropy(logits, targets)
                loss.backward()
                gradients = [parameter.grad for parameter in layer.parameters()]
                if (not torch.isfinite(loss)
                        or any(gradient is None or not torch.isfinite(gradient).all() for gradient in gradients)):
                    raise RuntimeError("Nonfinite loss or final-linear gradients")
                torch.nn.utils.clip_grad_norm_(layer.parameters(), 1.0)
                optimizer.step()
                total_loss += float(loss.detach()) * len(positions)
                updates += 1
            state = linear_state(layer)
            validation_logits = logits_for(features["validation"], state)
            validation_nll = float(functional.cross_entropy(validation_logits, validation_targets))
            if not math.isfinite(validation_nll):
                raise RuntimeError("Nonfinite validation objective")
            entry = {"learning_rate": learning_rate, "epoch": epoch,
                     "train_nll": total_loss / len(train_features), "validation_nll": validation_nll}
            trace.append(entry)
            print(json.dumps({"phase": "train", **entry}), flush=True)
            budget.sample("train", updates)
            if validation_nll < best_nll:
                best_nll, best_state, selection = validation_nll, state, entry.copy()
        del optimizer
        gc.collect()
        torch.mps.empty_cache()
    if best_state is None:
        raise RuntimeError("No finite validation-selected candidate")
    restore_linear(layer, best_state)
    if all(torch.equal(best_state[name], original[name]) for name in original):
        raise RuntimeError("Pretrained final-linear weights did not change")
    return best_state, selection, trace, updates, time.monotonic() - budget.training_started


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base", type=Path, required=True)
    parser.add_argument("--data", type=Path, default=Path("experiments/state-hints/corpus.jsonl"))
    parser.add_argument("--templates", type=Path, default=Path("experiments/state-hints/templates.json"))
    parser.add_argument("--plan", type=Path, required=True)
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args()
    started = time.monotonic()
    if sys.platform != "darwin" or not torch.backends.mps.is_available():
        raise RuntimeError("This pinned Apple Silicon experiment requires MPS; no CPU fallback")
    if importlib.metadata.version("laya") != "0.3.21" or torch.__version__.split("+")[0] != "2.14.0":
        raise RuntimeError("Use the existing pinned Laya 0.3.21/PyTorch 2.14.0 maintainer environment")
    if args.out.exists():
        raise ValueError("A new versioned local output directory is required")
    plan = verify_plan(args.plan, args.data, args.templates, args.base)
    families = load_templates(args.templates)
    rows = load_corpus(args.data, families)
    args.out.mkdir(parents=True)
    budget = Budget(args.out, plan, started)
    budget.check("extract")
    torch.set_num_threads(2)
    torch.manual_seed(plan["seed"])
    random.seed(plan["seed"])
    torch.mps.set_per_process_memory_fraction(0.25)
    cfg = json.loads((args.base / "rl_agent_config.json").read_text())
    tokenizer = AutoTokenizer.from_pretrained(args.base / "tokenizer", local_files_only=True)
    model = build_model(cfg, encoder_dir=str(args.base / "encoder"), pretrained=False)
    weights = load_file(args.base / "model.safetensors")
    model.load_state_dict(weights, strict=True, assign=True)
    del weights
    model.encoder.config.reference_compile = False
    model = model.float().to("mps")
    layer = configure_last_linear(model)
    original = linear_state(layer)
    frozen_before = frozen_hash(model)
    budget.sample("loaded", 0)
    features, extraction = {}, {}
    # Final-test forward is deferred until weights and calibration are locked.
    for split in ("train", "validation", "calibration"):
        features[split], extraction[split] = extract_features(model, tokenizer, rows[split], budget, original)
    selected, selection, trace, updates, training_seconds = fit(layer, original, features, rows, budget, plan)
    calibration_logits = logits_for(features["calibration"], selected)
    calibration_targets = torch.tensor(labels_for(rows["calibration"]), dtype=torch.long)
    temperature_grid = [0.5 + 0.05 * index for index in range(91)]
    temperature = min(temperature_grid, key=lambda value: float(functional.cross_entropy(
        calibration_logits / value, calibration_targets)))
    # The persisted delta keeps original parameter names and excludes every frozen tensor.
    delta = {TRAINABLE_NAMES[0]: selected["weight"], TRAINABLE_NAMES[1]: selected["bias"]}
    save_file(delta, args.out / "scorer.safetensors")
    if (args.out / "scorer.safetensors").stat().st_size > 8192:
        raise RuntimeError("Final-linear delta exceeds the 8 KiB experiment budget")
    locked = {"schema": "riido-statehint-teacher-lock-v1", "plan_sha256": file_hash(args.plan),
              "scorer_sha256": file_hash(args.out / "scorer.safetensors"),
              "selection": selection, "temperature": temperature,
              "confidence_threshold": 0.9, "test_evaluated": False}
    (args.out / "LOCK.json").write_text(json.dumps(locked, indent=2) + "\n")
    frozen_after = frozen_hash(model)
    if frozen_after != frozen_before:
        raise RuntimeError("Frozen pretrained parameters changed")
    persisted = load_file(args.out / "scorer.safetensors")
    if set(persisted) != set(TRAINABLE_NAMES):
        raise RuntimeError("Persisted delta contains unexpected parameters")
    reloaded = {"weight": persisted[TRAINABLE_NAMES[0]], "bias": persisted[TRAINABLE_NAMES[1]]}
    reload_error = float((logits_for(features["validation"], selected)
                          - logits_for(features["validation"], reloaded)).abs().max())
    if reload_error != 0.0:
        raise RuntimeError("Persisted delta reload parity failed")
    # Check one known validation item through the complete native MPS model too.
    # This is a validation parity check, never a second final-test measurement.
    restore_linear(layer, reloaded)
    probe = make_item(tokenizer, rows["validation"][0])
    probe_batch = collate_items([[probe]], tokenizer.pad_token_id)
    with torch.no_grad():
        probe_logits, _ = model(*(probe_batch[key].to("mps") for key in
                                  ("input_ids", "attention_mask", "marker_pos", "marker_mask", "qtype")))
    torch.mps.synchronize()
    native_reload_error = float((probe_logits.detach().cpu()[0]
                                - logits_for(features["validation"][0], reloaded)).abs().max())
    if native_reload_error > 0.001 or not math.isfinite(native_reload_error):
        raise RuntimeError("Complete MPS validation reload parity failed")
    # Extraction reconstructs baseline logits, so temporarily restore its last linear.
    restore_linear(layer, original)
    features["test"], extraction["test"] = extract_features(model, tokenizer, rows["test"], budget, original)
    restore_linear(layer, reloaded)
    baseline_temperature = upstream_temperature(cfg)
    evaluations = {}
    for split in ("validation", "calibration", "test"):
        base_logits = logits_for(features[split], original)
        tuned_logits = logits_for(features[split], reloaded)
        evaluations[split] = {
            "base_raw": summarize(base_logits, rows[split], 1.0),
            "base_upstream_calibrated": summarize(base_logits, rows[split], baseline_temperature),
            "tuned_raw": summarize(tuned_logits, rows[split], 1.0),
            "tuned_calibrated": summarize(tuned_logits, rows[split], temperature),
        }
    budget.check("complete")
    report = {
        "schema": "riido-statehint-teacher-results-v1",
        "origin": ORIGIN, "development_only": True,
        "limitations": [
            "2400 fictional variants represent 96 authored template families, not independent product truth",
            "Korean and English results are separated; the pretrained English checkpoint is not a validated Korean model",
            "Families are disjoint, but related concepts and bilingual paraphrases share a synthetic design",
            "One fixed option order is used; option-order generalization is not established",
            "A completion report does not prove an authorized product state transition",
            "Training the final linear reduces training work; inference still requires the full Laya backbone",
            "Step-end MPS samples miss transient peaks and overlap process RSS in unified memory",
        ],
        "method": "pretrained_frozen_laya_final_linear_cross_entropy_adamw",
        "runtime": "maintainer_reference_only",
        "model_id": "convaiinnovations/laya", "base_revision": BASE_REVISION,
        "base_sha256": BASE_SHA, "base_files": plan["base_files"],
        "data_sha256": file_hash(args.data), "templates_sha256": file_hash(args.templates),
        "source_sha256": file_hash(Path(__file__)), "plan_sha256": file_hash(args.plan),
        "reference_sha256": file_hash(Path(__file__).with_name("train_pilot.py")),
        "question": QUESTION, "intent_order": list(INTENTS),
        "trainable_parameters": 1025, "trainable_names": list(TRAINABLE_NAMES),
        "frozen_sha256_before": frozen_before, "frozen_sha256_after": frozen_after,
        "frozen_unchanged": frozen_before == frozen_after,
        "scorer_sha256": file_hash(args.out / "scorer.safetensors"),
        "scorer_bytes": (args.out / "scorer.safetensors").stat().st_size,
        "persisted_reload_maximum_logit_error": reload_error,
        "native_mps_validation_reload_calls": 1,
        "native_mps_validation_reload_maximum_logit_error": native_reload_error,
        "laya_version": importlib.metadata.version("laya"), "torch_version": torch.__version__,
        "device": "mps", "dtype": "float32", "cpu_fallback": False,
        "selection": selection, "learning_rates": plan["learning_rates"],
        "epochs_per_trial": plan["epochs"], "optimizer_updates": updates,
        "weight_decay": 0.01, "gradient_norm_limit": 1.0,
        "trace": trace, "calibration_temperature": temperature,
        "calibration_grid": {"minimum": 0.5, "maximum": 5.0, "step": 0.05},
        "confidence_threshold": 0.9,
        "split_rows": {split: len(values) for split, values in rows.items()},
        "split_groups": {split: len({row["group_id"] for row in values}) for split, values in rows.items()},
        "feature_extraction": extraction,
        "cached_training_wall_seconds": training_seconds,
        "total_wall_seconds": time.monotonic() - started,
        "memory_samples": budget.samples, "evaluations": evaluations,
        "test_forward_after_weight_temperature_lock": True,
    }
    (args.out / "results.json").write_text(json.dumps(report, indent=2) + "\n")
    (args.out / "rl_agent_config.delta.json").write_text(json.dumps({
        "schema": "riido-statehint-config-delta-v1", "base_sha256": BASE_SHA,
        "parameter_delta": "scorer.safetensors", "question": QUESTION,
        "temperature": [temperature, cfg["temperature"][1], cfg["temperature"][2]],
        "temperature_by_options": {"choice:6-10": temperature},
        "development_only": True,
    }, indent=2) + "\n")
    locked["test_evaluated"] = True
    (args.out / "TEST-COMPLETE.json").write_text(json.dumps(locked, indent=2) + "\n")
    print(json.dumps({"phase": "complete", "selection": selection,
                      "test": evaluations["test"]["tuned_calibrated"]}), flush=True)


if __name__ == "__main__":
    main()
