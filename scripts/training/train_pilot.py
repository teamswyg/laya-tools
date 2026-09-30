"""Bounded real Laya MPS head tuning on original synthetic development fixtures.

Not production qualification, not generative training. No uploads from this script.
"""
import argparse
import csv
import gc
import hashlib
import json
import math
import os
from pathlib import Path
import random
import re
import resource
import shutil
import subprocess
import time

os.environ['PYTORCH_ENABLE_MPS_FALLBACK'] = '0'
os.environ['USE_TF'] = '0'
os.environ['TOKENIZERS_PARALLELISM'] = 'false'
import torch
from safetensors.torch import load_file, save_file
from transformers import AutoTokenizer
from laya.common import build_model, build_sequence, collate_items, QTYPES
from head_policy import configure_head_only

BASE_SHA = '891102d372688fc2a094dac56a384bc537b87c63f21f9f3dac0be2b7cbc8d86c'
QUESTIONS = {
    'difficulty': {
        't': 'choice',
        'ins': "Select the minimum coding capability needed to complete the user's request correctly.",
        'crit': {
            'fast': 'a small, obvious edit, typo, formatting, or straightforward factual question',
            'standard': 'a bounded feature or bug fix with ordinary implementation and tests',
            'strong': 'difficult debugging, architecture, security, concurrency, migrations, or ambiguous multi-file work',
        },
    },
    'decomposition': {
        't': 'choice', 'ins': 'Choose the execution mode appropriate for this candidate task plan and its evidence.',
        'crit': {
            'keep_atomic': 'keep the task together; this split harms invariants or adds needless overhead',
            'split_sequential': 'useful subtasks require ordered outputs or shared state',
            'split_parallel': 'at least one stage has independent ready tasks; integrate afterward',
            'needs_context': 'candidate plan or dependency evidence is insufficient',
        },
    },
}


def file_hash(path):
    with open(path, 'rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


def frozen_hash(model):
    h = hashlib.sha256()
    for n, p in model.named_parameters():
        if not p.requires_grad:
            h.update(n.encode())
            h.update(p.detach().cpu().contiguous().numpy().tobytes())
    return h.hexdigest()


def head_state(model):
    return {n: p.detach().cpu().contiguous().clone() for n, p in model.named_parameters() if p.requires_grad}


def restore_head(model, state):
    with torch.no_grad():
        for n, p in model.named_parameters():
            if p.requires_grad:
                p.copy_(state[n].to(p.device))


def pressure():
    out = subprocess.check_output(['memory_pressure', '-Q'], text=True)
    return int(re.search(r'free percentage: (\d+)%', out).group(1))


def guard(started, minutes, root):
    if time.monotonic() - started > minutes * 60:
        raise RuntimeError('Wall-clock training budget exceeded; no automatic retry')
    if shutil.disk_usage(root).free < 30 * 1024**3:
        raise RuntimeError('Less than 30 GiB free disk; stopped')
    if pressure() < 20:
        raise RuntimeError('System memory availability below 20%; stopped')
    rss = int(subprocess.check_output(['ps', '-o', 'rss=', '-p', str(os.getpid())], text=True)) * 1024
    if rss > 5 * 1024**3:
        raise RuntimeError('Process RSS exceeds 5 GiB; stopped')


def metrics(logits, labels, temperature=1.0):
    p = torch.softmax(logits / temperature, -1)
    target = torch.tensor(labels)
    pred = p.argmax(-1)
    confidence = p.max(-1).values
    correct = pred.eq(target)
    accepted = confidence.ge(.9)
    ece = 0.0
    for i in range(10):
        mask = (confidence >= i / 10) & ((confidence < (i + 1) / 10) if i < 9 else (confidence <= 1))
        if mask.any():
            ece += float(mask.float().mean() * (confidence[mask].mean() - correct[mask].float().mean()).abs())
    return {'n': len(labels), 'correct': int(correct.sum()), 'accuracy': int(correct.sum()) / len(labels),
            'nll': float(torch.nn.functional.cross_entropy(logits / temperature, target)),
            'brier': float(((p - torch.nn.functional.one_hot(target, logits.shape[1]))**2).sum(-1).mean()),
            'ece_10_bins': ece, 'accepted': int(accepted.sum()),
            'accepted_correct': int((accepted & correct).sum()), 'coverage': int(accepted.sum()) / len(labels),
            'accepted_precision': int((accepted & correct).sum()) / int(accepted.sum()) if accepted.any() else None}


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--base', type=Path, required=True)
    ap.add_argument('--data', type=Path, default=Path('benchmarks/training/pilot-v1.tsv'))
    ap.add_argument('--task', choices=QUESTIONS, required=True)
    ap.add_argument('--out', type=Path, required=True)
    ap.add_argument('--epochs', type=int, choices=range(1, 6), default=2)
    ap.add_argument('--minutes', type=int, choices=range(1, 61), default=20)
    args = ap.parse_args()
    if args.out.exists():
        raise ValueError('Output must be a new versioned directory')
    args.out.mkdir(parents=True)
    started = time.monotonic()
    guard(started, args.minutes, args.out)
    if not torch.backends.mps.is_available():
        raise RuntimeError('MPS unavailable; no CPU fallback')
    if file_hash(args.base / 'model.safetensors') != BASE_SHA:
        raise ValueError('Original weight hash mismatch')
    torch.set_num_threads(2)
    torch.manual_seed(1729)
    random.seed(1729)
    torch.mps.set_per_process_memory_fraction(.25)
    cfg = json.loads((args.base / 'rl_agent_config.json').read_text())
    tok = AutoTokenizer.from_pretrained(args.base / 'tokenizer', local_files_only=True)
    model = build_model(cfg, encoder_dir=str(args.base / 'encoder'), pretrained=False)
    weights = load_file(args.base / 'model.safetensors')
    model.load_state_dict(weights, strict=True, assign=True)
    del weights
    model.encoder.config.reference_compile = False
    model = model.float().to('mps')
    names = configure_head_only(model)
    model.head_checkpointing = True
    original = head_state(model)
    frozen_before = frozen_hash(model)
    question = QUESTIONS[args.task]
    labels = list(question['crit'])
    rows = [r for r in csv.DictReader(args.data.open(), delimiter='\t') if r['task'] == args.task]
    if len({r['state'] for r in rows}) != len(rows):
        raise ValueError('Duplicate fixture text')
    splits = {s: [] for s in ('train', 'validation', 'calibration', 'test')}
    for i, r in enumerate(rows):
        ids, markers, stats = build_sequence(tok, r['state'], question, 512, 192, return_stats=True)
        uncut, _ = build_sequence(tok, r['state'], question, 8192, 192)
        if ids != uncut or stats['options_distinct'] != len(labels) or stats['tokens_per_option'] is not None:
            raise ValueError('Input or options truncated')
        item = {'ids': ids, 'markers': markers, 'qtype': QTYPES['choice'], 'label': labels.index(r['label']),
                'id': f'{args.task}-{i:03d}'}
        splits[r['split']].append(item)
    if not all(splits.values()):
        raise ValueError('All four partitions required')

    def forward(item):
        b = collate_items([[item]], tok.pad_token_id)
        return model(*(b[k].to('mps') for k in ('input_ids', 'attention_mask', 'marker_pos', 'marker_mask', 'qtype')))[0]

    def predict(split):
        model.eval()
        result = []
        with torch.no_grad():
            for item in splits[split]:
                guard(started, args.minutes, args.out)
                result.append(forward(item).detach().cpu()[0])
        return torch.stack(result)

    targets = {s: [x['label'] for x in items] for s, items in splits.items()}
    base_val = predict('validation')
    best_nll = float('inf')
    history = []
    losses = []
    samples = []
    best = None
    for lr in (1e-5, 5e-5):
        restore_head(model, original)
        torch.manual_seed(1729)
        optimizer = torch.optim.AdamW([p for p in model.parameters() if p.requires_grad], lr=lr, weight_decay=.01)
        for epoch in range(args.epochs):
            configure_head_only(model)
            order = list(range(len(splits['train'])))
            random.Random(1729 + epoch).shuffle(order)
            for pos in order:
                guard(started, args.minutes, args.out)
                item = splits['train'][pos]
                optimizer.zero_grad(set_to_none=True)
                logits = forward(item)
                loss = torch.nn.functional.cross_entropy(logits, torch.tensor([item['label']], device='mps'))
                loss.backward()
                grads = [p.grad for p in model.parameters() if p.requires_grad]
                if not torch.isfinite(loss) or any(g is None or not torch.isfinite(g).all() for g in grads):
                    raise RuntimeError('Non-finite loss or gradients')
                torch.nn.utils.clip_grad_norm_([p for p in model.parameters() if p.requires_grad], 1.0)
                optimizer.step()
                torch.mps.synchronize()
                losses.append(float(loss.detach()))
                samples.append({'mps_allocated': torch.mps.current_allocated_memory(), 'mps_driver': torch.mps.driver_allocated_memory()})
            val = metrics(predict('validation'), targets['validation'])
            candidate = {'lr': lr, 'epoch': epoch + 1, 'validation': val}
            history.append(candidate)
            print(json.dumps(candidate), flush=True)
            if val['nll'] < best_nll:
                best_nll, best = val['nll'], candidate
                save_file(head_state(model), args.out / 'head.safetensors')
            # Local-only resumable boundary, overwritten atomically; not uploaded.
            checkpoint = {'head': head_state(model), 'optimizer': optimizer.state_dict(), 'lr': lr, 'epoch': epoch + 1,
                          'cpu_rng': torch.get_rng_state(), 'mps_rng': torch.mps.get_rng_state(),
                          'task': args.task, 'data_sha256': file_hash(args.data), 'base_sha256': BASE_SHA}
            torch.save(checkpoint, args.out / 'resume.tmp')
            os.replace(args.out / 'resume.tmp', args.out / 'resume.pt')
            del checkpoint
        del optimizer
        gc.collect()
        torch.mps.empty_cache()
    restore_head(model, load_file(args.out / 'head.safetensors'))
    if frozen_hash(model) != frozen_before:
        raise RuntimeError('Frozen encoder/action weights changed')
    changed = any(not torch.equal(original[n], p.detach().cpu()) for n, p in model.named_parameters() if p.requires_grad)
    if not changed:
        raise RuntimeError('No head weights changed')
    cal = predict('calibration')
    temperatures = [0.5 + .05 * i for i in range(91)]
    temperature = min(temperatures, key=lambda t: metrics(cal, targets['calibration'], t)['nll'])
    selected_test = predict('test')
    # Reload persisted head and check prediction parity before publication.
    restore_head(model, load_file(args.out / 'head.safetensors'))
    reload_error = float((selected_test - predict('test')).abs().max())
    if reload_error > 1e-4:
        raise RuntimeError('Reload prediction parity failed')
    restore_head(model, original)
    base_test = predict('test')
    base_t = cfg.get('temperature_by_options', {}).get('choice:3-5', cfg['temperature'][0])
    result = {'schema_version': 1, 'status': 'experimental_synthetic_pilot', 'task': args.task,
              'base_revision': '55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851', 'base_sha256': BASE_SHA,
              'data_sha256': file_hash(args.data), 'seed': 1729, 'labels': labels, 'question': question,
              'split_counts': {s: len(v) for s, v in splits.items()}, 'max_len': 512, 'head_max_len': 192,
              'device': 'mps', 'dtype': 'float32', 'cpu_fallback': False, 'trainable_parameters': sum(p.numel() for p in model.parameters() if p.requires_grad),
              'trainable_names': list(names), 'frozen_weights_unchanged': True, 'head_weights_changed': changed,
              'base_validation': metrics(base_val, targets['validation']), 'candidates': history, 'selected': best,
              'temperature': temperature, 'calibration': metrics(cal, targets['calibration'], temperature),
              'test': metrics(selected_test, targets['test'], temperature), 'base_test': metrics(base_test, targets['test'], base_t),
              'test_predictions': [{'id': item['id'], 'expected': labels[item['label']], 'predicted': labels[int(z.argmax())],
                                    'probabilities': torch.softmax(z / temperature, -1).tolist()}
                                   for item, z in zip(splits['test'], selected_test)],
              'updates': len(losses), 'first_loss': losses[0], 'last_loss': losses[-1],
              'reload_max_logit_error': reload_error, 'optimizer_resume_verified': False,
              'elapsed_seconds': time.monotonic() - started, 'peak_process_rss_bytes': resource.getrusage(resource.RUSAGE_SELF).ru_maxrss,
              'step_end_mps_allocated_max_bytes': max(x['mps_allocated'] for x in samples),
              'step_end_mps_driver_max_bytes': max(x['mps_driver'] for x in samples),
              'limits': {'process_rss_gib': 5, 'min_free_disk_gib': 30, 'min_system_free_percent': 20, 'mps_memory_fraction': .25, 'minutes': args.minutes},
              'limitations': ['Original synthetic development data; not independent repository goldens',
                              'Tiny calibration/test partitions; no production claim', 'No end-to-end coding cost evaluation',
                              'Head replacement weights require exact base; not a standalone or Go ONNX artifact',
                              'Step-end MPS samples are not peaks; unified memory overlaps RSS', 'Saved optimizer checkpoint; resume training not yet validated']}
    (args.out / 'results.json').write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps({k: result[k] for k in ('task', 'test', 'base_test', 'elapsed_seconds', 'peak_process_rss_bytes')}), flush=True)


if __name__ == '__main__':
    main()
