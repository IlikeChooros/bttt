package bttt

import (
	"fmt"
	"time"
)

var _transpTable = _NewHashTable[HashEntry]()

func (e *Engine) _IterativeDeepening() {
	// Declare variables
	e.result.Nodes = 0
	pos := e.position
	alpha := MateValue
	beta := -MateValue
	score := 0
	bestscore := MateValue
	bestmove := posIllegal
	e.timer.Reset()
	e.stop.Store(false)

	// Don't start the search in a terminated position
	if pos.IsTerminated() {
		fmt.Println("terminated\nbestmove (none)")
		return
	}

	fmt.Println(pos.GenerateMoves().Slice())

	for d := 0; !e.stop.Load() && (e.limits.infinite || d < e.limits.depth); d++ {
		moves := pos.GenerateMoves()

		for _, m := range moves.Slice() {
			pos.MakeMove(m)
			score = e._NegaAlphaBeta(d, 0, alpha, beta)
			pos.UndoMove()

			// Now check if the timer has ended, if so this move wasn't fully searched,
			// so discard it's value
			if !e.limits.infinite && e.timer.IsEnd() {
				e.Stop()
			}

			// Check if we should stop searching the moves
			if e.stop.Load() {
				break
			}

			if score > bestscore {
				bestmove = m
				alpha = max(alpha, score)
			}

			if alpha >= beta {
				break
			}
		}

		deltatime := time.Since(e.timer.Start())
		fmt.Printf("info depth %d score %d nps %d nodes %d time %dms\n",
			d+1, score, // depth, score
			(e.result.Nodes*1000)/uint64(deltatime.Milliseconds()+1), // nps
			e.result.Nodes, deltatime.Milliseconds()) // time
	}

	// Print the result
	fmt.Printf("bestmove %s\n", bestmove.String())

	// Set the results
	e.result.Bestmove = bestmove
	e.result.Value = bestscore
}

func (e *Engine) _NegaAlphaBeta(depth, ply, alpha, beta int) int {

	e.result.Nodes++

	// Check if we calculated value of this node already, with requirement
	// of bigger or equal to depth of our current node's depth
	// strpos := e.position.Notation()
	oldAlpha := alpha
	if val, ok := _transpTable.Get(0); ok && val.depth >= depth {
		// Use the cached value
		if val.nodeType == Exact {
			return val.score
		} else if val.nodeType == LowerBound {
			alpha = max(alpha, val.score)
		} else {
			beta = min(beta, val.score)
		}

		if alpha >= beta {
			return val.score
		}
	}

	pos := e.position
	bestvalue := MateValue + depth
	value := 0
	bestmove := posIllegal

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
		value = -e._NegaAlphaBeta(depth-1, ply+1, -beta, -alpha)
		pos.UndoMove()

		if value > bestvalue {
			bestmove = m
			bestvalue = value
			alpha = max(alpha, value)
		}

		// Check the timer
		if !e.limits.infinite && e.timer.IsEnd() {
			return bestvalue
		}

		if alpha >= beta {
			break
		}
	}

	// Set the hash entry value
	newEntry := HashEntry{}
	newEntry.bestmove = bestmove
	newEntry.depth = depth
	// newEntry.hash = strpos

	if bestvalue >= beta {
		// Beta cutoff
		newEntry.nodeType = UpperBound
	}
	if bestvalue <= oldAlpha {
		// Lowerbound value
		newEntry.nodeType = LowerBound
	} else {
		newEntry.nodeType = Exact
	}

	return bestvalue
}
