package uttt

import (
	"math/rand"
)

// Generalized Monte-Carlo Tree Search algorithm

type Result int

// Will be called, when we choose this node, as it is the most promising to expand
type SelectionPolicy[T any] func(parent *NodeBase[T]) *NodeBase[T]

type NodeStats struct {
	Visits int
	Wins   Result
}

type NodeBase[T any] struct {
	NodeStats
	NodeSignature T
	// ID            int
	Children []NodeBase[T]
	Parent   *NodeBase[T]
	Terminal bool
}

type GameOperations[T any] interface {
	// Generate moves here, and add them as children to given node
	ExpandNode(*NodeBase[T])
	// Make a move on the internal position definition, with given
	// signature value (move)
	Traverse(T)
	// Go back up 1 time in the game tree (undo previous move, played in traverse)
	BackTraverse()
	// Function to make the playout, until terminal node is reached,
	// in case of UTTT, play random moves, until we reach draw/win/loss
	Rollout() Result
}

type MCTS[T any] struct {
	limits           *Limits
	timer            *_Timer
	selection_policy SelectionPolicy[T]
	root             *NodeBase[T]
	ops              GameOperations[T]
}

func NewMTCS[T any](
	selectionPolicy SelectionPolicy[T],
	operations GameOperations[T],
) *MCTS[T] {
	return &MCTS[T]{
		limits:           DefaultLimits(),
		timer:            _NewTimer(),
		selection_policy: selectionPolicy,
		root: &NodeBase[T]{
			NodeStats: NodeStats{
				Visits: 0,
				Wins:   0,
			},
			Children: nil,
			Parent:   nil,
		},
		ops: operations,
	}
}

func (mcts *MCTS[T]) Search() {
	// Setup
	mcts.timer.Reset()
	var node *NodeBase[T]

	for !mcts.timer.IsEnd() {
		// Choose the most promising node
		node = mcts.selection()
		// Get the result of the rollout/playout
		result := mcts.ops.Rollout()
		mcts.backpropagate(node, result)
	}
}

func (mcts *MCTS[T]) BestChild() T {
	// most wins
	wins := Result(MinValue)
	var signature T

	for _, child := range mcts.root.Children {
		if child.Wins > wins {
			signature = child.NodeSignature
		}
	}

	return signature
}

func (mcts *MCTS[T]) selection() *NodeBase[T] {
	node := mcts.root
	for node.Children != nil {
		node = mcts.selection_policy(node)
		mcts.ops.Traverse(node.NodeSignature)
	}

	// Add new children to this node, after finding leaf node
	if node.Visits > 0 && !node.Terminal {
		// Expand the node
		mcts.ops.ExpandNode(node)
		// Select child at random
		node = &node.Children[rand.Int()%len(node.Children)]
	}

	// return the candidate
	return node
}

func (mcts *MCTS[T]) backpropagate(node *NodeBase[T], result Result) {
	currentResult := result
	for node != nil {
		node.NodeStats.Wins += currentResult
		node.NodeStats.Visits += 1
		node = node.Parent
		mcts.ops.BackTraverse()
		currentResult = -currentResult
	}
}
