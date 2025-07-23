package main

import (
	"fmt"
	uttt "uttt/internal/engine"
)

func main() {
	// uttt.OptimizeHash(40, 120)
	// e := uttt.NewEngine(16)
	// e.SetLimits(*uttt.DefaultLimits().SetDepth(10))
	// e.Think(true)
	uttt.Init()
	pos := uttt.NewPosition()
	_ = pos.FromNotation(uttt.StartingPosition)
	mcts := uttt.NewUtttMCTS(*pos)
	mcts.SetLimits(uttt.DefaultLimits().SetThreads(4).SetMbSize(16))
	uttt.AsyncSearch(mcts)

	fmt.Printf("Tree size=%d\n", mcts.Size())

	bestchild := mcts.BestChild(mcts.Root())
	children := mcts.Root().Children
	for i := range len(children) {
		node := &children[i]
		fmt.Printf("\t%d. %s w=%d/v=%d (wr=%.2f) (lr=%.2f) (dr=%.2f)",
			i+1, node.NodeSignature, node.Wins, node.Visits,
			float64(node.Wins)/float64(node.Visits),
			float64(node.Losses)/float64(node.Visits),
			float64(node.Visits-int32(node.Wins)-int32(node.Losses))/float64(node.Visits),
		)

		if children[i].NodeSignature == bestchild.NodeSignature {
			fmt.Println(" (best)")
		} else {
			fmt.Println()
		}
	}

	// cli := uttt.NewCli()
	// cli.Start()
}
