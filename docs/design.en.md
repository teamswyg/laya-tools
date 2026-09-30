# How the router works

[한국어](design.ko.md) · English

Laya is not a code generator. It chooses among supplied options. Here it estimates whether a request needs a fast, standard, or strong coding model; the user's existing Codex performs the coding.

The flow is request → local Laya classification → selection policy → new Codex session. Policy does not simply trust the classifier: confidence below 0.9, truncated input, or unvalidated conditions such as Korean retain the strong model. An explicit user model override has priority.

A comment typo, an ordinary one-function bug, and a concurrency failure across files illustrate the three tiers. They are examples of criteria, not a promise that the classifier gets them right. It has not been trained/calibrated on dedicated difficulty labels for this project.

- `riidolaya route` only reports a recommendation and changes nothing in Codex.
- `riidolaya codex` starts a new Codex CLI task using the recommendation; it does not switch a live conversation.

Users configure actual model IDs, so changing model offerings does not require a new tool release. Leaving the strong model empty retains the existing Codex default. The tool does not separately read or store logins or API keys.

Code search is independent: keywords shortlist excerpts, Laya assesses relevance, and the tool returns paths and line ranges. This may reduce what a coding agent reads, but where keyword search already works, inference may add only latency. `--lexical` provides a comparison path.

People use text output; agents use JSONL or MCP. Persistent processes reuse the loaded model. No network server or startup service is installed automatically, and integration is opt-in.

Actual cost savings are unproven. A weak model that fails and needs a strong-model retry may cost more. Future evaluation should fix the coding tasks and compare success, total usage, duration, and retries. Current behavior is local retrieval/inference measurement and conservative recommendations, not an automatic completion-check/retry loop.

CI reviews development changes through tests, static checks, native inference, and secret scanning. Human review approval is not required. This concerns repository development, not relaxing the user's Codex execution permissions.
