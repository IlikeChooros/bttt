package uttt

import (
	"math"
	"math/rand"
)

// Actual UTTT mcts implementation
type UtttMCTS struct {
	MCTS[PosType]
}

const (
	explorationParam = 1.41
)

// Default selection of the node policy
var DefaultSelection SelectionPolicy[PosType] = func(parent *NodeBase[PosType]) *NodeBase[PosType] {

	max := float64(-1)
	index := 0

	for i, node := range parent.Children {
		// UCB 1 : visits/wins + C * sqrt(ln(parent_visits)/visits)
		ucb := float64(node.Visits)/float64(node.Wins) +
			explorationParam*math.Sqrt(math.Log(float64(node.Parent.Visits))/float64(node.Visits))

		if ucb > max {
			max = ucb
			index = i
		}
	}

	return &parent.Children[index]
}

func NewUtttMCTS() *UtttMCTS {
	return &UtttMCTS{
		MCTS: *NewMTCS(
			DefaultSelection,
			GameOperations[PosType](&UtttOperations{}),
		),
	}
}

type UtttOperations struct {
	position Position
}

func (mcts *UtttOperations) ExpandNode(node *NodeBase[PosType]) {
	moves := mcts.position.GenerateMoves()
	node.Children = make([]NodeBase[PosType], moves.size)

	for i, m := range moves.Slice() {
		mcts.position.MakeMove(m)
		isTerminal := mcts.position.IsTerminated()
		mcts.position.UndoMove()

		node.Children[i] = NodeBase[PosType]{
			NodeStats: NodeStats{
				Wins: 0, Visits: 0,
			},
			NodeSignature: m,
			Children:      nil,
			Parent:        node,
			Terminal:      isTerminal,
		}
	}
}

func (mcts *UtttOperations) Traverse(signature PosType) {
	mcts.position.MakeMove(signature)
}

func (mcts *UtttOperations) BackTraverse() {
	mcts.position.UndoMove()
}

// Play the game until a terminal node is reached
func (mcts *UtttOperations) Rollout() Result {
	var moves *MoveList
	var move PosType
	var result Result = 0
	var moveCount int = 0
	ourSide := mcts.position.Turn()

	for !mcts.position.IsTerminated() {
		moveCount++
		moves = mcts.position.GenerateMoves()

		// Choose at random move
		move = moves.moves[rand.Int31n(int32(moves.size))]
		mcts.position.MakeMove(move)
	}

	// If that's not a draw
	if mcts.position.termination != TerminationDraw {
		// We lost
		if mcts.position.Turn() == ourSide {
			result = -1
		} else {
			// We won
			result = 1
		}
	}

	// Undo the moves
	for range moveCount {
		mcts.position.UndoMove()
	}

	return result
}
