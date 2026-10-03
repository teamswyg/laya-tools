// Copyright 2026 teamswyg. SPDX-License-Identifier: Apache-2.0
package data35

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestReservedEmptyFilesSyncInOrder(t *testing.T) {
	for _, failAt := range []int{0, 1, 2} {
		t.Run([]string{"success", "data_sync_failure", "receipt_sync_failure"}[failAt], func(t *testing.T) {
			base, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(base, "attempt")
			var names []string
			r, err := reserveWithFileSync(out, Config{}, func(f *os.File) error {
				st, err := f.Stat()
				if err != nil || st.Size() != 0 {
					t.Fatal("reservation was not empty")
				}
				names = append(names, filepath.Base(f.Name()))
				if len(names) == failAt {
					return errors.New("synthetic_sync_failure")
				}
				return f.Sync()
			})
			if r != nil {
				defer r.Close()
			}
			wantCalls := 2
			wantError := ""
			if failAt == 1 {
				wantCalls, wantError = 1, "data_reservation_sync"
			}
			if failAt == 2 {
				wantError = "receipt_reservation_sync"
			}
			if len(names) != wantCalls || names[0] != "train.jsonl" || (len(names) == 2 && names[1] != "MATERIALIZATION.v1.json") {
				t.Fatal("empty file sync order")
			}
			if (err == nil) != (failAt == 0) || (err != nil && FailureCode(err) != wantError) {
				t.Fatal("sync failure was not propagated")
			}
			for _, name := range []string{"train.jsonl", "MATERIALIZATION.v1.json"} {
				b, err := os.ReadFile(filepath.Join(out, name))
				if err != nil || len(b) != 0 {
					t.Fatal("partial reservation was removed or changed")
				}
			}
			// No adoption, input loading or Reader is invoked in this control.
			if _, err := Reserve(out, Config{}); err == nil {
				t.Fatal("reserved attempt was reused")
			}
		})
	}
}
