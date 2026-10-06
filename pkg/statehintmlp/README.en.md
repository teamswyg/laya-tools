# Fixed contextual MLP experiment

An isolated Go model: the existing Contextual2048 feature representation,
16 ReLU hidden units and eight softmax outputs. It has 32,920 float32 trainable
parameters (131,680 parameter bytes). Its separate checksummed `RSM` version3
artifact is 131,872 bytes. No existing `statehint`, `statehintwide`, artifact,
driver or product-state API is changed.

```go
m := statehintmlp.NewModel() // fresh Glorot-PCG initialization, seed 1729
report, err := m.Fit(ownedSamples, statehintmlp.FitOptions{
    Epochs: 40, BatchSize: 32, LearningRate: .02,
    WeightDecay: .001, Seed: 1729,
})
if err != nil { /* handle training error */ }
_ = report
var workspace statehintmlp.Workspace
prediction, err := m.Predict("an original bounded input", &workspace)
_ = prediction
_ = err
```

By default, `Fit` uses hard eight-way cross-entropy and AdamW, averaging each minibatch's
gradient, correcting both moments, and decaying weights but not biases.
Initial weights use Glorot uniform limits `sqrt(6/(fanIn+fanOut))`; biases are
zero. Initialization and sample shuffling use separate PCG streams from the
recorded seed. Every valid Fit starts **fresh** weights/optimizer/steps and
temperature 1, including when called on a previously trained or loaded model.
It validates every sample before work and commits the new model only after
successful training. Invalid input or a numerical failure preserves the old
model. Initialization method and seed appear in the report and artifact.

Zero `FitOptions` values mean 40 epochs, batch 32, learning rate 0.02,
weight decay 0.001 and seed 1729. In particular, zero weight decay means the
default 0.001; this API does not express a no-decay experiment. This differs
from the wide API's zero-decay behavior. Set every recipe field explicitly
when making a controlled comparison. No sampling, calibration fitting,
selection, corpus IO, deployment or state mutation is implemented here.

## Optional training label smoothing

`FitOptions.LabelSmoothing` defaults to zero and accepts finite values from
0 to 0.2. Zero delegates to the original hard-CE arithmetic; a pre-change
owned toy artifact fingerprint checks byte compatibility. Positive alpha uses
the uniform eight-way target `q=(1-alpha)*one_hot+alpha/8`, cross-entropy
`-sum(q*log(p))` and gradient `p-q`. Soft loss uses shifted log-sum-exp to avoid
large common-logit cancellation. `FitReport.LabelSmoothing` records the value;
stored sample labels, initialization, shuffle, optimizer and inference do not
change. The v3 artifact remains 131,872 bytes and does not encode the training
objective: retain the FitReport and run recipe for objective provenance.

The fixed prospective comparison is hard CE versus alpha 0.05, giving a
target of 0.95625 for the declared class and 0.00625 for each other class.
There is no alpha sweep or automatic calibration. [Müller, Kornblith and
Hinton (2019)](https://proceedings.neurips.cc/paper_files/paper/2019/hash/f1748d6b0fd9d439f71450117eba2725-Abstract.html)
report empirical calibration/generalization benefits of label smoothing;
this supports a hypothesis, not a guarantee for this small model. It may
reduce wrong completion confidence and also reduce correct completion
coverage. Existing gates remain unchanged, exposed development results
remain development-only, and calibration/final-test data remain sealed.
Only owned numeric/toy tests have exercised this new option so far.

The bounded extractor is a private Apache-2.0 copy of
[the existing contextual extractor](../statehintwide/features.go). It keeps
UTF-8/4096-byte/NUL checks, Unicode lowercasing, word unigrams/bigrams and
character 2–5-grams, signed FNV-1a hashing, log-TF and L2 normalization.
Feature-copy drift is tested against **every one of 2048 bins** on six owned
fixtures recovered through public-v2 linear probe artifacts, without editing the old
extractor. This is an implementation parity check, not semantic evidence.

Only text becomes features. `Intent`/`Prediction` reuse the common API and
persisted eight-intent order. The learned/untrained `Source` wire values are
historically named `linear_softmax_*`; they are retained as compatibility tags
for the unchanged common evaluator, **not a claim that this model is linear**.
The v3 artifact and this package identify the MLP architecture. No-word input
returns a uniform posterior, unclear tie and `no_word_content` guard.

Predict is read-only and a reused caller Workspace requires no allocation.
Concurrent predictions need separate workspaces while the model is immutable;
Fit and SetTemperature require exclusive model ownership. There are no maps,
global caches, locks, native/GPU dependencies or implicit product actions.

Save/Load bind architecture dimensions, ReLU/initialization codes, feature
schema, intent-order hash, seed, temperature and steps. Load reads only the
fixed artifact plus one trailing-byte check and rejects truncation, extra data,
schema/version/reserved-byte drift, corruption and nonfinite parameters.
V1/v2/v3 loaders reject each other's artifacts; checksum verification detects
corruption and does not authenticate provenance. Generated model binaries
must remain outside source control.

Tests use owned artificial text/numerical fixtures: active/inactive ReLU
finite-difference gradients, stable large-logit CE, bias-free AdamW decay,
actual tiny learning, fresh deterministic/transactional Fit, concurrent
prediction, malformed artifacts and exact trained byte/prediction reload
parity. One controlled run on the existing training split produced 322/354
internal-development correct rows, but completion precision was 35/37 (94.59%),
below the 98% gate. The existing linear control remained selected. On the already
exposed diagnostic, the MLP reached 174/240 correct rows and only 1/15 English
completion families; both arms remained unqualified. No calibration/final-test
forward or promotion occurred. These are development observations, not fresh
evaluation or proof of semantic generalization. It may learn nonlinear
combinations of existing bins; it cannot recover discarded
order/hash information or establish language understanding from those tests.
Any future internal-dev data is already exposed development material, with
external validation diagnostic only and the final test sealed.

[한국어](README.ko.md)
