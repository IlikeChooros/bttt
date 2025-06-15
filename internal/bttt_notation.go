package bttt

import (
	"fmt"
	"strings"
)

// string notation for the big tic tac toe position
// Much like the FEN representation of a chessboard
func (p *Position) Notation() string {
	builder := strings.Builder{}

	board := p.position
	for rowIndex, row := range board {

		// In each row, we will loop through it, and append
		counter := 0
		for i := 0; i < 9; i++ {
			switch rowPiece := row[i]; rowPiece {
			case PieceCircle, PieceCross:
				// Write the counter, and current piece
				if counter > 0 {
					builder.WriteString(fmt.Sprintf("%d", counter))
					counter = 0
				}

				strPiece := "o"
				if rowPiece == PieceCross {
					strPiece = "x"
				}

				builder.WriteString(strPiece)
			default:
				// No piece on this square
				counter += 1
			}
		}

		// Check the counter
		if counter > 0 {
			builder.WriteString(fmt.Sprintf("%d", counter))
		}

		if rowIndex != 8 {
			builder.WriteString("/")
		}
	}

	return builder.String()
}

func (p *Position) FromNotation(notation string) {
	_FromNotation(p, notation)
}

// Create from notation position
func FromNotation(notation string) *Position {
	pos := NewPosition()
	_FromNotation(pos, notation)
	return pos
}

// Assign this position (from notation string) to given position object
func _FromNotation(pos *Position, notation string) {
	board := pos.position

	bigIndex := 0
	smallIndex := 0

	// Loop through the notation
	for _, v := range notation {
		switch v {
		case 'x', 'o':
			// If that's a piece, put it on the board
			board[bigIndex][smallIndex] = PieceFromRune(v)
			smallIndex++
		case '/':
			// New square, increase the bigIndex counter
			bigIndex++
			smallIndex = 0
		default:
			// Number, meaning skip given number of squares
			smallIndex += int(v - '0')
		}
	}
}
