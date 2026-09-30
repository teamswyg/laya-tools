"""Maintainer-only development patch projection; never execute patches."""
import argparse
import hashlib
import json
import os
from pathlib import Path

import pyarrow as pa
import pyarrow.compute as pc
import pyarrow.parquet as pq

from project_training_source import SOURCE_SHA256, ROWS

MEMBERSHIP_SHA256 = "5ceb4aa4681784431606f2ff4410c9225b9639d88a18f618b9de70fa2c5da1fd"
MAX_LINE = 16 << 20
MAX_TOTAL = 256 << 20
MAX_PATCH = 2 << 20


def selection(path, expected=MEMBERSHIP_SHA256, counts=(7335, 5686, 2402)):
    with path.open("rb") as stream:
        raw = stream.read((8 << 20) + 1)
    if len(raw) > 8 << 20 or hashlib.sha256(raw).hexdigest() != expected:
        raise ValueError("membership hash or size")
    rows = json.loads(raw)
    if not isinstance(rows, list) or len(rows) != sum(counts):
        raise ValueError("membership count")
    roles = {"train": 0, "validation": 0, "final": 0}
    all_ids, selected = set(), []
    for row in rows:
        role, task = row["Role"], row["Task"]
        ident = task["ID"]
        if role not in roles or not isinstance(ident, str) or not ident or ident in all_ids:
            raise ValueError("membership role or identity")
        all_ids.add(ident)
        roles[role] += 1
        if role != "final":
            selected.append(ident)
    if tuple(roles.values()) != counts:
        raise ValueError("membership role counts")
    return sorted(selected)


def project(source, membership, output):
    ids = selection(membership)
    return _project(source, output, ids, SOURCE_SHA256, ROWS)


def _project(source, output, ids, expected, rows):
    if not ids or len(set(ids)) != len(ids):
        raise ValueError("invalid selection")
    with source.open("rb") as stream:
        if hashlib.file_digest(stream, "sha256").hexdigest() != expected:
            raise ValueError("source hash")
    parquet = pq.ParquetFile(source)
    if parquet.metadata.num_rows != rows:
        raise ValueError("source rows")
    columns = ["instance_id", "patch"]
    for column in columns:
        if parquet.schema_arrow.field(column).type != pa.string():
            raise ValueError("source schema")
    wanted = pa.array(ids, type=pa.string())
    seen = set()
    total = 0
    fd = os.open(output, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    try:
        with os.fdopen(fd, "wb") as out:
            for batch in parquet.iter_batches(batch_size=32, columns=columns):
                mask = pc.is_in(batch.column(0), value_set=wanted)
                selected = batch.filter(mask)
                lengths = pc.binary_length(selected.column(1))
                oversized = pc.greater(lengths, MAX_PATCH)
                bounded = pa.table({
                    "instance_id": selected.column(0),
                    "patch": pc.if_else(oversized, pa.scalar(None, type=pa.string()), selected.column(1)),
                    "patch_bytes": lengths,
                    "oversized": oversized,
                })
                # Neither unselected nor oversized patch strings reach Python objects.
                for row in bounded.to_pylist():
                    ident, patch = row["instance_id"], row["patch"]
                    if ident in seen or row["patch_bytes"] is None or (not row["oversized"] and not isinstance(patch, str)):
                        raise ValueError("duplicate identity or invalid patch")
                    seen.add(ident)
                    line = (json.dumps(row, ensure_ascii=False, sort_keys=True) + "\n").encode()
                    total += len(line)
                    if len(line) > MAX_LINE or total > MAX_TOTAL:
                        raise ValueError(f"projection byte bound: line={len(line)}, total={total}")
                    out.write(line)
            if seen != set(ids):
                raise ValueError("missing selected identity")
    except BaseException:
        output.unlink(missing_ok=True)
        raise
    return total


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--membership", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    project(args.input, args.membership, args.output)
