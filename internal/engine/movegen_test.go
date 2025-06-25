package uttt

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
	// Check if move generation is working correctly -
	// and creates specified number of children nodes from the starting position

	// First 3 proved mathematically, next values looks good as well
	valid_nodes := []uint64{
		81, 720, 6336, 55080, 473256,
	}

	pos := NewPosition()
	for i, n := range valid_nodes {
		nodes := Perft(pos, i+1, false, true)

		// Check if the values match
		if nodes != n {
			t.Errorf(
				"Invalid Perft nodes (depth=%d), got=%d, want=%d",
				i+1, nodes, n,
			)
		}
	}
}

func BenchmarkMakeUndo(b *testing.B) {
	pos := NewPosition()
	m := MakeMove(B2, B2)

	for i := 0; i < b.N; i++ {
		pos.MakeMove(m)
		pos.UndoMove()
	}
}

func BenchmarkGenerateMoves(b *testing.B) {
	pos := NewPosition()

	notations := []string{
		StartingPosition,
		"1x7/2o6/x8/xoxoxo3/9/9/9/9/oo7 x 3",
		"3x5/o8/9/xox1x4/o3o4/8x/9/4o4/9 o 4",
		"9/2o6/1xo1x2x1/9/2x4o1/9/9/2o4x1/9 o 7",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// b.StopTimer()
		_ = pos.FromNotation(notations[i%len(notations)])
		// b.StartTimer()
		pos.GenerateMoves()
	}
}
