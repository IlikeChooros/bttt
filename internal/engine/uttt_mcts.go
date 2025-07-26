package uttt

import (
	"math/rand"
	"sync/atomic"
	"uttt/internal/mcts"
)

// Actual UTTT mcts implementation
type UtttMCTS struct {
	mcts.MCTS[PosType]
	ops UtttOperations
}

type UtttNode mcts.NodeBase[PosType]

func NewUtttMCTS(position Position) *UtttMCTS {
	uttt_ops := UtttOperations{position: position}
	ops := mcts.GameOperations[PosType](&uttt_ops)
	tree := &UtttMCTS{
		MCTS: *mcts.NewMTCS(
			mcts.DefaultSelection,
			ops,
			mcts.TerminalFlag(position.IsTerminated()),
		),
		ops: uttt_ops,
	}
	return tree
}

func (mcts *UtttMCTS) AsyncSearch() {
	mcts.MCTS.SearchMultiThreaded(&mcts.ops)
}

// Start the search
func (mcts *UtttMCTS) Search() {

	// Run the search
	mcts.SearchMultiThreaded(&mcts.ops)

	// Wait for the search to end
	mcts.Synchronize()
}

// Default selection
func (mcts *UtttMCTS) Selection() *mcts.NodeBase[PosType] {
	return mcts.MCTS.Selection(&mcts.ops)
}

// Default backprop
func (mcts *UtttMCTS) Backpropagate(node *mcts.NodeBase[PosType], result mcts.Result) {
	mcts.MCTS.Backpropagate(&mcts.ops, node, result)
}

func (mcts *UtttMCTS) Ops() mcts.GameOperations[PosType] {
	return &mcts.ops
}

func (mcts *UtttMCTS) Reset() {
	mcts.MCTS.Reset(&mcts.ops, mcts.ops.position.IsTerminated())
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

func (mcts *UtttMCTS) SearchResult(pvPolicy mcts.BestChildPolicy) SearchResult {
	pv, mate := mcts.Pv(pvPolicy)
	turn := 1
	result := SearchResult{
		Bestmove: mcts.RootSignature(),
		Nodes:    uint64(mcts.Nodes()),
		Nps:      uint64(mcts.Nps()),
		Depth:    mcts.MaxDepth(),
		Pv:       *pv,
	}

	if mcts.ops.position.Turn() == CircleTurn {
		turn = -1
	}

	// Set the score
	if mate {
		result.ScoreType = MateScore
		result.Value = pv.Size() * turn

		// If the game ends on our turn, we are losing
		if pv.Size()%2 == 0 {
			result.Value = -result.Value
		}
	} else {
		visits := int(atomic.LoadInt32(&mcts.Root.Visits))
		wins := int(atomic.LoadInt32(&mcts.Root.Wins))
		if mcts.Root != nil && visits > 0 {
			result.Value = 100 * wins / visits
		} else {
			result.Value = -1
		}
	}

	return result
}

// Get the principal variation (pv, isTerminal)
func (self *UtttMCTS) Pv(policy mcts.BestChildPolicy) (*MoveList, bool) {
	nodes, mate := self.MCTS.Pv(policy)
	pv := NewMoveList()
	for _, node := range nodes {
		pv.AppendMove(node.NodeSignature)
	}
	return pv, mate
}

type UtttOperations struct {
	position Position
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
	var result mcts.Result = 0
	var moveCount int = 0
	ourSide := ops.position.Turn()

	for !ops.position.IsTerminated() {
		moveCount++
		moves = ops.position.GenerateMoves()

		// Choose at random move
		move = moves.moves[rand.Int31n(int32(moves.size))]
		ops.position.MakeMove(move)
	}

	// If that's not a draw
	if ops.position.termination != TerminationDraw {
		// We lost
		if ops.position.Turn() == ourSide {
			result = -1
		} else {
			// We won
			result = 1
		}
	}

	// Undo the moves
	for range moveCount {
		ops.position.UndoMove()
	}

	return result
}

func (ops UtttOperations) Clone() mcts.GameOperations[PosType] {
	return mcts.GameOperations[PosType](&UtttOperations{position: ops.position.Clone()})
}
