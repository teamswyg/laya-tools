# A tiny shared network for development-work claim hints

The goal is inexpensive response-request, activity and completion-report suggestions. A report is a textual claim, not proof of completed work. Semantic-contrast data improved some research metrics in the linear model, but completion precision and private-comment usefulness remain insufficient. This study tests nonlinear reuse of the same input representation.

## Fixed comparison

Both new models use the unchanged1,680reviewed rows,2,048contextual features,40epochs,batch32,AdamW0.001/decay0.01,seed1729 and2,120updates, with fresh moments and no bias decay. The cold linear model starts at zero. The MLP has16shared ReLU units and three independently normalized three-state heads, with fixed Glorot input limit sqrt(6/(2048+16)) and per-head output limit sqrt(6/(16+3)). Initialization differs; data, order and update recipe are matched.

The prior4,240-update float candidate is descriptive only, not an equal-history control. Preserve T1/confidence.9/margin.05 and existing qualification requirements. This is neither a single nine-way label selector nor a generative LLM. It cannot restore absent attribution/scope/context merely by adding nonlinearity.

## Data and decision rule

Author, independently review, adjudicate, audit and seal a fresh360-row/180paired-family/60-group evaluation before model outputs. Never reuse old exposed evaluation, private work or calibration/test as training/selection material. Each locale's actual wording is judged independently; translations need not have forced equal targets. AI references are not human gold or product evidence.

Primary MLP-minus-cold-linear mean three-head CE uses paired six-row KO/EN groups,40/20contrast/general strata,2:1weights,10,000bootstrap draws,seed1729 and95%linearly interpolated percentiles. Research progress requires CI upper<0, lower unknown CE in both locales, no added true FP in any locale/head, and no general-stratum CE regression. No width/seed/epoch/threshold grid or third real Fit after results. The two-Fit research budget is separate from synthetic numerical tests and operational qualification.

## Go interface and measurement

The array-based `pkg/statehintclaimsmlp` stores32,937float32parameters with float64accumulators. Caller-owned reusable workspace permits immutable-model concurrency without model caches, maps or inference locks. RCM artifacts are131,972bytes. Separate artifact/model/workspace bytes, Go heap, process RSS and scoped timing; no size/test/CI result alone proves semantic quality or speedup. No GPU/MPS is required.

`riidolaya claims --model MODEL.rcm --jsonl` explicitly loads RCM and returns unqualified research JSON with architecture metadata. Preserve RSC/RQT compatibility and explicit choice; no automatic registration, download or app writes.

The maintainer `cmd/riido-statehint-claims-mlp-study` admits pinned bounded local regular files, rejects symlinks/path/inode aliases and drift, and uses a fresh private output. Check mode never opens/hashes evaluation or prior-reference bodies and performs no Load/Predict/Fit/output. Actual mode completes and persists both cold models before evaluation body access. Consumption/failure records and saved models persist without implicit retry; file/directory sync follows supported filesystem semantics.

Use Go1.27.1 and the runner help for exact pins/output flags. No weights, evaluation prose or raw profiles enter Git. Future ternary compression is separate; this float study makes no BitNet or compression-performance claim. [Glorot initialization](https://proceedings.mlr.press/v9/glorot10a.html) and [AdamW](https://arxiv.org/abs/1711.05101) are method references, not guarantees on this task.
