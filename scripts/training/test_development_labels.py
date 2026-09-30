import hashlib
import json
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import pyarrow as pa
import pyarrow.parquet as pq

from project_development_labels import selection, _project


class DevelopmentProjectionTests(unittest.TestCase):
    def test_role_selection(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "members"
            rows = [{"Role": role, "Task": {"ID": role}} for role in ("train", "validation", "final")]
            raw = json.dumps(rows).encode()
            path.write_bytes(raw)
            sha = hashlib.sha256(raw).hexdigest()
            self.assertEqual(selection(path, sha, (1, 1, 1)), ["train", "validation"])
            with self.assertRaises(ValueError):
                selection(path, "0" * 64, (1, 1, 1))
            with self.assertRaises(ValueError):
                selection(path, sha, (2, 1, 0))

    def test_projection_filters_and_fails_closed(self):
        with tempfile.TemporaryDirectory() as tmp:
            source = Path(tmp) / "data.parquet"
            output = Path(tmp) / "labels.jsonl"
            rows = [{"instance_id": "dev", "patch": "own synthetic patch"},
                    {"instance_id": "final", "patch": "DO NOT PROJECT"}]
            pq.write_table(pa.Table.from_pylist(rows), source)
            sha = hashlib.sha256(source.read_bytes()).hexdigest()
            _project(source, output, ["dev"], sha, 2)
            self.assertEqual(json.loads(output.read_text()), dict(rows[0], patch_bytes=len(rows[0]["patch"]), oversized=False))
            self.assertEqual(output.stat().st_mode & 0o777, 0o600)
            with self.assertRaises(FileExistsError):
                _project(source, output, ["dev"], sha, 2)
            self.assertTrue(output.exists())
            for ids in (["missing"], ["dev", "dev"]):
                failed = Path(tmp) / "failed"
                with self.assertRaises(ValueError):
                    _project(source, failed, ids, sha, 2)
                self.assertFalse(failed.exists())
            with patch("project_development_labels.MAX_PATCH", 1):
                oversized = Path(tmp) / "oversized"
                _project(source, oversized, ["dev"], sha, 2)
                item = json.loads(oversized.read_text())
                self.assertIsNone(item["patch"])
                self.assertTrue(item["oversized"])
                self.assertEqual(item["patch_bytes"], len(rows[0]["patch"]))
            for bound in ("MAX_LINE", "MAX_TOTAL"):
                failed = Path(tmp) / "failed"
                with patch("project_development_labels." + bound, 1):
                    with self.assertRaises(ValueError):
                        _project(source, failed, ["dev"], sha, 2)
                self.assertFalse(failed.exists())
            pq.write_table(pa.Table.from_pylist([rows[0], rows[0]]), source)
            sha = hashlib.sha256(source.read_bytes()).hexdigest()
            with self.assertRaises(ValueError):
                _project(source, Path(tmp) / "failed", ["dev"], sha, 2)
            self.assertFalse((Path(tmp) / "failed").exists())


if __name__ == "__main__":
    unittest.main()
