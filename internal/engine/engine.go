package uttt

import (
	"io"
	"os"
	"sync/atomic"
)

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
	print    bool
	pv       *MoveList
	writer   io.Writer
}

// Initialize the package
func Init() {
	_InitHashing()
}

// Get new engine instance
func NewEngine() *Engine {
	e := new(Engine)
	e.position = NewPosition()
	e.limits = DefaultLimits()
	e.result = SearchResult{}
	e.timer = _NewTimer()
	e.pv = NewMoveList()
	e.writer = os.Stdout
	return e
}

// Set the output stream of the engine, by default uses standard output
func (e *Engine) SetWriter(writer io.Writer) {
	e.writer = writer
}

// Starting seraching for the bestmove
func (e *Engine) Search() {
	// In the future, add some setup, maybe don't use 'main' thread
	go e.Think(true)
}

// Search the moves, in a blocking way
func (e *Engine) Think(print bool) SearchResult {
	e.print = print
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

func (e *Engine) Pv() *MoveList {
	return e.pv
}
