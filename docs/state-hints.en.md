# Content labels, emoji and state hints

[한국어](state-hints.ko.md) · [V3 research candidate](../experiments/state-hints-v3/README.en.md) · [V3 Wiki](https://github.com/teamswyg/laya-tools/wiki/State-Hints-V3-EN) · [Historical v1 training methods and frozen plan](../experiments/state-hints/TRAINING.md) · [Go results](../experiments/state-hints/results/go-v1.json) · [Laya results](../experiments/state-hints/results/laya-v1.json)

The first application scope is **label/emoji display proposals for Riido development progress, completion reports and questions**. A textual interpretation does not establish real completion or authority over the current work state. Its eight fixed intents are `question`, `blocker`, `reference`, `progress`, `completion_report`, `cancel_request`, `planned` and `unclear`. Model selection and reasoning-level adjustment are outside this feature's scope.

The latest Go V3 is a published research candidate with complete readback. On the new synthetic test of 200, correctness rose 152→158 versus the parent, but correct/all proposals changed 57/57→90/94 and completion mistakes rose 0→4; the candidate is not enabled by default. See the [pinned HF model](https://huggingface.co/JooYoon/riidolaya-statehint-go-v0.3/tree/cf6deca9eb2c9919b1bbbbe82ce3576453d67a40) and [full V3 record](../experiments/state-hints-v3/README.en.md). Its 760 original synthetic messages use AI-authored annotations, not product truth or independent human review. The test is now exposed development evidence.

Outputs are shadow proposals without executed writes. The [placement module](../pkg/statehintplacement/README.en.md) checks complete opaque work revisions and already-held text before and after classification. [PR160](https://github.com/teamswyg/laya-tools/pull/160) passed [Linux, macOS and final CI](https://github.com/teamswyg/laya-tools/actions/runs/37430259670) and merged through the bot. Its synthetic reader tests are distinct from real application owner reads, permission and active-catalog integration, which remain unverified.

## Use and application conditions

Install the lightweight dedicated CLI with Go 1.27.1. It provides the same classification feature as `riidolaya state-hint`.

```sh
go install github.com/teamswyg/laya-tools/cmd/riido-statehint@ddd090fe49230a1ec52d9919c3105ca327eddd66
riido-statehint --text "I am adjusting the explanation in the result view now." --json
riido-statehint --model ./statehint.rsh --jsonl < requests.jsonl
```

The [V3 Wiki](https://github.com/teamswyg/laya-tools/wiki/State-Hints-V3-EN) explains optional artifact download and SHA verification.

The classification command without `--model` uses an **unlearned rule baseline** without a model file. Its one-hot scores represent rule selection, rather than calibrated learned probabilities. A trained model is loaded from an explicitly supplied local `.rsh` file. There is no automatic model download. The basic JSONL request is `{"text":"content to classify"}` with an optional `context`. The CLI does not repeat the source text in its output.

Text must be valid UTF-8 and no more than 4,096 bytes. JSONL requests have a 16 KiB boundary and maximum nesting depth of 12. Ambiguous duplicate keys, unknown fields, trailing JSON values and invalid context are rejected. The trained classifier treats whitespace, punctuation and emoji-only input as `unclear` with the `no_word_content` reason.

The caller supplies current active label IDs mapped to intents. Emoji proposals use only canonical lowercase codes in the supplied candidate list and are for display. Existing matching labels or emoji produce no additional proposal. `unclear`, untrained models and results below the threshold also abstain. Plan confidence has a 0.9 floor; the CLI's default margin is 0.05.

### Existing generic state-plan interface: separate from the display pilot

The following preserves the compatibility interface; it does not authorize work-state mutation in the initial display pilot.

Generic states are `todo`, `active`, `done` and `cancelled`. A state proposal requires matching `progress`/`started`, `completion_report`/`completed` or `cancel_request`/`cancelled` evidence, plus matching work ID and current version. Ordinary content hints cannot reopen completed or cancelled work. Selecting the current state again is a state no-op. Conflicting events or text without event evidence cannot change a state.

`expected_version` and `command_id` are required, but **putting `trusted:true` in JSON does not establish permission or prove an event.** A future trusted adapter must verify event provenance, the current catalog, permission, current revision and command replay when applying a change. The present CLI neither supplies that authority nor executes writes.

## Historical v1: two training methods

The v1 methods, data, measurements and failures below remain preserved. Do not combine them with V3 results or treat them as current production quality.

| Method | Work actually performed | Scope |
|---|---|---|
| Small Go model | Hash Unicode lowercase words and character 2/3-grams into 1,024 bins; train an eight-intent linear softmax with cross-entropy and AdamW | Go content classification and shadow proposals |
| Existing Laya model | Freeze the original checkpoint's prefix; fine-tune only the 1,025 final `scorer.3.weight/bias` parameters with cross-entropy and AdamW | Maintainer MPS comparison |

The Go model uses fixed arrays and caller-owned workspaces, with no locks or shared mutable cache. Cloned weights support warm-start fitting. Each fit starts a fresh optimizer, while temperature is selected separately on calibration data. Its 8,200 parameters produce a 32,960-byte file. `.rsh` persists the feature schema, intent order, temperature and SHA-256, rejecting corruption, nonfinite values and trailing bytes. Model binaries are excluded from Git.

Laya uses the original English `convaiinnovations/laya` checkpoint with a pinned revision and file hashes. Its prefix remains in evaluation mode. Each sentence's final-linear input is extracted once and reused during training epochs. This lowers final-linear training work, but inference still needs the complete Laya backbone. The small Go model is neither an export of that backbone nor a distilled model.

Both methods select candidates by validation NLL and fit temperature on a separate calibration partition. Final-test evaluation follows the weight/temperature lock. The [training document](../experiments/state-hints/TRAINING.md) records the recipes and resource limits.

## Historical v1: data and results at that stage

The public dataset contains 2,400 original synthetic sentences: fictional name/number variants of 96 language-specific template families and 48 Korean/English concept pairs. It contains no real user content. Partitions are train 1,200 rows/48 families, validation 400/16, calibration 400/16 and test 400/16. Every partition remains synthetic development data; **these are not 2,400 independent product truths.**

Accuracy below means selecting the highest-scoring intent on the final test. Accepted rows have confidence at least 0.9 and a non-`unclear` prediction; they are not executed writes.

| Method | Correct/400 overall | Correct/200 Korean | Correct/200 English | Accepted/400 | Correct among accepted |
|---|---:|---:|---:|---:|---:|
| Rule baseline | 200 (50%) | 100 (50%) | 100 (50%) | 150 (37.5%) | 150/150 |
| Trained Go, temperature 0.7 | 317 (79.25%) | 170 (85%) | 147 (73.5%) | 136 (34%) | 136/136 |
| Original Laya, raw | 148 (37%) | 25 (12.5%) | 123 (61.5%) | 93 (23.25%) | 72/93 (77.4%) |
| Fine-tuned Laya, raw | 148 (37%) | 25 (12.5%) | 123 (61.5%) | 50 (12.5%) | 50/50 |
| Fine-tuned Laya, temperature 0.65 | 148 (37%) | 25 (12.5%) | 123 (61.5%) | 74 (18.5%) | 50/74 (67.6%) |

Go's overall accuracy exceeded the rules, while its accepted coverage was smaller. The 136/136 observation applies to these 16 test families and does not establish 100% operational precision.

Laya's raw test NLL fell from 1.8307 to 1.6071 without increasing correct predictions. Applying the calibration-selected temperature produced test NLL 1.7281, worse than raw tuned logits, and 24 of the 74 accepted predictions were wrong. These failures are retained: changes in NLL or calibration do not guarantee more correct answers.

The MPS Laya run took approximately 181.77 seconds and produced a 4,252-byte parameter delta. Frozen parameter hashes matched; cached final-layer reload error was zero. One validation item through the complete MPS model had reload error approximately `9.54e-7`. Go serialization parity covered predictions on all 2,400 inputs.

For one short sentence measured 2,000 times in the same warm process, Go p50 was 1.833µs, p95 1.875µs and allocations per call zero. This includes feature extraction and scoring only. JSON handling, plan generation, file reads, process startup and total RSS are excluded. The model struct's 32,816 bytes and workspace's 22,552 bytes are also not total memory consumption.

## Historical v1: observed failures and next steps at that stage

Go classified every `completion_report` test example as `progress`, giving **0/50 recall**. It also classified all 25 English `question` examples as `completion_report`. Completion-state judgment is therefore unqualified. Laya achieved 12.5% Korean test accuracy, with no overall accuracy improvement after fine-tuning, so this checkpoint is not adopted for deployment.

Next work should add distinct completion/progress, question, quotation and hypothetical examples, and create a new evaluation partition without turning this test into a selection criterion. A separate multilingual checkpoint or broader adapter training needs fresh provenance and language-specific evaluation. Actual Riido integration should begin with a trusted adapter's read and shadow observations, measuring incorrect label/state proposals and complete processing cost before adoption.
