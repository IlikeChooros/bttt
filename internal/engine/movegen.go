package uttt

import (
	"math/bits"
)

// Generate all possible moves in given position
func (pos *Position) GenerateMoves() *MoveList {
	movelist := NewMoveList()
	free := uint(0)

	// If there is no history, we can choose also the 'Big Index' position
	if pos.BigIndex() == int(PosIndexIllegal) {
		for bigIndex := range 9 {
			free = 0b111111111 ^ pos.bitboards[0][bigIndex] ^ pos.bitboards[1][bigIndex]
			for free != 0 {
				movelist.Append(bigIndex, bits.TrailingZeros(free))
				free &= free - 1
			}
		}
	} else {
		// Else we generate moves for the 'Big Index' position
		bi := pos.BigIndex()
		free = 0b111111111 ^ pos.bitboards[0][bi] ^ pos.bitboards[1][bi]
		for free != 0 {
			movelist.Append(bi, bits.TrailingZeros(free))
			free &= free - 1
		}
	}

	return movelist
}
