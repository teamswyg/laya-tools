package main

import (
	"bytes"
	"slices"
	"testing"
)

func TestRankingDigestFullOrderAndFraming(t *testing.T) {
	a := [3][]int{make([]int, 2400), {0, 1}, {1, 0}}
	for i := range a[0] {
		a[0][i] = i
	}
	digest := func(id string, orders [3][]int) []byte {
		d := newRankingDigest()
		d.add(id, orders)
		return d.h.Sum(nil)
	}
	x := digest("task", a)
	if !bytes.Equal(x, digest("task", a)) {
		t.Fatal("unstable digest")
	}
	b := a
	b[0] = slices.Clone(a[0])
	b[0][2398], b[0][2399] = b[0][2399], b[0][2398]
	if bytes.Equal(x, digest("task", b)) || bytes.Equal(x, digest("task2", a)) {
		t.Fatal("tail or identity missing")
	}
	b = a
	b[1], b[2] = b[2], b[1]
	if bytes.Equal(x, digest("task", b)) {
		t.Fatal("policy boundary missing")
	}
}
