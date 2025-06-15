package main

import (
	"bttt/internal"
	"fmt"
)

func main() {
	b := bttt.NewBoard()
	b.MakeMove(0, 5)

	fmt.Println(b.Board())
}
