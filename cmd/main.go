package main

import (
	bttt "bttt/internal"
	// "fmt"
)

func main() {
	// engine := bttt.NewEngine()
	// engine.SetLimits(*bttt.DefaultLimits().SetDepth(5))
	// pos := engine.Position()

	// positions := []string{
	// 	bttt.StartingPosition,
	// 	"1o7/2x6/o8/9/9/9/9/9/9 x -",
	// 	"9/9/9/7o1/4ox3/8o/9/4x4/x8 o -",
	// }

	// for _, notation := range positions {
	// 	err := pos.FromNotation(notation)

	// 	if err != nil {
	// 		fmt.Print(err)
	// 	} else {
	// 		// Show total number of nodes
	// 		n := bttt.Perft(engine.Position(), 5)
	// 		res := engine.Search()

	// 		// Percentage:
	// 		fmt.Printf("nodes searched %.4f\n", float32(res.Nodes)/float32(n))
	// 	}
	// }

	cli := bttt.NewCli()
	cli.Start()

}
