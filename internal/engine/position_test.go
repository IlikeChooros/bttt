package uttt

import (
	"reflect"
	"testing"
)

func TestPositionInit(t *testing.T) {
	p := NewPosition()

	// Run sub test, simply check if there is no thistory, and
	// and Big Index is not set
	test := func(t *testing.T) {
		if p.stateList.ValidSize() != 0 {
			t.Error("p.stateList.ValidSize() != 0")
		}

		if p.BigIndex() != int(PosIndexIllegal) {
			t.Errorf("p.BigIndex()=%d, want=%d", p.BigIndex(), PosIndexIllegal)
		}
	}

	t.Run("TestNewPosition", test)

	err := p.FromNotation(StartingPosition)
	if err != nil {
		t.Error(err)
	}

	t.Run("TestFromStartPos", test)
}

func TestMakeUndoStates(t *testing.T) {
	// Make sure that position state before making a move is the same
	// as making one and calling undo

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

	for i, txtpos := range notations {
		pos := NewPosition()
		mvpos := NewPosition()

		if err := pos.FromNotation(txtpos); err != nil {
			t.Error(err)
		} else {
			_ = mvpos.FromNotation(txtpos)

			// Make the moves
			for j, m := range moves[i] {
				if !mvpos.IsLegal(m) {
					t.Errorf("Move %s (at %d index, for pos index %d) is illegal", m.String(), j, i)
					break
				}
				mvpos.MakeMove(m)
			}

			// Undo the moves
			for range moves[i] {
				mvpos.UndoMove()
			}

			// Compare the states
			if mvpos.termination != pos.termination {
				t.Errorf("Termination ineq got=%v, want=%v", mvpos.termination, pos.termination)
			}
			if mvpos.hash != pos.hash {
				t.Errorf("Hash inequality got=%d, want=%d", mvpos.hash, pos.hash)
			}
			if !reflect.DeepEqual(mvpos.position, pos.position) {
				t.Errorf("Position inequality got=%v, want=%v", mvpos.position, pos.position)
			}
			if !reflect.DeepEqual(mvpos.GenerateMoves().Slice(), pos.GenerateMoves().Slice()) {
				t.Errorf("Legal moves ineq got=%v, want=%v", mvpos.GenerateMoves().Slice(), pos.GenerateMoves().Slice())
			}
			if !reflect.DeepEqual(mvpos.stateList.list, pos.stateList.list) {
				t.Errorf("Statelists ineq got=%v, want=%v", mvpos.stateList.list, pos.stateList.list)
			}
			if !reflect.DeepEqual(mvpos.bigPositionState, pos.bigPositionState) {
				t.Errorf("bigPositionStates ineq got=%v, want=%d", mvpos.bigPositionState, pos.bigPositionState)
			}
			if mvpos.Notation() != pos.Notation() {
				t.Errorf("Notation ineq got=%s, want=%s", mvpos.Notation(), pos.Notation())
			}
			if mvpos.IsTerminated() != pos.IsTerminated() {
				t.Errorf("Termination status ineq got=%t, want=%t", mvpos.IsTerminated(), pos.IsTerminated())
			}
		}
	}
}
