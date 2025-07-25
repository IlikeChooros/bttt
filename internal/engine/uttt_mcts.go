package uttt

import (
	"math/rand"
	"sync/atomic"
)

// Actual UTTT mcts implementation
type UtttMCTS struct {
	MCTS[PosType]
	ops UtttOperations
}

type UtttNode NodeBase[PosType]

func NewUtttMCTS(position Position) *UtttMCTS {
	uttt_ops := UtttOperations{position: position}
	ops := GameOperations[PosType](&uttt_ops)
	mcts := &UtttMCTS{
		MCTS: *NewMTCS(
			DefaultSelection,
			ops,
			TerminalFlag(position.IsTerminated()),
		),
		ops: uttt_ops,
	}
	return mcts
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
func (mcts *UtttMCTS) Selection() *NodeBase[PosType] {
	return mcts.MCTS.Selection(&mcts.ops)
}

// Default backprop
func (mcts *UtttMCTS) Backpropagate(node *NodeBase[PosType], result Result) {
	mcts.MCTS.Backpropagate(&mcts.ops, node, result)
}

func (mcts *UtttMCTS) Ops() GameOperations[PosType] {
	return &mcts.ops
}

func (mcts *UtttMCTS) Reset() {
	mcts.MCTS.Reset(&mcts.ops, bool(mcts.ops.position.Turn()), mcts.ops.position.IsTerminated())
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

func (mcts *UtttMCTS) SearchResult() SearchResult {
	pv, mate := mcts.Pv()
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
		visits := int(atomic.LoadInt32(&mcts.root.Visits))
		wins := int(atomic.LoadInt32(&mcts.root.Wins))
		if mcts.root != nil && visits > 0 {
			result.Value = 100 * wins / visits
		} else {
			result.Value = -1
		}
	}

	return result
}

// Get the principal variation (pv, isTerminal, lastnode)
func (mcts *UtttMCTS) Pv() (*MoveList, bool) {
	nodes, mate := mcts.MCTS.Pv(BestChildWinRate)
	pv := NewMoveList()
	for _, node := range nodes {
		pv.AppendMove(node.NodeSignature)
	}
	return pv, mate
}

type UtttOperations struct {
	position Position
}

func (ops *UtttOperations) ExpandNode(node *NodeBase[PosType]) uint64 {

	moves := ops.position.GenerateMoves()
	node.Children = make([]NodeBase[PosType], moves.size)

	for i, m := range moves.Slice() {
		ops.position.MakeMove(m)
		isTerminal := ops.position.IsTerminated()
		ops.position.UndoMove()

		node.Children[i] = *NewBaseNode(node, m, isTerminal)
	}

	return uint64(moves.size)
}

func (ops *UtttOperations) Traverse(signature PosType) {
	ops.position.MakeMove(signature)
}

func (ops *UtttOperations) BackTraverse() {
	ops.position.UndoMove()
}

// Play the game until a terminal node is reached
func (ops *UtttOperations) Rollout() Result {
	var moves *MoveList
	var move PosType
	var result Result = 0
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

func (ops UtttOperations) Clone() GameOperations[PosType] {
	return GameOperations[PosType](&UtttOperations{position: ops.position.Clone()})
}
