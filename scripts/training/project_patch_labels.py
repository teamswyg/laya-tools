"""Maintainer-only projection of pinned evaluation patches; never apply them."""
import argparse
import hashlib
import json
from pathlib import Path

import pyarrow as pa
import pyarrow.parquet as pq

from project_real_task_audit import SOURCES

COLUMNS = ("instance_id", "patch")


def project(source, path, output):
    expected, count = SOURCES[source]
    with path.open("rb") as stream:
        if hashlib.file_digest(stream, "sha256").hexdigest() != expected:
            raise ValueError("source hash mismatch")
    table = pq.ParquetFile(path)
    if table.metadata.num_rows != count:
        raise ValueError("row count mismatch")
    for name in COLUMNS:
        if table.schema_arrow.field(name).type != pa.string():
            raise ValueError("unexpected label projection field")
    # Exclusive 0600 output keeps raw patches private and avoids replacement.
    import os
    fd = os.open(output, os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o600)
    with os.fdopen(fd, "w", encoding="utf-8") as stream:
        for batch in table.iter_batches(batch_size=64, columns=list(COLUMNS)):
            for row in batch.to_pylist():
                if set(row) != set(COLUMNS) or any(not isinstance(v, str) for v in row.values()):
                    raise ValueError("invalid label projection row")
                stream.write(json.dumps(row, ensure_ascii=False, sort_keys=True) + "\n")


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", choices=SOURCES, required=True)
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    project(args.source, args.input, args.output)
