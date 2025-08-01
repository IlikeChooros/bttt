package uttt

import (
	"fmt"
	"strings"
	"testing"
	"uttt/internal/mcts"
)

func TestMates(t *testing.T) {
	// Test if the search correctly detects not immediate
	// mates (at depth >= 1)

	positions := []string{
		"xxx6/x1x6/xxx6/o3o3o/x1xoxooxo/o3o3o/ooo6/9/9 o 4",
		"xxx6/x1x6/xxx6/o3o3o/xoxoxooxo/o3o3o/ooo6/9/9 x 1",
		"xox1oo3/ooxox1xoo/1xo2o2o/1x2x1ox1/1x1ooo1x1/oxxox1o2/xxx3o2/1oo1xxo1o/x1xx2xxo x -",
	}

	// Positive means winning for the X, negative - O
	// for the opponent
	mate_depths := []int{
		2, 1, 1,
	}

	engine := NewEngine()
	engine.SetLimits(mcts.DefaultLimits().SetCycles(100))

	for i, pos := range positions {
		t.Run(fmt.Sprintf("Mates-%s", strings.ReplaceAll(pos, "/", "|")), func(t *testing.T) {
			if err := engine.SetNotation(pos); err != nil {
				t.Error(err)
				return
			}

			result := engine.Think()

			if result.ScoreType != MateScore {
				t.Errorf("ScoreType=%d, want=%d (%v, pv=%v)", result.ScoreType, MateScore, result, engine.Pv())
			}

			if result.Value != mate_depths[i] {
				t.Error("Expected other winning side, got=", result.Value, "want=", mate_depths[i])
			}
		})
	}
}
