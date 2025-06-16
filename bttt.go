package main

import (
	bttt "bttt/internal"
)

func main() {
	depths := []int{1, 2, 3, 4, 5, 6, 7}

	pos := bttt.NewPosition()
	for _, depth := range depths {
		bttt.Perft(pos, depth)
	}
}
