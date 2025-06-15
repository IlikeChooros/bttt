package bttt

// Main board struct
type Board struct {
	board     boardType  // 2d array of the pieces [bigIndex][smallIndex]
	turn      turnType   // current turn
	stateList *StateList // history of the board (for MakeMove, UndoMove)
}

// Create a heap-allocated, initialized board
func NewBoard() *Board {
	b := new(Board)
	b.Init()
	return b
}

/*		Board methods		*/

// Initialize the board
func (b *Board) Init() {
	b.turn = CrossTurn
	b.stateList = NewStateList()
}

// Getters
func (b *Board) Board() boardType {
	return b.board
}

func (b *Board) Turn() turnType {
	return b.turn
}

// Make a move on the board, switches the sides, and puts current piece
// on the board [bigIndex][smallIndex]
func (b *Board) MakeMove(bigIndex, smallIndex posType) {
	// Make sure the coordinates are correct
	if smallIndex > 8 || bigIndex > 8 {
		return
	}

	piece := Circle
	if !b.turn {
		piece = Cross
	}

	// Put that piece on the board
	b.board[bigIndex][smallIndex] = piece

	// Change the turn
	b.turn = !b.turn
}
