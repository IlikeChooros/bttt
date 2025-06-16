package bttt

import "fmt"

func (e *Engine) IterativeDeepening() {

	// Declare variables
	pos := e.position
	alpha := Value(-MateValue)
	beta := Value(MateValue)
	score := Value(0)
	bestscore := -MateValue
	bestmove := posIllegal

	for d := Depth(0); d < e.limits.depth; d++ {
		moves := pos.GenerateMoves()

		for _, m := range moves.moves {
			pos.MakeMove(m)
			score = alphaBeta(pos, alpha, beta)
			pos.UndoMove()

			alpha = max(alpha, score)

			if score > bestscore {
				bestmove = m
			}

			if beta >= alpha {
				break
			}
		}
	}

	// Print the result
	fmt.Printf("bestmove %s", bestmove.String())
}

func alphaBeta(pos *Position, alpha, beta Value) Value {
	return MateValue
}
