"""Maintainer-only Parquet projection; no training, scoring or public raw rows."""
import argparse
import hashlib
import json
from pathlib import Path

import pyarrow.parquet as pq

SOURCE_SHA256 = "cb78d8951e9623bc2460d9b29c3c4152724f28c4f685fc019e0aafa071d2b18f"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if hashlib.sha256(args.input.read_bytes()).hexdigest() != SOURCE_SHA256:
        raise ValueError("unexpected source revision")
    columns = {
        "repository_name": "repository",
        "func_path_in_repository": "path",
        "func_name": "name",
        "func_code_string": "code",
        "func_documentation_string": "query",
        "func_code_url": "url",
    }
    table = pq.read_table(args.input, columns=list(columns))
    if table.num_rows != 14291:
        raise ValueError("unexpected row count")
    with args.output.open("x", encoding="utf-8") as out:
        for batch in table.to_batches(max_chunksize=256):
            for row in batch.to_pylist():
                out.write(json.dumps({v: row[k] for k, v in columns.items()},
                                     ensure_ascii=False, sort_keys=True) + "\n")


if __name__ == "__main__":
    main()
