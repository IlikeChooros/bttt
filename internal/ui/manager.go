package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	engine "uttt/internal/engine"
)

// Manager handles user input and coordinates the engine and UI rendering.
type Manager struct {
	e      *engine.Engine
	board  *UltimateBoard
	limits *engine.Limits
}

// NewManager constructs a new Manager, initializes the engine, and sets up a board.
func NewManager() *Manager {
	engine.Init()
	m := &Manager{
		e:      engine.NewEngine(),
		board:  &UltimateBoard{},
		limits: engine.DefaultLimits().SetMovetime(1000),
	}
	return m
}

// Loop starts reading commands from stdin until user quits.
func (m *Manager) Loop() {
	scanner := bufio.NewScanner(os.Stdin)
	quitCmds := map[string]bool{"quit": true, "q": true, "exit": true, "e": true}

	// Initial render
	fmt.Print(CLEAR_SCREEN)
	m.updateBoard()
	m.board.RenderBoard()

	fmt.Println("[Manager] Enter commands (make, undo, getpos, eval, depth, nodes, movetime, infinite, stop). 'quit' to exit.")
	for {
		fmt.Print("\r> ", CLEAR_LINE_FROM_CURSOR)
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if quitCmds[line] {
			fmt.Println("Exiting manager loop.", CLEAR_LINE_FROM_CURSOR)
			break
		}
		if err := m.handleCommand(line); err != nil {
			PrintError("[Manager]", err.Error())
		}
	}
}

// handleCommand parses and executes a single line of user input.
func (m *Manager) handleCommand(line string) error {
	if line == "" {
		return nil
	}
	tokens := strings.Split(line, " ")

	switch tokens[0] {
	case "make": // e.g. "make A1B3"
		if len(tokens) < 2 {
			return fmt.Errorf("Usage: make <move>")
		}
		return m.handleMake(tokens[1:])
	case "undo":
		m.e.Position().UndoMove()
		m.updateBoard()
		m.board.RenderBoard()
	case "getpos":
		fmt.Print(m.e.Position().Notation(), CLEAR_LINE_FROM_CURSOR, CursorMoveVertical(1))
	case "eval": // Let the engine search and play the best move.
		m.e.SetLimits(*m.limits)
		result := m.e.Think(false)
		if result.Bestmove != engine.PosIllegal && m.e.Position().IsLegal(result.Bestmove) {
			m.e.Position().MakeMove(result.Bestmove)
			defer PrintOK("[Engine]", fmt.Sprintf("Engine played %v", result.Bestmove))
		} else {
			defer PrintError("[Engine]", "No valid best move found.")
		}
		m.updateBoard()
		m.board.RenderBoard()
	case "depth":
		return m.setDepth(tokens)
	case "nodes":
		return m.setNodes(tokens)
	case "movetime":
		return m.setMovetime(tokens)
	case "infinite":
		m.limits.SetInfinite(true)
		PrintOK("[Manager]", "Search set to infinite.")
	default:
		return fmt.Errorf("Unknown command: %s", tokens[0])
	}
	return nil
}

func (m *Manager) makeMove(move engine.PosType) {
	m.e.Position().MakeMove(move)
	if m.e.Position().IsTerminated() {
		defer PrintError("[Manager]", "Position is terminated")
	}
	m.updateBoard()
	m.board.RenderBoard()
}

// handleMake attempts to play user-specified moves, then waits briefly for engine response.
func (m *Manager) handleMake(moves []string) error {
	for _, mvtxt := range moves {
		mv := engine.MoveFromString(mvtxt)
		if mv == engine.PosIllegal {
			return fmt.Errorf("Invalid move notation: %s", mvtxt)
		}
		if !m.e.Position().IsLegal(mv) {
			return fmt.Errorf("Illegal move: %s", mvtxt)
		}
		m.makeMove(mv)
	}
	fmt.Print(CLEAR_SCREEN_FROM_CURSOR)

	// Let engine respond automatically
	m.e.SetLimits(*m.limits)
	res := m.e.Think(false)
	if res.Bestmove != engine.PosIllegal && m.e.Position().IsLegal(res.Bestmove) {
		m.makeMove(res.Bestmove)
	}
	fmt.Print(CLEAR_SCREEN_FROM_CURSOR)

	return nil
}

// setDepth updates engine depth limit.
func (m *Manager) setDepth(tokens []string) error {
	if len(tokens) < 2 {
		return fmt.Errorf("Usage: depth <n>")
	}
	depth, err := strconv.Atoi(tokens[1])
	if err != nil {
		return err
	}
	m.limits.SetDepth(depth)
	PrintOK("[Manager]", fmt.Sprintf("Depth set to %d.", depth))
	return nil
}

// setNodes updates engine node limit.
func (m *Manager) setNodes(tokens []string) error {
	if len(tokens) < 2 {
		return fmt.Errorf("Usage: nodes <n>")
	}
	nodes, err := strconv.Atoi(tokens[1])
	if err != nil {
		return err
	}
	m.limits.SetNodes(uint64(nodes))
	PrintOK("[Manager]", fmt.Sprintf("Nodes set to %d.", nodes))
	return nil
}

// setMovetime updates engine move time limit in milliseconds.
func (m *Manager) setMovetime(tokens []string) error {
	if len(tokens) < 2 {
		return fmt.Errorf("Usage: movetime <ms>")
	}
	mt, err := strconv.Atoi(tokens[1])
	if err != nil {
		return err
	}
	m.limits.SetMovetime(mt)
	PrintOK("[Manager]", fmt.Sprintf("Move time set to %d ms.", mt))
	return nil
}

// updateBoard copies data from the engine.Position() into the UI board.
func (m *Manager) updateBoard() {
	pos := m.e.Position()
	if pos.BigIndex() == int(engine.PosIndexIllegal) {
		m.board.BigIndex = -1
	} else {
		m.board.BigIndex = pos.BigIndex()
	}

	m.board.SetColors(m.e.Position().BigPositionState())

	board := pos.Position()
	for bi := 0; bi < 9; bi++ {
		for si := 0; si < 9; si++ {
			cell := board[bi][si]
			switch cell {
			case engine.PieceCross:
				m.board.Cells[bi][si] = 'X'
			case engine.PieceCircle:
				m.board.Cells[bi][si] = 'O'
			default:
				m.board.Cells[bi][si] = ' '
			}
		}
	}
}
