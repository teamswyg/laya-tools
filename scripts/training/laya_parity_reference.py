"""Maintainer-only, CPU implementation/quantization diagnostic; no quality labels.

Original synthetic inputs only. Never opens training, validation or reserve corpora.
Use installed reviewed Laya code; never execute Python fetched from a model repo.
"""
import argparse
import hashlib
import importlib.metadata
import json
import os
from pathlib import Path
import resource
import time

os.environ['TOKENIZERS_PARALLELISM'] = 'false'
os.environ['HF_HUB_OFFLINE'] = '1'
os.environ['USE_TF'] = '0'
import numpy as np
import onnxruntime as ort
import torch
from safetensors.torch import load_file
from transformers import PreTrainedTokenizerFast
import laya.common as common


def digest(path):
    with Path(path).open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


def guard(start):
    # macOS ru_maxrss is bytes. Sample between calls, not a hard native-call cap.
    if time.monotonic() - start > 900 or resource.getrusage(resource.RUSAGE_SELF).ru_maxrss > 8 * 1024**3:
        raise RuntimeError('diagnostic time/RSS guard reached')


def cases():
    states = [
        'File: sum.go\nfunc sum(a, b int) int { return a + b }',
        'File: cache.go\n// 캐시 검색\nfunc Get(key string) string { return key }',
        'File: whitespace.py\ndef tab(x):\n\treturn x\n',
        'File: unicode.go\n// cafe\u0301 café 🙂',
        'File: mask.txt\n[MASK] literal [SEP] marker',
        'File: long.go\n' + 'var x = 1\n' * 300,
    ]
    q = {'t': 'noul', 'ins': 'Is this source code relevant to the software change: "add a sum function"?'}
    return states, q


def main():
    p = argparse.ArgumentParser()
    p.add_argument('--model-dir', type=Path, required=True)
    p.add_argument('--source-dir', type=Path, required=True)
    p.add_argument('--out', type=Path, required=True)
    args = p.parse_args()
    if os.uname().sysname != 'Darwin':
        raise ValueError('RSS contract for this diagnostic requires macOS')
    if args.out.exists():
        raise ValueError('refuse to replace existing reference')
    started = time.monotonic()
    model_dir, source = args.model_dir, args.source_dir
    pins = json.loads(Path('experiments/laya-parity/assets-18.json').read_text())
    for key, path in [('onnx', model_dir/'model.onnx'), ('tokenizer', model_dir/'tokenizer.json'),
                      ('config', model_dir/'config.json'), ('source_weights', source/'model.safetensors'),
                      ('source_config', source/'rl_agent_config.json'), ('encoder_config', source/'encoder/config.json'),
                      ('common', common.__file__)]:
        if digest(path) != pins[key]:
            raise ValueError('hash mismatch: ' + key)
    versions = {name: importlib.metadata.version(name) for name in ['laya', 'torch', 'transformers', 'tokenizers', 'numpy', 'onnxruntime', 'safetensors']}
    if versions['onnxruntime'] != '1.30.0' or versions['laya'] != '0.3.21':
        raise ValueError('reference version mismatch')
    cfg = json.loads((model_dir/'config.json').read_text())
    source_cfg = json.loads((source/'rl_agent_config.json').read_text())
    if cfg != source_cfg:
        raise ValueError('INT8 and source config differ')
    tok = PreTrainedTokenizerFast(tokenizer_file=str(model_dir/'tokenizer.json'), cls_token='[CLS]', sep_token='[SEP]', mask_token='[MASK]', pad_token='[PAD]', unk_token='[UNK]')
    temp = common.clamp_temperature(cfg['temperature_by_options'].get('noul:2', cfg['temperature'][2]))
    so = ort.SessionOptions()
    so.intra_op_num_threads = 4
    so.inter_op_num_threads = 1
    so.enable_cpu_mem_arena = False
    so.enable_mem_pattern = False
    so.graph_optimization_level = ort.GraphOptimizationLevel.ORT_ENABLE_ALL
    session = ort.InferenceSession(str(model_dir/'model.onnx'), sess_options=so, providers=['CPUExecutionProvider'])
    rows = []
    states, q = cases()
    for state in states:
        for order in ([0, 1], [1, 0]):
            guard(started)
            ids, markers = common.build_sequence(tok, state, q, option_order=order)
            options = common.render_options(q)
            feed = {'input_ids': np.array([ids], dtype=np.int64), 'attention_mask': np.ones((1, len(ids)), dtype=np.int64),
                    'marker_pos': np.array([markers], dtype=np.int64), 'marker_mask': np.ones((1, 2), dtype=bool), 'qtype': np.array([2], dtype=np.int64)}
            logits = session.run(['logits'], feed)[0][0].astype(np.float64)
            z = (logits - logits.max()) / temp
            probs = np.exp(z) / np.exp(z).sum()
            rows.append(dict(State=state, Kind='noul', Instruction=q['ins'], Options=[options[i] for i in order], IDs=ids, Markers=markers, Probabilities=probs.tolist()))
    del session
    torch.set_num_threads(4)
    torch.set_num_interop_threads(1)
    torch.manual_seed(1729)
    model = common.build_model(source_cfg, encoder_dir=str(source/'encoder'), pretrained=False)
    model.load_state_dict(load_file(source/'model.safetensors'), strict=True, assign=True)
    model.encoder.config.reference_compile = False
    model = model.float().eval()
    with torch.inference_mode():
        for row in rows:
            guard(started)
            ids, markers = row['IDs'], row['Markers']
            logits = model(torch.tensor([ids]), torch.ones((1,len(ids)), dtype=torch.long), torch.tensor([markers]), torch.ones((1,2), dtype=torch.bool), torch.tensor([2]))[0][0].double()
            row['SourceProbabilities'] = torch.softmax(logits/temp, dim=-1).tolist()
    report = dict(Schema='riido-laya-parity-reference-v1', Versions=versions, Pins=pins, Cases=rows,
                  Seconds=time.monotonic()-started, PeakRSSBytes=resource.getrusage(resource.RUSAGE_SELF).ru_maxrss)
    args.out.write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps({'cases':len(rows), 'seconds':report['Seconds'], 'peak_rss_bytes':report['PeakRSSBytes']}))


if __name__ == '__main__':
    main()
