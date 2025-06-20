package bttt

type magics struct {
	mask  uint
	index int
}

// Setup for magic bitboards
func calculateMagics() {
	// What we need to is to calculate every possible hash combination
	// For the regular tic tac toe position (simply a bitboard of the position),
	// then randomly select an integer number and see if we can use it as hash
}
