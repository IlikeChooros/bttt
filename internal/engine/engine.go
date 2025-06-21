package bttt

import "sync/atomic"

/*
Main engine class, allowing user to make moves on the board,
search best move, based on given parameteres
*/
type Engine struct {
	position *Position
	limits   *Limits
	timer    *_Timer
	result   SearchResult
	stop     atomic.Bool
}

// Get new engine instance
func NewEngine() *Engine {
	e := new(Engine)
	e.position = NewPosition()
	e.limits = DefaultLimits()
	e.result = SearchResult{}
	e.timer = _NewTimer()
	return e
}

// Starting seraching for the bestmove
func (e *Engine) Search() SearchResult {
	// In the future, add some setup, maybe don't use 'main' thread
	e._IterativeDeepening()
	return e.result
}

// Get the position object
func (e *Engine) Position() *Position {
	return e.position
}

// Set the limits
func (e *Engine) SetLimits(limits Limits) {
	*e.limits = limits
	e.timer.Movetime(e.limits.movetime)
}

func (e *Engine) Stop() {
	e.stop.Store(true)
}
