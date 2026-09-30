"""Original-weight FP32 validation reference, not a user runtime or training job."""
import argparse
import importlib.metadata
import json
import os
from pathlib import Path
import resource
import time

from laya_parity_reference import digest, guard
import laya.common as common
import torch
from safetensors.torch import load_file
from transformers import PreTrainedTokenizerFast

PLAN_SHA = '4eb07b72f5029f8313dce98663d1dffb79b398c7685f70e9225a21b62b20399e'
PINS_SHA = 'fa8193110a1e4f4332176f901798c3cf35e839f9044b03e4fc9a71c99d6bdac9'


def request_hash(req):
    # Requests contain strings, integers, booleans and arrays, in Go field order.
    import hashlib
    text = json.dumps(req, ensure_ascii=False, separators=(',', ':'))
    for char in '<>&\u2028\u2029':
        text = text.replace(char, '\\u%04x' % ord(char))
    return hashlib.sha256(text.encode()).hexdigest()


def main():
    p = argparse.ArgumentParser()
    p.add_argument('--requests', type=Path, required=True)
    p.add_argument('--source-dir', type=Path, required=True)
    p.add_argument('--out', type=Path, required=True)
    a = p.parse_args()
    if os.uname().sysname != 'Darwin' or a.out.exists():
        raise ValueError('requires macOS and a new output file')
    if a.requests.stat().st_size > 16 * 1024**2:
        raise ValueError('requests too large')
    if digest('experiments/laya-original/plan-19.json') != PLAN_SHA or digest('experiments/laya-parity/assets-18.json') != PINS_SHA:
        raise ValueError('plan/assets changed')
    pins = json.loads(Path('experiments/laya-parity/assets-18.json').read_text())
    for key, path in [('source_weights', a.source_dir/'model.safetensors'), ('source_config', a.source_dir/'rl_agent_config.json'), ('encoder_config', a.source_dir/'encoder/config.json'), ('tokenizer', a.source_dir/'tokenizer/tokenizer.json'), ('common', common.__file__)]:
        if digest(path) != pins[key]:
            raise ValueError('asset changed: ' + key)
    data = json.loads(a.requests.read_text())
    if data['Schema'] != 'riido-source-requests-v1' or data['PlanSHA256'] != PLAN_SHA or data['ModelSHA256'] != pins['source_weights'] or len(data['Cases']) != 1106:
        raise ValueError('request contract mismatch')
    if data['SourceSHA256'] != '4d1a227fe3652992d1e98f43fa9dc1b8f54a6c27a1a24b9d182bdfb8cbc3274c' or data['MembershipSHA256'] != 'd90eda8e8a423eff44847ff46ef19b468a57ac2791d45c57e7e8ba39a4ee9902':
        raise ValueError('source/split mismatch')
    versions = {n: importlib.metadata.version(n) for n in ('laya','torch','transformers','tokenizers','safetensors')}
    if versions['laya'] != '0.3.21':
        raise ValueError('Laya reference changed')
    started = time.monotonic()
    torch.set_num_threads(4)
    torch.set_num_interop_threads(1)
    torch.manual_seed(1729)
    cfg = json.loads((a.source_dir/'rl_agent_config.json').read_text())
    tok = PreTrainedTokenizerFast(tokenizer_file=str(a.source_dir/'tokenizer/tokenizer.json'), cls_token='[CLS]', sep_token='[SEP]', mask_token='[MASK]', pad_token='[PAD]', unk_token='[UNK]')
    temperature = common.clamp_temperature(cfg['temperature_by_options'].get('noul:2',cfg['temperature'][2]))
    # Check every input before the first model inference; no labels are provided.
    for req in data['Cases']:
        if req['Kind'] != 'noul' or len(req['State']) > 16384 or len(req['Instruction']) > 8192:
            raise ValueError('input shape mismatch')
        q = {'t':'noul','ins':req['Instruction']}
        opts = common.render_options(q)
        order = [0,1] if req['Options'] == opts else [1,0]
        if req['Options'] != [opts[i] for i in order]:
            raise ValueError('option labels mismatch')
        ids, markers = common.build_sequence(tok,req['State'],q,option_order=order)
        if ids != req['IDs'] or markers != req['Markers']:
            raise ValueError('Go/upstream token parity mismatch')
    model = common.build_model(cfg, encoder_dir=str(a.source_dir/'encoder'), pretrained=False)
    model.load_state_dict(load_file(a.source_dir/'model.safetensors'), strict=True, assign=True)
    model.encoder.config.reference_compile = False
    model = model.float().eval()
    scores = []
    last = started
    with torch.inference_mode():
        for req in data['Cases']:
            guard(started)
            before = time.monotonic()
            ids, markers = req['IDs'], req['Markers']
            logits = model(torch.tensor([ids]),torch.ones((1,len(ids)),dtype=torch.long),torch.tensor([markers]),torch.ones((1,2),dtype=torch.bool),torch.tensor([2]))[0][0].double()
            probs = torch.softmax(logits/temperature,dim=-1)
            if not torch.isfinite(probs).all():
                raise ValueError('nonfinite probability')
            prediction = {'probabilities':probs.tolist(),'winner':int(probs.argmax()),'truncated':req['Truncated'],'tokens':len(ids),'ms':(time.monotonic()-before)*1000}
            scores.append({'RequestSHA256':request_hash(req),'Prediction':prediction})
            if time.monotonic()-last > 20:
                print(json.dumps({'completed_calls':len(scores),'total_calls':1106}),flush=True)
                last=time.monotonic()
    guard(started)
    result = {'Schema':'riido-source-scores-v1','RequestsSHA256':digest(a.requests),'PlanSHA256':PLAN_SHA,'ModelSHA256':pins['source_weights'],'Cases':scores,'TokenParityMatches':len(scores),'Seconds':time.monotonic()-started,'PeakRSSBytes':resource.getrusage(resource.RUSAGE_SELF).ru_maxrss,'Versions':versions}
    # Contains row-level scores; local-only, never publish to Git or HF.
    with a.out.open('x') as f:
        json.dump(result,f)
    os.chmod(a.out,0o600)
    print(json.dumps({'completed_calls':len(scores),'seconds':result['Seconds'],'peak_rss_bytes':result['PeakRSSBytes']}))


if __name__ == '__main__':
    main()
