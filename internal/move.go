package bttt

type Move struct {
	From, To posType
}

// Create a new move based on 'from' and 'to' position
func NewMove(from, to posType) *Move {
	return &Move{from, to}
}
