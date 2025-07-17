package uttt

import (
	"io"
	"os"
	"sync/atomic"
	"unsafe"
)

/*
Main engine class, allowing user to make moves on the board,
search best move, based on given parameteres
*/
type Engine struct {
	position   *Position
	limits     *Limits
	timer      *_Timer
	ttable     *HashTable[TTEntry]
	result     SearchResult
	stop       atomic.Bool
	print      bool
	isThinking atomic.Bool
	pv         *MoveList
	writer     io.Writer
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
	e.ttable = NewHashTable[TTEntry](16 * (1 << 20) / uint64(unsafe.Sizeof(TTEntry{})))
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
	e.isThinking.Store(true)
	e.print = print
	e._IterativeDeepening()
	e.isThinking.Store(false)
	return e.result
}

func (e *Engine) IsThinking() bool {
	return e.isThinking.Load()
}

// Get the position object
func (e *Engine) Position() *Position {
	return e.position
}

// Resets all search cache
func (e *Engine) NewGame() {
	// Wait for search to end
	for e.IsThinking() {
	}
	e.ttable.Clear()
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
