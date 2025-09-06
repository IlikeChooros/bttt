package ui

import "testing"

func TestManagerMakeMove(t *testing.T) {
	m := NewManager()

	m.updateBoard()
	_ = m.handleCommand("depth 10")
	_ = m.handleCommand("make B2b2")

	moves := m.e.Position().GenerateMoves().Slice()
	_ = m.handleMake([]string{moves[0].String()})
}
