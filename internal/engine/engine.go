package uttt

/*
Main engine class, allowing user to make moves on the board,
search best move, based on given parameteres
*/
type Engine struct {
	mcts   *UtttMCTS
	result SearchResult
}

// Initialize the package
func Init() {
	_InitHashing()
}

// Get new engine instance
func NewEngine() *Engine {
	e := &Engine{
		mcts: NewUtttMCTS(*NewPosition()),
	}
	e.result = SearchResult{}
	return e
}

// Starting seraching for the bestmove
func (e *Engine) Search() {
	// In the future, add some setup, maybe don't use 'main' thread
	go e.Think(true)
}

// Search the moves, in a blocking way
func (e *Engine) Think(print bool) SearchResult {
	e.mcts.Search()
	return e.mcts.SearchResult()
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
func (e *Engine) SetLimits(limits Limits) {
	e.mcts.SetLimits(limits)
}

func (e *Engine) Stop() {
	e.mcts.Stop()
}

func (e *Engine) Pv() *MoveList {
	pv, _ := e.mcts.Pv()
	return pv
}
