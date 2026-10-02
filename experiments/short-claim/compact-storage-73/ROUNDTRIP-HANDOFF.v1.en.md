# Storage roundtrip observation

The fixed 8390462-byte JSON became a private 1978694-byte compact file. Its 163758 metadata bytes and all four CSR FP64 bits/integers/null-empty states roundtripped exactly. Serialized reduction 76.4173%; existing uint16 width and diagnostic duplication are unchanged.

ROUNDTRIP-HANDOFF.v1.json contains exact aggregates/pins. Owned arrays+metadata payload 1978406B and original published payload1,928,154B use different accounting recipes. Go heap/whole RSS/file size are distinct; one sequential warm timing sample does not establish causal speed. Original JSON whitespace reconstruction is not claimed.

Pure codec race test1/vet1/build1 passed; stored roundtrip1/retry0. Original Projection/Features/Prepare/Fit/role/label/behavior/model API calls, HF and shared edits remain0. Raw JSON/compact/binary/host paths/logs stay private; new training/performance approval/final2,400 success remain false. Smaller storage may ease I/O/text parsing during128/512 expansion; no expansion experiment or loader integration occurred.
