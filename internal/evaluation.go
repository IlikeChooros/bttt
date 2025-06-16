package bttt

// import (
// 	"math/big"
// 	"math/bits"
// )

// horizontal, vertical and diagonal
var _winningPatterns [8]int = [8]int{
	0b111000000, 0b000111000, 0b000000111,
	0b100100100, 0b010010010, 0b001001001,
	0b100010001, 0b001010100,
}

// Assing value to each of the
var _pieceSquareTable [9]Value = [9]Value{
	5, 0, 5,
	0, 15, 0,
	5, 0, 5,
}

func InitEvaluation() {
}

func _EvaluateSquare(square []PieceType, ourPiece PieceType) Value {
	// Look for patterns
	eval := Value(0)
	enemy_bitboard := 0
	bitboard := 0
	square_table_eval := Value(0)

	// Write whole board into a bitboard
	for i, v := range square {
		// Evaluate square table evaluation
		if v == ourPiece {
			square_table_eval += _pieceSquareTable[i]
			bitboard |= (1 << i)
		} else if v != PieceNone {
			// Enemy
			square_table_eval -= _pieceSquareTable[i]
			enemy_bitboard |= (1 << i)
		}
	}

	// Look for patterns, and assign to each one value
	// Apply Hamming distance to each of the binary codes,
	// (winning pattern, our bitboard), then assign value to each result
	// When Enemy intersects our pattern: -5
	// When we need 3 moves to complete the pattern: 0
	// When we need 2 moves to resolve our pattern: 5
	// When we need 1 move to complete the pattern: 15
	// for _, pattern := range _winningPatterns {
	// 	// Evaluate our patterns
	// }

	eval += square_table_eval
	return eval
}

// Returns relative value of this position (meaning positive value are good for us, negative for the enemy)
func Evaluate(pos *Position) Value {
	// Assuming the position is NOT terminated
	eval := Value(0)
	ourPiece := PieceCircle
	if !pos.Turn() {
		ourPiece = PieceCross
	}

	for i := range pos.position {
		eval += _EvaluateSquare(pos.position[i][:], ourPiece)
	}

	return eval
}
