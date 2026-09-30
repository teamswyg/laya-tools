package hintsearch

import "fmt"

// InterleaveAfterPrefix retains an already emitted baseline prefix, then uses
// baseline-first interleave with those IDs removed. It only computes ordering:
// callers must defer constructing hints until an actual continuation arrives.
// Inputs are not mutated. This function makes no latency or token-cost promise.
func InterleaveAfterPrefix(baseline, hints []int, prefix int) ([]int, error) {
	if prefix < 0 || prefix > len(baseline) {
		return nil, fmt.Errorf("invalid emitted prefix")
	}
	order, err := InterleaveBaselineFirst(baseline, hints)
	if err != nil {
		return nil, err
	}
	var emitted [MaxDocuments]bool
	for _, id := range baseline[:prefix] {
		emitted[id] = true
	}
	n := 0
	for _, id := range order {
		if !emitted[id] {
			order[n] = id
			n++
		}
	}
	copy(order[prefix:], order[:n])
	copy(order[:prefix], baseline[:prefix])
	return order, nil
}
