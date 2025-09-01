package main

import (
	"fmt"
	uttt "uttt/internal/engine"
	"uttt/internal/mcts"
)

func main() {

	engine := uttt.NewEngine()
	engine.SetLimits(mcts.DefaultLimits().SetThreads(4).SetDepth(13).SetMbSize(16).SetMultiPv(3))
	engine.SetNotation("8o/9/x8/9/6x2/9/2o6/9/x8 o 0")

	fmt.Println(engine.Think())
	fmt.Println(engine.MultiPv())

	fmt.Printf("Collisions=%d (%.4f%%)\n",
		engine.Mcts().CollisionCount(),
		engine.Mcts().CollisionFactor()*100,
	)

	/*

		Before:

		UtttNode size = 88 bytes
		eval 0.45 depth 8 nps 3140824 nodes 1020768 cycles 83868 pv [A3b1 B1b2 B2c1 C1a1 A1a1 A1b2 B2b3 B3c3]
		[{0xc0000b2270 [A3b1 B1b2 B2c1 C1a1 A1a1 A1b2 B2b3 B3c3] false false} {0xc0000b2218 [A3a1 A1b2 B2a3 A3a2 A2c1 C1c1 C1a2 A2b2] false false} {0xc0000b2168 [A3b2 B2c3 C3a1 A1c2 C2b2 B2a2 A2a3] false false}]
		Collisions=29 (0.0346%)


		After:
		UtttNode size = 64 bytes
		eval 0.46 depth 8 nps 5169182 nodes 1530078 cycles 121080 pv [A3b1 B1b1 B1b3 B3b3 B3a3 A3b3 B3c1 C1b3]
		[{0xc0000b41c0 [A3b1 B1b1 B1b3 B3b3 B3a3 A3b3 B3c1 C1b3] false false} {0xc0000b4080 [A3c3 C3b2 B2b1 B1c1 C1c1 C1a1 A1b3 B3a3] false false} {0xc0000b4180 [A3a1 A1c1 C1b1 B1b3 B3b2 B2c3 C3b3] false false}]
		Collisions=1074 (0.8870%)
	*/

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
