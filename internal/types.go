package bttt

import "strings"

// Type defines for the position
type PieceType int8
type TurnType bool
type PosType uint8
type BoardType [9][9]PieceType
type PositionState uint8

// Type defines for search/limits
type ScoreType uint8
type HashEntry struct {
}

const (
	ValueScore ScoreType = 0
	MateScore  ScoreType = 1
)

// Struct holding information about the score value of the search
type SearchResult struct {
	Value     int
	ScoreType ScoreType
	Bestmove  PosType
	Nodes     uint64
}

const (
	MateValue = -1000000
)

// Enum for position
const (
	posIllegal      PosType = 255
	posIndexIllegal PosType = 15 // same as big/small index mask
)

const (
	PositionUnResolved PositionState = iota
	PositionDraw
	PositionCircleWon
	PositionCrossWon
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
	si, bi := pos.SmallIndex(), pos.BigIndex()

	if si >= 9 || bi >= 9 {
		return "(none)"
	}

	builder.WriteByte('A' + byte(bi%3))
	builder.WriteByte('0' + byte(3-bi/3))
	builder.WriteByte('a' + byte(si%3))
	builder.WriteByte('0' + byte(3-si/3))

	return builder.String()
}
