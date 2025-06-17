package bttt

import (
	"testing"
)

// Check if the initial position notation is valid
func TestInitNotation(t *testing.T) {
	// Simply test the initial notation (meaning no pieces on the board)
	init_notation := "9/9/9/9/9/9/9/9/9 o -"

	pos := NewPosition().Notation()

	if pos != init_notation {
		t.Errorf("Init notation %s, want: %s", pos, init_notation)
	}
}

func TestNotations(t *testing.T) {
	// Check if after setting given position, I will
	// get the same notation as the input one

	notations := []string{
		StartingPosition,
		"1o7/2x6/o8/9/9/9/9/9/9 x -",
		"9/9/9/7o1/4ox3/8o/9/4x4/x8 o -",
	}

	pos := NewPosition()
	for _, str := range notations {
		err := pos.FromNotation(str)

		if err != nil {
			t.Error(err)
		}

		if pos.Notation() != str {
			t.Errorf("Position was not set, got=%s, want=%s",
				pos.Notation(), str)
		}

		isCircleTurn := str[len(str)-3] == 'o'
		if pos.Turn() != TurnType(isCircleTurn) {
			t.Errorf("Invalid turn conversion, got=%t, want=%t ",
				pos.Turn(), isCircleTurn)
		}
	}
}

func TestInitPositionNotations(t *testing.T) {
	// Simply check if after making specified moves, from starting position
	// I will get given notation strings

	// In future I may add more of these
	notations := []string{
		"1o7/2x6/o8/9/9/9/9/9/9 x 0",
		"9/9/9/7o1/4ox3/8o/9/4x4/x8 o 0",
	}

	moves := [][]PosType{
		{MakeMove(0, 1), MakeMove(1, 2), MakeMove(2, 0)},
		{MakeMove(3, 7), MakeMove(7, 4), MakeMove(4, 4), MakeMove(4, 5), MakeMove(5, 8), MakeMove(8, 0)},
	}

	pos := NewPosition()

	for i, notation := range notations {

		err := pos.FromNotation("startpos")
		if err != nil {
			t.Error(err)
		}

		// Play the moves
		for _, m := range moves[i] {
			pos.MakeMove(m)
		}

		if pos.Notation() != notation {
			t.Errorf("Pos notation = %s, want %s", pos.Notation(), notation)
		}
	}
}
