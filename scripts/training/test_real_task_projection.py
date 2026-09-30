"""The audit projection must never copy answer/execution fields."""
import hashlib
import json
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import pyarrow as pa
import pyarrow.parquet as pq
import project_real_task_audit as audit


class ProjectionTests(unittest.TestCase):
    def test_whitelist_and_source_hash(self):
        with tempfile.TemporaryDirectory() as d:
            src, dst = Path(d) / "in.parquet", Path(d) / "out.jsonl"
            row = dict(instance_id="x", repo="a/b", base_commit="a" * 40,
                       problem_statement="Fix the bug", patch="reference solution",
                       eval_script="must never execute", FAIL_TO_PASS=["test"], PASS_TO_PASS=["test2"])
            pq.write_table(pa.Table.from_pylist([row]), src)
            sha = hashlib.sha256(src.read_bytes()).hexdigest()
            with patch.dict(audit.SOURCES, {"test": (sha, 1)}):
                audit.project("test", src, dst)
            self.assertEqual(set(json.loads(dst.read_text())), set(audit.COLUMNS))
            with patch.dict(audit.SOURCES, {"test": ("0" * 64, 1)}):
                with self.assertRaisesRegex(ValueError, "hash"):
                    audit.project("test", src, Path(d) / "bad.jsonl")


if __name__ == "__main__":
    unittest.main()
