package uttt

import (
	"math"
)

type Limits struct {
	depth    int
	nodes    uint64
	movetime int
	infinite bool
}

func DefaultLimits() *Limits {
	return &Limits{math.MaxInt, math.MaxInt, math.MaxInt32, true}
}

func (l *Limits) SetAll(depth int, nodes uint64, movetime int, infinite bool) {
	l.depth, l.nodes, l.movetime, l.infinite = depth, nodes, movetime, infinite
}

// Set the maximum depth of the search
func (l *Limits) SetDepth(depth int) *Limits {
	l.depth = depth
	l.infinite = false
	return l
}

// Set the maxiumum number of nodes engine can go through
func (l *Limits) SetNodes(nodes uint64) *Limits {
	l.nodes = nodes
	l.infinite = false
	return l
}

// Set the maximum time for engine to think
func (l *Limits) SetMovetime(movetime int) *Limits {
	l.movetime = movetime
	l.infinite = false
	return l
}

func (l *Limits) SetInfinite(infinite bool) {
	l.infinite = infinite
}
