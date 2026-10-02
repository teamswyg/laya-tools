# Conversion80 saved-result comparison

All 72 stored candidate observations match the 72 Wants sealed before execution. One Go native run observed the existing 3 requests and 9 code candidates against 24 original inputs. This was not Laya, GPU inference, new model training, or 72 independent requests.

The checker did not rerun original code or the observer. It compared every full literal, all 11 return channels, nil versus empty inputs, candidate/input order, Scan wrapper argument receiver, and unobserved panic text. Final, partial and stdout are byte-exact at SHA `ee6b993dd5495e901ea9b516f96278eee76821bdb1e41ffaaac0ba0f2433ab8c`.

Of 72 candidate dispatches, 69 returned normally and 3 expected ParsePanicOnError panics were recorded. Explicit Error observation on returned errors added 10 calls: 2 URN and 8 standard errorString calls. Thus 82 callbacks were tracked. The 56 inner original API mappings and 4 startup sites remain static evidence; their dynamically measured counts remain null. Error-valued panics incurred no additional Error/String/Format call.

The checker compared 37 closed file pins, including 20 originals, with the result. Only the frozen false-to-true token differs between draft and frozen plan. The initial reservation retained static56, zero execution counters and null inner/init counts. Root started once, retried zero times, exited zero and joined Wait, with no timeout or overflow.

The stored whole-child OS RSS was 19,529,728 B (18.625 MiB), and footprint was 16,728,544 B. OS real/user/system were 2.84/0.34/0.18 seconds; outside-controller wall was 2.8445567499999997 seconds. These include startup, loader, sync writes and candidates, rather than pure function cost. The 256 MiB Go heap setting is soft and does not establish a hard total-RSS cap.

One stdlib metadata checker attempt passed; zero failed. The receipt's 6,571 completed checks count decoding, shape and pin comparisons, rather than new semantic answers. The terminal count of 6,575 also includes final output fsync checks. This reviewer authored the observer/controller and saw the outcome beforehand. The review is nonblind saved-file mechanics QA, not an independent source oracle, runtime actor, blind semantic review or training approval. The original Wanted truth/role/weight/mask/label/Got nulls and preparation flags remain unchanged. Root's later SAT/eligibility judgments are separate records.

[Comparison](NUMERIC.v1.json) · [Receipt](RECEIPT.v1.json) · [Attempt ledger](LEDGER.v1.json)
