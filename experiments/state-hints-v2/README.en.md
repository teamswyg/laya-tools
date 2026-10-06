# Second content-hint PDCA

[한국어](README.md) · [Evaluation rubric](RUBRIC.md) · [Training rubric](TRAINING-RUBRIC.md) · [Complete results](results/run1.json)

Separate authors wrote320 original free-context training messages and40validation/40calibration/40sealed evaluation messages. There are no name/number filler variants. All cases are fictional development material, not independent product truth. The evaluator knew V1 failures; the training author did not inspect new evaluation files.

Four candidates used seed1729, AdamW,lr0.02,40epochs,batch32. The V1 warm start actually updated the published V1 weights with a fresh optimizer. Controls trained fresh1,024-bin weights and two2,048-bin variants. Only validation NLL selected a candidate; calibration fitted temperature0.5–5.0. The final40-case file was first read after the weight/temperature lock. Confidence0.9 remained fixed.

| Candidate | Validation correct/40 | NLL | Selected |
|---|---:|---:|---|
| V1 warm start |32|0.6791|yes|
| V1 fresh |31|0.9007|no|
| 2,048-bin simple |34|0.8082|no|
| 2,048-bin contextual |30|0.9522|no|

The selected model remains32,960bytes, with1,520→1,920training steps. On the new final set it scored34/40(85%) versus the unchanged parent's22/40(55%); Korean and English each17/20. Completion reports improved1/5→5/5 and questions0/5→4/5. Accepted non-unclear proposals at0.9 increased3/40→17/40; those17were correct in this small synthetic set. Five cases per intent cannot establish production100%precision or reliable completion recognition.

The frozen old/new rule controls scored17/40 and20/40; the new rule accepted19cases with2errors. Rule scores are not calibrated probabilities. Larger capacity/context features were not automatically better. The result supports further development of varied-message warm-start training, without isolating data diversity from weight updates.

The original [plan](plan-v2.json), complete results and lock are preserved. The later driver adds manifest/rubric/arm checks but is distinguished from original training source; no weights are reselected. Pooled and language metrics repeat predictions using the same locked model. Reading the final file once is not a single-forward claim. The final set is now exposed and must not select the next revision.

Runtime is Go and performs no state write. The [read-only catalog bridge](../../pkg/statehintcatalog/README.md) accepts opaque references from the trusted application reader. Native IDs,names,credentials stay private; missing canonical versions are never replaced with fake ones. Live Riido integration and permission/event verification remain unconnected.
