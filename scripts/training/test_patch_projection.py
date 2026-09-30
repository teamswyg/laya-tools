import hashlib
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import pyarrow as pa
import pyarrow.parquet as pq
import project_patch_labels as subject


class PatchProjection(unittest.TestCase):
    def test_private_patch_only_and_no_overwrite(self):
        with tempfile.TemporaryDirectory() as directory:
            source = Path(directory) / "input.parquet"
            output = Path(directory) / "labels.jsonl"
            pq.write_table(pa.table({"instance_id": ["original-1"], "patch": ["original patch"],
                                     "problem_statement": ["must not join label projection"],
                                     "eval_script": ["must never execute"]}), source)
            digest = hashlib.sha256(source.read_bytes()).hexdigest()
            with patch.dict(subject.SOURCES, {"fixture": (digest, 1)}):
                subject.project("fixture", source, output)
                self.assertEqual(json.loads(output.read_text()),
                                 {"instance_id": "original-1", "patch": "original patch"})
                self.assertEqual(output.stat().st_mode & 0o777, 0o600)
                with self.assertRaises(FileExistsError):
                    subject.project("fixture", source, output)
            with patch.dict(subject.SOURCES, {"fixture": ("0" * 64, 1)}):
                with self.assertRaisesRegex(ValueError, "hash"):
                    subject.project("fixture", source, Path(directory) / "wrong.jsonl")
