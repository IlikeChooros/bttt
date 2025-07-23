package uttt

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	_InitHashing()
	m.Run()
}

func TestBoolToInt(t *testing.T) {
	if _boolToInt(bool(CrossTurn)) != 1 {
		t.Errorf("CrossTurn =%d, want = 1", _boolToInt(bool(CrossTurn)))
	}
	if _boolToInt(bool(CircleTurn)) != 0 {
		t.Errorf("CircleTurn =%d, want = 0", _boolToInt(bool(CircleTurn)))
	}
}

func TestPositionInit(t *testing.T) {
	p := NewPosition()

	// Run sub test, simply check if there is no thistory, and
	// and Big Index is not set
	test := func(t *testing.T) {
		if p.stateList.ValidSize() != 0 {
			t.Error("p.stateList.ValidSize() != 0")
		}

		if p.BigIndex() != PosIndexIllegal {
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

func compare(mvpos, pos *Position) error {
	// Compare the states
	if mvpos.termination != pos.termination {
		return fmt.Errorf("Termination ineq got=%v, want=%v", mvpos.termination, pos.termination)
	}
	if mvpos.hash != pos.hash {
		return fmt.Errorf("Hash inequality got=%d, want=%d", mvpos.hash, pos.hash)
	}
	if !reflect.DeepEqual(mvpos.position, pos.position) {
		return fmt.Errorf("Position inequality got=%v, want=%v", mvpos.position, pos.position)
	}
	if !reflect.DeepEqual(mvpos.GenerateMoves().Slice(), pos.GenerateMoves().Slice()) {
		return fmt.Errorf("Legal moves ineq got=%v, want=%v", mvpos.GenerateMoves().Slice(), pos.GenerateMoves().Slice())
	}
	if !reflect.DeepEqual(mvpos.stateList.list, pos.stateList.list) {
		return fmt.Errorf("Statelists ineq got=%v, want=%v", mvpos.stateList.list, pos.stateList.list)
	}
	if !reflect.DeepEqual(mvpos.bigPositionState, pos.bigPositionState) {
		return fmt.Errorf("bigPositionStates ineq got=%v, want=%d", mvpos.bigPositionState, pos.bigPositionState)
	}
	if mvpos.Notation() != pos.Notation() {
		return fmt.Errorf("Notation ineq got=%s, want=%s", mvpos.Notation(), pos.Notation())
	}
	if mvpos.IsTerminated() != pos.IsTerminated() {
		return fmt.Errorf("Termination status ineq got=%t, want=%t", mvpos.IsTerminated(), pos.IsTerminated())
	}
	if !reflect.DeepEqual(mvpos.bitboards, pos.bitboards) {
		return fmt.Errorf("Bitboards ineq got=%v, want=%v", mvpos.bitboards, pos.bitboards)
	}
	if mvpos.BigIndex() != pos.BigIndex() {
		return fmt.Errorf("BigIndex ineq got=%d, want=%d", mvpos.BigIndex(), pos.BigIndex())
	}
	return nil
}

func TestMakeUndoStates(t *testing.T) {
	// Make sure that position state before making a move is the same
	// as making one and calling undo

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

	for i, txtpos := range notations {
		t.Run(fmt.Sprintf("SubTest-%s", strings.ReplaceAll(txtpos, "/", "|")), func(t *testing.T) {

			if pos, err := FromNotation(txtpos); err != nil {
				t.Error(err)
			} else {
				mvpos, _ := FromNotation(txtpos)

				// Make the moves
				for _, m := range moves[i] {
					if err := mvpos.MakeLegalMove(m); err != nil {
						t.Error(err)
						return
					}
				}

				// Undo the moves
				for range moves[i] {
					mvpos.UndoMove()
				}

				if err := compare(mvpos, pos); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestClone(t *testing.T) {
	notations := []string{
		StartingPosition,
		"1x7/2o6/x8/xoxoxo3/9/9/9/9/oo7 x 3",
		"3x5/o8/9/xox1x4/o3o4/8x/9/4o4/9 o 4",
		"9/1xx1o4/1ox1o4/9/1o2xx3/2o6/9/9/9 x 1",
	}

	for _, notation := range notations {
		t.Run(fmt.Sprintf("CloneTest-%s", strings.ReplaceAll(notation, "/", "|")), func(t *testing.T) {
			pos := NewPosition()
			if err := pos.FromNotation(notation); err != nil {
				t.Error(err)
				return
			}

			cloned := pos.Clone()

			if err := compare(pos, &cloned); err != nil {
				t.Error(err)
			}

			// Make sure that underlying memory is different
			// 1. Check the stateList pointer is different.
			if pos.stateList == cloned.stateList {
				t.Error("stateList pointers are identical - clone should allocate its own stateList")
			}

			// 2. Check that the underlying array of stateList.list is not shared.
			// The reflect package provides .Pointer() on slice headers.
			origListPtr := reflect.ValueOf(pos.stateList.list).Pointer()
			cloneListPtr := reflect.ValueOf(cloned.stateList.list).Pointer()
			if origListPtr == cloneListPtr {
				t.Error("stateList.list underlying array is shared - expected deep copy")
			}

			// 3. Optionally, check a couple of board cells (they are value types so already copied)
			// but here we verify that the address of the board is different.
			origBoardPtr := &pos.position[0][0]
			cloneBoardPtr := &cloned.position[0][0]
			// Since these are arrays, they will be stored in different positions if cloned properly.
			if origBoardPtr == cloneBoardPtr {
				t.Error("position board appears to be shared between clones")
			}
		})
	}
}
