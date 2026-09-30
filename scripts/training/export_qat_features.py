"""Export frozen, public QAT final features. Training/selection remain Go-only."""
import argparse
import json
import resource
import time
from pathlib import Path
import torch
from safetensors.torch import load_file
from fixed_tier import FixedTier, features
from pdca_train import setup
from train_pilot import file_hash
from publish_pilot import validate

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--base', type=Path, required=True)
p.add_argument('--package', type=Path, required=True)
p.add_argument('--out', type=Path, required=True)
a = p.parse_args()
if a.out.exists(): raise ValueError('New output required')
root = Path(__file__).resolve().parents[2] / 'experiments/ternary-qat'
plan = json.loads((root/'plan.json').read_text())
if file_hash(root/'final-families.json') != plan['final_families_sha256']: raise ValueError('Final data changed after registration')
validate(a.package)
provenance = json.loads((a.package/'PROVENANCE.json').read_text())
for f in provenance['base_files']:
    if file_hash(a.base/f['file']) != f['sha256']: raise ValueError('Base asset changed: '+f['file'])
if provenance['head_sha256'] != '29cc0c479fb60fc038ef9555b0e082f1c84555d353ce7f89c40eca989a365627': raise ValueError('Wrong parent')
rows = json.loads((root/'final-families.json').read_text())
start = time.monotonic()
model,tok = setup(a.base,2718)
model.requires_grad_(False);model.eval()
a.out.parent.mkdir(parents=True,exist_ok=True)
x = features(model,tok,rows,start,a.out.parent)
s = load_file(a.package/'head.safetensors')
head = FixedTier();head.load_state_dict(s);head.eval()
with torch.no_grad(): reference=(head(x)/.5).softmax(-1).tolist()
result=dict(schema='riido-qat-final-v1',origin='original_synthetic_ternary_qat_final',
 rows=[dict(id=r['id'],split='final',label=r['label'],feature=v,reference=q) for r,v,q in zip(rows,x.tolist(),reference)],
 provenance=dict(parent_head_sha256=provenance['head_sha256'],base_revision=provenance['base_revision'],base_sha256=provenance['base_sha256'],
 final_families_sha256=plan['final_families_sha256'],pooling='state-only non-special mean FP32; max512; no truncation',
 wall_seconds=time.monotonic()-start,peak_process_rss_bytes=resource.getrusage(resource.RUSAGE_SELF).ru_maxrss,mps_driver_bytes_at_end=torch.mps.driver_allocated_memory()))
a.out.write_text(json.dumps(result,allow_nan=False))
print(json.dumps(dict(cases=len(rows),**result['provenance'])))
