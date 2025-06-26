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

var _transpTable = _NewHashTable[HashEntry](1 << 21)

func (e *Engine) _printMsg(msg string) {
	if e.print {
		fmt.Print(msg)
	}
}

// Get the principal variation from the transpostion table
func (e *Engine) _LoadPv(rootmove PosType) {
	e.pv.Clear()
	depth := 0

	e.position.MakeMove(rootmove)

	// Go through the transposition table
	val, ok := _transpTable.Get(e.position.hash)
	for ; ok && depth < MaxDepth && val.Bestmove != PosIllegal; depth++ {
		e.pv.AppendMove(val.Bestmove)
		e.position.MakeMove(val.Bestmove)
		val, ok = _transpTable.Get(e.position.hash)
	}

	// Undo the moves
	for range depth {
		e.position.UndoMove()
	}

	e.position.UndoMove()
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
		e._LoadPv(e.result.Bestmove)
		deltatime := max(time.Since(e.timer.Start()).Milliseconds(), 1)
		e._printMsg(
			fmt.Sprintf("info depth %d score %s nps %d nodes %d time %dms pv %s\n",
				d+1, e.result.String(), // depth, score
				(e.result.Nodes*1000)/uint64(deltatime), // nps
				e.result.Nodes, deltatime, // nodes, time
				e.pv.String(), // pv
			))
	}

	// Print the result
	e._printMsg(fmt.Sprintf("bestmove %s\n", e.result.Bestmove.String()))
}

func (e *Engine) _NegaAlphaBeta(depth, ply, alpha, beta int) int {

	e.result.Nodes++

	// Check if we calculated value of this node already, with requirement
	// of bigger or equal to depth of our current node's depth

	oldAlpha := alpha
	hash := e.position.hash
	if val, ok := _transpTable.Get(hash); ok && val.Depth >= depth {
		// Use the cached value
		if val.NodeType == Exact {
			return val.Score
		} else if val.NodeType == LowerBound {
			alpha = max(alpha, val.Score)
		} else {
			beta = min(beta, val.Score)
		}

		if alpha >= beta {
			return val.Score
		}
	}

	pos := e.position
	bestvalue := MateValue - ply
	value := 0
	bestmove := PosIllegal

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
	newEntry.Bestmove = bestmove
	newEntry.Depth = depth
	newEntry.Hash = hash
	newEntry.Score = bestvalue

	if bestvalue >= beta {
		// Beta cutoff
		newEntry.NodeType = UpperBound
	}
	if bestvalue <= oldAlpha {
		// Lowerbound value
		newEntry.NodeType = LowerBound
	} else {
		newEntry.NodeType = Exact
	}

	_transpTable.Set(hash, newEntry)

	return bestvalue
}
