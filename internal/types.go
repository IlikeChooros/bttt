package bttt

import "strings"

// Type defines for the position
type PieceType int8
type TurnType bool
type PosType uint8
type BoardType [9][9]PieceType

// Type defines for search/limits
type Depth int
type Value int

const (
	MateValue Value = 1000000
)

// Enum for position
const (
	posIllegal      PosType = 255
	posIndexIllegal PosType = 15 // same as big/small index mask
)

// Enum for the piece type
const (
	PieceNone PieceType = iota
	PieceCircle
	PieceCross
)

// Enum for the turns
const (
	CircleTurn TurnType = true
	CrossTurn  TurnType = false
)

// Create piece from a rune
func PieceFromRune(square rune) PieceType {
	switch square {
	case 'x':
		return PieceCross
	case 'o':
		return PieceCircle
	default:
		return PieceNone
	}
}

// Get the big index of a move
func (pos PosType) BigIndex() PosType {
	return (pos & _moveBigIndexMask) >> 4
}

// Get the small index of tic tac toe board
func (pos PosType) SmallIndex() PosType {
	return pos & _moveSmallIndexMask
}

// Get string representation of the move, will contain
// a/b/c 1/2/3 as coorinates, for example big index = 7,
// small index = 2 -> <big index part><small index part>
// -> B1c3
//
//	     	A    B    C
//			 0 | 1 | 2	3
//			-----------
//			 3 | 4 | 5	2
//			-----------
//		     6 | 7 | 8	1
func (pos PosType) String() string {
	builder := strings.Builder{}
	sx, sy := byte(pos.SmallIndex()%3), byte(3-pos.SmallIndex()/3)
	bx, by := byte(pos.BigIndex()%3), byte(3-pos.BigIndex()/3)

	builder.WriteByte('A' + bx)
	builder.WriteByte('0' + by)
	builder.WriteByte('a' + sx)
	builder.WriteByte('0' + sy)

	return builder.String()
}
