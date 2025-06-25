package uttt

import (
	"math/bits"
)

func (pos *Position) _GenerateMoves(movelist *MoveList, bigIndex int) {
	// Go through each cell, and pick empty ones
	free := 0b111111111 ^ pos.bitboards[0][bigIndex] ^ pos.bitboards[1][bigIndex]
	for free != 0 {
		movelist.Append(bigIndex, bits.TrailingZeros(free))
		free &= free - 1
	}
}

// Generate all possible moves in given position
func (pos *Position) GenerateMoves() *MoveList {
	movelist := NewMoveList()

	// If there is no history, we can choose also the 'Big Index' position
	if pos.BigIndex() == int(PosIndexIllegal) {
		for i := range 9 {
			pos._GenerateMoves(movelist, i)
		}
	} else {
		// Else we generate moves for the 'Big Index' position
		pos._GenerateMoves(movelist, pos.BigIndex())
	}

	return movelist
}
