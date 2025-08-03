package uttt

import (
	"math/rand"
	"time"
	"uttt/internal/mcts"
)

// Actual UTTT mcts implementation
type UtttMCTS struct {
	mcts.MCTS[PosType]
	ops *UtttOperations
}

type UtttNode mcts.NodeBase[PosType]

func NewUtttMCTS(position Position) *UtttMCTS {
	uttt_ops := newUtttOps(position)
	ops := mcts.GameOperations[PosType](uttt_ops)
	tree := &UtttMCTS{
		MCTS: *mcts.NewMTCS(
			mcts.UCB1,
			ops,
			mcts.TerminalFlag(position.IsTerminated()),
		),
		ops: uttt_ops,
	}
	return tree
}

func (mcts *UtttMCTS) AsyncSearch() {
	mcts.MCTS.SearchMultiThreaded(mcts.ops)
}

// Start the search
func (mcts *UtttMCTS) Search() {

	// Run the search
	mcts.SearchMultiThreaded(mcts.ops)

	// Wait for the search to end
	mcts.Synchronize()
}

// Default selection
func (mcts *UtttMCTS) Selection() *mcts.NodeBase[PosType] {
	return mcts.MCTS.Selection(mcts.ops, rand.New(rand.NewSource(time.Now().UnixNano())), 0)
}

// Default backprop
func (mcts *UtttMCTS) Backpropagate(node *mcts.NodeBase[PosType], result mcts.Result) {
	mcts.MCTS.Backpropagate(mcts.ops, node, result)
}

func (mcts *UtttMCTS) Ops() mcts.GameOperations[PosType] {
	return mcts.ops
}

func (mcts *UtttMCTS) Reset() {
	mcts.MCTS.Reset(mcts.ops, mcts.ops.position.IsTerminated())
}

// Set the position
func (mcts *UtttMCTS) SetPosition(position Position) {
	mcts.ops.position = position
	mcts.Reset()
}

func (mcts *UtttMCTS) SetNotation(notation string) error {
	defer mcts.Reset()
	return mcts.ops.position.FromNotation(notation)
}

func ToSearchResult(stats mcts.ListenerTreeStats[PosType]) SearchResult {
	result := SearchResult{
		Bestmove: stats.BestMove,
		Nodes:    0,
		Nps:      stats.Nps,
		Depth:    stats.Maxdepth,
		Cycles:   int32(stats.Cycles),
		Pv:       stats.Pv,
	}

	// Set the score
	if stats.Terminal {
		if stats.Draw {
			result.ScoreType = ValueScore
			result.Value = 50
		} else {
			result.ScoreType = MateScore
			result.Value = len(stats.Pv)

			// If the game ends on our turn, we are losing
			if result.Value%2 == 0 {
				result.Value = -result.Value
			}
		}
	} else {
		result.ScoreType = ValueScore
		result.Value = int(100 * stats.Eval)
	}

	return result
}

func (mcts *UtttMCTS) SearchResult(pvPolicy mcts.BestChildPolicy) SearchResult {
	pv, terminal, draw := mcts.Pv(pvPolicy)
	turn := 1
	result := SearchResult{
		Bestmove: mcts.RootSignature(),
		Nodes:    uint64(mcts.Nodes()),
		Nps:      uint64(mcts.Nps()),
		Depth:    mcts.MaxDepth(),
		Cycles:   mcts.Root.Visits(),
		Pv:       pv,
	}

	if mcts.ops.position.Turn() == CircleTurn {
		turn = -1
	}

	// Set the score
	if terminal {
		if draw {
			result.ScoreType = ValueScore
			result.Value = 50
		} else {
			result.ScoreType = MateScore
			result.Value = len(pv) * turn

			// If the game ends on our turn, we are losing
			if len(pv)%2 == 0 {
				result.Value = -result.Value
			}
		}
	} else {
		visits := float64(mcts.Root.Visits())
		wins := float64(mcts.Root.Outcomes())
		if mcts.Root != nil && visits > 0 {
			result.Value = int(100 * wins / visits)
		} else {
			result.Value = -100
		}
	}

	return result
}

type UtttOperations struct {
	position Position
	rootSide TurnType
	random   *rand.Rand
}

func newUtttOps(pos Position) *UtttOperations {
	return &UtttOperations{
		position: pos,
		rootSide: pos.Turn(),
		random:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (ops *UtttOperations) Reset() {
	ops.rootSide = ops.position.Turn()
}

func (ops *UtttOperations) ExpandNode(node *mcts.NodeBase[PosType]) uint32 {

	moves := ops.position.GenerateMoves()
	node.Children = make([]mcts.NodeBase[PosType], moves.size)

	for i, m := range moves.Slice() {
		ops.position.MakeMove(m)
		isTerminal := ops.position.IsTerminated()
		ops.position.UndoMove()

		node.Children[i] = *mcts.NewBaseNode(node, m, isTerminal)
	}

	return uint32(moves.size)
}

func (ops *UtttOperations) Traverse(signature PosType) {
	ops.position.MakeMove(signature)
}

func (ops *UtttOperations) BackTraverse() {
	ops.position.UndoMove()
}

// Play the game until a terminal node is reached
func (ops *UtttOperations) Rollout() mcts.Result {
	var moves *MoveList
	var move PosType
	var result mcts.Result = 0.5
	var moveCount int = 0
	leafTurn := ops.position.Turn()

	for !ops.position.IsTerminated() {
		moveCount++
		moves = ops.position.GenerateMoves()

		// Choose at random move
		move = moves.moves[ops.random.Int31n(int32(moves.size))]
		ops.position.MakeMove(move)
	}

	// If that's not a draw
	if t := ops.position.termination; t == TerminationCircleWon && leafTurn == CircleTurn ||
		t == TerminationCrossWon && leafTurn == CrossTurn {
		result = 1.0
	} else if t != TerminationDraw {
		result = 0.0
	}

	// Undo the moves
	for range moveCount {
		ops.position.UndoMove()
	}

	return result
}

func (ops UtttOperations) Clone() mcts.GameOperations[PosType] {
	return mcts.GameOperations[PosType](&UtttOperations{
		position: ops.position.Clone(),
		rootSide: ops.rootSide,
		random:   rand.New(rand.NewSource(time.Now().UnixMicro())),
	})
}
