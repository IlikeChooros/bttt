package uttt

import (
	"fmt"
	"math"
	"math/rand"
	"sync/atomic"
	"unsafe"
)

// Generalized Monte-Carlo Tree Search algorithm

const (
	virtualLoss = 10
)

// Result of the rollout, either -1 (loss), 0 (draw), or 1 (win)
type Result int

// Will be called, when we choose this node, as it is the most promising to expand
// Warning: when using NodeStats fields, must use atomic operations (Load, Store)
// since the search may be multi-threaded (tree parallelized)
type SelectionPolicy[T comparable] func(parent *NodeBase[T]) *NodeBase[T]

type NodeStats struct {
	Visits int32
	Wins   int32
	Losses int32
}

type NodeBase[T comparable] struct {
	NodeStats
	NodeSignature T
	// ID            int
	Children []NodeBase[T]
	Parent   *NodeBase[T]
	Terminal bool
	expanded int32 // atomic flag
}

type GameOperations[T comparable] interface {
	// Generate moves here, and add them as children to given node
	ExpandNode(*NodeBase[T]) uint64
	// Make a move on the internal position definition, with given
	// signature value (move)
	Traverse(T)
	// Go back up 1 time in the game tree (undo previous move, played in traverse)
	BackTraverse()
	// Function to make the playout, until terminal node is reached,
	// in case of UTTT, play random moves, until we reach draw/win/loss
	Rollout() Result
}

type TreeStats struct {
	// size     atomic.Int32
	maxdepth atomic.Int32
	nps      atomic.Uint32
	nodes    atomic.Uint32
}

type MCTS[T comparable] struct {
	TreeStats
	stop             atomic.Bool
	limits           *Limits
	timer            *_Timer
	selection_policy SelectionPolicy[T]
	root             *NodeBase[T]
	size             uint64
}

func NewMTCS[T comparable](
	selectionPolicy SelectionPolicy[T],
	operations GameOperations[T],
) *MCTS[T] {
	mcts := &MCTS[T]{
		TreeStats:        TreeStats{},
		limits:           DefaultLimits(),
		timer:            _NewTimer(),
		selection_policy: selectionPolicy,
		root:             &NodeBase[T]{},
		size:             1,
	}

	mcts.size += operations.ExpandNode(mcts.root)
	return mcts
}

// This function only sets the limits, resets the counters, and the stop flag
// doesn't actually start the search
func (mcts *MCTS[T]) setupSearch() {
	// Setup
	mcts.timer.Movetime(mcts.limits.movetime)
	mcts.timer.Reset()
	mcts.nodes.Store(0)
	mcts.stop.Store(false)
}

func (mcts *MCTS[T]) search(ops GameOperations[T]) {

	var maxcount uint64 = math.MaxUint64
	if !mcts.limits.InfiniteSize() {
		maxcount = uint64(mcts.limits.byteSize) / (uint64(unsafe.Sizeof(*mcts.root)))
	}

	var node *NodeBase[T]

	for !mcts.timer.IsEnd() &&
		!mcts.stop.Load() &&
		mcts.Nodes() <= uint32(mcts.limits.nodes) &&
		atomic.LoadUint64(&mcts.size) < maxcount {

		// Choose the most promising node
		node = mcts.selection(ops)
		// Get the result of the rollout/playout
		result := ops.Rollout()
		mcts.backpropagate(ops, node, result)
		// Store the nps
		mcts.nps.Store(mcts.nodes.Load() / uint32(mcts.timer.Deltatime()))
	}

	mcts.stop.Store(true)
}

func (mcts *MCTS[T]) IsThinking() bool {
	return !mcts.stop.Load()
}

func (mcts *MCTS[T]) Stop() {
	mcts.stop.Store(true)
}

func (mcts *MCTS[T]) MaxDepth() int {
	return int(mcts.maxdepth.Load())
}

func (mcts *MCTS[T]) Nps() uint32 {
	return mcts.nps.Load()
}

func (mcts *MCTS[T]) Nodes() uint32 {
	return mcts.nodes.Load()
}

func (mcts *MCTS[T]) SetLimits(limits *Limits) {
	mcts.limits = limits
}

func (mcts *MCTS[T]) String() string {
	str := fmt.Sprintf("MCTS={Size=%d, Stats:{MaxDepth=%d, Nps=%d, Nodes=%d}, Stop=%v",
		mcts.Size(), mcts.MaxDepth(), mcts.Nps(), mcts.Nodes(), mcts.stop.Load())
	str += fmt.Sprintf(", Root=%v, Root.Children=%v", mcts.root, mcts.root.Children)
	return str
}

func countTreeNodes[T comparable](node *NodeBase[T]) int {
	nodes := 1
	for i := range node.Children {
		if len(node.Children[i].Children) > 0 {
			nodes += countTreeNodes(&node.Children[i])
		} else {
			nodes += 1
		}
	}

	return nodes
}

func (mcts *MCTS[T]) Size() int {
	// Count every node in the tree
	return countTreeNodes(mcts.root)
}

// Remove previous tree
func (mcts *MCTS[T]) Reset() {
	// Make new root node
	mcts.root = &NodeBase[T]{}
}

// Select new root (play given move on the board, and update the tree)
func (mcts *MCTS[T]) MakeRoot(signature T) bool {
	index := -1
	for i := range mcts.root.Children {
		if mcts.root.Children[i].NodeSignature == signature {
			index = i
		}
	}

	if index == -1 {
		return false
	}

	// Create a completely new node to avoid any lingering references
	selectedChild := &mcts.root.Children[index]
	newRoot := &NodeBase[T]{
		NodeStats:     selectedChild.NodeStats,
		NodeSignature: selectedChild.NodeSignature,
		Children:      selectedChild.Children, // Keep the subtree
		Parent:        nil,                    // No parent
		Terminal:      selectedChild.Terminal,
	}

	// Update parent pointers in children
	for i := range newRoot.Children {
		newRoot.Children[i].Parent = newRoot
	}

	mcts.root = newRoot
	return true
}

func (mcts *MCTS[T]) Root() *NodeBase[T] {
	return mcts.root
}

func (mcts *MCTS[T]) RootSignature() T {
	return mcts.BestChild(mcts.root).NodeSignature
}

// Return best child, based on the number of visits
func (mcts *MCTS[T]) BestChild(node *NodeBase[T]) *NodeBase[T] {
	// most visits
	// maxVisits := 0
	// var bestChild *NodeBase[T]
	// Choose the one with highest number of visits
	// for i := 0; i < len(node.Children); i++ {
	// 	if int(node.Children[i].Visits) > maxVisits {
	// 		maxVisits = int(node.Children[i].Visits)
	// 		bestChild = &node.Children[i]
	// 	}
	// }

	const minVisitsThreshold = 0.4
	bestChild := &node.Children[0]
	bestWinRate := float64(node.Children[0].Wins) / float64(node.Children[0].Visits)

	// Get max visits out the children
	maxVisits := 0
	for i := 0; i < len(node.Children); i++ {
		maxVisits = max(int(node.Children[i].Visits), maxVisits)
	}

	// Go through the children
	for i := 1; i < len(node.Children); i++ {
		child := &node.Children[i]
		if child.Visits > int32(minVisitsThreshold*float64(maxVisits)) {
			winRate := float64(child.Wins) / float64(child.Visits)
			if winRate > bestWinRate {
				bestWinRate = winRate
				bestChild = child
			}
		}
	}

	return bestChild
}

func (mcts *MCTS[T]) selection(ops GameOperations[T]) *NodeBase[T] {
	node := mcts.root
	depth := 0
	for node.Children != nil {
		node = mcts.selection_policy(node)
		ops.Traverse(node.NodeSignature)
		depth++
		mcts.nodes.Add(1)
		// Apply virtual loss
		// atomic.AddInt32(&node.Visits, virtualLoss)
	}

	// Add new children to this node, after finding leaf node
	if atomic.LoadInt32(&node.Visits) > 0 && !node.Terminal {
		// Expand the node, only if needed (expand flag is 0)
		if atomic.CompareAndSwapInt32(&node.expanded, 0, 1) {
			atomic.AddUint64(&mcts.size, ops.ExpandNode(node))
			// Now it's expanded
			atomic.StoreInt32(&node.expanded, 2)
		}

		// Currently expanding
		for len(node.Children) == 0 && atomic.LoadInt32(&node.expanded) == 1 {
		}

		// Already expanded
		if len(node.Children) > 0 && atomic.LoadInt32(&node.expanded) == 2 {
			// Select child at random
			node = &node.Children[rand.Int()%len(node.Children)]
			// Apply again virtual loss
			// atomic.AddInt32(&node.Visits, virtualLoss)
			// Traverse to this child
			ops.Traverse(node.NodeSignature)
			depth++
			mcts.nodes.Add(1)
		}
	}

	// Set the 'max depth'
	if depth > int(mcts.maxdepth.Load()) {
		mcts.maxdepth.Store(int32(depth))
	}

	// return the candidate
	return node
}

func (mcts *MCTS[T]) backpropagate(ops GameOperations[T], node *NodeBase[T], result Result) {
	currentResult := result

	for node != nil {

		// node.Mutex.Lock()
		if currentResult > 0 {
			// node.Wins += 1
			atomic.AddInt32(&node.Wins, 1)
		} else if currentResult < 0 {
			// node.Losses += 1
			atomic.AddInt32(&node.Losses, 1)
		}

		// Reverse virtual loss
		// atomic.AddInt32(&node.Visits, -virtualLoss+1)
		atomic.AddInt32(&node.Visits, 1)

		node = node.Parent
		ops.BackTraverse()
		mcts.nodes.Add(1)
		currentResult = -currentResult
	}
}
