package uttt

import (
	"math"
	"math/rand"
	"sync/atomic"
)

// Actual UTTT mcts implementation
type UtttMCTS struct {
	MCTS[PosType]
	ops UtttOperations
}

const (
	explorationParam = 0.7
)

// Default selection of the node policy
var DefaultSelection SelectionPolicy[PosType] = func(node *NodeBase[PosType]) *NodeBase[PosType] {

	max := float64(-1)
	index := 0
	parent_visits := atomic.LoadInt32(&node.Visits)

	for i := 0; i < len(node.Children); i++ {
		// Get the variables
		visits := atomic.LoadInt32(&node.Children[i].Visits)

		// Pick the unvisited one
		if visits <= virtualLoss {
			// Return pointer to the child
			return &node.Children[i]
		}

		wins := atomic.LoadInt32(&node.Children[i].Wins)

		// UCB 1 : wins/visits + C * sqrt(ln(parent_visits)/visits)
		ucb := float64(wins)/float64(visits) +
			explorationParam*math.Sqrt(math.Log(float64(parent_visits))/float64(visits))

		if ucb > max {
			max = ucb
			index = i
		}
	}

	return &node.Children[index]
}

func NewUtttMCTS(position Position) *UtttMCTS {
	uttt_ops := UtttOperations{position: position}
	ops := GameOperations[PosType](&uttt_ops)
	mcts := &UtttMCTS{
		MCTS: *NewMTCS(
			DefaultSelection,
			ops,
		),
		ops: uttt_ops,
	}

	// Check if the root node is terminal
	mcts.root.Terminal = position.IsTerminated()
	return mcts
}

func (mcts *UtttMCTS) AsyncSearch() {
	mcts.setupSearch()
	threads := max(1, mcts.limits.nThreads)
	for range threads {
		go mcts.search(&UtttOperations{position: mcts.ops.position.Clone()})
	}
}

// Start the search
func (mcts *UtttMCTS) Search() {
	mcts.setupSearch()

	// If that's multi-threaded search
	if mcts.limits.nThreads > 1 {
		for i := 0; i < mcts.limits.nThreads; i++ {
			go mcts.search(&UtttOperations{position: mcts.ops.position.Clone()})
		}

		// Wait for the search to end
		for mcts.IsThinking() {
			continue
		}
	} else {
		mcts.search(&mcts.ops)
	}
}

// Default selection
func (mcts *UtttMCTS) Selection() *NodeBase[PosType] {
	return mcts.selection(&mcts.ops)
}

// Default backprop
func (mcts *UtttMCTS) Backpropagate(node *NodeBase[PosType], result Result) {
	mcts.backpropagate(&mcts.ops, node, result)
}

func (mcts *UtttMCTS) SetNotation(notation string) error {
	mcts.Reset()
	return mcts.ops.position.FromNotation(notation)
}

// Get the principal variation
func (mcts *UtttMCTS) GetPv() *MoveList {
	pv := NewMoveList()

	node := mcts.root

	// Simply select 'best child' until we don't have any children
	// or the node is nil
	for node != nil && len(node.Children) > 0 {
		node = mcts.BestChild(node)
		pv.AppendMove(node.NodeSignature)
	}

	return pv
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

		node.Children[i] = NodeBase[PosType]{
			NodeStats:     NodeStats{},
			NodeSignature: m,
			Children:      nil,
			Parent:        node,
			Terminal:      isTerminal,
		}
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
