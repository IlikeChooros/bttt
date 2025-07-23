package uttt

import (
	"math"
)

type Limits struct {
	depth    int
	nodes    uint64
	movetime int
	infinite bool
	nThreads int
	byteSize int64
}

func DefaultLimits() *Limits {
	return &Limits{math.MaxInt, math.MaxInt, -1, true, 1, -1}
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

func (l *Limits) SetThreads(threads int) *Limits {
	l.nThreads = max(threads, 1)
	return l
}

func (l *Limits) SetMbSize(mbsize int64) *Limits {
	return l.SetByteSize(mbsize * (1 << 20))
}

func (l *Limits) SetByteSize(bytesize int64) *Limits {
	l.byteSize = bytesize
	return l
}

func (l *Limits) InfiniteSize() bool {
	return l.byteSize == -1
}
