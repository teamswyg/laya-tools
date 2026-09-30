"""Maintainer-only identity projection; no patches or dataset code are executed."""
import argparse
import hashlib
import json
import os
from pathlib import Path

import pyarrow as pa
import pyarrow.parquet as pq

SOURCE_SHA256 = "0ba403a7060af657e1f0476937a703726c07da15b759836694b16e553705a10f"
ROWS = 19008
COLUMNS = ("instance_id", "repo", "base_commit", "problem_statement")


def project(path, output, expected=SOURCE_SHA256, rows=ROWS):
    with path.open("rb") as f:
        if hashlib.file_digest(f, "sha256").hexdigest() != expected:
            raise ValueError("source hash mismatch")
    parquet = pq.ParquetFile(path)
    if parquet.metadata.num_rows != rows:
        raise ValueError("row count mismatch")
    for column in COLUMNS:
        if parquet.schema_arrow.field(column).type != pa.string():
            raise ValueError("unexpected identity schema")
    fd = os.open(output, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    with os.fdopen(fd, "w", encoding="utf-8") as out:
        total = 0
        for batch in parquet.iter_batches(batch_size=128, columns=list(COLUMNS)):
            for row in batch.to_pylist():
                if any(not isinstance(v, str) for v in row.values()):
                    raise ValueError("invalid identity value")
                line = json.dumps(row, ensure_ascii=False, sort_keys=True) + "\n"
                size = len(line.encode("utf-8"))
                total += size
                if size >= 1 << 20 or total > 96 << 20:
                    raise ValueError("projection byte limit")
                out.write(line)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    project(args.input, args.output)
