package uttt

import (
	"fmt"
	"slices"
	"time"
)

// Monte carlo tree search

func AsyncSearch(mcts *UtttMCTS) {
	go mcts.Search()

	// Periodically print messages
	var nps, nodes uint32
	var depth int
	for mcts.IsThinking() {
		Nps := mcts.Nps()
		n := mcts.Nodes()
		d := mcts.MaxDepth()

		if d != depth || nps != Nps || nodes != n {
			nps, nodes, depth = Nps, n, d
			fmt.Printf("\rinfo depth %d nps %d nodes %d", depth, nps, nodes)
		}
	}

	fmt.Printf("\rinfo depth %d nps %d nodes %d pv %s\n", depth, nps, nodes, mcts.GetPv().String())
}

// Mate value for the search
const (
	MaxDepth          = 64
	MateValue         = -1000000
	MinValue          = MateValue - MaxDepth
	MateTresholdValue = -MateValue
)

func (e *Engine) _PrintMsg(msg string) {
	if e.print {
		fmt.Fprint(e.writer, msg)
	}
}

// func (e *Engine) _GetPvMove(ply int) PosType {
// 	if ply >= int(e.pv.size) {
// 		return PosIllegal
// 	}
// 	return e.pv.moves[ply]
// }

// Get the principal variation from the transpostion table
func (e *Engine) _LoadPv(rootmove PosType, maxdepth int) {
	e.pv.Clear()
	depth := 0

	e.position.MakeMove(rootmove)
	e.pv.AppendMove(rootmove)

	// Go through the transposition table
	val, ok := e.ttable.Get(e.position.hash)
	for ; ok && depth < maxdepth && val.Bestmove != PosIllegal; depth++ {
		// Generate legal moves and see if that's a valid move
		if !slices.Contains(e.position.GenerateMoves().Slice(), val.Bestmove) {
			break
		}

		e.pv.AppendMove(val.Bestmove)
		e.position.MakeMove(val.Bestmove)
		val, ok = e.ttable.Get(e.position.hash)
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
	bestscore := MinValue
	e.stop.Store(false)

	moves := pos.GenerateMoves().Slice()

	// Don't start the search in a terminated position
	if pos.IsTerminated() {
		e._PrintMsg(fmt.Sprintf("terminated %v\nbestmove (none)\n", pos.termination))
		return
	}

	e.timer.Reset()
	for d := 0; !e.stop.Load() && (d < MaxDepth && (e.limits.infinite || d < e.limits.depth)); d++ {
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

		// Reset
		bestscore = MinValue

		// Get the number of milliseconds since the start
		e._LoadPv(e.result.Bestmove, d+1)
		deltatime := max(time.Since(e.timer.Start()).Milliseconds(), 1)
		e.result.Nps = (e.result.Nodes * 1000) / uint64(deltatime)
		e.result.Depth = d + 1
		e._PrintMsg(
			fmt.Sprintf("info depth %d score %s nps %d nodes %d time %dms pv %s\n",
				d+1, e.result.String(), // depth, score
				e.result.Nps, e.result.Nodes, deltatime, // nps, nodes, time
				e.pv.String(), // pv
			))
	}

	// Print the result
	e._PrintMsg(fmt.Sprintf("bestmove %s\n", e.result.Bestmove.String()))
}

func (e *Engine) _NegaAlphaBeta(depth, ply, alpha, beta int) int {

	e.result.Nodes++

	// Check if we calculated value of this node already, with requirement
	// of bigger or equal to depth of our current node's depth

	oldAlpha := alpha
	hash := e.position.hash
	if val, ok := e.ttable.Get(hash); ok && val.Depth >= depth {
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
	moves := pos.GenerateMoves()
	// MoveOrdering(moves, e.position, e._GetPvMove(ply), ply)

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
	newEntry := TTEntry{}
	newEntry.Bestmove = bestmove
	newEntry.Depth = depth
	newEntry.Hash = hash
	newEntry.Score = bestvalue

	if bestvalue >= beta {
		// Beta cutoff
		newEntry.NodeType = UpperBound
		// _UpdateHistory(_boolToInt(bool(e.position.Turn())), bestmove, depth*depth)
	}
	if bestvalue <= oldAlpha {
		// Lowerbound value
		newEntry.NodeType = LowerBound
	} else {
		newEntry.NodeType = Exact
	}

	e.ttable.Set(hash, newEntry)

	return bestvalue
}
