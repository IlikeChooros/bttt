package mcts

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Default node selection policy (upper confidence bound)
func UCB1[T MoveLike](parent, root *NodeBase[T]) *NodeBase[T] {

	// Is that's a terminal node, simply return itself, there is no children anyway
	// and on the rollout we will exit early, since the position is terminated
	if parent.Terminal() {
		return parent
	}

	max := float64(-1)
	index := 0
	parentVisits := parent.Visits()

	// If this flag is set to true, we will use children's lose rate
	// as exploitation value
	var child *NodeBase[T]

	for i := 0; i < len(parent.Children); i++ {

		// Get the variables
		child = &parent.Children[i]
		visits, vl := child.GetVvl()
		actualVisits := visits - vl

		// Pick the unvisited one
		if actualVisits == 0 {
			// Return pointer to the child
			return child
		}

		// If root's turn == child's turn, then we maximize the wins, else
		// minimize the losses
		wins := child.Outcomes()

		// UCB 1 : wins/visits + C * sqrt(ln(parent_visits)/visits)
		// ucb1 = epliotation + exploration
		// Since we assume the game is zero-sum, we want to expand the tree's nodes
		// that have best value according to the root, meaning we want don't want to look
		// too much at our opponents best moves, for example forced mate in 2, since there is a chance
		// we can avoid it. If we don't, the engine will tunnel vision on opponent's best moves, discarding
		// our winning chances, due to the specific implementation of this tree
		ucb1 := float64(wins)/float64(visits) +
			threadExploration*math.Sqrt(2*math.Log(float64(parentVisits))/float64(visits))

		if ucb1 > max {
			max = ucb1
			index = i
		}
	}

	return &parent.Children[index]
}

// Use when started multi-threaded search and want it to synchronize with this thread
func (mcts *MCTS[T]) Synchronize() {
	mcts.wg.Wait()
}

// Run multi-treaded search, to wait for the result, call Synchronize
func (mcts *MCTS[T]) SearchMultiThreaded(ops GameOperations[T]) {
	mcts.setupSearch()
	threads := max(1, mcts.Limiter.Limits().NThreads)

	// if threads >= 8 {
	// 	threadExploration = (0.75 + 1/float64(threads)) * ExplorationParam
	// } else if threads >= 4 {
	// 	threadExploration = (0.25 + math.Sqrt(float64(threads))/3) * ExplorationParam
	// } else {
	// 	threadExploration = ExplorationParam
	// }
	VirtualLoss = 1 + (int32(threads)-1)*20

	fmt.Printf("Exp=%f, VL=%d\n", threadExploration, VirtualLoss)
	// VirtualLoss = 5

	for id := range threads {
		mcts.wg.Add(1)
		go mcts.Search(ops.Clone(), id)
	}
}

// This function only sets the limits, resets the counters, and the stop flag
// doesn't actually start the search
func (mcts *MCTS[T]) setupSearch() {
	// Setup
	// mcts.timer.Movetime(mcts.Limiter.Limits.Movetime)
	// mcts.timer.Reset()
	mcts.Limiter.Reset()
	mcts.nodes.Store(0)
	mcts.nps.Store(0)
	mcts.maxdepth.Store(0)
	// mcts.stop.Store(false)
}

// Actual search function implementation, simply calls:
//
// 1. selection - to choose the most promising node
//
// 2. rollout - to simulate the user-defined game, and get the result of a playout
//
// 3. backpropagate - to increment counters up to the root
//
// Until runs out of the allocated time, nodes, or memory
func (mcts *MCTS[T]) Search(ops GameOperations[T], threadId int) {
	defer mcts.wg.Done()

	threadRand := rand.New(rand.NewSource(time.Now().UnixNano() + int64(threadId)))

	if mcts.Root.Terminal() {
		return
	}

	var node *NodeBase[T]

	for mcts.Limiter.Ok(mcts.Nodes(), mcts.Size(), uint32(mcts.MaxDepth()), uint32(mcts.Root.Visits())) {

		// Choose the most promising node
		node = mcts.Selection(ops, threadRand)
		// Get the result of the rollout/playout
		result := ops.Rollout()
		mcts.Backpropagate(ops, node, result)
		// Store the nps
		mcts.nps.Store(mcts.nodes.Load() * 1000 / mcts.Limiter.Elapsed())
	}

	// Synchronize all threads
	mcts.Limiter.Stop()
}

// Selects next child to expand, by user-defined selection policy
func (mcts *MCTS[T]) Selection(ops GameOperations[T], threadRand *rand.Rand) *NodeBase[T] {

	// Apply virtual loss (for compabality in backpropagation)
	// mcts.Root.virtualLoss.Add(virtualLoss)
	// atomic.AddInt32(&mcts.Root.Visits, virtualLoss)
	// mcts.Root.AddVvl(virtualLoss, VirtualLoss)

	node := mcts.Root
	depth := 0
	for node.Expanded() {
		node = mcts.selection_policy(node, mcts.Root)
		ops.Traverse(node.NodeSignature)
		depth++
		mcts.nodes.Add(1)

		// Apply virtual loss
		node.AddVvl(VirtualLoss, VirtualLoss)
	}

	// Add new children to this node, after finding leaf node
	if node.RealVisits() > 0 && !node.Terminal() {
		// Expand the node, only if needed (expand flag is 0)
		if mcts.Limiter.Expand() && node.CanExpand() {
			mcts.size.Add(ops.ExpandNode(node))
			// Now update it's state
			node.FinishExpanding()
		}

		// Currently expanding
		for node.Expanding() {
		}

		// Already set
		if node.Expanded() {
			// Select child at random
			node = &node.Children[threadRand.Int31n(int32(len(node.Children)))]
			// Traverse to this child
			ops.Traverse(node.NodeSignature)
			depth++
			mcts.nodes.Add(1)
			// Apply again virtual loss
			node.AddVvl(VirtualLoss, VirtualLoss)
		}
	}

	// Set the 'max depth'
	if depth > int(mcts.maxdepth.Load()) {
		mcts.maxdepth.Store(int32(depth))
	}

	// return the candidate
	return node
}

// Increment the counters (wins/losses/visits) along the tree path
func (mcts *MCTS[T]) Backpropagate(ops GameOperations[T], node *NodeBase[T], result Result) {
	currentResult := result

	for node != nil {

		// Reverse virtual loss for non-root
		if node.Parent != nil {
			node.AddVvl(1-VirtualLoss, -VirtualLoss)
		} else {
			node.AddVvl(1, 0)
		}

		// Add the outcome
		node.AddOutcome(result)

		// Backpropagate
		node = node.Parent
		ops.BackTraverse()
		mcts.nodes.Add(1)
		currentResult = -currentResult
	}
}
