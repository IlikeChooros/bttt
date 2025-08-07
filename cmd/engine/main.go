package main

import (
	"fmt"
	// "log"
	uttt "uttt/internal/engine"
	"uttt/internal/mcts"
	// "net/http"
	// _ "net/http/pprof"
)

func main() {
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()
	// pos := uttt.NewPosition()
	// _ = pos.FromNotation(uttt.StartingPosition)
	// mcts := uttt.NewUtttMCTS(*pos)
	// mcts.SetLimits(*uttt.DefaultLimits().SetThreads(4).SetMbSize(32))
	// mcts.Search()
	engine := uttt.NewEngine()
	engine.SetLimits(mcts.DefaultLimits().SetThreads(4).SetDepth(13).SetMbSize(16).SetCycles(10000).SetMultiPv(3))
	// engine.SetNotation("1xxx1xxx1/oo1o1xox1/oxox1x1ox/oo1x1o1oo/o1o1x2x1/o1o2ox1o/oxxx2o2/1xox1xoo1/2xxx3o o 3")
	engine.SetNotation("8o/9/x8/9/6x2/9/2o6/9/x8 o 0")

	// for {
	fmt.Println(engine.Think())
	fmt.Println(engine.MultiPv())
	// }

	// engine.Position().MakeMove(uttt.MoveFromString("A2b2"))
	// engine.SetNotation(engine.Position().Notation())
	// fmt.Println(engine.Think())

	// tree := engine.Mcts()

	// size, nodesize := tree.Size(), uint32(unsafe.Sizeof(uttt.UtttNode{}))
	// fmt.Printf("Tree size=%d - %d bytes\n", size, size*nodesize)

	// bestchild := tree.BestChild(tree.Root, mcts.BestChildWinRate)
	// children := tree.Root.Children
	// for i := range len(children) {
	// 	node := &children[i]
	// 	fmt.Printf("\t%d. %s w=%d/v=%d (wr=%.2f) (lr=%.2f) (dr=%.2f)",
	// 		i+1, node.NodeSignature, node.Wins.Load(), node.Visits(),
	// 		float64(node.Wins.Load())/float64(node.Visits()),
	// 		float64(node.Losses.Load())/float64(node.Visits()),
	// 		float64(node.Visits()-int32(node.Wins.Load())-int32(node.Losses.Load()))/float64(node.Visits()),
	// 	)

	// 	if bestchild != nil && children[i].NodeSignature == bestchild.NodeSignature {
	// 		fmt.Println(" (best)")
	// 	} else {
	// 		fmt.Println()
	// 	}
	// }

	// cli := uttt.NewCli()
	// cli.Start()
}
