# Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
"""Owned tokenizer/forward stubs only; no actual corpus, model or ML imports."""
from array import array
import copy
import importlib.util
import json
import math
import os
from pathlib import Path
import sys
import tempfile
from types import SimpleNamespace
from unittest.mock import patch, Mock
import unittest

spec = importlib.util.spec_from_file_location("owned_feature_module", Path(__file__).with_name("extract_training.py"))
m = importlib.util.module_from_spec(spec)
sys.modules[spec.name] = m
spec.loader.exec_module(m)


class StubBudget:
    def check(self, phase, completed):
        return None


class OwnedTokenizer:
    def __init__(self, fail_at=None):
        self.seen = []
        self.fail_at = fail_at

    def prepare(self, text):
        self.seen.append(text)
        if self.fail_at == len(self.seen):
            raise m.ReferenceError("owned truncation stub")
        return {"ids": [1, len(text), 3], "markers": list(range(8)), "qtype": 0}


class OwnedBackend:
    """Fixed arithmetic feature bytes; not a pretrained forward or accuracy proof."""
    def __init__(self, corrupt=False):
        self.calls, self.seen = 0, []
        values = array("f", [0.25] * (8 * 1024))
        if corrupt:
            values[0] = math.nan
        if sys.byteorder != "little":
            values.byteswap()
        self.data = values.tobytes()

    def forward(self, item):
        self.calls += 1
        self.seen.append(item)
        return m.Forward(self.data)


def owned_rows():
    rows = []
    for label in m.INTENTS:
        for i in range(105):
            family = f"owned-feature-{label}-{i:03d}"
            for locale in ("ko", "en"):
                rows.append({"id": family + "-" + locale, "family_id": family,
                    "leakage_group_id": family, "partition": "train", "locale": locale,
                    "wording_index": 1, "role": "prose", "semantic_applicable": True,
                    "text": f"Owned synthetic tokenizer fixture {label} {i} {locale}",
                    "expected_intent": label, "ontology_freeze_sha256": m.RUBRIC_SHA,
                    "license": "Apache-2.0"})
    return rows


class FeatureTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.previous = Path.cwd()
        os.chdir(self.temp.name)

    def tearDown(self):
        os.chdir(self.previous)
        self.temp.cleanup()

    def fixed_plan(self):
        return {"schema": "riidolaya-laya-frozen-features-shared-go-probe-plan-v1",
            "status": "FIXED_BEFORE_EXTRACTION_OR_HEAD_FIT", "baseSHA": m.BASE_SHA,
            "baseRevision": m.BASE_REVISION, "frozenReferencePlanSHA": m.REFERENCE_PLAN_SHA,
            "extraction": {"input": "train840 original1,680rowsONLY pinned586f...660e",
                "encoderCalls":1680,"oneMPSWorker":True,"frozenBackbone":True,
                "requiresFullNoTruncation":True,"featureBytes":55050240,
                "maxTotalOutputBytes":67108864,"maxRSSBytes":5368709120,
                "maxMPSFraction":.25,"minAvailableFraction":.2,"noDownloadsOrCPUFallback":True},
            "calibrationAccess":False,"finalTestAccess":False,"promotionAllowed":False,"fitOnMPSAllowed":False}

    def test_fixed_plan_and_actual_orchestration_stop_before_model_on_truncation(self):
        plan = self.fixed_plan()
        m.validate_plan(plan)
        for key, value in (("encoderCalls",1681),("requiresFullNoTruncation",False),("noDownloadsOrCPUFallback",False)):
            changed = copy.deepcopy(plan)
            changed["extraction"][key] = value
            with self.assertRaises(m.ReferenceError):
                m.validate_plan(changed)
        reference = {"instructionAndOptions":m.QUESTION,"intentOrder":list(m.INTENTS),
                     "base":{"configPins":m.CONFIG_PINS}}
        rows = owned_rows()
        def owned_read_pin(path, *args, **kwargs):
            if str(path)=="owned-plan":return json.dumps(plan).encode()
            if str(path)=="owned-reference":return json.dumps(reference).encode()
            if str(path)=="owned-train":return b"\n".join(json.dumps(r).encode() for r in rows)
            if str(path).endswith("rl_agent_config.json"):return b"{}"
            return None
        backend = Mock(side_effect=AssertionError("backbone must not be created"))
        args = SimpleNamespace(plan="owned-plan",plan_sha256="0"*64,
            reference_plan="owned-reference",base="owned-base",train="owned-train",
            check=False,out=str(m.ANCHOR/"must-not-exist"))
        with patch.object(m,"read_pin",owned_read_pin),patch.object(m,"Budget",StubBudget), \
                patch.object(m,"LocalTokenizer",lambda base:OwnedTokenizer(fail_at=1680)), \
                patch.object(m,"LayaBackend",backend):
            with self.assertRaises(m.ReferenceError):
                m.run(args)
        backend.assert_not_called()
        self.assertFalse((m.ANCHOR/"must-not-exist").exists())

    def test_full_stream_exact_bytes_sidecar_order_and_original_targets(self):
        rows = owned_rows()
        parsed = m.load_training(b"\n".join(json.dumps(r).encode() for r in rows))
        self.assertEqual(parsed, rows)
        tokenizer, backend = OwnedTokenizer(), OwnedBackend()
        items = m.preflight_rows(rows, tokenizer)
        output = m.PrivateOutput(m.ANCHOR / "owned-stream")
        report = m.extract_rows(rows, items, backend, output, StubBudget())
        self.assertEqual(backend.calls, 1680)
        self.assertEqual(tokenizer.seen, [r["text"] for r in rows])
        self.assertEqual(backend.seen, items)
        self.assertEqual(report["feature_bytes"], 55050240)
        self.assertEqual((output.path / "features.f32le").stat().st_size, 55050240)
        sidecar = json.loads((output.path / "sidecar.json").read_text())
        self.assertEqual(sidecar["rows_count"], 1680)
        self.assertEqual(sidecar["shape"], [1680, 8, 1024])
        self.assertEqual(sidecar["feature_sha256"], report["feature_sha256"])
        for original, metadata in zip(rows, sidecar["rows"]):
            self.assertEqual(set(metadata), {"id", "locale", "text_sha256", "expected_intent"})
            self.assertEqual(metadata["id"], original["id"])
            self.assertEqual(metadata["expected_intent"], original["expected_intent"])
        self.assertLess(output.used, 64 * 1024 ** 2)
        for p in output.path.iterdir():
            self.assertEqual(p.stat().st_mode & 0o777, 0o600)

    def test_all_tokenizer_preflight_before_backbone_no_truncation_skip(self):
        rows = owned_rows()
        tokenizer, backend = OwnedTokenizer(fail_at=1680), OwnedBackend()
        with self.assertRaises(m.ReferenceError):
            items = m.preflight_rows(rows, tokenizer)
            m.extract_rows(rows, items, backend, m.PrivateOutput(m.ANCHOR / "must-not-exist"), StubBudget())
        self.assertEqual(backend.calls, 0)
        self.assertFalse(m.ANCHOR.exists())
        self.assertEqual(len(tokenizer.seen), 1680)

    def test_targets_do_not_enter_tokenizer_or_forward_items(self):
        rows = owned_rows()
        changed = copy.deepcopy(rows)
        for row in changed:
            row["expected_intent"] = "unclear"
            row["declared_unit"] = "Changed annotation only."
        left, right = OwnedTokenizer(), OwnedTokenizer()
        self.assertEqual(m.preflight_rows(rows, left), m.preflight_rows(changed, right))
        self.assertEqual(left.seen, right.seen)
        self.assertNotIn("torch", sys.modules)
        self.assertNotIn("laya", sys.modules)

    def test_nonfinite_capture_partial_stage_has_no_success_sidecar(self):
        rows = owned_rows()
        items = m.preflight_rows(rows, OwnedTokenizer())
        backend = OwnedBackend(corrupt=True)
        output = m.PrivateOutput(m.ANCHOR / "failed")
        with self.assertRaises(m.ReferenceError):
            m.extract_rows(rows, items, backend, output, StubBudget())
        self.assertEqual(backend.calls, 1)
        self.assertFalse((output.path / "sidecar.json").exists())
        with self.assertRaises(m.ReferenceError):
            m.extract_rows(rows[:-1], items, OwnedBackend(), output, StubBudget())

    def test_train_only_exact_pairs_pins_and_cache_cap(self):
        rows = owned_rows()
        changed = copy.deepcopy(rows)
        changed[0]["partition"] = "validation"
        with self.assertRaises(m.ReferenceError):
            m.load_training(b"\n".join(json.dumps(r).encode() for r in changed))
        changed = copy.deepcopy(rows)
        changed[1]["expected_intent"] = "blocker"
        with self.assertRaises(m.ReferenceError):
            m.load_training(b"\n".join(json.dumps(r).encode() for r in changed))
        Path("owned").write_bytes(b"owned file")
        self.assertEqual(m.read_pin("owned", m.sha(b"owned file"), 10), b"owned file")
        for pin, limit in (("0" * 64, 10), (m.sha(b"owned file"), 9)):
            with self.assertRaises(m.ReferenceError):
                m.read_pin("owned", pin, limit)
        self.assertEqual(m.check_cache_budget()["reserved_output_bytes"], 64 * 1024 ** 2)
        p = Path(".cache/statehint/v4-foundation")
        p.mkdir(parents=True)
        with (p / "owned-sparse-cache").open("wb") as stream:
            stream.truncate(449 * 1024 ** 2)
        with self.assertRaises(m.ReferenceError):
            m.check_cache_budget()


if __name__ == "__main__":
    unittest.main()
