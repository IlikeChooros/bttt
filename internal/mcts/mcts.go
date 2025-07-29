package mcts

import (
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Generalized Monte-Carlo Tree Search algorithm

var VirtualLoss int32 = 0

// Result of the rollout, should range from [0, 1] 0 being loss from the root's perspective
// and 1 being a win
type Result float64
type MoveLike comparable
type BestChildPolicy int

var ExplorationParam float64 = 0.75
var threadExploration float64 = ExplorationParam // exploartion factor, but scaled up for multi-threaded search

const (
	BestChildMostVisits BestChildPolicy = iota
	BestChildWinRate
)

// Will be called, when we choose this node, as it is the most promising to expand
// Warning: when using NodeStats fields, must use atomic operations (Load, Store)
// since the search may be multi-threaded (tree parallelized)
type SelectionPolicy[T MoveLike] func(parent, root *NodeBase[T]) *NodeBase[T]

// visits/virutal loss/win/loss count of the node,
// wins, and losses should be accessed only with atomic operations
// However to read the visit and virtual loss counts, use the methods
type NodeStats struct {
	// This is visit counter, it cannot be read by atomic, use GetVvl() Visits() to properly read this value
	visits int32

	// Current virtual loss applied to visits, it always meets condition: visits - virtualLoss >= 0.
	// Read this value ONLY with GetVvl() or VirtualLoss() methods
	virtualLoss int32

	outcomes Result // float64 value of compounded outcomes for this node

	// Synchornizes read/write on visits, virtual loss and outcomes
	statsMutex sync.RWMutex
}

type GameFlags uint8

const (
	TurnMask     GameFlags = 1
	TerminalMask GameFlags = 2
)

type NodeBase[T MoveLike] struct {
	NodeStats
	NodeSignature T
	Children      []NodeBase[T]
	Parent        *NodeBase[T]
	GameFlags     GameFlags
	state         atomic.Uint32 // atomic flag for 'expanded'
}

func newRootNode[T MoveLike](terminated bool) *NodeBase[T] {
	return &NodeBase[T]{
		Children:  nil,
		GameFlags: TerminalFlag(terminated),
	}
}

func NewBaseNode[T MoveLike](parent *NodeBase[T], signature T, terminated bool) *NodeBase[T] {
	return &NodeBase[T]{
		NodeSignature: signature,
		Children:      nil,
		Parent:        parent,
		GameFlags:     TerminalFlag(terminated) | ((parent.GameFlags & TurnMask) ^ TurnMask), // flip the turn
	}
}

func (node *NodeBase[T]) AvgOutcome() Result {
	node.statsMutex.RLock()
	defer node.statsMutex.RUnlock()
	return node.outcomes / Result(node.visits)
}

func (node *NodeBase[T]) Outcomes() Result {
	node.statsMutex.RLock()
	defer node.statsMutex.RUnlock()
	return node.outcomes
}

func (node *NodeBase[T]) AddOutcome(result Result) {
	node.statsMutex.Lock()
	node.outcomes += result
	node.statsMutex.Unlock()
}

func (node *NodeBase[T]) Visits() int32 {
	node.statsMutex.RLock()
	defer node.statsMutex.RUnlock()

	return node.visits
}

func (node *NodeBase[T]) VirtualLoss() int32 {
	node.statsMutex.RLock()
	defer node.statsMutex.RUnlock()

	return node.virtualLoss
}

// Get both visits and virtual loss (to avoid situtation one of them is modified)
// returns (visits, virtual loss)
func (node *NodeBase[T]) GetVvl() (int32, int32) {
	node.statsMutex.RLock()
	defer node.statsMutex.RUnlock()

	return node.visits, node.virtualLoss
}

// Returns visits - virtual loss
func (node *NodeBase[T]) RealVisits() int32 {
	visits, virtualLoss := node.GetVvl()
	return visits - virtualLoss
}

// Adds VirtuaLoss to both visits and virtual loss counters
func (node *NodeBase[T]) AddVvl(visits, virtualLoss int32) {
	node.statsMutex.Lock()

	node.visits += visits
	node.virtualLoss += virtualLoss

	node.statsMutex.Unlock()
}

// Sets visits and virtual loss of this node to specified value
func (node *NodeBase[T]) SetVvl(visits, virtualLoss int32) {
	node.statsMutex.Lock()

	node.visits = visits
	node.virtualLoss = virtualLoss

	node.statsMutex.Unlock()
}

// Reads the game flags, and return wheter the node is terminal
func (node *NodeBase[T]) Terminal() bool {
	return node.GameFlags&TerminalMask == TerminalMask
}

func (node *NodeBase[T]) Turn() bool {
	return node.GameFlags&TurnMask == TurnMask
}

func (node *NodeBase[T]) SetFlag(flag GameFlags) {
	node.GameFlags |= flag
}

func TurnFlag(turn bool) GameFlags {
	flag := GameFlags(0)
	if turn {
		flag |= 1
	}
	return flag
}

func TerminalFlag(terminal bool) GameFlags {
	flag := GameFlags(0)
	if terminal {
		flag |= 2
	}
	return flag
}

// Same as asking if the node has chidlren
func (node *NodeBase[T]) Expanded() bool {
	return node.state.Load() == 2
}

// See if currenlty node is being expanded
func (node *NodeBase[T]) Expanding() bool {
	return node.state.Load() == 1
}

// Should be called when we want to expand this node,
// if it's possible, sets the internal flag to 'currently expanding'
func (node *NodeBase[T]) CanExpand() bool {
	return node.state.CompareAndSwap(0, 1)
}

// After successful 'CanExpand' call, use this function to set
// the state of the node to 'expanded'
func (node *NodeBase[T]) FinishExpanding() {
	node.state.CompareAndSwap(1, 2)
}

type GameOperations[T MoveLike] interface {
	// Generate moves here, and add them as children to given node
	ExpandNode(parent *NodeBase[T]) uint32
	// Make a move on the internal position definition, with given
	// signature value (move)
	Traverse(T)
	// Go back up 1 time in the game tree (undo previous move, which was played in traverse)
	BackTraverse()
	// Function to make the playout, until terminal node is reached,
	// in case of UTTT, play random moves, until we reach draw/win/loss
	Rollout() Result

	// Reset game state to current internal position, called after changing
	// position, for example using SetNotation function in engine
	Reset()
	// Clone itself, without any shared memory with the other object
	Clone() GameOperations[T]
}

type TreeStats struct {
	// size     atomic.Int32
	maxdepth atomic.Int32
	nps      atomic.Uint32
	nodes    atomic.Uint32
}

type MCTS[T MoveLike] struct {
	TreeStats
	Limiter          LimiterLike
	selection_policy SelectionPolicy[T]
	Root             *NodeBase[T]
	size             atomic.Uint32
	wg               sync.WaitGroup
}

func NewMTCS[T MoveLike](
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

	// Set IsThinking to false
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
func countTreeNodes[T MoveLike](node *NodeBase[T]) int {
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

// Remove previous tree & update game ops state
func (mcts *MCTS[T]) Reset(ops GameOperations[T], isTerminated bool) {
	// Discard running search
	if mcts.IsThinking() {
		mcts.Stop()
		mcts.Synchronize()
	}

	// Reset game state and make new root
	ops.Reset()
	mcts.Root = newRootNode[T](isTerminated)
	mcts.size.Store(1)
	mcts.Root.state.Store(2)

	if !isTerminated {
		mcts.size.Add(ops.ExpandNode(mcts.Root))
	}
}

func (mcts *MCTS[T]) RootSignature() T {
	var signature T
	if bestChild := mcts.BestChild(mcts.Root, BestChildMostVisits); bestChild != nil {
		signature = bestChild.NodeSignature
	}
	return signature
}

// Return best child, based on the number of visits
func (mcts *MCTS[T]) BestChild(node *NodeBase[T], policy BestChildPolicy) *NodeBase[T] {
	var bestChild *NodeBase[T]
	var child *NodeBase[T]

	// DEBUG
	// rootTurn := mcts.Root.Turn()
	// childSameAsRootTurn := mcts.Root.Turn() != node.Turn()
	// if node.Turn() == rootTurn {
	// 	fmt.Println("Root's turn")
	// } else {
	// 	fmt.Println("Enemy's turn")
	// }

	// for i := range node.Children {
	// 	ch := &node.Children[i]
	// 	fmt.Printf("%d. %v v=%d (wr=%.2f)\n",
	// 		i+1, ch.NodeSignature, ch.Visits(),
	// 		float64(ch.Outcomes())/float64(ch.Visits()),
	// 	)
	// }

	switch policy {
	case BestChildMostVisits:
		maxVisits := 0
		for i := 0; i < len(node.Children); i++ {
			child = &node.Children[i]
			if v := int(child.Visits()); v > maxVisits && v > 0 {
				maxVisits = int(child.Visits())
				bestChild = child
			}
		}
	case BestChildWinRate:
		// the child we choose should have at least 20% of the max visit count (from the neighbours)
		const (
			minVisitsPercentageThreshold = 0
			minVisitsThreshold           = 10
		)

		bestWinRate := -1.0

		// Get max visits out the children
		maxVisits := 0
		for i := 0; i < len(node.Children); i++ {
			maxVisits = max(int(node.Children[i].Visits()), maxVisits)
		}

		// Go through the children
		for i := 0; i < len(node.Children); i++ {
			child = &node.Children[i]
			real := child.RealVisits()
			if real > minVisitsThreshold && real > int32(minVisitsPercentageThreshold*float64(maxVisits)) {

				// We optimize the winning chances, looking from the root's perspective
				var winRate float64 = float64(child.Outcomes()) / float64(child.Visits())

				if winRate > bestWinRate {
					bestWinRate = winRate
					bestChild = child
				}
			}
		}
	}

	// if bestChild != nil {
	// 	fmt.Println("Chose", bestChild.NodeSignature)
	// }

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
