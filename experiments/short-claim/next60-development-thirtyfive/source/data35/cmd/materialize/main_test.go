// Copyright 2026 teamswyg. SPDX-License-Identifier: Apache-2.0
package main

import (
	"errors"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/shortclaimdata"
)

func TestReaderErrorReturnAccountingWithoutReaderCall(t *testing.T) {
	v, err := bridge(shortclaimdata.Example{}, errors.New("synthetic_reader_error"))
	if err == nil || err.Error() != "reader_failed" || !v.ReaderReturned || v.ValidateAttempted || v.Valid || v.Count != 0 {
		t.Fatal("returned Reader error must retain one return and zero extra validations")
	}
}
