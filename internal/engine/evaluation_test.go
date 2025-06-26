package uttt

import (
	"testing"
)

func TestEvalInitPosition(t *testing.T) {
	// Check if the initial position's evaluation is 0
	pos := NewPosition()

	if eval := Evaluate(pos); eval != 0 {
		t.Errorf("Evaluate(init)=%d, want=0", eval)
	}
}

func TestSquareEvaluation(t *testing.T) {
	O, X, n := PieceCircle, PieceCross, PieceNone

	// Test if the _evaluateSquare function is working as intented
	squares := [][9]PieceType{
		{
			O, X, n,
			X, O, n,
			n, O, n,
		},
		{
			O, X, X,
			X, X, n,
			O, O, n,
		},
		{
			O, n, X,
			n, n, n,
			O, n, X,
		},
		{
			O, X, X,
			X, X, n,
			O, O, n,
		},
	}

	const (
		Draw    = 0
		Winning = 1
		Losing  = 2
	)

	// Setup states corresponding to each position
	states := []int{
		Winning, Winning, Draw, Losing,
	}

	pieces := []PieceType{
		PieceCircle, PieceCross, PieceCircle, PieceCircle,
	}

	for i, square := range squares {
		eval := evaluateSquare(toBitboards(square, pieces[i]))

		// Setup actual relative state of the square
		state := Draw
		if eval > 0 {
			state = Winning
		} else if eval < 0 {
			state = Losing
		}

		// States should match
		if state != states[i] {
			t.Errorf("evaluateSquare(%v)=%d, got state=%d, want=%d",
				square, eval, state, states[i])
		} else {
			t.Logf("evaluateSquare(%d)=%d", i, eval)
		}
	}
}

func TestEvalAdvantageousPositions(t *testing.T) {
	positions := []string{
		// O - X (big square position)
		// X O -
		// - O -
		"x1o1oxox1/o7x/ox2xo1x1/xoox2xo1/x1o1oxox1/4ox3/xxo6/x1o1oxox1/o3oxx2 o 8",
	}

	pos := NewPosition()
	for _, position := range positions {
		err := pos.FromNotation(position)
		if err != nil {
			t.Error(err)
			continue
		}

		if eval := Evaluate(pos); eval <= 0 {
			t.Errorf("Evaluate(%s)=%d, want > 0", position, eval)
		} else {
			t.Logf("Evaluate(%s)=%d", position, eval)
		}
	}
}

func TestEvalSymmetry(t *testing.T) {
	// Make sure that evaluation function is symmetric, meaning
	// evaluation if evalution from cross's perspective = -evaluation from circle's perspective
	notations := []string{
		StartingPosition,
		"1x7/2o6/x8/xoxoxo3/9/9/9/9/oo7 x 3",
		"3x5/o8/9/xox1x4/o3o4/8x/9/4o4/9 o 4",
		"9/2o6/1xo1x2x1/9/2x4o1/9/9/2o4x1/9 o 7",
	}

	pos := NewPosition()
	for _, notation := range notations {
		if err := pos.FromNotation(notation); err != nil {
			t.Error(err)
		} else {
			crossEval, circleEval := 0, 0

			if pos.Turn() == CrossTurn {
				crossEval = Evaluate(pos)
				pos.stateList.Last().turn = CrossTurn
				circleEval = Evaluate(pos)
			} else {
				circleEval = Evaluate(pos)
				pos.stateList.Last().turn = CircleTurn
				crossEval = Evaluate(pos)
			}

			if crossEval != -circleEval {
				t.Errorf("%s, eval mismatch X=%d, O=%d", notation, crossEval, circleEval)
			}
		}
	}
}
