package bttt

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

func (pos *Position) SetupBoardState() {
	// Check each small square, and set proper big square state
	for i, square := range pos.position {
		pos.bigPositionState[i] = _checkSquareTermination(square)
	}
}

// Check if given 'small' square is terminated\
// TODO: use bitboards for this function
func _checkSquareTermination(square [9]PieceType) PositionState {

	// Check winning conditions for all patterns
	for _, pattern := range _patterns {
		// Check this pattern, and resolve it
		if v := square[pattern[0]]; v == square[pattern[1]] &&
			square[pattern[1]] == square[pattern[2]] &&
			v != PieceNone {

			state := PositionCircleWon
			// Check if that terminates that board, meaning one of the sides won
			if v == PieceCross {
				state = PositionCrossWon
			}
			return state
		}
	}

	// Check draw conditions
	is_filled := true
	for i := 0; i < 9 && is_filled; i++ {
		is_filled = square[i] != PieceNone
	}

	// Fully filled, and no outcome, meaning that's a draw
	if is_filled {
		return PositionDraw
	}

	// Unresolved
	return PositionUnResolved
}

func (pos *Position) CheckTerminationPattern() {
	// Check if we are in a terminated state of the board
	// Assuming we correctly updated 'bigPositionState'

	// Check draw condition
	is_draw := true
	if pos.bigPositionState[_patterns[0][0]] != pos.bigPositionState[_patterns[0][1]] ||
		pos.bigPositionState[_patterns[0][1]] != pos.bigPositionState[_patterns[0][2]] ||
		pos.bigPositionState[_patterns[0][0]] != PositionDraw {
		is_draw = false
	}

	// Check draw condition
	if is_draw {
		pos.termination = TerminationDraw
		return
	}

	// Check winning conditions for all patterns
	for _, pattern := range _patterns {
		// Check this pattern, and resolve it
		if v := pos.bigPositionState[pattern[0]]; v == pos.bigPositionState[pattern[1]] &&
			pos.bigPositionState[pattern[1]] == pos.bigPositionState[pattern[2]] &&
			v != PositionUnResolved && v != PositionDraw {

			// Check if that terminates that board, meaning one of the sides won
			if v == PositionCircleWon {
				pos.termination = TerminationCircleWon
			} else {
				pos.termination = TerminationCrossWon
			}
			return // Exit the function
		}
	}

	// This is neither a draw or a win, so there is no termination
	pos.termination = TerminationNone
}
