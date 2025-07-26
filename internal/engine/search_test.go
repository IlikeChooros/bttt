package uttt

import (
	"testing"
	"uttt/internal/mcts"
)

func TestMates(t *testing.T) {
	// Test if the search correctly detects not immediate
	// mates (at depth >= 1)

	positions := []string{
		"xxx6/x1x6/xxx6/o3o3o/x1xoxooxo/o3o3o/ooo6/9/9 o 4",
		"xxx6/x1x6/xxx6/o3o3o/xoxoxooxo/o3o3o/ooo6/9/9 x 1",
	}

	// Positive means winning for the X, negative - O
	// for the opponent
	mate_depths := []int{
		2, 1,
	}

	engine := NewEngine()
	engine.SetLimits(mcts.DefaultLimits().SetNodes(400))

	for i, pos := range positions {
		if err := engine.SetNotation(pos); err != nil {
			t.Error(err)
			continue
		}

		result := engine.Think(false)

		if result.ScoreType != MateScore {
			t.Errorf("ScoreType=%d, want=%d (%v, pv=%v)", result.ScoreType, MateScore, result, engine.Pv())
			continue
		}

		if result.Value != mate_depths[i] {
			t.Error("Expected other winning side, got=", result.Value, "want=", mate_depths)
		}
	}
}
