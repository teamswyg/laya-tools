package main

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
)

// rankingDigest fingerprints complete ordered results without publishing paths
// or per-task records. Length framing prevents ambiguous concatenation.
type rankingDigest struct {
	h   hash.Hash
	buf [8192]byte
}

func newRankingDigest() *rankingDigest { return &rankingDigest{h: sha256.New()} }
func (d *rankingDigest) number(n int) {
	binary.LittleEndian.PutUint64(d.buf[:8], uint64(n))
	d.h.Write(d.buf[:8])
}
func (d *rankingDigest) add(id string, orders [3][]int) {
	d.number(len(id))
	d.h.Write([]byte(id))
	for policy, order := range orders {
		d.number(policy)
		d.number(len(order))
		for len(order) > 0 {
			n := min(len(order), len(d.buf)/8)
			for i, v := range order[:n] {
				binary.LittleEndian.PutUint64(d.buf[i*8:], uint64(v))
			}
			d.h.Write(d.buf[:n*8])
			order = order[n:]
		}
	}
}
