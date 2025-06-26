package main

import (
	uttt "uttt/internal/engine"
)

func main() {
	// uttt.OptimizeHash(40, 120)
	// e := uttt.NewEngine()
	// e.SetLimits(*uttt.DefaultLimits().SetDepth(10))
	// e.Think(true)
	cli := uttt.NewCli()
	cli.Start()
}
