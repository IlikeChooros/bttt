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

func HashSmallBoard(o_bitboard, x_bitboard uint) uint {
	// Since the o's and x's are by themselves hashes of the position
	// This is as simple as shifting one to the left by 9
	return (o_bitboard & 0b111111111) | ((x_bitboard << 9) & 0b1111111111000000000)
}

func setupTranspTable() {

}
