package bttt

import (
	"fmt"
	"testing"
)

func HelperTestReversingMoves(
	pos *Position,
	expectedNotation string,
	moves []PosType,
) error {
	// Make the moves
	nmoves := len(moves)
	for _, m := range moves {
		pos.MakeMove(m)
	}

	// Undo them
	for i := 0; i < nmoves; i++ {
		pos.UndoMove()
	}

	// Check if we got the same state as before
	if pos.Notation() != expectedNotation {
		return fmt.Errorf("Notation, got = %s, want: %s", pos.Notation(), expectedNotation)
	}

	// Check if we don't have any history
	if pos.stateList.ValidSize() != 0 {
		return fmt.Errorf("Got history, meaning it didn't reverse all moves")
	}

	return nil
}

func TestMakeUndoMoves(t *testing.T) {
	// Starting from given root position, check if, after making a few moves, we can restore
	// the initial position by calling repeatedly UndoMove

	notations := []string{
		StartingPosition, // starting position
		"1o7/2x6/o8/9/9/9/9/9/9 x 1",
		"9/9/9/7o1/4ox3/8o/9/4x4/x8 o 4",
	}

	// Array of LEGAL moves to make in each position
	moves := [][]PosType{
		{MakeMove(4, 4), MakeMove(4, 3), MakeMove(3, 8), MakeMove(8, 1), MakeMove(1, 4)},
		{MakeMove(1, 1), MakeMove(1, 8), MakeMove(8, 7), MakeMove(7, 0)},
		{MakeMove(4, 0), MakeMove(0, 2), MakeMove(2, 1), MakeMove(1, 5)},
	}

	pos := NewPosition()

	// Init position (meaning we didn't call the .FromNotation)
	t.Run("Init position", func(t *testing.T) {
		helperErr := HelperTestReversingMoves(pos, StartingPosition, moves[0])
		if helperErr != nil {
			t.Error(helperErr)
		}
	})

	// Go through all other moves, also starting position, but
	// initialized with .FromNotation
	for i, strpos := range notations {
		t.Run(fmt.Sprintf("%d::%s", i, strpos), func(test *testing.T) {

			err := pos.FromNotation(strpos)

			if err != nil {
				t.Error(err)
			}

			helperErr := HelperTestReversingMoves(pos, strpos, moves[i])

			if helperErr != nil {
				t.Error(helperErr)
			}
		})
	}
}

func TestPerft(t *testing.T) {
	// First 3 proved mathematically, next values looks good as well
	valid_nodes := []uint64{
		81, 720, 6336, 55080, 473256,
	}

	pos := NewPosition()
	for i, n := range valid_nodes {
		nodes := Perft(pos, i+1)

		// Check if the values match
		if nodes != n {
			t.Errorf(
				"Invalid Perft nodes (depth=%d), got=%d, want=%d",
				i+1, nodes, n,
			)
		}
	}
}
