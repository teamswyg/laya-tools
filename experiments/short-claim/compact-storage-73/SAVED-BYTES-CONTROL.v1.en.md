# Saved-byte baseline comparison

The same fixed pretty source 8390462B → whitespace-only minified JSON 4697570B → existing compact 1978694B. Formatting-only reduction 44.0130%; binary storage reduction against minified 57.8783%. Original pretty-relative 76.4173% combines both effects. Existing roundtrip results were neither modified nor repeated.

One DefaultCompression gzip per form: pretty 594236B, minified 400256B, compact 328206B. Exact SHA/source pins/predeclared plan/separate counts are in SAVED-BYTES-CONTROL.v1.json. No gzip decoding, speed or RSS measurement. Compressed cache/transfer bytes do not automatically resolve decoded JSON64MiB caps/owned payload/RSS.

Helper1 successful/failures0; json.Compact1/gzip3/retries0. Existing roundtrip repeat, compact/gzip decoders, original APIs/models/fitting/roles/new labels/final reads/shared edits/HF remain0. Raw outputs/host paths/helper source stay private. Training/performance approval/final2,400 success remain false.
