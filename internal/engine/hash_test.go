package uttt

import (
	"fmt"
	"strings"
	"testing"
)

func TestHashUpdate(t *testing.T) {
	notations := []string{
		StartingPosition,
		"1x7/2o6/x8/xoxoxo3/9/9/9/9/oo7 x 3",
		"3x5/o8/9/xox1x4/o3o4/8x/9/4o4/9 o 4",
		"9/1xx1o4/1ox1o4/9/1o2xx3/2o6/9/9/9 x 1",
	}

	moves := [][]PosType{
		{MakeMove(B2, A3), MakeMove(A3, A3), MakeMove(A3, C1), MakeMove(C1, A3), MakeMove(A3, C2),
			MakeMove(C2, A3), MakeMove(A3, C3), MakeMove(C3, A3), MakeMove(B2, B2)},
		{MakeMove(A2, C1), MakeMove(C1, C3)},
		{MakeMove(B2, C1), MakeMove(C1, C1), MakeMove(C1, A2), MakeMove(A2, C1)},
		{MakeMove(B3, A3), MakeMove(A3, B2), MakeMove(B2, A2), MakeMove(A2, B2),
			MakeMove(B1, B2), MakeMove(C3, B1), MakeMove(B1, B3), MakeMove(C2, B2), MakeMove(B1, B1)},
	}

	pos := NewPosition()

	for i, notation := range notations {
		t.Run(fmt.Sprintf("SubTest-%d-%s", i, strings.ReplaceAll(notation, "/", "|")), func(t *testing.T) {
			if err := pos.FromNotation(notation); err != nil {
				t.Error(err)
			} else {
				before := pos.hash

				// Go through the moves, and undo them
				for _, m := range moves[i] {
					if err := pos.MakeLegalMove(m); err != nil {
						t.Error(err)
						return
					}
				}

				for range moves[i] {
					pos.UndoMove()
				}

				// Hash should be the same
				after := pos.hash

				if before != after {
					t.Errorf("%s Hash mismatch: got=%d, want=%d", notation, after, before)
				}
			}
		})
	}
}
