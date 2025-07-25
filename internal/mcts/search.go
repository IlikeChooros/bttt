package mcts

import (
	"math/rand"
	"sync/atomic"
)

// Use when started multi-threaded search and want it to synchronize with this thread
func (mcts *MCTS[T]) Synchronize() {
	mcts.wg.Wait()
}

// Run multi-treaded search, to wait for the result, call Synchronize
func (mcts *MCTS[T]) SearchMultiThreaded(ops GameOperations[T]) {
	mcts.setupSearch()
	threads := max(1, mcts.Limiter.Limits().NThreads)

	for range threads {
		mcts.wg.Add(1)
		go mcts.Search(ops.Clone())
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
// 2. rollout - to simulate the user-defined game, and get the result of the playout
//
// 3. backpropagate - to increment counters up to the root
//
// Until runs out of the allocated time, nodes, or memory
func (mcts *MCTS[T]) Search(ops GameOperations[T]) {
	defer mcts.wg.Done()

	if mcts.Root.Terminal() {
		return
	}

	var node *NodeBase[T]

	for mcts.Limiter.Ok(mcts.Nodes(), mcts.Size(), uint32(mcts.MaxDepth())) {

		// Choose the most promising node
		node = mcts.Selection(ops)
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
func (mcts *MCTS[T]) Selection(ops GameOperations[T]) *NodeBase[T] {
	node := mcts.Root
	depth := 0

	// Currently expanding
	for node.state.Load() == 1 {
	}

	for len(node.Children) > 0 && node.state.Load() == 2 {
		node = mcts.selection_policy(node)
		ops.Traverse(node.NodeSignature)
		depth++
		mcts.nodes.Add(1)
		// Apply virtual loss
		// atomic.AddInt32(&node.Visits, virtualLoss)
	}

	// Add new children to this node, after finding leaf node
	if atomic.LoadInt32(&node.Visits) > 0 && !node.Terminal() {
		// Expand the node, only if needed (expand flag is 0)
		if mcts.Limiter.Expand() && node.state.CompareAndSwap(0, 1) {
			mcts.size.Add(ops.ExpandNode(node))
			// Now it's state
			node.state.Store(2)
		}

		// Currently expanding
		for node.state.Load() == 1 {
		}

		// Already set
		if len(node.Children) > 0 && node.state.Load() == 2 {
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

// Increment the counters (wins/losses/visits) along the tree path
func (mcts *MCTS[T]) Backpropagate(ops GameOperations[T], node *NodeBase[T], result Result) {
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
