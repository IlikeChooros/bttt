package uttt

import (
	"fmt"
	"time"
)

// Mate value for the search
const (
	MateValue         = -1000000
	MaxDepth          = 64
	MateTresholdValue = -MateValue
)

// var _transpTable = _NewHashTable[HashEntry](1 << 20)
func (e *Engine) _printMsg(msg string) {
	if e.print {
		fmt.Print(msg)
	}
}

func (e *Engine) _IterativeDeepening() {

	// Declare variables
	e.result.Nodes = 0
	pos := e.position
	alpha := MateValue
	beta := -MateValue
	score := 0
	bestscore := MateValue
	e.stop.Store(false)

	moves := pos.GenerateMoves().Slice()

	// Don't start the search in a terminated position
	if pos.IsTerminated() {
		e._printMsg(fmt.Sprintf("terminated %v\nbestmove (none)\n", pos.termination))
		return
	}

	e.timer.Reset()
	for d := 0; !e.stop.Load() && (e.limits.infinite || d < e.limits.depth); d++ {

		for _, m := range moves {
			pos.MakeMove(m)
			score = -e._NegaAlphaBeta(d, 1, -beta, -alpha)
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
				e.result.SetValue(score, e.position.Turn())
				e.result.Bestmove = m
				bestscore = score
				alpha = max(alpha, score)

				// That's a mate, go back
				if e.result.ScoreType == MateScore {
					e.Stop()
					break
				}
			}

			if alpha >= beta {
				break
			}
		}

		// Get the number of milliseconds since the start
		deltatime := max(time.Since(e.timer.Start()).Milliseconds(), 1)
		e._printMsg(
			fmt.Sprintf("info depth %d score %s nps %d nodes %d time %dms\n",
				d+1, e.result.String(), // depth, score
				(e.result.Nodes*1000)/uint64(deltatime), // nps
				e.result.Nodes, deltatime), // nodes, time
		)
	}

	// Print the result
	e._printMsg(fmt.Sprintf("bestmove %s\n", e.result.Bestmove.String()))
}

func (e *Engine) _NegaAlphaBeta(depth, ply, alpha, beta int) int {

	e.result.Nodes++

	// Check if we calculated value of this node already, with requirement
	// of bigger or equal to depth of our current node's depth

	// oldAlpha := alpha
	// hash := e.position.Hash()
	// if val, ok := _transpTable.Get(hash); ok && val.depth >= depth {
	// 	// Use the cached value
	// 	if val.nodeType == Exact {
	// 		return val.score
	// 	} else if val.nodeType == LowerBound {
	// 		alpha = max(alpha, val.score)
	// 	} else {
	// 		beta = min(beta, val.score)
	// 	}

	// 	if alpha >= beta {
	// 		return val.score
	// 	}
	// }

	pos := e.position
	bestvalue := MateValue - ply
	value := 0
	// bestmove := PosIllegal

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
	moves := pos.GenerateMoves().Slice()
	for _, m := range moves {

		pos.MakeMove(m)
		value = -e._NegaAlphaBeta(depth-1, ply+1, -beta, -alpha)
		pos.UndoMove()

		if value > bestvalue {
			// bestmove = m
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
	// newEntry := HashEntry{}
	// newEntry.bestmove = bestmove
	// newEntry.depth = depth
	// newEntry.hash = hash
	// newEntry.score = bestvalue

	// if bestvalue >= beta {
	// 	// Beta cutoff
	// 	newEntry.nodeType = UpperBound
	// }
	// if bestvalue <= oldAlpha {
	// 	// Lowerbound value
	// 	newEntry.nodeType = LowerBound
	// } else {
	// 	newEntry.nodeType = Exact
	// }

	// _transpTable.Set(hash, newEntry)

	return bestvalue
}
