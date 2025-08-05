package uttt

import (
	"uttt/internal/mcts"
)

/*
Main engine class, allowing user to make moves on the board,
search best move, based on given parameteres
*/
type Engine struct {
	mcts   *UtttMCTS
	policy mcts.BestChildPolicy
}

// Initialize the package
func Init() {
	_InitHashing()
}

// Get new engine instance
func NewEngine() *Engine {
	return &Engine{
		mcts:   NewUtttMCTS(*NewPosition()),
		policy: mcts.BestChildMostVisits,
	}
}

func (e *Engine) SetBestChildPolicy(policy mcts.BestChildPolicy) {
	e.policy = policy
}

// Starting seraching for the bestmove
func (e *Engine) Search() {
	// In the future, add some setup, maybe don't use 'main' thread
	go e.Think()
}

// Search the moves, in a blocking way
func (e *Engine) Think() SearchResult {
	e.mcts.Search()
	return e.mcts.SearchResult(e.policy)
}

func (e *Engine) IsThinking() bool {
	return e.mcts.IsThinking()
}

func (e *Engine) SetNotation(notation string) error {
	return e.mcts.SetNotation(notation)
}

// Get the position object
func (e *Engine) Position() *Position {
	return &e.mcts.ops.position
}

// Resets all search cache
func (e *Engine) NewGame() {
	e.mcts.Reset()
}

// Set the limits
func (e *Engine) SetLimits(limits *mcts.Limits) {
	e.mcts.SetLimits(limits)
}

func (e *Engine) Stop() {
	e.mcts.Stop()
}

func (e *Engine) Pv() *MoveList {
	pv, _, _ := e.mcts.Pv(e.mcts.Root, e.policy, false)
	return ToMoveList(pv)
}

func (e *Engine) MultiPv() []mcts.PvResult[PosType] {
	return e.mcts.MultiPv(e.policy)
}

func (e *Engine) Mcts() *UtttMCTS {
	return e.mcts
}
