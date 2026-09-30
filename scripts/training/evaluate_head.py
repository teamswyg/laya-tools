"""Maintainer CPU reference for a published head replacement. Not the Go runtime."""
import argparse
import csv
import json
import os
from pathlib import Path
os.environ['USE_TF'] = '0'
import torch
from safetensors.torch import load_file
from transformers import AutoTokenizer
from laya.common import build_model, build_sequence, collate_items, QTYPES
from train_pilot import BASE_SHA, file_hash
from publish_pilot import validate


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--base', type=Path, required=True)
    p.add_argument('--package', type=Path, required=True)
    p.add_argument('--text')
    p.add_argument('--check-reference', action='store_true')
    args = p.parse_args()
    validate(args.package)
    if file_hash(args.base / 'model.safetensors') != BASE_SHA:
        raise ValueError('Exact original checkpoint required')
    torch.set_num_threads(2)
    cfg = json.loads((args.base / 'rl_agent_config.json').read_text())
    result = json.loads((args.package / 'results.json').read_text())
    model = build_model(cfg, encoder_dir=str(args.base / 'encoder'), pretrained=False)
    model.load_state_dict(load_file(args.base / 'model.safetensors'), strict=True, assign=True)
    model = model.float()
    head = load_file(args.package / 'head.safetensors')
    expected = {n for n, _ in model.named_parameters() if n.startswith(('head.', 'type_emb.', 'scorer.'))}
    if set(head) != expected:
        raise ValueError('Head parameter set mismatch')
    model.load_state_dict(head, strict=False)
    model.eval()
    model.encoder.config.reference_compile = False
    tok = AutoTokenizer.from_pretrained(args.base / 'tokenizer', local_files_only=True)
    def predict(text):
        ids, markers, stats = build_sequence(tok, text, result['question'], 512, 192, return_stats=True)
        full, _ = build_sequence(tok, text, result['question'], 8192, 192)
        if ids != full or stats['tokens_per_option'] is not None:
            raise ValueError('Input truncated; abstain')
        batch = collate_items([[{'ids': ids, 'markers': markers, 'qtype': QTYPES['choice']}]], tok.pad_token_id)
        with torch.no_grad():
            logits, _ = model(*(batch[k] for k in ('input_ids', 'attention_mask', 'marker_pos', 'marker_mask', 'qtype')))
            return torch.softmax(logits[0] / result['temperature'], -1)
    if args.check_reference:
        with (args.package / 'training-data.tsv').open() as f:
            rows = [r for r in csv.DictReader(f, delimiter='\t') if r['task'] == result['task']]
        by_id = {f"{result['task']}-{i:03d}": r for i, r in enumerate(rows)}
        error, flips = 0., 0
        for item in result['test_predictions']:
            probs = predict(by_id[item['id']]['state'])
            error = max(error, float((probs - torch.tensor(item['probabilities'])).abs().max()))
            flips += result['labels'][int(probs.argmax())] != item['predicted']
        print(json.dumps({'task': result['task'], 'cases': len(result['test_predictions']),
                          'cpu_mps_max_probability_error': error, 'winner_flips': flips}))
        if error > .001 or flips:
            raise RuntimeError('CPU/MPS reference parity failed')
        return
    if not args.text:
        p.error('--text or --check-reference required')
    probs = predict(args.text)
    print(json.dumps({'status': 'experimental_reference_only', 'probabilities': dict(zip(result['labels'], probs.tolist())),
                      'winner': result['labels'][int(probs.argmax())], 'automatic_execution_allowed': False}, indent=2))


if __name__ == '__main__':
    main()
