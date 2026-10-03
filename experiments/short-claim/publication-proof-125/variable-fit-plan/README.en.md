# Preparing a Go trainer for variable request counts

The Fit79 source was found. All11 driver files and16 current core dependencies match historical pins. Reuse the trainer and replace fixed79-parent/235-candidate guards with a frozen variable-input contract. This is design only: no implementation migration, training or new weights.

Keep seed1729,8192dimensions,FP32 scoring,50epochs,batch128,LR0.1,L2 0.0001,pair0 and the earliest strict minimum validation NLL. Existing fp32 rounds scoring coefficients to float32, while shadow weights, gradients and losses remain float64. Rewriting all arithmetic would be a different experiment.

Read33 and later approximately60 development requests one bounded row at a time; join only Root-selected known candidates. Exclude U-only candidates with null labels/weights in the evidence ledger. Root-adopted knownF+U negatives may retain their labels. Never infer no-answer from all false. 79+33=112 is arithmetic, not independent requests.

Preserve family roles, old validation/calibration columns, masks and zero-weight rows. Groups82/83 exceed33: use sorted group-role slices, fixed[8] candidate slots and[3] role aggregates, not group IDs as parent-array indexes. Only request/candidate text becomes features; supervision/IDs/roles/counts stay metadata.

Before Fit require actual materialization/Go reader evidence, semantic dedup and family/role closure against79, separately frozen approximately60 coverage/roles, exact joined membership and current resource evidence. Counts alone never trigger Fit. Exposed validation stays development regression/selection, not new held-out evidence. Protected2400 is unread and is not added as a prerequisite for every development Fit.

Reuse oneFitCheckpoint, complete50epoch trace and four baselines/5% necessary utility/nondegradedTop1/Top3. Replace fixed79 counts and76..78 train-origin range with Root membership. New seeds/features/losses/ternary remain separate. Fit3/8/model3/16 and inactive failed models are unchanged.

Keep CSR/contiguous arrays and one worker. Training time/OS RSS/Go heap/sparse bytes differ from later normalization/features/scoring/ranking measurements. The32,792-byte model format estimates disk storage, not RAM. Repeated corpus Fits for profiling still count; fake callbacks and actual synthetic test-fitting need separate records.

See[MIGRATION](MIGRATION.v1.json),[measurement contract](PROFILE.v1.json) and[source pins](SOURCE-PINS.v1.json). Agent114's prior after30 plan is separately pinned and preserved. This nonblind author wrote related comparers/materializer preparations but did not author Fit79.
