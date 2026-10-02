# Joining role assignment and loss exclusions69

[한국어](README.ko.md) · [Detailed findings](MASK-ROLE-AUDIT.en.md) · [Go projection handoff](GO-BRIDGE-HANDOFF.en.md)

Actual roles66 and caption supervision67 were joined to the original72 requests and216 candidates. A mask controls whether a caption contributes to training loss. Masked candidates remain in the original candidate list and function truth; unknown truth remains null.

| Role | Eligible positive | Eligible negative | Masked known | Unknown | Eligible groups |
|---|---:|---:|---:|---:|---:|
| train |20|52|18|42|9|
| validation |10|25|1|12|3|
| calibration |5|18|4|9|3|

Train group52 has all known candidates masked, reducing eligible train groups from10 to9. Roles and masks are unchanged; the existing9/3/3 preparation minima are retained. Positive-bearing groups7/3/2 are descriptive and introduce no new minimum. Calibration remains separate for a subsequent abstention-policy review.

One read-only metadata aggregation passed792 binding checks without rerunning original roles, behavior, features or fitting. Preserve the [actual result](results-69.json), [reservation](INVOCATION-69.json), [ledger](ATTEMPT-LEDGER.json) and [copy ledger](QA-COPY-LEDGER-92.v1.json). The four new upstream requests have not been appended to these72 requests.
