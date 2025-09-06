package ui

import "fmt"

// ANSI escape codes and colors
const (
	RESET = "\033[0m"
	BOLD  = "\033[1m"

	CURSOR_HOME              = "\033[H"
	CLEAR_SCREEN             = "\033[2J"
	CLEAR_SCREEN_FROM_CURSOR = "\033[J"
	CLEAR_LINE               = "\033[K"
	CLEAR_LINE_FROM_CURSOR   = "\033[K"

	// Colors for X and O
	FG_X          = "\033[38;2;157;249;169m" // Light green
	FG_O          = "\033[38;2;247;140;131m" // Light red
	FG_RGB_FORMAT = "\033[38;2;%d;%d;%dm"
	FG_DEFAULT    = "\033[39m"

	BG_DEFAULT    = "\033[49m"
	BG_RGB_FORMAT = "\033[48;2;%d;%d;%dm"
	BG_SELECTED   = "\033[48;2;128;128;128m"
)

func _helperCursorMove(positive, negative rune, n int) string {
	if n == 0 {
		return ""
	}
	if n > 0 {
		return fmt.Sprintf("\033[%d%c", n, positive)
	}
	return fmt.Sprintf("\033[%d%c", -n, negative)
}

func CursorMoveVertical(n int) string {
	return _helperCursorMove('A', 'B', n)
}

func CursorMoveHorizontal(n int) string {
	return _helperCursorMove('C', 'D', n)
}
