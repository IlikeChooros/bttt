package bttt

// Type defines for the position
type PieceType int8
type TurnType bool
type PosType uint8
type BoardType [9][9]PieceType

// Enum for position
const (
	posIllegal PosType = 255
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
