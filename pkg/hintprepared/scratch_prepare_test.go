// SPDX-License-Identifier: Apache-2.0
package hintprepared

import (
	"math"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

// Mixed large/short rows exercise copying out of reused borrowed storage.
// Existing tests cover model formats and ranking; these use no model or View.
func TestPrepareScratchRowsAreCopiedAndTight(t *testing.T) {
	in := apiInput(8)
	in.Request = "a b c d e f g h i j k l m n o p q r s t u v w x y z aa ab ac ad ae af"
	for i := range in.Candidates {
		in.Candidates[i].Text = "c"
		if i%2 == 0 {
			in.Candidates[i].Text = in.Request
		}
	}
	v := validated(t, in)
	p := prepared(t, v)
	check := func(owner *Prepared, snapshot shortclaim.Prepared) {
		if cap(owner.features) != len(owner.features) || int(owner.offsets[owner.count]) != len(owner.features) {
			t.Fatal("non-tight feature backing")
		}
		for i := 0; i < snapshot.Count; i++ {
			want := hintlearn.Features(snapshot.Request, snapshot.Candidates[i].Text)
			row := owner.features[owner.offsets[i]:owner.offsets[i+1]]
			if len(row) != len(want) {
				t.Fatal("borrowed row tail copied")
			}
			for j, f := range want {
				if row[j].Index != f.Index || math.Float64bits(row[j].Value) != math.Float64bits(f.Value) {
					t.Fatal("another row overwrote these feature bits")
				}
			}
		}
	}
	check(p, v.Prepared())
	next := copyAPIInput(in)
	for i := range next.Candidates {
		next.Candidates[i].Text = in.Candidates[7-i].Text
	}
	w := validated(t, next)
	q := prepared(t, w)
	check(q, w.Prepared())
	check(p, v.Prepared())
	if &p.features[0] == &q.features[0] {
		t.Fatal("constructors share retained feature backing")
	}
}

func TestPrepareScratchKeepsPublicErrorPrecedence(t *testing.T) {
	base := validated(t, apiInput(2)).Prepared()
	bad := [3]shortclaim.Prepared{base, base, base}
	bad[0].Count, bad[0].Schema = 0, "other"
	bad[1].Schema, bad[1].Request = "other", strings.Repeat("a", 513)
	bad[2].NormalizedRequest = "forged"
	bad[2].Candidates[1].ID = bad[2].Candidates[0].ID
	wants := [3]error{shortclaim.ErrCandidateCount, shortclaim.ErrSchema, shortclaim.ErrDuplicateID}
	for i, input := range bad {
		if owner, err := Prepare(input); owner != nil || err != wants[i] {
			t.Fatal("feature prepass bypassed public diagnostics", i, err)
		}
	}
}
