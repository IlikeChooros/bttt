package main

import (
	"fmt"
	"unsafe"
	uttt "uttt/internal/engine"
	"uttt/internal/mcts"
)

func main() {
	// uttt.OptimizeHash(40, 120)
	// e := uttt.NewEngine(16)
	// e.SetLimits(*uttt.DefaultLimits().SetDepth(10))
	// e.Think(true)
	// uttt.Init()
	// pos := uttt.NewPosition()
	// _ = pos.FromNotation(uttt.StartingPosition)
	// mcts := uttt.NewUtttMCTS(*pos)
	// mcts.SetLimits(*uttt.DefaultLimits().SetThreads(4).SetMbSize(32))
	// mcts.Search()
	engine := uttt.NewEngine()
	engine.SetLimits(mcts.DefaultLimits().SetThreads(8).SetDepth(9))
	fmt.Println(engine.Think())
	tree := engine.Mcts()

	size, count, nodesize := tree.Size(), tree.Count(), uint32(unsafe.Sizeof(uttt.UtttNode{}))
	fmt.Printf("Tree size=%d - %d bytes (count=%d - %d bytes)\n", size, size*nodesize, count, count*int(nodesize))

	bestchild := tree.BestChild(tree.Root, mcts.BestChildWinRate)
	children := tree.Root.Children
	for i := range len(children) {
		node := &children[i]
		fmt.Printf("\t%d. %s w=%d/v=%d (wr=%.2f) (lr=%.2f) (dr=%.2f)",
			i+1, node.NodeSignature, node.Wins, node.Visits,
			float64(node.Wins)/float64(node.Visits),
			float64(node.Losses)/float64(node.Visits),
			float64(node.Visits-int32(node.Wins)-int32(node.Losses))/float64(node.Visits),
		)

		if bestchild != nil && children[i].NodeSignature == bestchild.NodeSignature {
			fmt.Println(" (best)")
		} else {
			fmt.Println()
		}
	}

	// cli := uttt.NewCli()
	// cli.Start()
}
