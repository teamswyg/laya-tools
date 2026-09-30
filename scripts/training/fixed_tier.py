"""Frozen Laya encoder + small fixed-tier classifier. Not a replacement Laya choice head."""
import argparse
import itertools
import json
import resource
import time
from pathlib import Path
import os
os.environ['PYTORCH_ENABLE_MPS_FALLBACK'] = '0'
os.environ['USE_TF'] = '0'
os.environ['TOKENIZERS_PARALLELISM'] = 'false'
import torch
from safetensors.torch import load_file, save_file
from pdca_train import setup, load_data, evaluate, detailed, assess, keyword_reference
from train_pilot import file_hash, frozen_hash, guard, metrics

LABELS = ['fast', 'standard', 'strong']
FORMAT = 'laya-encoder-fixed-tier-linear-v1'


class FixedTier(torch.nn.Sequential):
    def __init__(self, width=1024):
        super().__init__(torch.nn.LayerNorm(width), torch.nn.Linear(width, 3))


def state_feature(model, tok, text):
    """State-only mean pooling; no question/options or corpus-fitted normalization."""
    b = tok(text, return_tensors='pt', truncation=False)
    if b['input_ids'].shape[1] > 512:
        raise ValueError('State exceeds 512 tokens; no silent truncation')
    device = next(model.parameters()).device
    ids, mask = b['input_ids'].to(device), b['attention_mask'].to(device)
    content = mask.bool() & ids.ne(tok.cls_token_id) & ids.ne(tok.sep_token_id)
    if not content.any(): raise ValueError('Empty state')
    with torch.no_grad():
        hidden = model.encoder(input_ids=ids, attention_mask=mask).last_hidden_state
        feature = (hidden * content[..., None]).sum(1) / content.sum(1, keepdim=True)
    if not torch.isfinite(feature).all(): raise ValueError('Non-finite encoder feature')
    return feature[0]


def features(model, tok, rows, start, out):
    result = []
    for row in rows:
        guard(start, 20, out)
        result.append(state_feature(model, tok, row['state']).cpu())
    return torch.stack(result)


def state(head):
    return {k:v.detach().cpu().contiguous().clone() for k,v in head.state_dict().items()}


def logits(head, x):
    head.eval()
    with torch.no_grad():
        return head(x.to(next(head.parameters()).device)).cpu()


def train(args, plan, rows):
    if args.out.exists(): raise ValueError('New output directory required')
    args.out.mkdir(parents=True)
    start = time.monotonic()
    model, tok = setup(args.base, args.seed)
    model.requires_grad_(False); model.eval()
    frozen = frozen_hash(model)
    training = [r for r in rows if r['split']=='train']
    validation = [r for r in rows if r['split']=='validation']
    train_x = features(model,tok,training,start,args.out).to('mps')
    val_x = features(model,tok,validation,start,args.out).to('mps')
    train_y = torch.tensor([r['label'] for r in training],device='mps')
    val_y = [r['label'] for r in validation]
    torch.manual_seed(args.seed)
    head = FixedTier(train_x.shape[1]).to('mps')
    if sum(p.numel() for p in head.parameters()) != plan['head_parameter_count']:
        raise ValueError('Unexpected classifier dimensions')
    initial = state(head)
    records, best, selected = [], float('inf'), None
    driver_max = 0
    updates = 0
    for lr in plan['learning_rates']:
        head.load_state_dict(initial)
        optimizer = torch.optim.AdamW(head.parameters(),lr=lr,weight_decay=plan['weight_decay'])
        for epoch in range(plan['max_epochs']):
            if epoch % 10 == 0: guard(start,20,args.out)
            head.train()
            order = torch.randperm(len(training),generator=torch.Generator().manual_seed(args.seed+epoch)).to('mps')
            total = 0.
            for ids in order.split(plan['batch_size']):
                optimizer.zero_grad(set_to_none=True)
                loss = torch.nn.functional.cross_entropy(head(train_x[ids]),train_y[ids])
                loss.backward()
                if not torch.isfinite(loss) or any(not torch.isfinite(p.grad).all() for p in head.parameters()):
                    raise RuntimeError('Non-finite classifier update')
                torch.nn.utils.clip_grad_norm_(head.parameters(),1.)
                optimizer.step()
                total += float(loss.detach())*len(ids); updates += 1
            torch.mps.synchronize()
            driver_max = max(driver_max,torch.mps.driver_allocated_memory())
            val = metrics(logits(head,val_x),val_y)
            rec = dict(lr=lr,epoch=epoch+1,training_loss=total/len(training),validation=val,selection_score=val['nll'])
            records.append(rec)
            if epoch == 0 or (epoch+1)%10 == 0: print(json.dumps(rec),flush=True)
            if val['nll'] < best:
                best,selected=val['nll'],rec
                save_file(state(head),args.out/'head.safetensors')
    learned = load_file(args.out/'head.safetensors')
    changed = any(not torch.equal(learned[k],v) for k,v in initial.items())
    if not changed or frozen_hash(model) != frozen: raise RuntimeError('Training integrity failure')
    report = dict(cycle=plan['cycle'],format=FORMAT,seed=args.seed,selection=selected,history=records,
                  updates=updates,train_rows=len(training),validation_rows=len(validation),
                  head_sha256=file_hash(args.out/'head.safetensors'),
                  plan_sha256=file_hash(args.data/'plan.json'),data_sha256=file_hash(args.data/'families.json'),
                  legacy_train_sha256=file_hash(args.data.parent/'pilot-v1.tsv'),
                  frozen_weights_unchanged=True,head_weights_changed=True,final_test_used_for_selection=False,
                  trainable_parameters=sum(p.numel() for p in head.parameters()),
                  feature_width=train_x.shape[1],feature_cache_bytes=(train_x.numel()+val_x.numel())*4,
                  elapsed_seconds=time.monotonic()-start,peak_process_rss_bytes=resource.getrusage(resource.RUSAGE_SELF).ru_maxrss,
                  sampled_driver_max_bytes=driver_max)
    (args.out/'selection.json').write_text(json.dumps(report,indent=2)+'\n')


def check(args,plan,rows):
    s=json.loads((args.out/'selection.json').read_text())
    if s['seed'] != args.seed or s['format'] != FORMAT: raise ValueError('Wrong frozen candidate')
    if (args.out/'evaluation.json').exists(): raise ValueError('Refusing to overwrite final')
    for key,path in [('head_sha256',args.out/'head.safetensors'),('plan_sha256',args.data/'plan.json'),('data_sha256',args.data/'families.json')]:
        if s[key] != file_hash(path): raise ValueError('Frozen input mismatch')
    start=time.monotonic()
    model,tok=setup(args.base,args.seed)
    model.requires_grad_(False);model.eval()
    test=[r for r in rows if r['split']=='test'];cal=[r for r in rows if r['split']=='calibration']
    legacy=[dict(id=r['id'],state=r['prompt'],label=LABELS.index(r['expected']))
            for r in json.loads(Path('benchmarks/routing-golden.json').read_text()) if r['language']=='en']
    temperatures=[.5+i*.05 for i in range(91)]
    def fit(z): return min(temperatures,key=lambda t:metrics(z,[r['label'] for r in cal],t)['nll'])
    base_t=fit(evaluate(model,tok,cal,start,args.out))
    baseline=detailed(evaluate(model,tok,test,start,args.out),test,base_t)
    base_legacy=detailed(evaluate(model,tok,legacy,start,args.out),legacy,base_t)
    head=FixedTier(s['feature_width']).to('mps')
    head.load_state_dict(load_file(args.out/'head.safetensors'))
    cal_x=features(model,tok,cal,start,args.out)
    test_x=features(model,tok,test,start,args.out)
    legacy_x=features(model,tok,legacy,start,args.out)
    temperature=fit(logits(head,cal_x))
    z=logits(head,test_x)
    head.load_state_dict(load_file(args.out/'head.safetensors'))
    reload_error=float((z-logits(head,test_x)).abs().max())
    if reload_error>1e-4: raise RuntimeError('Reload mismatch')
    tuned=detailed(z,test,temperature);tuned_legacy=detailed(logits(head,legacy_x),legacy,temperature)
    # Fixed labels are output semantics, never an input choice list. Check only mapping.
    flips=0
    for order in itertools.permutations(range(3)):
        if order == (0,1,2): continue
        serialized=z[:,list(order)]
        restored=serialized[:,[order.index(i) for i in range(3)]]
        flips+=int(restored.argmax(-1).ne(z.argmax(-1)).sum())
    rate=flips/(5*len(test))
    passed=assess(plan,baseline,tuned,base_legacy,tuned_legacy,rate)
    report=dict(seed=args.seed,format=FORMAT,selected_head_sha256=s['head_sha256'],
                plan_sha256=s['plan_sha256'],data_sha256=s['data_sha256'],
                base_temperature=base_t,temperature=temperature,baseline=baseline,tuned=tuned,
                base_legacy=base_legacy,tuned_legacy=tuned_legacy,permutation_comparisons=5*len(test),
                permutation_flips=flips,permutation_flip_rate=rate,
                permutation_method='fixed-label output mapping; structural invariance, NOT raw Laya choice robustness',
                reload_max_logit_error=reload_error,criteria=passed,passed=all(passed.values()),
                keyword_reference=dict(correct=sum(keyword_reference(r['state'])==r['label'] for r in test),n=len(test)),
                limits='New fixed-class architecture; authored English small adaptive evaluation, not production/cost evidence')
    (args.out/'evaluation.json').write_text(json.dumps(report,indent=2)+'\n')
    print(json.dumps(dict(seed=args.seed,passed=report['passed'],criteria=passed,accuracy=tuned['accuracy'],
                         base_accuracy=baseline['accuracy'],coverage=tuned['coverage'],precision=tuned['accepted_precision'])),flush=True)


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('mode',choices=['train','check'])
    ap.add_argument('--base',type=Path,required=True);ap.add_argument('--data',type=Path,required=True)
    ap.add_argument('--out',type=Path,required=True);ap.add_argument('--seed',type=int,choices=[1729,2718],required=True)
    args=ap.parse_args();plan=json.loads((args.data/'plan.json').read_text())
    if plan['format'] != FORMAT: raise ValueError('Wrong architecture plan')
    (train if args.mode=='train' else check)(args,plan,load_data(args.data))


if __name__=='__main__':main()
