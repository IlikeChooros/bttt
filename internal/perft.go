package bttt

import (
	"fmt"
	"time"
)

func Perft(pos *Position, depth int) uint64 {
	now := time.Now()
	nodes := uint64(0)

	// Measure the time execution with defer
	defer func() {
		fmt.Printf("Nodes: %d (depth=%d, time=%v)\n", nodes, depth, time.Since(now))
	}()

	// Call the actual perft function and return the number of nodes
	nodes = _Perft(pos, depth)
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
