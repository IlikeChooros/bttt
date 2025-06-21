package bttt

type BoardState struct {
	move PosType
	turn TurnType
}

// Stores the history of the position as a slice of BoardState
type StateList struct {
	list []BoardState
}

// Get new StateList object
func NewStateList() *StateList {
	sl := new(StateList)
	sl.Init()
	return sl
}

// Initialize the state list, for example after calling 'Clear'
func (sl *StateList) Init() {
	sl.list = make([]BoardState, 0, 10)
	sl.Append(posIllegal, CircleTurn)
}

// Append new state
func (sl *StateList) Append(move PosType, turn TurnType) {
	sl.list = append(sl.list, BoardState{move, turn})
}

// Reset all states (remove them)
func (sl *StateList) Clear() {
	sl.list = nil
	sl.Init()
}

// Remove last state
func (sl *StateList) Remove() {
	sl.list = sl.list[:len(sl.list)-1]
}

// Get actual size of the history
func (sl *StateList) ValidSize() int {
	return len(sl.list) - 1
}

// Get the last element of the state list (current state of the board)
func (sl *StateList) Last() *BoardState {
	return &sl.list[len(sl.list)-1]
}

// Get last move's Big Index
func (sl *StateList) BigIndex() PosType {
	return sl.Last().move.BigIndex()
}

// Same as SmallIndex
func (sl *StateList) NextBigIndex() PosType {
	return sl.Last().move.SmallIndex()
}
