package uttt

import (
	"fmt"
	"unsafe"
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

// Mate values for Score type
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
	Nps       uint64
	Depth     int
	Cycles    int32
	Pv        MoveList
}

func (s SearchResult) String() string {
	return fmt.Sprintf("eval %s depth %d nps %d nodes %d cycles %d pv %s",
		s.StringValue(), s.Depth, s.Nps, s.Nodes, s.Cycles, s.Pv.String())
}

// Get the string representation of the value
func (s SearchResult) StringValue() string {
	if s.ScoreType == MateScore {
		return fmt.Sprintf("%dM", s.Value)
	} else if s.Value == -1 {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", float32(s.Value)/100.0)
}

// Fast bool to int conversion
func _boolToInt(v bool) int {
	return int(*(*byte)(unsafe.Pointer(&v)))
}

// Enum for position
const (
	PosIllegal      PosType = 255
	PosIndexIllegal PosType = 15 // same as big/small index mask
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
	CircleTurn TurnType = false
	CrossTurn  TurnType = true
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
