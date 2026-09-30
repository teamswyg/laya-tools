# Project purpose and licensing

[한국어](https://github.com/teamswyg/laya-tools/wiki/Licensing-and-Project-KO) · [Home](https://github.com/teamswyg/laya-tools/wiki)

## Why use this tool?

The project explores whether teams can own model selection as a quality/budget policy. Using a small model is not automatically economical: failed work, retries, and latency count. We want measurable decisions, not an assumption that either a provider or a local classifier always makes the right choice.

The repository/module is `laya-tools`; the command is `riidolaya`. Laya classifies choices, while Codex performs coding when explicitly launched. Repository selection remains a preview. Riido-daemon integration, actual savings, and live-session model switching are not established features.

## Can I use or redistribute it?

Original project code uses Apache-2.0. Identified third-party components use Apache-2.0, MIT, or BSD-3-Clause terms. Modification, commercial use, and redistribution are possible under those conditions, including notice obligations. A Go port does not remove upstream copyrights.

When redistributing, preserve the bundled LICENSE, NOTICE, and licenses directory. Model bundles include their own license, attribution, provenance, and modification notices; keep those too. Do not present this tool as officially endorsed by Laya, laya.tools, or an upstream author.

Earlier model/CLI bundles had notice omissions. Corrected models-v2 and current binaries include the missing notices; older releases also have a supplement. The [full audit](https://github.com/teamswyg/laya-tools/blob/main/docs/license-audit.en.md) explains exactly what changed.

Public licensing does not independently prove rights to every training datum or patent. The audit is a technical distribution review, not a blanket legal warranty. Separate contractual warranties or private-data fine-tuning require additional review.

## Credits

Thanks to [Laya and its SDK](https://github.com/NandhaKishorM/laya), the independent [laya.tools](https://laya.tools/) directory, [system-one-router](https://github.com/mmornati/system-one-router), and [pi-pignon](https://github.com/siiick/pi-pignon). We adapted candidate policy and cache-switch logic with pinned sources and notices. Other projects' speed/accuracy claims are not our results.

[Adaptation details](https://github.com/teamswyg/laya-tools/blob/main/docs/ecosystem.en.md) · [NOTICE](https://github.com/teamswyg/laya-tools/blob/main/NOTICE)

## How changes are reviewed

PRs must pass CI: formatting, tests/race checks, vet, native inference, secrets, and license inventory. Trusted same-repository PRs can merge without human reviewer approval. This is a development merge policy, not permission to bypass Codex's execution controls.

Wiki originals live in [docs/wiki](https://github.com/teamswyg/laya-tools/tree/main/docs/wiki). Changes should update both languages, pass repository CI, and then be published to the Wiki. Direct Wiki edits are not automatically protected by the repository's main-branch rules. No extra credential or automatic publishing service was installed for this documentation.
