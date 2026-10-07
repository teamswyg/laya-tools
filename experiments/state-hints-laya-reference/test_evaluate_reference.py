# Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
"""Owned stub-forward tests. No Laya/Torch imports, actual model proof or data."""
from array import array
import copy
import hashlib
import importlib.util
import json
import math
import os
from pathlib import Path
import sys
import tempfile
import unittest

MODULE_PATH = Path(__file__).with_name("evaluate_reference.py")
spec = importlib.util.spec_from_file_location("owned_reference_test_module", MODULE_PATH)
m = importlib.util.module_from_spec(spec)
sys.modules[spec.name] = m
spec.loader.exec_module(m)


class StubBudget:
    def __init__(self):
        self.samples = []

    def check(self, phase, completed):
        self.samples.append((phase, completed))


class OwnedStubBackend:
    """Text-coded arithmetic stub, explicitly not a pretrained-model execution."""
    def __init__(self, wrong_parity=False, truncated=False, nonfinite=False):
        self.calls, self.seen = 0, []
        self.wrong_parity, self.truncated, self.nonfinite = wrong_parity, truncated, nonfinite

    def forward(self, text):
        self.calls += 1
        self.seen.append(text)
        option = m.INTENTS.index(text.split()[3])
        features = array("f", [0.0] * (m.OPTIONS * m.FEATURES))
        features[option * m.FEATURES] = 100.0
        if self.nonfinite:
            features[1] = math.nan
        base = [20.0 if i == option else 0.0 for i in range(8)]
        delta = [10.0 if i == option else 0.0 for i in range(8)]
        native = list(base)
        if self.wrong_parity:
            native[0] += 1.0
        if sys.byteorder != "little":
            features.byteswap()
        return m.Forward(features.tobytes(), base, delta, native, 128, self.truncated)


def owned_rows():
    rows = []
    for label in m.INTENTS:
        for i in range(15):
            family = f"owned-reference-{label}-{i:02d}"
            for locale in ("ko", "en"):
                rows.append({"schema": "statehint-v4-original-train-seed-row-v1",
                    "id": family + "-" + locale, "family_id": family,
                    "leakage_group_id": family, "partition": "validation", "locale": locale,
                    "wording_index": 1, "role": "prose", "semantic_applicable": True,
                    "text": f"Owned stub category {label} arithmetic fixture {i} {locale}",
                    "declared_unit": "Owned synthetic test unit, not a product case.",
                    "expected_intent": label, "ontology_freeze_sha256": m.RUBRIC_SHA,
                    "license": "Apache-2.0"})
    return rows


class ReferenceTests(unittest.TestCase):
    def setUp(self):
        self.previous = Path.cwd()
        self.temp = tempfile.TemporaryDirectory()
        os.chdir(self.temp.name)

    def tearDown(self):
        os.chdir(self.previous)
        self.temp.cleanup()

    def output(self, name="owned-run"):
        return m.PrivateOutput(m.ANCHOR / name)

    def test_complete_once_captured_features_two_heads_and_body_free_scores(self):
        rows = owned_rows()
        self.assertEqual(m.load_validation(b"\n".join(json.dumps(x).encode() for x in rows)), rows)
        backend, budget, output = OwnedStubBackend(), StubBudget(), self.output()
        result = m.evaluate_rows(rows, backend, output, budget, 1.0000158548355103)
        self.assertEqual(backend.calls, 240)
        self.assertEqual(backend.seen, [r["text"] for r in rows])
        self.assertEqual(result["cached_features_bytes"], 7864320)
        cache = (output.path / "features.f32le").read_bytes()
        self.assertEqual(len(cache), 7864320)
        self.assertEqual(hashlib.sha256(cache).hexdigest(), result["cached_features_sha256"])
        scores = json.loads((output.path / "scores.json").read_text())
        self.assertEqual(scores["intent_order"], list(m.INTENTS))
        self.assertFalse(scores["training_steps_known"])
        self.assertIsNone(scores["training_steps"])
        self.assertEqual(scores["base_temperature"], 1.0000158548355103)
        self.assertEqual(scores["delta_temperature"], .65)
        self.assertEqual(len(scores["rows"]), 240)
        for row in scores["rows"]:
            self.assertEqual(set(row), {"id", "locale", "text_sha256", "base_logits", "delta_logits", "base_probabilities", "delta_probabilities", "truncated"})
            for name in ("base_probabilities", "delta_probabilities"):
                self.assertEqual(len(row[name]), 8)
                self.assertAlmostEqual(sum(row[name]), 1.0)
        self.assertLess(output.used, m.OUTPUT_LIMIT)
        self.assertEqual(output.path.stat().st_mode & 0o777, 0o700)
        for path in output.path.iterdir():
            self.assertEqual(path.stat().st_mode & 0o777, 0o600)

    def test_expected_labels_units_and_scope_cannot_change_forward_scores(self):
        rows = owned_rows()
        changed = copy.deepcopy(rows)
        for r in changed:
            r["expected_intent"] = "unclear"
            r["declared_unit"] = "Changed evaluation-only metadata."
        a, b = self.output("first"), self.output("second")
        m.evaluate_rows(rows, OwnedStubBackend(), a, StubBudget(), 1.0)
        m.evaluate_rows(changed, OwnedStubBackend(), b, StubBudget(), 1.0)
        self.assertEqual((a.path / "scores.json").read_bytes(), (b.path / "scores.json").read_bytes())
        self.assertEqual((a.path / "features.f32le").read_bytes(), (b.path / "features.f32le").read_bytes())

    def test_no_second_forward_retries_or_partial_success_on_invalid_capture(self):
        for kwargs in ({"wrong_parity": True}, {"truncated": True}, {"nonfinite": True}):
            with self.subTest(kwargs=kwargs):
                output = self.output("failed-" + next(iter(kwargs)))
                backend = OwnedStubBackend(**kwargs)
                with self.assertRaises(m.ReferenceError):
                    m.evaluate_rows(owned_rows(), backend, output, StubBudget(), 1.0)
                self.assertEqual(backend.calls, 1)
                self.assertFalse((output.path / "scores.json").exists())
        output = self.output("too-many")
        with self.assertRaises(m.ReferenceError):
            m.evaluate_rows(owned_rows() + [owned_rows()[0]], OwnedStubBackend(), output, StubBudget(), 1.0)
        self.assertFalse((output.path / "scores.json").exists())

    def test_stable_softmax_finiteness_temperature_and_original_order(self):
        self.assertEqual(m.probabilities([10000.0] * 8, .65), [.125] * 8)
        p = m.probabilities([10000.0, -10000.0] + [0.0] * 6, .65)
        self.assertEqual(p[0], 1.0)
        for logits, t in (([math.nan] * 8, 1), ([0.0] * 7, 1), ([0.0] * 8, 0), ([0.0] * 8, math.inf)):
            with self.assertRaises(m.ReferenceError):
                m.probabilities(logits, t)
        self.assertEqual(m.historical_temperature({"temperature": [9], "temperature_by_options": {"choice:6-10": 1.0000158548355103}}), 1.0000158548355103)

    def test_bounded_exact_pins_exclusive_outputs_and_validation_only(self):
        data = b"owned temporary input"
        Path("owned").write_bytes(data)
        pin = m.sha(data)
        self.assertEqual(m.read_pin("owned", pin, len(data)), data)
        for expected, limit in (("0" * 64, len(data)), (pin, len(data) - 1)):
            with self.assertRaises(m.ReferenceError):
                m.read_pin("owned", expected, limit)
        Path("link").symlink_to("owned")
        with self.assertRaises(m.ReferenceError):
            m.read_pin("link", pin, len(data))
        with self.assertRaises(m.ReferenceError):
            m.local("../private")
        out = self.output("exclusive")
        out.write("same", b"original")
        with self.assertRaises(FileExistsError):
            out.write("same", b"replacement")
        with self.assertRaises(FileExistsError):
            self.output("exclusive")
        out.used = m.OUTPUT_LIMIT
        with self.assertRaises(m.ReferenceError):
            out.write("overflow", b"x")
        rows = owned_rows()
        for partition in ("train", "calibration", "test"):
            changed = copy.deepcopy(rows)
            changed[0]["partition"] = partition
            with self.assertRaises(m.ReferenceError):
                m.load_validation(b"\n".join(json.dumps(x).encode() for x in changed))
        rows[1]["expected_intent"] = "blocker"
        with self.assertRaises(m.ReferenceError):
            m.load_validation(b"\n".join(json.dumps(x).encode() for x in rows))

class FixedPlanTests(unittest.TestCase):
    def plan(self):
        return {"schema": "riidolaya-frozen-laya-reference-plan-v1",
            "status": "FIXED_BEFORE_ANY_NEW_REFERENCE_MODEL_CALL",
            "instructionAndOptions": copy.deepcopy(m.QUESTION), "intentOrder": list(m.INTENTS),
            "base": {"revision": m.BASE_REVISION, "sha256": m.BASE_SHA,
                     "bytes": m.BASE_BYTES, "configPins": dict(m.CONFIG_PINS)},
            "delta": {"sha256": m.DELTA_SHA, "bytes": m.DELTA_BYTES,
                      "modifiedNames": ["scorer.3.weight", "scorer.3.bias"]},
            "input": {"families": 120, "rows": 240, "sha256": m.VALIDATION_SHA,
                      "partition": "validation", "previouslyExposed": True, "textOnly": True},
            "forwardBudget": {"oneMPSWorker": True, "encoderCalls": 240,
                "sourceBaseAndDeltaFromSameCapturedFeatures": True,
                "maxPeakRSSBytes": 5 * 1024 ** 3, "maxMPSTotalMemoryFraction": .25,
                "minSystemAvailableMemoryFraction": .2, "cpuFallbackAllowed": False,
                "missingCacheDownloadAllowed": False},
            "temperatures": {"deltaLocked": .65,
                "baseHistoricalConfig": "choice:6-10 defaultcalibration recordedbeforecalls; no sweep",
                "raw1Diagnostic": False},
            "gates": {"confidence": .9, "margin": .05},
            "outputs": {"cachedFeaturesBytes": m.FEATURE_CACHE_BYTES, "maxOutputBytes": m.OUTPUT_LIMIT},
            "calibrationAccess": False, "finalTestAccess": False,
            "privateTaskInputsAllowed": False, "modelSelected": False,
            "modelPromoted": False, "fitUpdates": 0, "originalV1OntologyDiffersFromV4": True}

    def test_prospective_fixed_options_scope_and_no_sweep(self):
        m.validate_plan(self.plan())
        for mutate in (
            lambda p: p["instructionAndOptions"]["crit"].update({"completion_report": "Changed criterion"}),
            lambda p: p["temperatures"].update({"raw1Diagnostic": True}),
            lambda p: p["forwardBudget"].update({"encoderCalls": 241}),
            lambda p: p["forwardBudget"].update({"cpuFallbackAllowed": True}),
            lambda p: p.update({"finalTestAccess": True}),
            lambda p: p.update({"fitUpdates": 1}),
        ):
            plan = self.plan()
            mutate(plan)
            with self.assertRaises(m.ReferenceError):
                m.validate_plan(plan)
        self.assertNotIn("torch", sys.modules)
        self.assertNotIn("laya", sys.modules)


if __name__ == "__main__":
    unittest.main()
