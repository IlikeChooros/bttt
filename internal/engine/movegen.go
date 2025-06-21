package bttt

func (pos *Position) _GenerateMoves(movelist *MoveList, bigIndex int) {
	// Go through each cell, and pick empty ones
	for position, v := range pos.position[bigIndex] {
		if v == PieceNone {
			movelist.Append(bigIndex, position)
		}
	}
}

// Generate all possible moves in given position
func (pos *Position) GenerateMoves() *MoveList {
	movelist := NewMoveList()

	// If there is no history, we can choose also the 'Big Index' position
	if pos.BigIndex() == int(posIndexIllegal) {
		for i := 0; i < 9; i++ {
			pos._GenerateMoves(movelist, i)
		}
	} else {
		// Else we generate moves for the 'Big Index' position
		pos._GenerateMoves(movelist, pos.BigIndex())
	}

	return movelist
}
