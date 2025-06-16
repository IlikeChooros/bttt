package bttt

import (
	"testing"
)

func TestPositionInit(t *testing.T) {
	p := NewPosition()
	test := func(t *testing.T) {
		if p.stateList.ValidSize() != 0 {
			t.Error("p.stateList.ValidSize() != 0")
		}

		if p.BigIndex() != int(posIndexIllegal) {
			t.Errorf("p.BigIndex()=%d, want=%d", p.BigIndex(), posIndexIllegal)
		}
	}

	t.Run("TestNewPosition", test)

	err := p.FromNotation(StartingPosition)
	if err != nil {
		t.Error(err)
	}

	t.Run("TestFromStartPos", test)
}
