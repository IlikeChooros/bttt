package bttt

import utils "bttt/internal/utils"

// To optimize hashing the position for search, we first calculate EVERY possible hash
// for the regular tic tac toe board, meaning we get total of 3^9 = 19,683 positions (including 'invalid' ones)
// which takes about 14 bits per position to achieve perfect hashing for the small board
// Therefore, to create a perfect hasing table for the 'big' board, we should have at least
// 14 * 9 = 126 bits, thus I defined the uint128 type with basic bitwise arithmetics
type HashTable[T any] map[utils.Uint128]T

// Transposition table
var _transpTable HashTable[HashEntry]

// Hash the small board
func HashSmallBoard(bitboard uint) {

}

func setupTranspTable() {

}
