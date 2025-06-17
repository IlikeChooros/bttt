package bttt

import (
	"math/bits"
)

// horizontal, vertical and diagonal patterns as bitboards
var _winningPatterns [8]uint = [...]uint{
	0b111000000, 0b000111000, 0b000000111,
	0b100100100, 0b010010010, 0b001001001,
	0b100010001, 0b001010100,
}

// Assing value to each of the
var _pieceSquareTable [9]Value = [...]Value{
	5, 0, 5,
	0, 15, 0,
	5, 0, 5,
}

// Multiplier, for each 'small' square
var _bigSquareTableFactors [9]float32 = [...]float32{
	1.0, 0.8, 1.0,
	0.8, 1.6, 0.8,
	1.0, 0.8, 1.0,
}

var pow2table [6]int = [...]int{
	1, 2, 4, 8, 16, 32,
}

// Look for patterns, and assign to each one value
// Assign value to each result
// When Enemy intersects our pattern: -5
// When we need 3 moves to complete the pattern: 0
// When we need 2 moves to resolve our pattern: 5
// When we need 1 move to complete the pattern: 15
func _evaluatePattern(pattern, bitboard, enemy_bitboard uint) Value {
	// Evaluate our patterns
	pattern_eval := Value(0)
	our_count := bits.OnesCount(pattern & bitboard)
	intersection := (pattern & bitboard) ^ pattern

	// Enemy itersects our pattern, we can't resolve it
	if intersection&enemy_bitboard != 0 {
		pattern_eval -= 5
	} else {
		pattern_eval += Value((pow2table[our_count] - 1) * 5)
	}

	return pattern_eval
}

func _evaluateSquare(square []PieceType, ourPiece PieceType) Value {
	// Look for patterns
	eval := Value(0)
	enemy_bitboard := uint(0)
	bitboard := uint(0)
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

	// Evaluate patterns
	pattern_eval := Value(0)
	for _, pattern := range _winningPatterns {
		// Evaluate our patterns
		pattern_eval += _evaluatePattern(pattern, bitboard, enemy_bitboard)
		pattern_eval -= _evaluatePattern(pattern, enemy_bitboard, bitboard)
	}

	// Add up the evaluation
	eval += pattern_eval
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
		eval += Value(float32(_evaluateSquare(pos.position[i][:], ourPiece)) * _bigSquareTableFactors[i])
	}

	return eval
}
