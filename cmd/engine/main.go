package main

import (
	uttt "uttt/internal/engine"
)

func main() {
	// uttt.Init()
	// e := uttt.NewEngine()
	// e.SetLimits(*uttt.DefaultLimits().SetDepth(10))
	// e.Think(true)
	cli := uttt.NewCli()
	cli.Start()
}
