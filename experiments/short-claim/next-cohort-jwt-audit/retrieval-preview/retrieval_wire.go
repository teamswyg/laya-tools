// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import "encoding/json"

const retrievalStdoutLimit = 192 << 10

type retrievalWireCode string

const (
	retrievalWireOK             retrievalWireCode = ""
	retrievalWireMarshalUnknown retrievalWireCode = "unknown_output_marshal"
	retrievalWireBoundUnknown   retrievalWireCode = "unknown_output_bound"
)

// Prospective final runtime guard, not a replacement for a complete frozen
// worst-shape qualification before the first JWT API call. Caller supplies only
// its owned frozen output structs, not input-controlled custom MarshalJSON.
// No partial record is emitted here. Nonempty code is a failed experiment;
// caller must retain first status/empty stdout/fixed stderr and never discard
// rows to turn it into success or retry the original worker.
func marshalRetrievalBounded(out any) ([]byte, retrievalWireCode) {
	raw, err := json.Marshal(out)
	if err != nil {
		return nil, retrievalWireMarshalUnknown
	}
	if len(raw)+1 > retrievalStdoutLimit {
		return nil, retrievalWireBoundUnknown
	}
	return append(raw, '\n'), retrievalWireOK
}
