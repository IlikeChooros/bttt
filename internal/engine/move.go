package uttt

const (
	_moveBigIndexMask   = 0b11110000
	_moveSmallIndexMask = 0b1111
)

type MoveList struct {
	moves [9 * 9]PosType
	size  uint8
}

// Make a new move list struct
func NewMoveList() *MoveList {
	mv := new(MoveList)
	mv.size = 0
	return mv
}

// Create a move, based on big and small indexes
func MakeMove(bigIndex, smallIndex int) PosType {
	return PosType((smallIndex & _moveSmallIndexMask) | ((bigIndex << 4) & _moveBigIndexMask))
}

// Reset the movelist, simply sets the size to 0
func (ml *MoveList) Clear() {
	ml.size = 0
}

// Get the actual slice of valid moves
func (ml *MoveList) Slice() []PosType {
	return ml.moves[0:ml.size]
}

// Appends a new move to the list of moves
func (ml *MoveList) Append(bigIndex, smallIndex int) {
	ml.moves[ml.size] = MakeMove(bigIndex, smallIndex)
	ml.size++
}
