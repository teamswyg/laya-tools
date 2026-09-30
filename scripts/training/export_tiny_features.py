"""Maintainer-only export: public PDCA-06 fixtures -> frozen MPS features.
No arbitrary corpus argument: only reviewed, original synthetic repository data.
Go owns head inference, compression, selection, metrics and timing.
"""
import argparse
import json
import time
import resource
from pathlib import Path
import torch
from safetensors.torch import load_file
from fixed_tier import FixedTier, features
from pdca_train import setup, load_data
from train_pilot import file_hash
from publish_pilot import validate

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--base', type=Path, required=True)
p.add_argument('--package', type=Path, required=True)
p.add_argument('--out', type=Path, required=True)
a = p.parse_args()
if a.out.exists(): raise ValueError('Output must be new')
validate(a.package)
if file_hash(a.package / 'head.safetensors') != '29cc0c479fb60fc038ef9555b0e082f1c84555d353ce7f89c40eca989a365627':
    raise ValueError('Expected pinned v0.2 head')
root = Path(__file__).resolve().parents[2]
rows = load_data(root / 'benchmarks/training/pdca-06')
start = time.monotonic()
model, tok = setup(a.base, 2718)
model.requires_grad_(False); model.eval()
a.out.parent.mkdir(parents=True, exist_ok=True)
x = features(model, tok, rows, start, a.out.parent)
s = load_file(a.package / 'head.safetensors')
head = FixedTier(); head.load_state_dict(s); head.eval()
with torch.no_grad(): probs = (head(x) / .5).softmax(-1).tolist()
source = dict(gamma=s['0.weight'].tolist(), beta=s['0.bias'].tolist(), weight=s['1.weight'].flatten().tolist(), bias=s['1.bias'].tolist(), temperature=.5)
public = dict(schema='riido-tiny-features-v1', origin='original_synthetic_pdca06_development', source=source,
    rows=[dict(id=r['id'], split=r['split'], label=r['label'], feature=v, reference=q) for r,v,q in zip(rows,x.tolist(),probs)],
    provenance=dict(parent_head_sha256=file_hash(a.package/'head.safetensors'), families_sha256=file_hash(root/'benchmarks/training/pdca-06/families.json'),
                    wall_seconds=time.monotonic()-start, peak_process_rss_bytes=resource.getrusage(resource.RUSAGE_SELF).ru_maxrss,
                    mps_driver_bytes_at_end=torch.mps.driver_allocated_memory()))
a.out.write_text(json.dumps(public, allow_nan=False))
print(json.dumps(dict(cases=len(rows), **public['provenance'])))
