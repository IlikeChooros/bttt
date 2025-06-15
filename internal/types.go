package bttt

// Type defines for the board
type pieceType int8
type turnType bool
type posType uint8
type boardType [9][9]pieceType

// Enum for the piece type
const (
	None pieceType = iota
	Circle
	Cross
)

// Enum for the turns
const (
	CicrleTurn turnType = true
	CrossTurn  turnType = false
)
