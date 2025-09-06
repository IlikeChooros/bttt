package ui

import (
	"fmt"
	uttt "uttt/_pkg/engine"
)

// UltimateBoard represents a 9x9 board for Ultimate Tic Tac Toe.
// Each cell is either 'X', 'O', or ' ' for empty.
type UltimateBoard struct {
	Cells          [9][9]rune
	BigIndex       int
	BigSquareColor [9]string
}

func (board *UltimateBoard) SetColors(states [9]uttt.PositionState) {
	// If unresolved, set the DEFAULT_BG
	for i, state := range states {
		switch state {
		case uttt.PositionCircleWon:
			board.BigSquareColor[i] = fmt.Sprintf(BG_RGB_FORMAT, 201, 140, 141)
		case uttt.PositionCrossWon:
			board.BigSquareColor[i] = fmt.Sprintf(BG_RGB_FORMAT, 197, 222, 164)
		case uttt.PositionDraw:
			board.BigSquareColor[i] = fmt.Sprintf(BG_RGB_FORMAT, 97, 97, 97)
		case uttt.PositionUnResolved:
			board.BigSquareColor[i] = BG_DEFAULT
		}
	}

	if board.BigIndex != -1 {
		board.BigSquareColor[board.BigIndex] = BG_SELECTED
	}
}

// RenderBoard clears the terminal and renders the Ultimate Tic Tac Toe board.
func (board *UltimateBoard) RenderBoard() {
	// Clear screen and move cursor to home to prevent flicker.
	fmt.Print(CURSOR_HOME)
	// fmt.Print(CLEAR_SCREEN)

	// Print header row
	fmt.Println("   a  b  c a  b  c a  b  c")
	horizontalDivider := "+-------+-------+-------+"

	offsetBigIndex, offsetSmallIndex := 0, 0
	for i := 0; i < 9; i++ {
		// Print horizontal divider every 3 rows.
		offsetSmallIndex += 3
		if i%3 == 0 {
			fmt.Println("  ", horizontalDivider, CLEAR_LINE_FROM_CURSOR)
			if i != 0 {
				offsetBigIndex += 3
			}
			offsetSmallIndex = 0
		}

		// Row label (1-indexed)
		line := fmt.Sprintf("%d |", 3-(i%3))
		for j := 0; j < 9; j++ {
			// Add vertical divider every 3 columns.
			if j != 0 && j%3 == 0 {
				line += " " + BG_DEFAULT + "|"
			}

			bigIndex, smallIndex := offsetBigIndex+(j/3), offsetSmallIndex+(j%3)
			bckgr := board.BigSquareColor[bigIndex]
			cell := board.Cells[bigIndex][smallIndex]
			symbol := " "
			switch cell {
			case 'X':
				symbol = FG_X + "X" + RESET
			case 'O':
				symbol = FG_O + "O" + RESET
			}
			line += bckgr + " " + symbol + bckgr
		}
		line += " " + BG_DEFAULT + "|"
		if i%3 == 1 {
			line += fmt.Sprintf(" %d", 3-(i/3))
		}
		line += CLEAR_LINE_FROM_CURSOR
		fmt.Println(line)
	}
	// Final horizontal divider
	fmt.Println("  ", horizontalDivider, CLEAR_LINE_FROM_CURSOR)
	fmt.Println("      A       B       C", CLEAR_LINE_FROM_CURSOR)
}

// Print prints a message with ANSI format without appending a newline.
func Print(msg string) {
	fmt.Print(msg)
}

// PrintError prints an error message with predefined error colors.
func PrintError(mainMsg, desc string) {
	// Using red color for errors (using FG_O as error color).
	fmt.Print("\r", FG_O, mainMsg, " ", desc, RESET, CLEAR_LINE, CursorMoveVertical(1))
}

// PrintOK prints an OK message with predefined ok colors.
func PrintOK(mainMsg, desc string) {
	// Using green color for OK messages (using FG_X as ok color).
	fmt.Print("\r", FG_X, mainMsg, " ", desc, RESET, CLEAR_LINE, CursorMoveVertical(1))
}
