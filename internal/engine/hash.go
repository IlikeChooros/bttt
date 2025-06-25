package uttt

import "math/rand"

type _HashEntry interface {
	empty() bool
}

type HashEntry struct {
	depth    int
	hash     uint64
	score    int
	nodeType EntryNodeType
	bestmove PosType
}

func (h HashEntry) empty() bool {
	return h.hash == 0
}

const (
	Exact      EntryNodeType = iota // Exact value of the node (a pvs node)
	LowerBound                      // It's value if <= alpha
	UpperBound                      // This node caused a beta-cutoff (beta >= alpha)
)

// Hash table implementation
type _HashTable[T _HashEntry] struct {
	internal []T
	size     uint64
}

func _NewHashTable[T _HashEntry](size uint64) *_HashTable[T] {
	ht := _HashTable[T]{}
	ht.internal = make([]T, size)
	ht.size = size
	return &ht
}

func (self *_HashTable[T]) _key(value uint64) uint64 {
	return value % self.size
}

func (self *_HashTable[T]) Get(key uint64) (T, bool) {
	val := self.internal[self._key(key)]
	return val, !val.empty()
}

func (self *_HashTable[T]) Set(key uint64, value T) {
	self.internal[self._key(key)] = value
}

func (self *_HashTable[T]) SetSize(size uint64) {
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

var _hashSmallBoard = [2][9][9]uint64{} // [0] -> X [1] -> O (none -> empty square)
var _hashBigBoard = [3][9]uint64{}      // [0] -> X [1] -> O, [2] -> Draw (none -> unresolved)
var _hashTurn uint64
var _hashBigIndex = [9]uint64{}

// Initialize hashes for Zobrist like approach
func _InitHashing() {
	gen := rand.New(rand.NewSource(27))

	// Hashes for the O's and X's position
	for i := 0; i < 2; i++ {
		for j := 0; j < 9; j++ {
			for k := 0; k < 9; k++ {
				_hashSmallBoard[i][j][k] = gen.Uint64()
			}
		}
		for j := 0; j < 9; j++ {
			_hashBigBoard[i][j] = gen.Uint64()
		}
	}

	// Get hashes for 'big index'
	for i := 0; i < 9; i++ {
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
		XIndex = 0
		OIndex = 1
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

		hash ^= _hashBigBoard[stateIndex][i]
	}

	// Hash all smaller boards state
	for bi := 0; bi < 9; bi++ {
		for si := 0; si < 9; si++ {
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
	if pos.BigIndex() != int(PosIndexIllegal) {
		hash ^= _hashBigIndex[pos.BigIndex()]
	}

	return hash
}
