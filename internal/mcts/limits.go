package mcts

import (
	"encoding/json"
	"math"
	"strings"
)

type Limits struct {
	Depth    int
	Nodes    uint32
	Movetime int
	Infinite bool
	NThreads int
	ByteSize int64
}

func (l Limits) String() string {
	builder := strings.Builder{}
	_ = json.NewEncoder(&builder).Encode(l)
	return builder.String()
}

const (
	DefaultDepthLimit    int    = math.MaxInt
	DefaultNodeLimit     uint32 = math.MaxInt32
	DefaultMovetimeLimit int    = -1
	DefaultByteSizeLimit int64  = -1
)

func DefaultLimits() *Limits {
	return &Limits{DefaultDepthLimit, DefaultNodeLimit, DefaultMovetimeLimit, true, 1, DefaultByteSizeLimit}
}

// Set the maximum depth of the search
func (l *Limits) SetDepth(depth int) *Limits {
	l.Depth = depth
	l.Infinite = false
	return l
}

// Set the maxiumum number of nodes engine can go through
func (l *Limits) SetNodes(nodes uint32) *Limits {
	l.Nodes = nodes
	l.Infinite = false
	return l
}

// Set the maximum time for engine to think
func (l *Limits) SetMovetime(movetime int) *Limits {
	l.Movetime = movetime
	l.Infinite = false
	return l
}

func (l *Limits) SetInfinite(infinite bool) {
	l.Infinite = infinite
}

func (l *Limits) SetThreads(threads int) *Limits {
	l.NThreads = max(threads, 1)
	return l
}

func (l *Limits) SetMbSize(mbsize int) *Limits {
	return l.SetByteSize(int64(mbsize) * (1 << 20))
}

func (l *Limits) SetByteSize(bytesize int64) *Limits {
	l.ByteSize = bytesize
	l.Infinite = false
	return l
}

func (l *Limits) InfiniteSize() bool {
	return l.ByteSize == -1
}
