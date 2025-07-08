package uttt

import (
	"testing"
)

func TestMates(t *testing.T) {
	// Test if the search correctly detects not immediate
	// mates (at depth > 1)

	positions := []string{
		"Xxxx6/x1x6/Xxxx6/Oo3o3o/x1xoxooxo/Oo3o3o/Oooo6/9/9 o 4",
	}

	// Positive means winning for the X, negative - O
	// for the opponent
	mate_depths := []int{
		2,
	}

	engine := NewEngine()
	engine.SetLimits(*DefaultLimits().SetDepth(6))

	for i, pos := range positions {
		if err := engine.Position().FromNotation(pos); err != nil {
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
