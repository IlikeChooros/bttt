package bttt

import "math"

type Limits struct {
	depth    int
	nodes    uint64
	movetime int
}

func DefaultLimits() *Limits {
	return &Limits{4, math.MaxInt, math.MaxInt}
}

func (l *Limits) SetAll(depth int, nodes uint64, movetime int) {
	l.depth, l.nodes, l.movetime = depth, nodes, movetime
}

// Set the maximum depth of the search
func (l *Limits) SetDepth(depth int) *Limits {
	l.depth = depth
	return l
}

// Set the maxiumum number of nodes engine can go through
func (l *Limits) SetNodes(nodes uint64) *Limits {
	l.nodes = nodes
	return l
}

// Set the maximum time for engine to think
func (l *Limits) SetMovetime(movetime int) *Limits {
	l.movetime = movetime
	return l
}
