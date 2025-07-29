package mcts

import (
	"math"
	"sync/atomic"
	"unsafe"
)

type LimiterLike interface {
	// Set the limits
	SetLimits(*Limits)
	// Get the limits
	Limits() *Limits
	// Get elapsed time (from the last 'Reset' call)
	Elapsed() uint32
	// Set the stop signal, will cause to exit search if set to true
	SetStop(bool)
	// Get the stop signal
	Stop() bool
	// Reset the limiter's flags, called on search setup
	Reset()
	// Wheter the tree can grow
	Expand() bool
	// Wheter the search should stop, called in the main search loop
	Ok(nodes, size, depth, cycles uint32) bool
}

type Limiter struct {
	limits     *Limits
	Timer      *_Timer
	nodeSize   uint32
	maxSize    uint32
	expand     atomic.Bool
	stop       atomic.Bool
	areSetMask int
}

func NewLimiter(nodesize uint32) *Limiter {
	limiter := &Limiter{
		limits:   DefaultLimits(),
		Timer:    _NewTimer(),
		nodeSize: nodesize,
	}

	limiter.expand.Store(true)
	return limiter
}

func (l *Limiter) SetStop(v bool) {
	l.stop.Store(v)
}

func (l *Limiter) Stop() bool {
	return l.stop.Load()
}

func (l *Limiter) SetLimits(limits *Limits) {
	l.limits = limits
}

func (l *Limiter) Limits() *Limits {
	return l.limits
}

func (l *Limiter) Elapsed() uint32 {
	return uint32(l.Timer.Deltatime())
}

func (l *Limiter) Reset() {
	l.Timer.Movetime(l.limits.Movetime)
	l.Timer.Reset()
	l.stop.Store(false)
	l.expand.Store(true)

	// Calculate 'nodes' based on memory
	if l.limits.ByteSize != DefaultByteSizeLimit {
		l.maxSize = uint32(l.limits.ByteSize) / l.nodeSize
	} else {
		l.maxSize = math.MaxUint32
	}

	// Pre-calculate 'are set' limit mask, see 'Ok' method for more explanation
	l.areSetMask = toMask(l.Timer.IsSet(), 1) |
		toMask(l.limits.Nodes != DefaultNodeLimit, 2) |
		toMask(l.limits.ByteSize != DefaultByteSizeLimit, 3) |
		toMask(l.limits.Depth != DefaultDepthLimit, 4) |
		toMask(l.limits.Cycles != DefaultCyclesLimit, 5)
}

func (l *Limiter) Expand() bool {
	return l.expand.Load()
}

func toMask(val bool, offset int) int {
	return int(*(*byte)(unsafe.Pointer(&val))) << offset
}

func (l *Limiter) Ok(nodes, size, depth, cycles uint32) bool {
	// return !l.stop.Load() && nodes < l.limits.Nodes && !l.Timer.IsEnd() && size < l.maxSize
	stop := l.stop.Load()
	if l.limits.Infinite {
		return !stop
	}

	limitMask := 0
	const (
		StopMask   = 1
		TimeMask   = 2
		NodesMask  = 4
		MemoryMask = 8
		DepthMask  = 16
		CyclesMask = 32
	)

	limitMask |= toMask(stop, 0)
	limitMask |= toMask(l.Timer.IsEnd(), 1)
	limitMask |= toMask(l.limits.Nodes <= nodes, 2)
	limitMask |= toMask(l.maxSize <= size, 3)
	limitMask |= toMask(l.limits.Depth <= int(depth), 4)
	limitMask |= toMask(l.limits.Cycles <= cycles, 5)

	// Hierachy of stop signals
	// 1. stop
	// 2. Movetime
	// 3. Memory
	// 4. Depth

	// Check the combos:
	// (Time/Node/Cycles or any combination of them) AND memory limit ->
	// if memory is exhausted, disable expanding of the tree and wait for the other limitation/s
	if (l.areSetMask&MemoryMask) == MemoryMask && (l.areSetMask&(TimeMask|NodesMask|CyclesMask)) != 0 {
		// Memory exhausted
		if limitMask&MemoryMask == MemoryMask {
			l.expand.Store(false)
			limitMask ^= MemoryMask // remove memory limitation
		}
	}

	// Time + Nodes, Time + Depth, Time + Nodes + Depth is natural
	return limitMask == 0
}
