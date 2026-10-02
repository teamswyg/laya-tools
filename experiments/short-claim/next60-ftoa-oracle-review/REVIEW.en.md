# ftoa Want review

No static numeric error was found in six Wants. The reviewer authored none of the Wants, oracle, observer or candidates. This is nonblind source review, without build or execution evidence.

Exact1.125 scales to112.5 cents: keep even112, yielding1.12. Exact1.375 scales to137.5: round to even138, yielding1.38. Represented binary64 values for1.999 and1.25e-7 are safely inside the rounding intervals for2 and0.000000125. Negative zero canonicalizes to0; +Inf yields nonfinite. Predicted bits are not observed compiler outputs.

The oracle uses exact bits-to-rational integer arithmetic and quotient/remainder parity, rather than candidate-rendered strings. Input selection still shares the Go compiler/runtime. FormatFloat's f precision specifies places after the decimal point, not significant digits.

The original API returns only a string. Preserve unavailable error channel; do not invent nil error/success. Typed authored APIs are separate. Oracle checks6 and planned candidate observations18 remain one request.

Calls and qualified additions remain0; role/label/weight are null. Compiler selection, compatible outside controller, CI/resource and Root freezes remain pending. Future empty/partial output after a kill means unknown calls, not zero. No original/Want/seal changes or blanket licensing/training clearance.

See [exact review](ORACLE-REVIEW.v1.json) and [ledger](LEDGER.v1.json).
