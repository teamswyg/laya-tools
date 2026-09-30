"""Adapt an accepted two-seed PDCA experiment to the existing release contract.

Select by the pre-registered validation objective only. Refuse a failed seed or mismatched frozen inputs.
"""
import argparse
import csv
import json
from pathlib import Path
import shutil
import torch
from safetensors import safe_open
from safetensors.torch import load_file
from pdca_train import load_data
from train_pilot import BASE_SHA, QUESTIONS, file_hash


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--runs', type=Path, nargs=2, required=True)
    ap.add_argument('--base', type=Path, required=True)
    ap.add_argument('--out', type=Path, required=True)
    ap.add_argument('--data', type=Path, default=Path('benchmarks/training/pdca-01'))
    args=ap.parse_args()
    if args.out.exists(): raise ValueError('New output required')
    plan=json.loads((args.data/'plan.json').read_text())
    entries=[]
    for run in args.runs:
        selection=json.loads((run/'selection.json').read_text())
        evaluation=json.loads((run/'evaluation.json').read_text())
        if evaluation['seed'] != selection['seed']: raise ValueError('Seed identity mismatch')
        if not evaluation['passed'] or not all(evaluation['criteria'].values()):
            raise ValueError('Both pre-registered seed checks must pass')
        for name, path in [('plan_sha256',args.data/'plan.json'),('data_sha256',args.data/'families.json')]:
            if selection[name]!=file_hash(path) or evaluation[name]!=selection[name]: raise ValueError('Frozen input mismatch')
        if selection['head_sha256'] != file_hash(run/'head.safetensors') or evaluation['selected_head_sha256'] != selection['head_sha256']:
            raise ValueError('Changed candidate')
        entries.append((run,selection,evaluation))
    if sorted(e[1]['seed'] for e in entries) != sorted(plan['seeds']): raise ValueError('Required independent seeds missing')
    run,s,e=min(entries,key=lambda x:x[1]['selection'].get('selection_score', x[1]['selection']['validation']['nll']))
    head=load_file(run/'head.safetensors')
    if file_hash(args.base/'model.safetensors')!=BASE_SHA: raise ValueError('Wrong base')
    with safe_open(args.base/'model.safetensors',framework='pt') as base:
        changed=any(not torch.equal(value,base.get_tensor(name).float()) for name,value in head.items())
    if not changed: raise ValueError('No learned change')
    args.out.mkdir(parents=True)
    shutil.copyfile(run/'head.safetensors',args.out/'head.safetensors')
    rows=load_data(args.data)
    with (args.out/'training-data.tsv').open('w') as f:
        w=csv.DictWriter(f,fieldnames=['task','split','label','state','id','family'],delimiter='\t',lineterminator='\n')
        w.writeheader()
        for row in rows:
            w.writerow({**row,'task':'difficulty','label':['fast','standard','strong'][row['label']]})
    labels=['fast','standard','strong']
    counts={k:sum(r['split']==k for r in rows) for k in ['train','validation','calibration','test']}
    description=(f"{plan['cycle']} trained learning rates {plan['learning_rates']} for up to "
                 f"{plan['max_epochs']} epochs on two seeds; {plan.get('train_orders_per_update',1)} "
                 f"option orders per update. Selection: {plan['selection']}. "
                 "Both seeds passed pre-registered checks; release selection uses the pre-registered validation objective only.")
    if plan.get('initial_head_source_cycle'):
        description += (f" Warm-started from the matching {plan['initial_head_source_cycle']} seed head; "
                        "fresh optimizer, not an optimizer-resume equivalence claim.")
    dataset=(f"{len(rows)} authored English cases: {counts['train']} train, "
             f"{counts['validation']} validation, {counts['calibration']} calibration, "
             f"{counts['test']} held-out test. Training includes 24 prior training examples. "
             "Prior viewed tests become development data; the final test uses new families. "
             "Labels and style remain authored and synthetic; related concepts may recur. "
             "Baseline and candidate calibration use the same grid and calibration set. "
             "This establishes only an initial small-evaluation result, not repository generalization.")
    report={'schema_version':1,'status':'initial_authored_evaluation_improvement_not_production','task':'difficulty',
            'base_revision':'55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851','base_sha256':BASE_SHA,
            'data_sha256':file_hash(args.out/'training-data.tsv'),'seed':s['seed'],'labels':labels,'question':QUESTIONS['difficulty'],
            'split_counts':{k:sum(r['split']==k for r in rows) for k in ['train','validation','calibration','test']},
            'max_len':512,'head_max_len':192,'device':'mps','dtype':'float32','cpu_fallback':False,
            'trainable_parameters':sum(v.numel() for v in head.values()),'frozen_weights_unchanged':s['frozen_weights_unchanged'],
            'head_weights_changed':changed,'selected':s['selection'],'temperature':e['temperature'],
            'test':e['tuned'],'base_test':e['baseline'],
            'test_predictions':[{**p,'expected':labels[p['expected']],'predicted':labels[p['predicted']]} for p in e['tuned']['predictions']],
            'reload_max_logit_error':e['reload_max_logit_error'],'optimizer_resume_verified':False,
            'elapsed_seconds':s['elapsed_seconds'],'peak_process_rss_bytes':s['peak_process_rss_bytes'],
            'step_end_mps_driver_max_bytes':s['sampled_driver_max_bytes'],
            'training_description':description,
            'dataset_description':dataset,
            'pdca_plan':plan,'confirmation_results':[x[2] for x in entries]}
    (args.out/'results.json').write_text(json.dumps(report,indent=2)+'\n')
    print(json.dumps({'selected_seed':s['seed'],'accuracy':e['tuned']['accuracy'],'coverage':e['tuned']['coverage']}))


if __name__=='__main__': main()
