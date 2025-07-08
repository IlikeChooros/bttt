package uttt

import (
	"fmt"
	"strings"
)

// Number of sections, seperated by space in the notation
const _notationNumberOfSections int = 3

// string notation for the big tic tac toe position
// Much like the FEN representation of a chessboard
// Will result in something like this:
//
//	X/X/X/X/X/X/X/X/X <turn> <big index>
//
// where `X` is one small square string, saves board
// position's data, same as FEN, but instead of chess pieces
// we have got 'o' and 'x'
//
// For example, let X be:
//
//	o | x | x
//
// ----------
//
//	x | o |
//
// ----------
//
//	o |   |
//
// then X format string would be:
//
//	oxxxo1o2
//
// However, if this board is terminated, we add at the begining
// a flag indicating it's state (see that in the example above there is no termination flag)
//
//	o | x | x
//
// ----------
//
//	o | x |
//
// ----------
//
//	o | x |
//
// Here both O and X could win, however it only counts for the side that achieved it first
// Format of this position could be:
//
//	Xoxxox1ox1
//	Ooxxox1ox1
//
// In the first case, the X would win on that square, and on the other one - O.
// If there is a draw - D flag is used
//
// <turn> - either 'o' or 'x'
//
// <big index> - where should current player make move on the
// big plane, it is an integer between 0 and 9, or - if player can move anywhere
//
// Examples:
//
// * 9/9/9/9/9/9/9/9/9 x -
//
// * 9/9/9/7x1/4xo3/8x/9/4o4/o8 x 0
func (p *Position) Notation() string {
	builder := strings.Builder{}

	board := p.position
	for rowIndex, row := range board {

		// In each row, we will generate the small square string
		counter := 0
		switch p.bigPositionState[rowIndex] {
		case PositionCircleWon:
			builder.WriteByte('O')
		case PositionCrossWon:
			builder.WriteByte('X')
		case PositionDraw:
			builder.WriteByte('D')
		}

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

	// Add the turn
	builder.WriteByte(' ')
	if p.Turn() == CircleTurn {
		builder.WriteByte('o')
	} else {
		builder.WriteByte('x')
	}

	// Add the BigIndex
	builder.WriteByte(' ')
	if p.BigIndex() == PosIndexIllegal {
		builder.WriteByte('-')
	} else {
		builder.WriteByte('0' + byte(p.BigIndex()))
	}

	return builder.String()
}

// Create the position from given notation string, will reset current state,
// load current position and setup termination flags
func (p *Position) FromNotation(notation string) error {
	// Reset history and generated moves
	p.Reset()

	if notation == "startpos" {
		notation = StartingPosition
	}

	return _FromNotation(p, notation)
}

// Create from notation position
func FromNotation(notation string) *Position {
	pos := NewPosition()
	_ = _FromNotation(pos, notation) // ignore the error return value
	return pos
}

// Assign this position (from notation string) to given position object
func _FromNotation(pos *Position, notation string) error {
	// TODO: make this more robust
	board := &pos.position

	bigIndex := 0
	smallIndex := 0

	// Assert we have a valid structure
	const numSlash = 8
	slashCounter := 0
	seprarationIndexes := [2]int{-1, -1}
	for i, v := range notation {
		if v == '/' {
			slashCounter++
		} else if v == ' ' && slashCounter == numSlash {
			if seprarationIndexes[0] == -1 {
				seprarationIndexes[0] = i
			} else {
				seprarationIndexes[1] = i
			}
		}
	}

	// Check the counters
	if slashCounter != numSlash || seprarationIndexes[0] == -1 {
		return fmt.Errorf(
			"Invalid notation structure, expected %d slashes and 1 separation, got = %d",
			numSlash, slashCounter,
		)
	}

	// Loop through the notation
	for i := 0; i < seprarationIndexes[0]; i++ {
		switch v := rune(notation[i]); v {
		case 'x', 'o':
			// If that's a piece, put it on the board
			board[bigIndex][smallIndex] = PieceFromRune(v)
			smallIndex++
		case 'D':
			pos.bigPositionState[bigIndex] = PositionDraw
		case 'O':
			pos.bigPositionState[bigIndex] = PositionCircleWon
		case 'X':
			pos.bigPositionState[bigIndex] = PositionCrossWon
		case '-':
			pos.bigPositionState[bigIndex] = PositionUnResolved
		case '/':
			// Small index must be 9, before moving on to next square
			if smallIndex != 9 {
				return fmt.Errorf("Invalid number of squares within bigIndex=%d", bigIndex)
			}

			// New square, increase the bigIndex counter
			bigIndex++
			smallIndex = 0
		default:

			if '0' <= v && v <= '9' {
				// Number, meaning skip given number of squares
				smallIndex += int(v - '0')

				if smallIndex > 9 {
					return fmt.Errorf("Invalid number of skip squares %d, at index = %d", smallIndex, i)
				}

			} else {
				return fmt.Errorf("Invalid notation = %s, at token = %d (%c)", notation, i, v)
			}
		}
	}

	// Read the side
	if v := notation[seprarationIndexes[0]+1]; v == 'o' || v == 'x' {
		pos.stateList.Last().turn = v != 'x'
	} else {
		return fmt.Errorf("Invalid side character %c", v)
	}

	// Read the big index counter
	if v := notation[seprarationIndexes[1]+1]; v >= '0' && v <= '8' {
		// Set NextBigIndex to given value v
		pos.stateList.Last().move = MakeMove(0, int(v-'0'))
	} else if v != '-' {
		return fmt.Errorf("Invalid big index %c, expected a digit 0-8", v)
	}

	// Setup the position state
	pos.SetupBoardState()
	pos.CheckTerminationPattern()
	return nil
}
