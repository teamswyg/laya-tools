# Preserve primary abort cause and signal outcome separately

This source overlays the temporary CI controller while preserving sealed native-wire3 source and historical receipts. The first macOS failure lost stderr overflow behind a subsequent group-signal failure. Its old receipt has no errno, so the exact OS cause remains unknown.

The first **accepted abort request** preserves receiver/stderr/deadline cause. The receipt separately records one group-signal attempt, accepted/no-matching-group/failed outcome, and numeric errno when known. An accepted signal does not prove the whole group was killed. EPERM, EINVAL and unknown errors remain failures.

Results are read after receiver, stderr and watcher joins. Child Start and Wait attempts remain one each. Any abort remains failure; stderr retention stays32KiB. Competing requests cannot rewrite the cause or retry signaling.

New controls cover result classes, wrapped errno, concurrent losers, joining the unfinished winner, panic(nil), missing causes and self-exiting owned children with stderr/receiver/deadline faults. If a real OS signal fails and the child continues running, Wait may remain delayed. These fake controls do not prove termination despite arbitrary permission failures. This change adds no retry or alternate force-termination mechanism.

Reproduce with Go1.27.1: `bash scripts/verify-next60-native-wire3.sh`. It starts no original library, Laya or training. [한국어](README.ko.md).
