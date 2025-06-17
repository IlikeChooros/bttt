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

// value representing winning on the 'big square' of the board
var _winningBigSquareValue Value = 50

// Multiplier, for each 'small' square
var _bigSquareTableFactors [9]float32 = [...]float32{
	1.0, 0.8, 1.0,
	0.8, 1.6, 0.8,
	1.0, 0.8, 1.0,
}

var pow2table [6]int = [...]int{
	1, 2, 4, 8, 16, 32,
}

// Convert given 'small square' with given 'ourPiece' parameter, into (our bitboard, enemy bitboard)
func _toBitboards(square [9]PieceType, ourPiece PieceType) (bitboard, enemy_bitboard uint) {
	// Write whole board into a bitboard
	for i, v := range square {
		// Evaluate square table evaluation
		if v == ourPiece {
			bitboard |= (1 << i)
		} else if v != PieceNone {
			// Enemy
			enemy_bitboard |= (1 << i)
		}
	}

	return bitboard, enemy_bitboard
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

func _evaluateSquare(square [9]PieceType, ourPiece PieceType) Value {
	// Look for patterns
	eval := Value(0)
	bitboard, enemy_bitboard := _toBitboards(square, ourPiece)
	square_table_eval := Value(0)

	// Calculate the piece square table for each side
	temp, enemytemp := bitboard, enemy_bitboard
	for temp != 0 {
		square_table_eval += _pieceSquareTable[bits.TrailingZeros(temp)]
		temp &= (temp - 1)
	}

	for enemytemp != 0 {
		square_table_eval -= _pieceSquareTable[bits.TrailingZeros(enemytemp)]
		enemytemp &= (enemytemp - 1)
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
	winningState := PositionCircleWon

	if !pos.Turn() {
		ourPiece = PieceCross
		winningState = PositionCrossWon
	}

	// Evaluate whole board
	for i := range pos.position {
		value := Value(0)

		if state := pos.bigPositionState[i]; state == PositionUnResolved {
			// Evaluate unresolved square
			value += _evaluateSquare(pos.position[i], ourPiece)
		} else {
			// Assign value by the square state
			if state != PositionDraw {
				if state == winningState {
					value = _winningBigSquareValue
				} else {
					value = -_winningBigSquareValue
				}
			}
		}

		// Add product of value and it's factor
		eval += Value(float32(value) * _bigSquareTableFactors[i])
	}

	return eval
}
