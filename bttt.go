package main

import (
	bttt "bttt/internal"
	"fmt"
)

func main() {
	// depths := []int{1, 2, 3, 4, 5, 6, 7}

	pos := bttt.NewPosition()
	// for _, depth := range depths {
	// 	bttt.Perft(pos, depth)
	// }

	positions := []string{
		bttt.StartingPosition,
		"1o7/2x6/o8/9/9/9/9/9/9 x -",
		"9/9/9/7o1/4ox3/8o/9/4x4/x8 o -",
	}

	for _, notation := range positions {
		err := pos.FromNotation(notation)

		if err != nil {
			fmt.Print(err)
		} else {
			fmt.Printf("%s, eval=%d\n", notation, bttt.Evaluate(pos))
		}
	}
}
