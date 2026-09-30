import hashlib
import json
import tempfile
import unittest
from pathlib import Path

import pyarrow as pa
import pyarrow.parquet as pq

from project_training_source import project


class TrainingProjectionTests(unittest.TestCase):
    def test_whitelist_hash_permissions_and_no_overwrite(self):
        with tempfile.TemporaryDirectory() as tmp:
            source = Path(tmp) / "source.parquet"
            out = Path(tmp) / "out.jsonl"
            row = dict(instance_id="test", repo="example/repo", base_commit="a" * 40,
                       problem_statement="Own synthetic query", patch="must not project",
                       hints_text="must not project", eval_script="must not execute")
            pq.write_table(pa.Table.from_pylist([row]), source)
            sha = hashlib.sha256(source.read_bytes()).hexdigest()
            with self.assertRaises(ValueError):
                project(source, out, "0" * 64, 1)
            self.assertFalse(out.exists())
            with self.assertRaises(ValueError):
                project(source, out, sha, 2)
            project(source, out, sha, 1)
            result = json.loads(out.read_text())
            self.assertEqual(set(result), {"instance_id", "repo", "base_commit", "problem_statement"})
            self.assertEqual(out.stat().st_mode & 0o777, 0o600)
            with self.assertRaises(FileExistsError):
                project(source, out, sha, 1)


if __name__ == "__main__":
    unittest.main()
