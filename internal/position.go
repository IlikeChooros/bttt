package bttt

// Main position struct
type Position struct {
	position  BoardType  // 2d array of the pieces [bigIndex][smallIndex]
	stateList *StateList // history of the position (for MakeMove, UndoMove)
	moves     *MoveList  // all possible moves in given position
}

// Create a heap-allocated, initialized Big Tic Tac Toe position
func NewPosition() *Position {
	b := new(Position)
	b.Init()
	return b
}

// Initialize the position
func (b *Position) Init() {
	b.stateList = NewStateList()
	b.moves = NewMoveList()
}

// Getters
func (b *Position) Position() BoardType {
	return b.position
}

func (b *Position) Turn() TurnType {
	return b.stateList.Last().turn
}

func (p *Position) BigIndex() int {
	return int(p.stateList.NextBigIndex())
}

// Get pointer to the MoveList struct, holding generated legal moves
func (p *Position) Moves() *MoveList {
	return p.moves
}

// Make a move on the position, switches the sides, and puts current piece
// on the position [bigIndex][smallIndex], accepts any move
func (p *Position) MakeMove(move PosType) {
	smallIndex := move.SmallIndex()
	bigIndex := move.BigIndex()

	// Make sure the coordinates are correct
	if smallIndex > 8 || bigIndex > 8 {
		return
	}

	// Choose the piece, based on the current side to move
	piece := PieceCircle
	lastState := p.stateList.Last()
	if !lastState.turn {
		piece = PieceCross
	}

	// Put that piece on the position
	p.position[bigIndex][smallIndex] = piece

	// Append new state
	p.stateList.Append(move, !p.Turn())
}

// Undo last move, from the state list
func (p *Position) UndoMove() {
	if p.stateList.ValidSize() == 0 {
		return
	}

	// Get the coordiantes
	lastState := p.stateList.Last()
	smallIndex := lastState.move.SmallIndex()
	bigIndex := lastState.move.BigIndex()

	// Remove that piece from it's square
	p.position[bigIndex][smallIndex] = PieceNone

	// Restore current state
	p.stateList.Remove()
}
