package uttt

import (
	"cmp"
	"slices"
)

// Move ordering struct
type _OrderedMove struct {
	move  PosType
	value int
}

func _NewOrderedMove(move, pvmove PosType, turn, ply int) _OrderedMove {
	ordmv := _OrderedMove{
		move:  move,
		value: 0,
	}

	// const (
	// 	Multiplier = 10000
	// 	PVBias     = 8 * Multiplier
	// )

	bi, si := move.BigIndex(), move.SmallIndex()

	// ordmv.value += _pieceSquareTable[si] * 3
	// ordmv.value += int(_bigSquareTableFactors[bi] * 10)
	ordmv.value += _historyHeuristic[turn][bi][si]

	return ordmv
}

// Compare the moves and return -1, 0, or 1 if
func _CmpOrderedMoves(lhs, rhs _OrderedMove) int {
	// Descending ordering
	return -cmp.Compare(lhs.value, rhs.value)
}

func MoveOrdering(ml *MoveList, pos *Position, pvmove PosType, ply int) {
	ordered := make([]_OrderedMove, ml.size)
	turn := _boolToInt(bool(pos.Turn()))

	for i, m := range ml.Slice() {
		ordered[i] = _NewOrderedMove(m, pvmove, turn, ply)
	}

	// Sort the moves
	slices.SortStableFunc(ordered, _CmpOrderedMoves)

	// Reorder the movelist
	for i := range ml.size {
		ml.moves[i] = ordered[i].move
	}
}
