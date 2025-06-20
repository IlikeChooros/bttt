package bttt

import "math/rand"

// Hash table implementation
type _HashTable[T any] struct {
	internal map[uint64]T
}

func _NewHashTable[T any]() *_HashTable[T] {
	ht := _HashTable[T]{}
	ht.internal = make(map[uint64]T)
	return &ht
}

func (self *_HashTable[T]) Get(key uint64) (T, bool) {
	val, ok := self.internal[key]
	return val, ok
}

func (self *_HashTable[T]) Set(key uint64, value T) {
	self.internal[key] = value
}

var _hashSmallBoard = [2][9]uint64{} // [0] -> X [1] -> O
var _hashBigBoard = [2][9]uint64{}   // [0] -> X [1] -> O
var _hashTurn uint64

// I will use similar approach to Zobrist hashing:
// Generate random number (a hash) for each square (different for X's and O's) on the regular
// tic tac toe board. If the 'piece' is present XOR with the main hash. The same approach is used
// for the big tic tac toe board. To distinguish the turn, also generate hash for it (if the turn is 'X' then XOR,
// otherwise do nothing)
func initHashing() {
	for i := 0; i < 2; i++ {
		for j := 0; j < 9; j++ {
			_hashSmallBoard[i][j] = rand.Uint64()
			_hashBigBoard[i][j] = rand.Uint64()
		}
	}

	_hashTurn = rand.Uint64()
}

func (pos *Position) Hash() uint64 {
	// for i, state := range pos.bigPositionState {
	// 	if state == Position
	// }

	return pos.hash
}
