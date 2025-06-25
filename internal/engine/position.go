package uttt

// Constants

const (
	StartingPosition string = "9/9/9/9/9/9/9/9/9 x -"
)

// Main position struct
type Position struct {
	position         BoardType // 2d array of the pieces [bigIndex][smallIndex]
	bitboards        [2][9]uint
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
	p.termination = TerminationNone
	p.hash = 0

	// Zero the board
	for i := range p.position {
		for j := range p.position[i] {
			p.position[i][j] = PieceNone
		}
	}

	// Zero the bigPositionState
	for i := range p.bigPositionState {
		p.bigPositionState[i] = PositionUnResolved
		p.bitboards[0][i] = 0
		p.bitboards[1][i] = 0
	}
}

// Convert given 'small square' with given 'ourPiece' parameter, into (our bitboard, enemy bitboard)
func toBitboards(square [9]PieceType, ourPiece PieceType) (bitboard, enemy_bitboard uint) {
	// Write whole board into a bitboard
	for i, v := range square {
		// Evaluate square table evaluation
		if v == ourPiece {
			bitboard |= (1 << i)
		} else if v != PieceNone {
			// Enemy
			enemy_bitboard |= (1 << i)
		}
	}

	return bitboard, enemy_bitboard
}

// Make sure bitboards represent the same position as the 2d arrays
func (p *Position) MatchBitboards() {
	for i, square := range p.position {
		p.bitboards[1][i], p.bitboards[0][i] = toBitboards(square, PieceCross)
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
	if p.termination != TerminationNone {
		return
	}

	smallIndex := move.SmallIndex()
	bigIndex := move.BigIndex()

	// Make sure the coordinates are correct
	if smallIndex > 8 || bigIndex > 8 {
		return
	}

	// Choose the piece, based on the current side to move
	piece := PieceCross
	lastState := p.stateList.Last()
	posStateBefore := p.bigPositionState[bigIndex]

	// Meaning last turn, cross made a move, so now it's circle's turn
	if lastState.turn == CrossTurn {
		piece = PieceCircle
	}

	// Put that piece on the position
	p.position[bigIndex][smallIndex] = piece
	p.bitboards[_boolToInt(bool(p.Turn()))][bigIndex] ^= (1 << smallIndex)

	// Update Big board state
	if p.bigPositionState[bigIndex] == PositionUnResolved {
		p.bigPositionState[bigIndex] = _checkSquareTermination(
			p.bitboards[1][bigIndex], p.bitboards[0][bigIndex],
		)
	}

	// Append new state
	p.stateList.Append(move, !lastState.turn, posStateBefore)
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
	p.bitboards[_boolToInt(bool(p.Turn()))][bigIndex] ^= (1 << smallIndex)

	// Restore bigPositionState
	p.bigPositionState[bigIndex] = lastState.thisPositionState

	// Restore termination
	p.termination = TerminationNone

	// Restore current state
	p.stateList.Remove()
}

// Get the 'big position state'
func (p *Position) BigPositionState() [9]PositionState {
	return p.bigPositionState
}

// Check if given move is legal
func (p *Position) IsLegal(move PosType) bool {

	bi, si := move.BigIndex(), move.SmallIndex()
	if p.BigIndex() != int(PosIndexIllegal) && bi != PosType(p.BigIndex()) {
		return false
	}

	// Index out of range or this square is occupied
	if bi >= 9 || si >= 9 || p.position[bi][si] != PieceNone {
		return false
	}

	return true
}
