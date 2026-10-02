# Fit79 guard v2 preparation

This separate preparation fixes two omissions in the v1 storage checks. The original sources, HANDOFF and binary remain unchanged. This is not a training result or model approval.

First, a resource policy amendment must have frozen: true and the exact state frozen_resource_policy_for_one_data_effect79_stage. A pending proposal with only an audit SHA filled in is rejected. Its current_registry_bytes and SHA must match the sole retained-serialized-audit-registry evidence pin. The exact pinned audit JSON is read and checked for its schema, snapshot completeness, per-path counting basis, entry count and byte sum. Every audited ID / bytes / SHA must appear in the current consumption Assets inventory. Additional new entries are allowed.

Second, every registered_model_artifacts path / bytes / SHA must also appear in the budgeted Assets inventory. A present, correctly hashed model cannot be omitted from storage accounting.

The pinned audit describes 2,776 paths and 206,707,066 B, including one empty regular file. Empty files still count as inventory entries; an asset-only path verifies their zero size and empty-content SHA. The positive-size rule for scientific input readPin is unchanged. Audit completeness is limited to its fixed selector and snapshot, not the entire machine or deleted assets.

The main / fit / utility / seed / role code is unchanged. The 16 public source pins match commit 01ce47dcca16c8f82022c97354d82105b6819875 / CI 37015530233. The new private CLI was compiled separately on the local machine; that public CI did not test this private CLI.

Synthetic race checks passed 23 top-level tests and 45 subtests with zero failures or skips. Vet, formatting, the CGO 0 / trimpath build and GoList-selected source verification also passed. Tests use synthetic DTOs, files and callbacks. Actual corpus Fit, Score, Features, Project, RoleAssign, model Encode / Decode and driver main / help executions remain zero.

Root's actual frozen amendment, the real 79-parent projection, runtime QA, the full consumption inventory and independent review remain pending. The historical 64 MiB excess stays a failure. New-stage data / models / numeric records remain capped at 64 MiB; sparse payload stays 64 MiB, CPU threads 1, observed RSS 256 MiB and timeout 300 seconds. The default retained limit remains 64 MiB; the optional 512 MiB branch requires Root's exact frozen policy pin.

Public-safe evidence archives exact Go source as .go.txt. Binaries, model bodies, the host-path-bearing go.mod and raw build logs are not published; only their sizes and SHA values are retained.
