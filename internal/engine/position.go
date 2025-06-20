package bttt

// Constants

const (
	StartingPosition string = "9/9/9/9/9/9/9/9/9 o -"
)

// Main position struct
type Position struct {
	position         BoardType        // 2d array of the pieces [bigIndex][smallIndex]
	bigPositionState [9]PositionState // Array of uint8's, where each one means, either cross, circle or no one won on that square
	stateList        *StateList       // history of the position (for MakeMove, UndoMove)
	termination      Termination
	hash             uint64 // Current hash of the position
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
}

func (p *Position) Reset() {
	p.stateList.Clear()

	// Zero the board
	for i := range p.position {
		for j := range p.position[i] {
			p.position[i][j] = PieceNone
		}
	}
}

// Getters
func (b *Position) Position() BoardType {
	return b.position
}

func (b *Position) Turn() TurnType {
	return !b.stateList.Last().turn
}

func (p *Position) BigIndex() int {
	return int(p.stateList.NextBigIndex())
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
	piece := PieceCross
	lastState := p.stateList.Last()

	// Meaning last turn, cross made a move, so now it's circle's turn
	if !lastState.turn {
		piece = PieceCircle
	}

	// Put that piece on the position
	p.position[bigIndex][smallIndex] = piece

	// Update Big board state
	p.bigPositionState[bigIndex] = _checkSquareTermination(p.position[bigIndex])

	// Append new state
	p.stateList.Append(move, !lastState.turn)
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
