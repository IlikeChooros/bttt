package mcts

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Generalized Monte-Carlo Tree Search algorithm

const (
	virtualLoss = 10
)

// Result of the rollout, either -1 (loss), 0 (draw), or 1 (win)
type Result int

type BestChildPolicy int

var ExplorationParam float64 = 0.55

const (
	BestChildMostVisits BestChildPolicy = iota
	BestChildWinRate
)

// Will be called, when we choose this node, as it is the most promising to expand
// Warning: when using NodeStats fields, must use atomic operations (Load, Store)
// since the search may be multi-threaded (tree parallelized)
type SelectionPolicy[T comparable] func(parent *NodeBase[T]) *NodeBase[T]

// Default selection of the node policy (with ucb 1 value)
func DefaultSelection[T comparable](node *NodeBase[T]) *NodeBase[T] {

	// Is that's a terminal node, simply return itself, there is no children anyway
	// and on the rollout we will exit early, since the position is terminated
	if node.Terminal() {
		return node
	}

	max := float64(-1)
	index := 0
	parent_visits := atomic.LoadInt32(&node.Visits)

	for i := 0; i < len(node.Children); i++ {
		// Get the variables
		visits := atomic.LoadInt32(&node.Children[i].Visits)

		// Pick the unvisited one
		if visits == 0 {
			// Return pointer to the child
			return &node.Children[i]
		}

		wins := atomic.LoadInt32(&node.Children[i].Wins)

		// UCB 1 : wins/visits + C * sqrt(ln(parent_visits)/visits)
		ucb := float64(wins)/float64(visits) +
			ExplorationParam*math.Sqrt(math.Log(float64(parent_visits))/float64(visits))

		if ucb > max {
			max = ucb
			index = i
		}
	}

	return &node.Children[index]
}

// visit/win/loss count of the node, should be accessed only with atomic operations
type NodeStats struct {
	Visits int32 // Incremented each time this node is on a path: selected node - root
	Wins   int32 // From this node's perspective (not root's!), number of wins we get from this position forward
	Losses int32 // Same as wins, but it counts number of losses
}

type GameFlags uint8

const (
	TurnMask     GameFlags = 1
	TerminalMask GameFlags = 2
)

type NodeBase[T comparable] struct {
	NodeStats
	NodeSignature T
	Children      []NodeBase[T]
	Parent        *NodeBase[T]
	GameFlags     GameFlags
	state         atomic.Int32 // atomic flag for 'expanded'
}

func NewBaseNode[T comparable](parent *NodeBase[T], signature T, terminated bool) *NodeBase[T] {
	return &NodeBase[T]{
		NodeSignature: signature,
		Children:      nil,
		Parent:        parent,
		GameFlags:     TerminalFlag(terminated),
	}
}

// Reads the game flags, and return wheter the node is terminal
func (node *NodeBase[T]) Terminal() bool {
	return node.GameFlags&TerminalMask == TerminalMask
}

func (node *NodeBase[T]) SetFlag(flag GameFlags) {
	node.GameFlags |= flag
}

func TerminalFlag(terminal bool) GameFlags {
	flag := GameFlags(0)
	if terminal {
		flag |= 2
	}
	return flag
}

type GameOperations[T comparable] interface {
	// Generate moves here, and add them as children to given node
	ExpandNode(*NodeBase[T]) uint32
	// Make a move on the internal position definition, with given
	// signature value (move)
	Traverse(T)
	// Go back up 1 time in the game tree (undo previous move, played in traverse)
	BackTraverse()
	// Function to make the playout, until terminal node is reached,
	// in case of UTTT, play random moves, until we reach draw/win/loss
	Rollout() Result

	// Clone itself, without any shared memory with the other object
	Clone() GameOperations[T]
}

type TreeStats struct {
	// size     atomic.Int32
	maxdepth atomic.Int32
	nps      atomic.Uint32
	nodes    atomic.Uint32
}

type MCTS[T comparable] struct {
	TreeStats
	Limiter          LimiterLike
	selection_policy SelectionPolicy[T]
	Root             *NodeBase[T]
	size             atomic.Uint32
	wg               sync.WaitGroup
}

func NewMTCS[T comparable](
	selectionPolicy SelectionPolicy[T],
	operations GameOperations[T],
	flags GameFlags,
) *MCTS[T] {
	mcts := &MCTS[T]{
		TreeStats:        TreeStats{},
		Limiter:          LimiterLike(NewLimiter(uint32(unsafe.Sizeof(NodeBase[T]{})))),
		selection_policy: selectionPolicy,
		Root:             &NodeBase[T]{GameFlags: flags},
	}

	mcts.Limiter.Stop()
	mcts.Root.state.Store(2)
	mcts.size.Store(1 + operations.ExpandNode(mcts.Root))
	return mcts
}

func (mcts *MCTS[T]) IsThinking() bool {
	return !mcts.Limiter.Stop()
}

func (mcts *MCTS[T]) Stop() {
	mcts.Limiter.SetStop(true)
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
	mcts.Limiter.SetLimits(limits)
}

func (mcts *MCTS[T]) Limits() *Limits {
	return mcts.Limiter.Limits()
}

func (mcts *MCTS[T]) String() string {
	str := fmt.Sprintf("MCTS={Size=%d, Stats:{MaxDepth=%d, Nps=%d, Nodes=%d}, Stop=%v",
		mcts.Size(), mcts.MaxDepth(), mcts.Nps(), mcts.Nodes(), !mcts.IsThinking())
	str += fmt.Sprintf(", Root=%v, Root.Children=%v", mcts.Root, mcts.Root.Children)
	return str
}

// Helper function to count tree nodes
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

// Get the size of the tree (by counting)
func (mcts *MCTS[T]) Count() int {
	return countTreeNodes(mcts.Root)
}

// Get the size of the tree
func (mcts *MCTS[T]) Size() uint32 {
	// Count every node in the tree
	return mcts.size.Load()
}

// Remove previous tree
func (mcts *MCTS[T]) Reset(ops GameOperations[T], isTerminated bool) {
	// Discard running search
	if mcts.IsThinking() {
		mcts.Stop()
		mcts.Synchronize()
	}

	// Make new root
	var signatureNull T
	mcts.Root = NewBaseNode(nil, signatureNull, isTerminated)
	mcts.size.Store(1)
	mcts.Root.state.Store(2)

	if !isTerminated {
		mcts.size.Add(ops.ExpandNode(mcts.Root))
	}
}

// Select new root (play given move on the board, and update the tree)
func (mcts *MCTS[T]) MakeRoot(signature T) bool {
	index := -1
	for i := range mcts.Root.Children {
		if mcts.Root.Children[i].NodeSignature == signature {
			index = i
		}
	}

	if index == -1 {
		return false
	}

	// Create a completely new node to avoid any lingering references
	selectedChild := &mcts.Root.Children[index]
	newRoot := &NodeBase[T]{
		NodeStats:     selectedChild.NodeStats,
		NodeSignature: selectedChild.NodeSignature,
		Children:      selectedChild.Children, // Keep the subtree
		Parent:        nil,                    // No parent
		GameFlags:     selectedChild.GameFlags,
	}

	// Update parent pointers in children
	for i := range newRoot.Children {
		newRoot.Children[i].Parent = newRoot
	}

	mcts.Root = newRoot

	// Update the counters
	mcts.size.Store(uint32(mcts.Count()))
	return true
}

func (mcts *MCTS[T]) RootSignature() T {
	var signature T
	if bestChild := mcts.BestChild(mcts.Root, BestChildWinRate); bestChild != nil {
		signature = bestChild.NodeSignature
	}
	return signature
}

// Return best child, based on the number of visits
func (mcts *MCTS[T]) BestChild(node *NodeBase[T], policy BestChildPolicy) *NodeBase[T] {
	var bestChild *NodeBase[T]
	var child *NodeBase[T]

	switch policy {
	case BestChildMostVisits:
		maxVisits := 0
		for i := 0; i < len(node.Children); i++ {
			child = &node.Children[i]
			if int(child.Visits) > maxVisits {
				maxVisits = int(child.Visits)
				bestChild = child
			}
			// Always a choose winning terminal node (by definition, when position terminates
			// on opponents turn, previous player won)
			if child.Terminal() {
				bestChild = child
				break
			}
		}
	case BestChildWinRate:
		// the child we choose should have at least 20% of the max visit count (from the neighbours)
		const minVisitsThreshold = 0.2
		bestWinRate := -1.0

		// Get max visits out the children
		maxVisits := 0
		for i := 0; i < len(node.Children); i++ {
			maxVisits = max(int(node.Children[i].Visits), maxVisits)
		}

		// Go through the children
		for i := 0; i < len(node.Children); i++ {
			child = &node.Children[i]
			if child.Visits > virtualLoss && child.Visits > int32(minVisitsThreshold*float64(maxVisits)) {

				// We choose a move that minimizes winnning changes of our opponent
				var winRate float64 = float64(child.Losses) / float64(child.Visits)

				if winRate > bestWinRate {
					bestWinRate = winRate
					bestChild = child
				}
			}

			// Always choose a terminating move (meaning we terminated the position after this move
			// which, in most board games, means we won)
			if child.Terminal() {
				bestChild = child
				break
			}
		}
	}

	return bestChild
}

// Get the principal variation (ie. the best sequence of moves)
// based on given best child policy
func (mcts *MCTS[T]) Pv(policy BestChildPolicy) ([]*NodeBase[T], bool) {
	if mcts.Root == nil {
		return nil, false
	}

	pv := make([]*NodeBase[T], 0, mcts.MaxDepth())
	node := mcts.Root
	mate := false

	// Simply select 'best child' until we don't have any children
	// or the node is nil
	for len(node.Children) > 0 {
		node = mcts.BestChild(node, policy)
		if node == nil {
			break
		}

		pv = append(pv, node)

		// If that's a terminal node, we got a mate score
		if node.Terminal() {
			mate = true
			break
		}
	}

	return pv, mate
}
