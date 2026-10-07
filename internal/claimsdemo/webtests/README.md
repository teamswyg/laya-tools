# Client interaction regression fixtures

Run from the repository root with Node.js 22 or newer:

```sh
node --test internal/claimsdemo/webtests/*.test.cjs
```

The suite executes the shipped `web/app.js` unchanged with Node's built-in test
runner and VM. It uses a small DOM fixture whose required IDs, head names,
probability rows and example buttons are checked against the shipped HTML.
Deterministic time and deferred HTTP/JSON responses cover input debounce, aborted
and stale requests, Korean IME composition, UTF-8 byte boundaries, late server
limits, clear/whitespace reset, invalid responses, and raw candidates versus final
abstention. All input strings and model responses are authored public fixtures.

Node is a maintainer/CI test dependency only. The demo application and inference
server remain Go; this suite needs no installed packages, browser download, model
artifact or training. It does not test browser layout, browser-specific event
ordering, real HTTP transport or research-model semantic quality. Go HTTP tests
and a real-browser demo serve those separate purposes.
