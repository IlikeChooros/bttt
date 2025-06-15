package bttt

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

// Get the big index of a move
func (pos PosType) BigIndex() PosType {
	return pos & _moveBigIndexMask
}

// Get the small index of tic tac toe board
func (pos PosType) SmallIndex() PosType {
	return pos & _moveSmallIndexMask
}

// Appends a new move to the list of moves
func (ml *MoveList) Append(bigIndex, smallIndex int) {
	ml.moves[ml.size] = MakeMove(bigIndex, smallIndex)
	ml.size++
}
