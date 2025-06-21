package bttt

import (
	"strings"
)

// Type defines for the position
type PieceType int8
type TurnType bool
type PosType uint8 // Also used as move representation
type BoardType [9][9]PieceType
type PositionState uint8
type EntryNodeType uint8

// Type defines for search/limits
type ScoreType uint8
type HashEntry struct {
	depth    int
	hash     string
	score    int
	nodeType EntryNodeType
	bestmove PosType
}

const (
	Exact      EntryNodeType = iota // Exact value of the node (a pvs node)
	LowerBound                      // It's value if <= alpha
	UpperBound                      // This node caused a beta-cutoff (beta >= alpha)
)

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

// Enum for the squares (same for the smaller ones)
const (
	A3 PosType = iota
	B3
	C3
	A2
	B2
	C2
	A1
	B1
	C1
)

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
	builder.WriteByte('3' - byte(bi/3))
	builder.WriteByte('a' + byte(si%3))
	builder.WriteByte('3' - byte(si/3))

	return builder.String()
}

func MoveFromString(str string) PosType {
	if str == "(none)" || len(str) != 4 {
		return posIllegal
	}

	// Helper function to make sure the coordinates are withing the range
	_cmp := func(i int, letter byte) bool {
		return (str[i] >= letter && str[i] <= letter+2) &&
			(str[i+1] >= '1' && str[i+1] <= '3')
	}

	if _cmp(0, 'A') && _cmp(2, 'a') {
		return MakeMove(
			int((str[0]-'A')+('3'-str[1])*3),
			int((str[2]-'a')+('3'-str[3])*3))
	}

	return posIllegal
}
