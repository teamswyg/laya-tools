// Package claimtree implements a bounded research regression tree. It emits a
// fallible score, not a calibrated probability or an action authorization.
package claimtree

import (
	"cmp"
	"fmt"
	"math"
	"slices"
)

const Dimension = 16
const MaxNodes = 31

type Vector [Dimension]float64

// Model uses parallel fixed arrays. Once fitted/validated, it is immutable;
// each reader owns its input, so scoring needs no locks or allocations.
type Model struct {
	Nodes       int
	Feature     [MaxNodes]int8 // -1 denotes a leaf
	Left, Right [MaxNodes]uint8
	Cut, Value  [MaxNodes]float64
}

// Score requires a model that passed Validate and finite input features.
func (m *Model) Score(x Vector) float64 {
	i := 0
	for m.Feature[i] >= 0 {
		if x[m.Feature[i]] <= m.Cut[i] {
			i = int(m.Left[i])
		} else {
			i = int(m.Right[i])
		}
	}
	return m.Value[i]
}

func (m *Model) Validate() error {
	if m.Nodes < 1 || m.Nodes > MaxNodes {
		return fmt.Errorf("invalid node count")
	}
	var seen [MaxNodes]bool
	var walk func(int, int) error
	walk = func(i, depth int) error {
		if i < 0 || i >= m.Nodes || seen[i] || depth > 4 {
			return fmt.Errorf("invalid tree topology")
		}
		seen[i] = true
		if !finite(m.Value[i]) || math.Abs(m.Value[i]) > 150 || !finite(m.Cut[i]) {
			return fmt.Errorf("invalid tree value")
		}
		f := m.Feature[i]
		if f == -1 {
			if m.Left[i] != 0 || m.Right[i] != 0 {
				return fmt.Errorf("leaf has children")
			}
			return nil
		}
		if f < 0 || f >= Dimension || int(m.Left[i]) <= i || int(m.Right[i]) <= i {
			return fmt.Errorf("invalid branch")
		}
		if err := walk(int(m.Left[i]), depth+1); err != nil {
			return err
		}
		return walk(int(m.Right[i]), depth+1)
	}
	if err := walk(0, 0); err != nil {
		return err
	}
	for i := 0; i < m.Nodes; i++ {
		if !seen[i] {
			return fmt.Errorf("unreachable node")
		}
	}
	return nil
}
func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }

// Fit reads training rows only. Integer page targets are bounded to avoid
// overflow. Split ties are deterministic: feature order, then cut order.
func Fit(x []Vector, y []float64, depth, minLeaf int) (Model, error) {
	var m Model
	if len(x) == 0 || len(x) != len(y) || len(x) > 4096 || depth < 1 || depth > 4 || minLeaf < 1 || minLeaf > 4096 {
		return m, fmt.Errorf("invalid training shape/config")
	}
	for i, row := range x {
		if !finite(y[i]) || math.Abs(y[i]) > 150 {
			return m, fmt.Errorf("invalid target")
		}
		for _, v := range row {
			if !finite(v) || math.Abs(v) > 1e6 {
				return m, fmt.Errorf("invalid feature")
			}
		}
	}
	ids := make([]int, len(x))
	for i := range ids {
		ids[i] = i
	}
	var build func([]int, int) int
	build = func(rows []int, level int) int {
		node := m.Nodes
		m.Nodes++
		m.Feature[node] = -1
		sum := 0.0
		for _, i := range rows {
			sum += y[i]
		}
		m.Value[node] = sum / float64(len(rows))
		if level == depth || len(rows) < 2*minLeaf {
			return node
		}
		bestGain := 0.0
		bestFeature := -1
		bestCut := 0.0
		ordered := slices.Clone(rows)
		for f := 0; f < Dimension; f++ {
			slices.SortFunc(ordered, func(a, b int) int {
				if c := cmp.Compare(x[a][f], x[b][f]); c != 0 {
					return c
				}
				return cmp.Compare(a, b)
			})
			leftSum := 0.0
			for j := 1; j < len(ordered); j++ {
				leftSum += y[ordered[j-1]]
				if j < minLeaf || len(ordered)-j < minLeaf || x[ordered[j-1]][f] == x[ordered[j]][f] {
					continue
				}
				rightSum := sum - leftSum
				gain := leftSum*leftSum/float64(j) + rightSum*rightSum/float64(len(ordered)-j) - sum*sum/float64(len(ordered))
				if gain > bestGain {
					bestGain = gain
					bestFeature = f
					bestCut = x[ordered[j-1]][f]
				}
			}
		}
		if bestFeature < 0 {
			return node
		}
		left := make([]int, 0, len(rows))
		right := make([]int, 0, len(rows))
		for _, i := range rows {
			if x[i][bestFeature] <= bestCut {
				left = append(left, i)
			} else {
				right = append(right, i)
			}
		}
		m.Feature[node] = int8(bestFeature)
		m.Cut[node] = bestCut
		m.Left[node] = uint8(build(left, level+1))
		m.Right[node] = uint8(build(right, level+1))
		return node
	}
	build(ids, 0)
	return m, m.Validate()
}
