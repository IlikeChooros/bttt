package bttt

import (
	"testing"
)

func TestPositionInit(t *testing.T) {
	p := NewPosition()

	if p.stateList.ValidSize() != 0 {
		t.Error("p.stateList.ValidSize() != 0")
	}
}
