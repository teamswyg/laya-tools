# Three claim hints for development comments

This study seeks inexpensive response-request, activity and completion-report hints for Riido development comments. A real private diagnostic found the previous eight-intent parent confidently classifying a reminder as completion. We preserve that failure and test a new target suited to co-occurring claims instead of fitting thresholds to the observation.

Each attribute independently emits `true`, `false` or `unknown`. Different substeps can be reported complete and in progress together. False means the corresponding positive claim is absent, not that the task is unfinished. Results distinguish the raw `winner`, emitted `state` and `unknown_reason`: learned semantic uncertainty is different from confidence-based abstention.

This is a **research preview**. Confidence is not an accuracy guarantee. Outputs neither verify completion nor authorize labels, emoji or task-state writes. They do not replace the original classifier or its quality gates. Actor, activity scope and provenance are not invented from the numeric outputs; a real adapter must check them separately.

## Usage for people and agents

Build with Go 1.27.1 and supply an explicit local research model. No automatic download or default registration occurs. RSC is distinct from RSH, RSM and RSP artifacts.

```sh
go build -trimpath -o riidolaya ./cmd/riidolaya
./riidolaya claims --model MODEL.rsc --text 'The type check passed and I am running regression tests now.' --json
```

JSONL loads the model once for a warm stream. Each line contains exactly one `text` field. Keep private inputs and outputs local.

```sh
printf '%s\n' '{"text":"An original fictional development comment."}' |
  ./riidolaya claims --model MODEL.rsc --jsonl
```

The three named outputs are `response_requested`, `current_activity_claimed` and `completion_claimed`, with per-head probabilities, raw winner and abstention reasons. Both `semantic_quality_qualified` and `mutation_executed` remain false. Invalid or out-of-scope requests produce a reason without echoing input prose.

## Training and evidence

The [English rubric](RUBRIC.en.md) targets development-work reports. The initial review of broad-domain original synthetic text found runtime descriptions being counted as work completion and conflicting traces being counted as unclear work claims. Preserve those annotations as history; no model was fitted to them. Prepare development-comment data separately instead of converting the old eight labels. Review both locales separately with evidence and uncertainty. AI reference annotation is not human gold.

The first model uses three categorical linear heads over the unchanged 2,048-bin contextual word/character representation. It minimizes mean three-head cross entropy with fresh AdamW: temperature 1, 40 epochs, batch 32, learning rate 0.001, weight decay 0.01, seed 1729. This directly trains new weights; it is not fine-tuning Laya or the existing parent's weights.

The artifact format is 73,988 bytes, which is not whole-process memory. CPU-only Go inference reads shared activated features once for nine output columns. Models are read-only during inference and workspaces are caller-owned, without global caches or inference locks. No SIMD or measured speedup is claimed without evidence.

Split new development-comment groups into training, validation, calibration and test before authoring text. Preserve the broad-domain data and its 663/177 split as separate history. AI development evaluation does not replace real-comment accuracy. Original calibration/final-test data are neither opened nor trained on. Keep raw score records and profiles local; only licensed original synthetic material and safe aggregates may be published as research assets. Model binaries belong outside Git.

Three output states are distinct from 1.58-bit weights. Establish float32 semantic quality first; weight compression and Laya-representation training remain separate experiments. Shout-out to [laya.tools](https://laya.tools/).
