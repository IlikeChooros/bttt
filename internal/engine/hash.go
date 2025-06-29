package uttt

import "math/rand"

type HashEntryLike interface {
	comparable
	valid(hash uint64) bool
	depth() int
}

type HashEntryBase struct {
	Depth int
	Hash  uint64
}

// Main hash entry struct, storing depth, hash, score, node type and bestmove.
// Used for transposition table
type TTEntry struct {
	HashEntryBase
	Score    int
	Bestmove PosType
	NodeType EntryNodeType
}

// Interface requirements
func (h HashEntryBase) valid(expected uint64) bool {
	return h.Hash != 0 && h.Hash == expected
}

func (h HashEntryBase) depth() int {
	return h.Depth
}

// Types of the hash entry nodes
const (
	Exact      EntryNodeType = iota // Exact value of the node (a pv node)
	LowerBound                      // It's value if <= alpha
	UpperBound                      // This node caused a beta-cutoff (beta >= alpha)
)

// Hash table implementation
type HashTable[T HashEntryLike] struct {
	internal []T
	size     uint64
}

func NewHashTable[T HashEntryLike](size uint64) *HashTable[T] {
	ht := HashTable[T]{}
	ht.internal = make([]T, size)
	ht.size = size
	return &ht
}

// Internal key hash function, simply modulo applied on the value
func (self *HashTable[T]) _key(value uint64) uint64 {
	return value % self.size
}

// Get the key, returns (value, exists)
func (self *HashTable[T]) Get(key uint64) (T, bool) {
	val := self.internal[self._key(key)]

	// If T is a pointer type, an uninitialized entry will be nil.
	// We must check for this before calling any methods on `val`.
	var zero T
	if val == zero {
		return val, false
	}

	// An entry exists, now check if it's logically valid (Hash != 0 && Hash == key).
	return val, val.valid(key)
}

// Set given key's value
func (self *HashTable[T]) Set(key uint64, value T) {
	val, ok := self.Get(key)

	// Empty, simply set the value
	if !ok {
		self.internal[self._key(key)] = value
	} else {
		// See it current depth is greater, than the value's
		if val.depth() < value.depth() {
			self.internal[self._key(key)] = value
		}
	}
}

// Immiediate replacement strategy
func (self *HashTable[T]) SetForced(key uint64, value T) {
	self.internal[self._key(key)] = value
}

// Set new size, will either shrink the internal buffer, or grow
func (self *HashTable[T]) SetSize(size uint64) {
	if self.size >= size {
		self.internal = self.internal[:size]
	} else {
		internalCopy := make([]T, size)
		copy(internalCopy, self.internal)
		self.internal = nil
		self.internal = internalCopy
	}

	self.size = size
}

// Get the load factor (fraction of non empty entries in the whole table)
func (self *HashTable[T]) LoadFactor() float64 {
	lf := 0.0

	var zero T
	for i, entry := range self.internal {
		if entry == zero {
			lf -= (lf) / (float64(i + 1))
		} else {
			lf += (1 - lf) / (float64(i + 1))
		}
	}

	return lf
}

// Clear the hash table
func (self *HashTable[T]) Clear() {
	var zero T
	for i := range self.internal {
		self.internal[i] = zero
	}
}

var _seedHash int64 = 677881139300273795
var _hashSmallBoard = [2][9][9]uint64{} // [0] -> O [1] -> X (none -> empty square)
var _hashBigPosState = [3][9]uint64{}   // [0] -> O [1] -> X, [2] -> Draw (none -> unresolved)
var _hashTurn uint64                    // Use if turn == CrossTurn
var _hashBigIndex = [9]uint64{}

// Initialize hashes for Zobrist like approach
func _InitHashing() {
	gen := rand.New(rand.NewSource(_seedHash))

	// Hashes for the O's and X's position
	for i := range 2 {
		for j := range 9 {
			for k := range 9 {
				_hashSmallBoard[i][j][k] = gen.Uint64()
			}
		}
		for j := range 9 {
			_hashBigPosState[i][j] = gen.Uint64()
		}
	}

	// Get hashes for 'big index'
	for i := range 9 {
		_hashBigIndex[i] = gen.Uint64()
	}

	_hashTurn = gen.Uint64()
}

// I will use similar approach to Zobrist hashing:
// Generate random number (a hash) for each square (different for X's and O's) on the regular
// tic tac toe board. If the 'piece' is present XOR with the main hash. The same approach is used
// for the big tic tac toe board. To distinguish the turn, also generate hash for it (if the turn is 'X' then XOR,
// otherwise do nothing)
func (pos *Position) Hash() uint64 {
	var hash uint64 = 0

	const (
		OIndex = 0
		XIndex = 1
	)

	// Hash 'big' position state
	for i, state := range pos.bigPositionState {
		if state == PositionUnResolved {
			continue
		}

		stateIndex := XIndex
		if state == PositionCircleWon {
			stateIndex = OIndex
		} else if state == PositionDraw {
			stateIndex = 2
		}

		hash ^= _hashBigPosState[stateIndex][i]
	}

	// Hash all smaller boards state
	for bi := range 9 {
		for si := range 9 {
			if piece := pos.position[bi][si]; piece != PieceNone {
				if piece == PieceCircle {
					hash ^= _hashSmallBoard[OIndex][bi][si]
				} else {
					hash ^= _hashSmallBoard[XIndex][bi][si]
				}
			}
		}
	}

	// Hash turn
	if pos.Turn() == CrossTurn {
		hash ^= _hashTurn
	}

	// Hash big Index
	if pos.BigIndex() != PosIndexIllegal {
		hash ^= _hashBigIndex[pos.BigIndex()]
	}

	return hash
}
