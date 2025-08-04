package mcts

type ListenerTreeStats[T MoveLike] struct {
	BestMove T
	Eval     float64
	Maxdepth int
	Cycles   int
	TimeMs   int
	Nps      uint64
	Pv       []T
	Terminal bool
	Draw     bool
}

// Convert TreeStats to 'ListenerTreeStats' struct
func toListenerStats[T MoveLike](tree *MCTS[T]) ListenerTreeStats[T] {
	pv, terminal, draw := tree.Pv(BestChildMostVisits)
	return ListenerTreeStats[T]{
		BestMove: tree.RootSignature(),
		Eval:     float64(tree.RootScore()),
		Maxdepth: int(tree.MaxDepth()),
		Cycles:   int(tree.Root.Visits()),
		TimeMs:   int(tree.Limiter.Elapsed()),
		Nps:      uint64(tree.Nps()),
		Pv:       pv,
		Terminal: terminal,
		Draw:     draw,
	}
}

// Listener function callback, will recieve current tree statistics, like
// max depth of tree, number of iterations so far
type ListenerFunc[T MoveLike] func(ListenerTreeStats[T])

type StatsListener[T MoveLike] struct {
	// called when 'max depth' increases, receives new max depth
	onDepth ListenerFunc[T]

	// called every one full iteration, receives total number of cycles
	onCycle ListenerFunc[T]

	// called when the search stops (either by limiter or 'stop' signal)
	onStop ListenerFunc[T]
}

// Attach new on max depth change callback, will be called only be the main search thread
func (listener *StatsListener[T]) OnDepth(onDepth ListenerFunc[T]) *StatsListener[T] {
	listener.onDepth = onDepth
	return listener
}

// Attach new on iteration increase callback
func (listener *StatsListener[T]) OnCycle(onCycle ListenerFunc[T]) *StatsListener[T] {
	listener.onCycle = onCycle
	return listener
}

// Attach 'on stop signal' callback, executed when the search has ended,
func (listener *StatsListener[T]) OnStop(onStop ListenerFunc[T]) *StatsListener[T] {
	listener.onStop = onStop
	return listener
}

// Invoke the listener's callback
func listenerInvoke[T MoveLike](f ListenerFunc[T], tree *MCTS[T]) {
	f(toListenerStats(tree))
}
