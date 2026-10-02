# Standalone input reading: file reduction and memory reduction

JSON and compact were each read once in a fresh process. All24 column bit/count/null digests and metadata matched. JSON uses two passes to check retained sizes before array allocation; compact uses the earlier codec. Original fitting, features and labels were neither repeated nor changed.

Peak RSS was JSON18.23MiB versus compact18.41MiB: **no RAM saving was demonstrated**. Compact had smaller files and cumulative allocations but slightly higher final Go heap. One warm-file, fixed-order, whole-child sample does not establish a general speedup. It is also not directly comparable with the previous82.3MiB whole roundtrip.

For users, compact remains an optional file-cache experiment. Default loaders and input caps are unchanged. JSON minification and gzip already reduce storage/transfer size; evidence does not yet justify adopting a dedicated default format. This sample does not bound whole-process RSS at128,512 or2,400 requests.

[Full numbers and limitations](RESULTS-READER74.en.md) · [Actual controller record](actual/controller.results.json) · [Saved-record review](independent-runtime/REVIEW.v1.en.md) · [Public copy ledger](SAFE-COPY-LEDGER.v1.json).

Host paths from original plans are replaced in separate derived public views, which are not executable plans. Binaries, original input blobs, raw logs and private helpers remain outside Git. `.go.txt` source archives are references, not public runtime integration. Root added this README after the32-file stage and did not retroactively alter its ledger.

[한국어](README.ko.md)
