package main

import (
	ui "uttt/internal/ui"
)

func main() {
	man := ui.NewManager()
	man.Loop()
}
