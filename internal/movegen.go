package bttt

func (pos *Position) _GenerateMoves(bigIndex int) {
	// Go through each cell, and pick empty ones
	for position, v := range pos.position[bigIndex] {
		if v == PieceNone {
			pos.moves.Append(bigIndex, position)
		}
	}
}

// Generate all possible moves in given position
func (pos *Position) GenerateMoves() {

	// If there is no history, we can choose also the 'Big Index' position
	if pos.stateList.ValidSize() == 0 {
		for i := 0; i < 9; i++ {
			pos._GenerateMoves(i)
		}
		return
	}

	// Else we generate moves for the 'Big Index' position
	pos._GenerateMoves(pos.BigIndex())
}
