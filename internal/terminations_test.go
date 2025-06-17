package bttt

import (
	"fmt"
	"testing"
)

func TestInitTermination(t *testing.T) {
	// Simply test if the initial position, has properly set the
	// TerminationNone flag
	pos := NewPosition()
	if pos.IsTerminated() {
		t.Errorf("pos.IsTerminated()=%t, want=%t (term=%+v)", pos.IsTerminated(), false, pos.termination)
	}

	if pos.termination != TerminationNone {
		t.Errorf("pos.termination=%d, want=%d", pos.termination, TerminationNone)
	}
}

func TestTerminatedPositions(t *testing.T) {
	// Test if given terminated position are correctly detected
	notations := []string{
		StartingPosition, // no termination
		// x o x Draw position
		// x o x
		// o x o
		"xoxxoxoxo/xoxxoxoxo/xoxxoxoxo/xoxxoxoxo/xoxxoxoxo/xoxxoxoxo/xoxxoxoxo/xoxxoxoxo/xoxxoxoxo o 0",
		// x x x Impossible position: Winning position for X
		"xxx6/xxx6/xxx6/9/9/9/9/9/9 x 0",
		// o o o Impossible position: Winning position for O
		"ooo6/9/ooo6/9/ooo6/xoxo5/ooo6/9/ooo6 x 0",
	}

	terminations := []Termination{
		TerminationNone,
		TerminationDraw,
		TerminationCrossWon,
		TerminationCircleWon,
	}

	pos := NewPosition()
	for i, str := range notations {
		t.Run(fmt.Sprintf("SubTest-%s", str), func(ts *testing.T) {
			// Catch invalid notation string
			err := pos.FromNotation(str)
			if err != nil {
				ts.Error(err)
			} else {
				// Test the position
				if term := terminations[i]; pos.Termination() != term {
					ts.Errorf("pos.Termination()=%d, want=%d", pos.Termination(), term)
				}
			}
		})
	}
}

func TestSmallSquareTermination(t *testing.T) {
	// Check if the _checkSquareTermination function is working properly

	O, X, n := PieceCircle, PieceCross, PieceNone

	squares := [][9]PieceType{
		{
			O, X, n,
			X, X, O,
			X, O, n,
		},
		{
			O, X, n,
			X, X, O,
			X, X, O,
		},
		{
			O, X, n,
			O, O, O,
			X, n, X,
		},
		{
			O, X, O,
			O, X, O,
			X, O, X,
		},
	}

	states := []PositionState{
		PositionUnResolved,
		PositionCrossWon,
		PositionCircleWon,
		PositionDraw,
	}

	for i, square := range squares {
		// Run each position as a sub test
		t.Run(fmt.Sprintf("SquareTest-%d", i+1), func(t *testing.T) {
			state := _checkSquareTermination(square)

			if state != states[i] {
				t.Errorf("_checkSquareTermination(%v)=%d, want=%d", square, state, states[i])
			}
		})
	}
}
