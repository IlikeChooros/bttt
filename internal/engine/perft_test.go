package uttt

import (
	"testing"
)

func BenchmarkPerft(b *testing.B) {
	pos := NewPosition()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Perft(pos, (i&3)+1, false, false)
	}
}
