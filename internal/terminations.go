package bttt

type Termination int

const (
	TerminationNone      Termination = 0
	TerminationCircleWon Termination = 1
	TerminationCrossWon  Termination = 2
	TerminationDraw      Termination = 4
	TerminationResigned  Termination = 8
)

// Set the termination flag
func (p *Position) SetTermination(t Termination) {
	p.termination = t
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

func (pos *Position) CheckTerminationPattern() {
	// Check if we are in a terminated state of the board
	// Assuming we correctly updated 'bigPositionState'
	patterns := [8][3]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8},
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8},
		{0, 4, 8}, {2, 4, 6},
	}

	// Check draw condition
	is_draw := true
	if pos.bigPositionState[patterns[0][0]] != pos.bigPositionState[patterns[0][1]] ||
		pos.bigPositionState[patterns[0][1]] != pos.bigPositionState[patterns[0][2]] ||
		pos.bigPositionState[patterns[0][0]] != PositionDraw {
		is_draw = false
	}

	// Check draw condition
	if is_draw {
		pos.termination = TerminationDraw
		return
	}

	// Check winning conditions for all patterns
	for _, pattern := range patterns {
		// Check this pattern, and resolve it
		if v := pos.bigPositionState[pattern[0]]; v == pos.bigPositionState[pattern[1]] &&
			pos.bigPositionState[pattern[1]] == pos.bigPositionState[pattern[2]] {

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
