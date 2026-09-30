"""Build an allowlisted, non-production head-only model package from local results."""
import argparse
import hashlib
import json
from pathlib import Path
import shutil


def digest(p):
    with p.open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--run', type=Path, required=True)
    ap.add_argument('--base', type=Path, required=True)
    ap.add_argument('--encoder-card', type=Path, required=True)
    ap.add_argument('--output', type=Path, required=True)
    ap.add_argument('--source-revision', default='uncommitted-preview-not-publishable')
    args = ap.parse_args()
    root = Path(__file__).resolve().parents[2]
    r = json.loads((args.run / 'results.json').read_text())
    args.output.mkdir(parents=True, exist_ok=False)
    files = {
        'head.safetensors': args.run / 'head.safetensors', 'results.json': args.run / 'results.json',
        'training-data.tsv': root / 'benchmarks/training/pilot-v1.tsv', 'LICENSE': root / 'LICENSE',
        'requirements.txt': root / 'scripts/training/requirements-pilot-macos-arm64.txt',
        'UPSTREAM_MODEL_CARD.md': args.base / 'README.md', 'ENCODER_MODEL_CARD.md': args.encoder_card,
    }
    for dest, src in files.items():
        shutil.copyfile(src, args.output / dest)
    provenance = {'license': 'apache-2.0', 'dataset_origin': 'original_synthetic', 'production_ready': False,
                  'base_model': 'convaiinnovations/laya', 'base_revision': r['base_revision'], 'base_sha256': r['base_sha256'],
                  'encoder': 'answerdotai/ModernBERT-large', 'encoder_card_revision': '45bb4654a4d5aaff24dd11d4781fa46d39bf8c13',
                  'sdk': 'laya==0.3.21', 'format': 'fp32-head-replacement-not-lora',
                  'head_sha256': digest(files['head.safetensors']), 'data_sha256': digest(files['training-data.tsv']),
                  'source_repository': 'https://github.com/teamswyg/laya-tools',
                  'source_revision': args.source_revision,
                  'training_code_sha256': digest(root / 'scripts/training/train_pilot.py'),
                  'base_files_sha256': {str(p.relative_to(args.base)): digest(p) for p in args.base.rglob('*') if p.is_file()}}
    (args.output / 'PROVENANCE.json').write_text(json.dumps(provenance, indent=2) + '\n')
    notice = '''Experimental riidolaya head fine-tune. Copyright 2026 teamswyg contributors. Apache-2.0.
Derived from Laya by Convai Innovations: https://huggingface.co/convaiinnovations/laya
Original encoder/tokenizer: ModernBERT-large by Answer.AI and collaborators, Apache-2.0:
https://huggingface.co/answerdotai/ModernBERT-large
SDK: NandhaKishorM/laya, Apache-2.0. No endorsement or affiliation implied.
Original model cards are preserved separately; their claims are not claims about this fine-tune.
This package contains replacement head parameters, not the encoder, tokenizer, SDK or training framework.
'''
    (args.output / 'NOTICE').write_text(notice)
    (args.output / 'MODIFICATIONS.md').write_text('Changed head/type_emb/scorer parameters using supervised choice labels on original synthetic development examples. Encoder and action head frozen. Temperature fitted on a small separate development calibration partition. Stored FP32 replacement parameters, not LoRA matrices or a standalone model. See results.json for regressions and limitations.\n')
    test, base = r['test'], r['base_test']
    card = f'''---
license: apache-2.0
base_model: convaiinnovations/laya
language:
- en
tags:
- experimental
- decision-model
- mps
- head-only
- riidolaya
---
# riidolaya {r['task']} — experimental MPS pilot

**Research/development artifact. Not a production model, not a standalone checkpoint, not a LoRA adapter, not a drop-in Go ONNX model.** No automatic coding, plan generation, model switching or task dispatch is enabled by this release.

학습 실험 공개용입니다. 기본 모델로 권장하지 않습니다. 원본 가중치와 결합해야 하며 기존 Go ONNX 파일을 대체할 수 없습니다. 코드/계획을 생성하는 모델이 아니라 주어진 선택지를 평가합니다.

## What changed

The original Laya encoder and action/escalation head stayed frozen; {r['trainable_parameters']:,} head/type_emb/scorer parameters were eligible for updates on real Apple MPS, FP32, no CPU fallback. Two learning rates (1e-5, 5e-5) and up to five epochs were compared by validation NLL. Selected: learning rate {r['selected']['lr']}, epoch {r['selected']['epoch']}. The deployed part is about 100 MiB, avoiding a duplicate full encoder per specialist.

Choices: {', '.join(r['labels'])}. Exact question and option ordering are in results.json. Calibrated temperature: {r['temperature']}. Frozen-weight integrity and saved-head reload checks passed.

## Evidence and limits

Development test: tuned **{test['correct']}/{test['n']}**, original **{base['correct']}/{base['n']}**; confidence >=0.9 accepts **{test['accepted']}/{test['n']}**. Do not describe zero coverage or a regression as improved routing. Partition sizes: {json.dumps(r['split_counts'])}.

All 88 dataset entries are newly authored English synthetic examples, Apache-2.0. They contain no private code, customer prompts or third-party dataset. Labels reflect a rubric, not real downstream model outcomes. Related conceptual templates occur across splits. A first short run was observed before extending training, so these are development partitions, **not an untouched final holdout**. Calibration/test samples are tiny. Baseline uses original calibration while the candidate uses fitted calibration; probability-metric changes do not isolate weight-learning gains. No cost-saving claim, no Korean validation, no independent repository goldens, no ONNX/Go parity. Optimizer resume is not verified.

Process peak RSS: {r['peak_process_rss_bytes']/1024**3:.2f} GiB; step-end sampled MPS driver maximum: {r['step_end_mps_driver_max_bytes']/1024**3:.2f} GiB. These overlap in unified memory and must not be added. Driver samples are not continuous peaks. Measured loop including loading, evaluation, guards, hashes and checkpoint writes: {r['elapsed_seconds']:.2f} s; not a throughput benchmark.

## Maintainer loading

Use the scripts at [teamswyg/laya-tools](https://github.com/teamswyg/laya-tools/tree/main/scripts/training). Runtime users continue using the existing Go binary; Python here is only a research reference.

1. Install this package's requirements.txt in an isolated Python environment (tested macOS arm64 / Python 3.14.7).
2. Download the original `convaiinnovations/laya` revision `{r['base_revision']}`: model.safetensors, rl_agent_config.json, encoder/config.json, tokenizer/tokenizer.json and tokenizer/tokenizer_config.json. Verify PROVENANCE.json hashes. No remote code execution is required.
3. Download this entire release folder using the pinned Hub commit, retaining MANIFEST.json and notices. Set BASE_DIR and PACKAGE_DIR to those directories.
4. From the tools repository run:

```sh
python scripts/training/evaluate_head.py --base "$BASE_DIR" --package "$PACKAGE_DIR" --text 'Fix one spelling mistake in a help string.'
```

For decomposition supply both the request and a candidate plan with evidence. There is no standalone `AutoModel.from_pretrained` recipe for this head-only artifact. The loader converts base weights to FP32 **before** applying FP32 head replacements. It verifies the exact base SHA and file manifest. It uses the same SDK formatter, option order and calibration as training. It never executes the predicted action.

## Reproduction and licensing

Full original fixtures, per-candidate results, requirements, SHA manifest, source revisions, modified modules and license/attribution are included. Training loop and dataset documentation live in the source repository. Original source model cards declare Apache-2.0; full Apache terms and attribution are retained. No external noncommercial/share-alike training dataset was added. This is a review of declared source licenses, not a warranty of all upstream pretraining provenance.

원본·학습 데이터 출처, 수정 내용, 해시를 보존했습니다. 학습 성능이 부족한 결과도 공개하며 기존 모델을 덮어쓰지 않습니다. 다음 단계는 실제 공개 저장소의 작업군 분리 데이터, 근거 기반 레이블, 독립 평가, ONNX/Go 검증입니다.
'''
    (args.output / 'README.md').write_text(card)
    manifest = {p.name: digest(p) for p in args.output.iterdir() if p.is_file()}
    (args.output / 'MANIFEST.json').write_text(json.dumps(manifest, indent=2) + '\n')
    print('Prepared', r['task'], 'experimental package')


if __name__ == '__main__':
    main()
