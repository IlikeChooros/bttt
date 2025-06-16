package bttt

/*
Main engine class, allowing user to make moves on the board,
search best move, based on given parameteres
*/
type Engine struct {
	position *Position
	limits   Limits
}
