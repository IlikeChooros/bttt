package main

import (
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
	mcts.SetLimits(uttt.DefaultLimits().SetMovetime(1000))
	uttt.AsyncSearch(mcts)

	// cli := uttt.NewCli()
	// cli.Start()
}
