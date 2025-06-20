package bttt

import (
	"testing"
)

func TestMakeMove(t *testing.T) {
	// Example: bigIndex = 2, smallIndex = 3
	move := MakeMove(2, 3)
	expectedSmall := PosType(3 & _moveSmallIndexMask)
	expectedBig := PosType((2 << 4) & _moveBigIndexMask)
	expected := expectedBig | expectedSmall

	if move != expected {
		t.Errorf("MakeMove(2, 3) = %v; want %v", move, expected)
	}
}

func TestBigIndex(t *testing.T) {
	move := MakeMove(5, 7)
	big := move.BigIndex()
	expectedBig := PosType(5)

	if big != expectedBig {
		t.Errorf("BigIndex() = %v; want %v", big, expectedBig)
	}
}

func TestSmallIndex(t *testing.T) {
	move := MakeMove(5, 7)
	small := move.SmallIndex()
	expectedSmall := PosType(7 & _moveSmallIndexMask)

	if small != expectedSmall {
		t.Errorf("SmallIndex() = %v; want %v", small, expectedSmall)
	}
}

func TestAppend(t *testing.T) {
	mv := NewMoveList()

	if mv.size != 0 {
		t.Errorf("NewMoveList size = %v; want 0", mv.size)
	}

	mv.Append(1, 4)
	if mv.size != 1 {
		t.Errorf("After Append, size = %v; want 1", mv.size)
	}

	expectedMove := MakeMove(1, 4)
	if mv.moves[0] != expectedMove {
		t.Errorf("First move = %v; want %v", mv.moves[0], expectedMove)
	}
}
