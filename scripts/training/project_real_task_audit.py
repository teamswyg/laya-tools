"""Maintainer-only pinned Parquet projection. Never execute dataset scripts."""
import argparse
import hashlib
import json
from pathlib import Path

import pyarrow as pa
import pyarrow.parquet as pq

SOURCES = {
    "full": ("d4f5a245c75319fa8240c540674958c4d491e82edf274b144d43836bdcbc4567", 2294),
    "multilingual": ("92abca7cb527b41a9f66d03a26ce441ff7319e3a49f985998fd56be4bb9b08b2", 300),
}
COLUMNS = ("instance_id", "repo", "base_commit", "problem_statement")


def project(source, path, output):
    expected, count = SOURCES[source]
    with path.open("rb") as f:
        if hashlib.file_digest(f, "sha256").hexdigest() != expected:
            raise ValueError("source hash mismatch")
    parquet = pq.ParquetFile(path)
    if parquet.metadata.num_rows != count:
        raise ValueError("row count mismatch")
    for column in COLUMNS:
        if parquet.schema_arrow.field(column).type != pa.string():
            raise ValueError("unexpected projection field type")
    for column in ("FAIL_TO_PASS", "PASS_TO_PASS"):
        field_type = parquet.schema_arrow.field(column).type
        if not pa.types.is_list(field_type) or field_type.value_type != pa.string():
            raise ValueError("unexpected test-list schema")
    with output.open("x", encoding="utf-8") as out:
        for batch in parquet.iter_batches(batch_size=128, columns=list(COLUMNS)):
            for row in batch.to_pylist():
                if set(row) != set(COLUMNS) or any(not isinstance(v, str) for v in row.values()):
                    raise ValueError("invalid projected row")
                out.write(json.dumps(row, ensure_ascii=False, sort_keys=True) + "\n")


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--source", required=True, choices=SOURCES)
    p.add_argument("--input", required=True, type=Path)
    p.add_argument("--output", required=True, type=Path)
    a = p.parse_args()
    project(a.source, a.input, a.output)


if __name__ == "__main__":
    main()
