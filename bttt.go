package main

import (
	"bttt/internal"
	"fmt"
)

func main() {
	p := bttt.NewPosition()

	fmt.Println(p.Moves())
	fmt.Println(p.Position())

	p.GenerateMoves()
	fmt.Println(p.Moves())

	fmt.Println(p.Notation())
}
