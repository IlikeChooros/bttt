package uttt

type Termination int

const (
	TerminationNone      Termination = 0
	TerminationCircleWon Termination = 1
	TerminationCrossWon  Termination = 2
	TerminationDraw      Termination = 4
	TerminationResigned  Termination = 8
)

var _patterns = [8][3]int{
	{0, 1, 2}, {3, 4, 5}, {6, 7, 8},
	{0, 3, 6}, {1, 4, 7}, {2, 5, 8},
	{0, 4, 8}, {2, 4, 6},
}

// Set the termination flag
func (p *Position) SetTermination(t Termination) {
	p.termination = t
}

// Get the termination reason (after, calling IsTerminated, or CheckTerminationPattern)
func (p *Position) Termination() Termination {
	return p.termination
}

// Check if the whole board is terminated
func (p *Position) IsTerminated() bool {
	if p.termination != TerminationNone {
		return true
	}

	// Evaluate termination
	p.CheckTerminationPattern()
	return p.termination != TerminationNone
}

// Check if given slice is filled with items other than 'none'
func _isFilled[T comparable](arr []T, none T) bool {
	is_filled := true
	for i := 0; is_filled && i < len(arr); i++ {
		is_filled = arr[i] != none
	}
	return is_filled
}

// Check if given 'small' square is terminated
// TODO: use bitboards for this function
func _checkSquareTermination(crossbb, circlebb uint) PositionState {

	for _, pattern := range _winningPatterns {
		if crossbb&pattern == pattern {
			return PositionCrossWon
		}
		if circlebb&pattern == pattern {
			return PositionCircleWon
		}
	}

	if (crossbb | circlebb) == 0b111111111 {
		return PositionDraw
	}
	return PositionUnResolved

	// // Check winning conditions for all patterns
	// for _, pattern := range _patterns {
	// 	// Check this pattern, and resolve it
	// 	if v := square[pattern[0]]; v == square[pattern[1]] &&
	// 		square[pattern[1]] == square[pattern[2]] &&
	// 		v != PieceNone {

	// 		state := PositionCircleWon
	// 		// Check if that terminates that board, meaning one of the sides won
	// 		if v == PieceCross {
	// 			state = PositionCrossWon
	// 		}
	// 		return state
	// 	}
	// }

	// // Check draw conditions
	// // Fully filled, and no outcome, meaning that's a draw
	// if _isFilled(square[:], PieceNone) {
	// 	return PositionDraw
	// }

	// // Unresolved
	// return PositionUnResolved
}

func (pos *Position) CheckTerminationPattern() {
	// Check if we are in a terminated state of the board
	// Assuming we correctly updated 'bigPositionState'

	// Check first draw condition:
	// If our current BigIndex board,
	// Is fully filled, thus no move is possible
	if bi := int(pos.BigIndex()); bi != int(PosIndexIllegal) &&
		((pos.bitboards[0][bi] | pos.bitboards[1][bi]) == 0b111111111) {
		pos.termination = TerminationDraw
		return
	}

	// Check winning conditions for all patterns
	for _, pattern := range _patterns {
		// Check this pattern, and resolve it
		if v := pos.bigPositionState[pattern[0]]; v == pos.bigPositionState[pattern[1]] &&
			pos.bigPositionState[pattern[1]] == pos.bigPositionState[pattern[2]] &&
			v != PositionUnResolved && v != PositionDraw {

			// Got a winner
			if v == PositionCircleWon {
				pos.termination = TerminationCircleWon
			} else {
				pos.termination = TerminationCrossWon
			}
			return // Exit the function
		}
	}

	// Check other draw condition
	// If there is no winner, check if all of the squares position
	// aren't unresolved, if so that means we got a draw
	is_draw := _isFilled(pos.bigPositionState[:], PositionUnResolved)
	if is_draw {
		pos.termination = TerminationDraw
	} else {
		// This is neither a draw or a win, so there is no termination
		pos.termination = TerminationNone
	}
}
