package bttt

import (
	"fmt"
	"time"
)

func (e *Engine) _IterativeDeepening() {

	// Declare variables
	e.result.Nodes = 0
	pos := e.position
	alpha := MateValue
	beta := -MateValue
	score := 0
	bestscore := MateValue
	bestmove := posIllegal
	beginTime := time.Now()

	if pos.IsTerminated() {
		fmt.Println("bestmove (none)")
		return
	}

	for d := 0; d < e.limits.depth; d++ {
		moves := pos.GenerateMoves()

		for _, m := range moves.moves {
			pos.MakeMove(m)
			score = e.negaAlphaBeta(d, 0, alpha, beta)
			pos.UndoMove()

			alpha = max(alpha, score)

			if score > bestscore {
				bestmove = m
			}

			if alpha >= beta {
				break
			}
		}

		fmt.Printf("info depth %d score %d nodes %d time %s\n", d, score, e.result.Nodes, time.Since(beginTime).String())
	}

	// Print the result
	fmt.Printf("bestmove %s\n", bestmove.String())

	// Set the results
	e.result.Bestmove = bestmove
	e.result.Value = bestscore
}

func (e *Engine) negaAlphaBeta(depth, ply, alpha, beta int) int {

	e.result.Nodes++
	pos := e.position
	bestvalue := MateValue + depth
	value := 0
	// bestmove := posIllegal

	// Check if that's terminated node, if so return according value
	if pos.IsTerminated() {
		if pos.termination == TerminationDraw {
			bestvalue = 0 // Draw value is 0
		}

		return bestvalue
	}

	// If we reach the terminating depth, return the static evaluation of the position
	if depth <= 0 {
		return Evaluate(pos)
	}

	// Go through the moves
	moves := pos.GenerateMoves()
	for _, m := range moves.Slice() {
		pos.MakeMove(m)
		value = -e.negaAlphaBeta(depth-1, ply+1, -beta, -alpha)
		pos.UndoMove()

		if value > bestvalue {
			// bestmove = m
			bestvalue = value
			alpha = max(alpha, value)
		}

		if alpha >= beta {
			break
		}
	}

	return bestvalue
}
