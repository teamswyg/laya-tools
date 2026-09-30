"""PDCA: validation-only candidate selection, then a separate sealed evaluation.

Authored family-held-out research, not downstream coding success. Local MPS only.
"""
import argparse
import csv
import itertools
import json
import random
import time
import resource
import os
from pathlib import Path
os.environ['PYTORCH_ENABLE_MPS_FALLBACK'] = '0'
os.environ['USE_TF'] = '0'
os.environ['TOKENIZERS_PARALLELISM'] = 'false'
import torch
from safetensors.torch import load_file, save_file
from transformers import AutoTokenizer
from laya.common import build_model, build_sequence, collate_items, QTYPES
from head_policy import configure_head_only
from train_pilot import BASE_SHA, QUESTIONS, file_hash, frozen_hash, head_state, restore_head, guard, metrics


def load_data(root):
    families = json.loads((root / 'families.json').read_text())
    if len({x['family'] for x in families}) != len(families):
        raise ValueError('Family leakage')
    rows = []
    for f in families:
        if len(f['cases']) != 3:
            raise ValueError('Every family needs three labelled cases')
        for label, text in enumerate(f['cases']):
            rows.append({'id': f"{f['family']}-{label}", 'family': f['family'], 'split': f['split'], 'label': label, 'state': text})
    with (root.parent / 'pilot-v1.tsv').open() as stream:
        for i, r in enumerate(csv.DictReader(stream, delimiter='\t')):
            if r['task'] == 'difficulty' and r['split'] == 'train':
                rows.append({'id': f'legacy-train-{i}', 'family': 'legacy-train', 'split': 'train',
                             'label': ['fast', 'standard', 'strong'].index(r['label']), 'state': r['state']})
    if len({x['state'] for x in rows}) != len(rows):
        raise ValueError('Duplicate texts')
    return rows


def setup(base, seed):
    if not torch.backends.mps.is_available(): raise RuntimeError('MPS required')
    if file_hash(base / 'model.safetensors') != BASE_SHA: raise ValueError('Wrong original weights')
    torch.set_num_threads(2)
    torch.manual_seed(seed)
    torch.mps.set_per_process_memory_fraction(.25)
    cfg = json.loads((base / 'rl_agent_config.json').read_text())
    tok = AutoTokenizer.from_pretrained(base / 'tokenizer', local_files_only=True)
    model = build_model(cfg, encoder_dir=str(base / 'encoder'), pretrained=False)
    model.load_state_dict(load_file(base / 'model.safetensors'), strict=True, assign=True)
    model.encoder.config.reference_compile = False
    model = model.float().to('mps')
    configure_head_only(model)
    model.head_checkpointing = True
    return model, tok


def make_item(tok, row, order=(0, 1, 2)):
    ids, markers, stats = build_sequence(tok, row['state'], QUESTIONS['difficulty'], 512, 192,
                                         option_order=list(order), return_stats=True)
    full, _ = build_sequence(tok, row['state'], QUESTIONS['difficulty'], 8192, 192, option_order=list(order))
    if ids != full or stats['tokens_per_option'] is not None:
        raise ValueError('Truncated input: ' + row['id'])
    return {'ids': ids, 'markers': markers, 'qtype': QTYPES['choice'], 'label': order.index(row['label'])}


def forward(model, tok, item):
    b = collate_items([[item]], tok.pad_token_id)
    return model(*(b[k].to('mps') for k in ('input_ids', 'attention_mask', 'marker_pos', 'marker_mask', 'qtype')))[0]


def evaluate(model, tok, rows, started, out, order=(0, 1, 2)):
    model.eval()
    values = []
    with torch.no_grad():
        for r in rows:
            guard(started, 20, out)
            z = forward(model, tok, make_item(tok, r, order)).cpu()[0]
            values.append(z[[order.index(i) for i in range(3)]])
    return torch.stack(values)


def train(args, plan, rows):
    if args.out.exists(): raise ValueError('New output directory required')
    args.out.mkdir(parents=True)
    start = time.monotonic()
    guard(start, 20, args.out)
    model, tok = setup(args.base, args.seed)
    frozen = frozen_hash(model)
    original = head_state(model)
    training = [r for r in rows if r['split'] == 'train']
    validation = [r for r in rows if r['split'] == 'validation']
    target = [r['label'] for r in validation]
    best_loss, selected = float('inf'), None
    records = []
    updates = 0
    driver_max = 0
    permutations = list(itertools.permutations(range(3)))
    for lr in plan['learning_rates']:
        restore_head(model, original)
        torch.manual_seed(args.seed)
        optimizer = torch.optim.AdamW([p for p in model.parameters() if p.requires_grad], lr=lr, weight_decay=.01)
        for epoch in range(plan['max_epochs']):
            configure_head_only(model)
            rng = random.Random(args.seed + epoch)
            order = list(training); rng.shuffle(order)
            loss_total = 0.
            for row in order:
                guard(start, 20, args.out)
                optimizer.zero_grad(set_to_none=True)
                if plan.get('train_orders_per_update', 1) == 3:
                    orders = [(0,1,2),(1,2,0),(2,0,1)] if epoch % 2 == 0 else [(0,2,1),(2,1,0),(1,0,2)]
                else:
                    orders = [rng.choice(permutations)]
                mean_loss = 0.
                for permutation in orders:
                    item = make_item(tok, row, permutation)
                    logits = forward(model, tok, item)
                    loss = torch.nn.functional.cross_entropy(logits, torch.tensor([item['label']], device='mps'))
                    if not torch.isfinite(loss): raise RuntimeError('Non-finite loss')
                    (loss / len(orders)).backward()
                    mean_loss += float(loss.detach()) / len(orders)
                parameters = [p for p in model.parameters() if p.requires_grad]
                if not torch.isfinite(loss) or any(p.grad is None or not torch.isfinite(p.grad).all() for p in parameters):
                    raise RuntimeError('Non-finite update')
                torch.nn.utils.clip_grad_norm_(parameters, 1.)
                optimizer.step()
                torch.mps.synchronize()
                driver_max = max(driver_max, torch.mps.driver_allocated_memory())
                loss_total += mean_loss; updates += 1
            val_orders = permutations if plan.get('validation_all_permutations') else [(0,1,2)]
            val_logits = torch.cat([evaluate(model,tok,validation,start,args.out,order) for order in val_orders])
            val = metrics(val_logits, target * len(val_orders))
            record = {'lr': lr, 'epoch': epoch+1, 'training_loss': loss_total / len(training), 'validation': val}
            records.append(record)
            print(json.dumps(record), flush=True)
            if val['nll'] < best_loss:
                best_loss, selected = val['nll'], record
                save_file(head_state(model), args.out / 'head.safetensors')
        del optimizer
        torch.mps.empty_cache()
    if frozen_hash(model) != frozen: raise RuntimeError('Frozen parameter mutation')
    report = {'cycle': plan['cycle'], 'seed': args.seed, 'selection': selected, 'history': records,
              'updates': updates, 'train_rows': len(training), 'validation_rows': len(validation),
              'head_sha256': file_hash(args.out / 'head.safetensors'), 'plan_sha256': file_hash(args.data / 'plan.json'),
              'data_sha256': file_hash(args.data / 'families.json'), 'legacy_train_sha256': file_hash(args.data.parent / 'pilot-v1.tsv'),
              'frozen_weights_unchanged': True, 'final_test_used_for_selection': False,
              'elapsed_seconds': time.monotonic()-start, 'peak_process_rss_bytes': resource.getrusage(resource.RUSAGE_SELF).ru_maxrss,
              'sampled_driver_max_bytes': driver_max}
    (args.out / 'selection.json').write_text(json.dumps(report, indent=2)+'\n')


def detailed(z, rows, temperature):
    target = [r['label'] for r in rows]
    m = metrics(z, target, temperature)
    prob = torch.softmax(z / temperature, -1)
    pred = prob.argmax(-1)
    strong = torch.tensor(target).eq(2)
    m['strong_recall'] = float(pred[strong].eq(2).float().mean())
    m['strong_to_fast'] = int(pred[strong].eq(0).sum())
    m['predictions'] = [{'id': r['id'], 'expected': r['label'], 'predicted': int(p.argmax()), 'probabilities': p.tolist()}
                        for r, p in zip(rows, prob)]
    return m


def keyword_reference(text):
    """Cheap, transparent comparison only; not a production routing policy."""
    text = text.lower()
    cosmetic = ('typo', 'spelling', 'punctuation', 'prose', 'caption', 'heading',
                'comment', 'gofmt', 'formatter', 'rename a private', 'replace one')
    bounded = ('one', 'single', 'supplied', 'without', 'only', 'no ')
    if any(x in text for x in cosmetic) and any(x in text for x in bounded):
        return 0
    if any(x in text for x in ('concurrent', 'race', 'deadlock', 'distributed', 'crash',
                               'cross-tenant', 'rollback', 'native handle', 'exactly-once',
                               'ownership', 'authorization', 'reproducib')):
        return 2
    return 1


def check(args, plan, rows):
    selection = json.loads((args.out / 'selection.json').read_text())
    if selection['seed'] != args.seed: raise ValueError('Seed does not match frozen selection')
    if (args.out / 'evaluation.json').exists(): raise ValueError('Evaluation already exists; do not overwrite')
    if selection['plan_sha256'] != file_hash(args.data/'plan.json') or selection['data_sha256'] != file_hash(args.data/'families.json'):
        raise ValueError('Plan/data changed after selection')
    if selection['head_sha256'] != file_hash(args.out/'head.safetensors'): raise ValueError('Candidate changed')
    start = time.monotonic()
    model, tok = setup(args.base, args.seed)
    test = [r for r in rows if r['split'] == 'test']
    cal = [r for r in rows if r['split'] == 'calibration']
    legacy = [{'id': r['id'], 'state': r['prompt'], 'label': ['fast','standard','strong'].index(r['expected'])}
              for r in json.loads(Path('benchmarks/routing-golden.json').read_text()) if r['language']=='en']
    temperatures = [.5+i*.05 for i in range(91)]
    def fit(z):
        return min(temperatures, key=lambda t: metrics(z, [r['label'] for r in cal], t)['nll'])
    base_t = fit(evaluate(model,tok,cal,start,args.out))
    base_z = evaluate(model,tok,test,start,args.out)
    baseline = detailed(base_z,test,base_t)
    base_legacy = detailed(evaluate(model,tok,legacy,start,args.out),legacy,base_t)
    restore_head(model, load_file(args.out/'head.safetensors'))
    temperature = fit(evaluate(model,tok,cal,start,args.out))
    z = evaluate(model,tok,test,start,args.out)
    restore_head(model, load_file(args.out/'head.safetensors'))
    reload_error = float((z-evaluate(model,tok,test,start,args.out)).abs().max())
    if reload_error > 1e-4: raise RuntimeError('Saved head reload parity failed')
    tuned = detailed(z,test,temperature)
    tuned_legacy = detailed(evaluate(model,tok,legacy,start,args.out),legacy,temperature)
    flips = 0
    for order in itertools.permutations(range(3)):
        if order == (0,1,2): continue
        permuted = evaluate(model,tok,test,start,args.out,order)
        flips += int(permuted.argmax(-1).ne(z.argmax(-1)).sum())
    flip_rate = flips/(5*len(test))
    c = plan['criteria']
    passed = {
        'accuracy': tuned['accuracy'] >= c['minimum_accuracy'],
        'no_accuracy_regression': tuned['accuracy'] >= baseline['accuracy'],
        'strong_recall': tuned['strong_recall'] >= c['minimum_strong_recall'],
        'no_strong_to_fast': tuned['strong_to_fast'] == 0,
        'coverage': tuned['coverage'] >= c['minimum_coverage'],
        'accepted_precision': tuned['accepted_precision'] is not None and tuned['accepted_precision'] >= c['minimum_accepted_precision'],
        'gain': tuned['accuracy'] - baseline['accuracy'] >= c['improvement_any']['accuracy_absolute_gain'] or
                tuned['brier'] <= baseline['brier'] * (1-c['improvement_any']['brier_relative_reduction']),
        'permutation': flip_rate <= c['max_permutation_winner_flip_rate'],
        'legacy': tuned_legacy['accuracy'] >= base_legacy['accuracy']-c['max_legacy_accuracy_regression'],
    }
    report = {'seed': args.seed, 'selected_head_sha256': selection['head_sha256'], 'plan_sha256': selection['plan_sha256'],
              'data_sha256': selection['data_sha256'], 'base_temperature': base_t, 'temperature': temperature,
              'baseline': baseline, 'tuned': tuned, 'base_legacy': base_legacy, 'tuned_legacy': tuned_legacy,
              'permutation_comparisons': 5*len(test), 'permutation_flips': flips, 'permutation_flip_rate': flip_rate,
              'reload_max_logit_error': reload_error,
              'criteria': passed, 'passed': all(passed.values()),
              'keyword_reference': {'correct': sum(keyword_reference(r['state']) == r['label'] for r in test),
                                    'n': len(test), 'note': 'Uncalibrated keyword comparison; no confidence or production claim'},
              'limits': 'Authored synthetic family holdout, small sample; not downstream coding or production qualification'}
    (args.out/'evaluation.json').write_text(json.dumps(report,indent=2)+'\n')
    print(json.dumps({'seed':args.seed,'criteria':passed,'passed':report['passed'], 'accuracy':tuned['accuracy'],
                      'base_accuracy':baseline['accuracy'],'coverage':tuned['coverage'],'precision':tuned['accepted_precision']}),flush=True)


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('mode',choices=['train','check'])
    ap.add_argument('--base',type=Path,required=True)
    ap.add_argument('--data',type=Path,default=Path('benchmarks/training/pdca-01'))
    ap.add_argument('--out',type=Path,required=True)
    ap.add_argument('--seed',type=int,choices=[1729,2718],required=True)
    args=ap.parse_args()
    plan=json.loads((args.data/'plan.json').read_text());rows=load_data(args.data)
    (train if args.mode=='train' else check)(args,plan,rows)


if __name__=='__main__': main()
