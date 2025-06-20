package bttt

import (
	"fmt"
	"time"
)

func Perft(pos *Position, depth int, valid bool, print bool) uint64 {
	now := time.Now()
	nodes := uint64(0)

	// Measure the time execution with defer
	defer func() {
		if print {
			fmt.Printf("Nodes: %d (%.1f Mnps)\n", nodes, float64(nodes)/float64(time.Since(now).Microseconds()))
		}
	}()

	// Call the actual perft function and return the number of nodes
	if valid {
		if !pos.IsTerminated() {
			nodes = _ValidPerft(pos, depth)
		}
	} else {
		nodes = _Perft(pos, depth)
	}

	return nodes
}

// Count total number of VALID positions up to given depth
func _ValidPerft(position *Position, depth int) uint64 {
	if depth == 0 {
		return 0
	}

	if depth == 1 {
		return uint64(position.GenerateMoves().size)
	}

	nodes := uint64(0)
	moves := position.GenerateMoves()

	for _, m := range moves.Slice() {
		position.MakeMove(m)

		if !position.IsTerminated() {
			nodes += _ValidPerft(position, depth-1)
		} else {
			// Reset the flag
			position.SetTermination(TerminationNone)
		}
		position.UndoMove()
	}

	return nodes
}

// Get total number of positions (nodes) at certain depth
func _Perft(position *Position, depth int) uint64 {
	// Default case
	if depth == 0 {
		return 0
	}

	// Optimized
	if depth == 1 {
		return uint64(position.GenerateMoves().size)
	}

	// Simply count number of 'children' nodes from this position, recursively
	nodes := uint64(0)
	moves := position.GenerateMoves()

	for _, m := range moves.Slice() {
		position.MakeMove(m)
		nodes += _Perft(position, depth-1)
		position.UndoMove()
	}

	return nodes
}
