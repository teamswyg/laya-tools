# Shared-16 claim MLP source prototype

`statehintclaimsmlp` implements pure-Go, float32 `2048 contextual sparse inputs
→ 16 shared ReLU units → 3 independent 3-state categorical heads`. The extractor
is exactly `statehintwide.ExtractContextual`; head/state/source/prediction types
are aliases to `statehintclaims`. Each head has its own three-way softmax and
float64 dot-product, loss and probability accumulation. There is no generative
output, qualification, task/application-state authority or runtime registration.

The parameter arrays contain 32,937 float32 values: 32,912 weights and 25 biases.
`Model.Predict`/`Scores` are read-only and require a caller-owned `Workspace`.
Concurrent callers use separate workspaces. `Clone` and `Parameters` return owned
value copies. No mutable weight importer, global cache, map, lock or hot-path
allocation is provided. `Metadata` always reports `qualified=false` and
`state_authority=false`. Predict keeps the existing claim gates and precedence:
no word content, untrained, semantic unknown, confidence below .9, margin below
.05. T is fixed at 1; no-word predictions are uniform with an unknown winner.

`Fit(samples, trainingWorkspace)` has one fixed cold recipe: 40 epochs, batches
of 32 (including a shorter last batch), AdamW learning rate .001, weight decay
.01 on both weight blocks, no bias decay, seed 1729, hard categorical targets,
and mean three-head cross-entropy. ReLU's derivative at zero is zero. Every Fit
resets the working model, both moments and step count, with no warm start,
hyperparameter flags, smoothing, temperature tuning or retained text/features
corpus. All targets, bounded UTF-8 and nonempty word content validate before
updates; errors leave the receiver unchanged. 1,680 supplied samples imply
`40 × ceil(1680/32) = 2,120` updates; this arithmetic is not a real-data run.

Fresh Glorot initialization uses PCG `(1729, 1729 XOR 0x696e697472636d31)` and
draws `float32((2*u-1)*limit)` in input-feature/hidden, then hidden/head/state
order. Input limit is `sqrt(6/(2048+16))`; each independent output head's limit is
`sqrt(6/(16+3))`. Both bias blocks start at zero. Shuffle uses a separate PCG
stream `(1729, 1729 XOR 0x9e3779b97f4a7c15)`, matching the cold linear claim
recipe's sample ordering. The linear control starts at zero, whereas this MLP
starts at Glorot weights, so initial states are not identical; the data recipe
and update recipe are matched. This package does not inherit the old eight-way
MLP defaults of rate .02/decay .001 or its weights.

`RCM\0` v1 is a distinct little-endian artifact: 192-byte header, 131,748-byte
float32 payload and 32-byte SHA-256, totaling 131,972 bytes. The header binds
dimensions, ReLU/init codes, seed, steps, feature-schema hash, ordered head/state
hash, numeric-contract hash, .9/.05/T1, training samples, parameter/weight counts
and zero reserved bytes. Payload order is input, hidden bias, output, output
bias. Load requires exact length, checksum, finite parameters, seed 1729 and the
fixed training-sample/step relation (untrained means both counts zero). Counts
and dimensions never control allocation. SHA-256 detects corruption and does
not authenticate a publisher. Optimizer, corpus and parent weights are absent.

Only original synthetic tests establish numerical correctness, serialization,
read-only concurrency and allocation behavior. No real corpus is fitted or
queried, and no quality, speed, storage advantage or calibrated-performance
claim follows from these tests. Ternary/BitNet adapters are outside this package.
