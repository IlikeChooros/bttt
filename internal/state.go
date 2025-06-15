package bttt

type BoardState struct {
	move     Move
	turn     turnType
	bigIndex posType
}

// Stores the history of the board as a slice of BoardState
type StateList struct {
	list []BoardState
}

// Get new StateList object
func NewStateList() *StateList {
	sl := new(StateList)
	sl.list = make([]BoardState, 0, 10)
	return sl
}

// Append new state
func (sl *StateList) Append(state BoardState) {
	sl.list = append(sl.list, state)
}

// Remove last state
func (sl *StateList) Remove() {
	sl.list = sl.list[:len(sl.list)-1]
}
