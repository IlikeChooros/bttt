package uttt

import (
	"testing"
)

func TestHashUpdate(t *testing.T) {
	notations := []string{
		StartingPosition,
		"1x7/2o6/x8/xoxoxo3/9/9/9/9/oo7 x 3",
		"3x5/o8/9/xox1x4/o3o4/8x/9/4o4/9 o 4",
		"9/2o6/1xo1x2x1/9/2x4o1/9/9/2o4x1/9 o 7",
	}

	moves := [][]PosType{
		{MakeMove(B2, A3), MakeMove(A3, A3), MakeMove(A3, C1), MakeMove(C1, A3), MakeMove(A3, C2),
			MakeMove(C2, A3), MakeMove(A3, C3), MakeMove(C3, A3), MakeMove(A3, B2)},
		{MakeMove(A2, C1), MakeMove(C1, C3)},
		{MakeMove(B2, C1), MakeMove(C1, C1), MakeMove(C1, A2), MakeMove(A2, C1)},
		{MakeMove(B1, A1), MakeMove(A1, B2), MakeMove(B2, B2), MakeMove(B2, B3),
			MakeMove(B3, B2), MakeMove(B2, A3), MakeMove(A3, A1), MakeMove(A1, C3),
			MakeMove(C3, A1), MakeMove(A1, A1)},
	}

	pos := NewPosition()

	for i, notation := range notations {
		if err := pos.FromNotation(notation); err != nil {
			t.Error(err)
		} else {
			before := pos.Hash()

			// Go through the moves, and undo them
			for _, m := range moves[i] {
				pos.MakeMove(m)
			}

			for range moves[i] {
				pos.UndoMove()
			}

			// Hash should be the same
			after := pos.Hash()

			if before != after {
				t.Errorf("Hash mismatch: got=%d, want=%d", after, before)
			}
		}
	}
}
