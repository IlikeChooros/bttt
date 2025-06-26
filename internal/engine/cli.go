package uttt

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Cli struct {
	engine *Engine
}

// Get new cli object (pointer)
func NewCli() *Cli {
	cli := new(Cli)
	cli.engine = NewEngine()
	return cli
}

// Start the cli loop
func (cli *Cli) Start() {
	defer fmt.Println("Exiting...")

	exit_flags := [4]string{
		"e", "q", "exit", "quit",
	}

	// Initialize the lib
	Init()

	var arg string
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Ultimate Tic Tac Toe engine")
	for {
		scanner.Scan()
		arg = scanner.Text()

		// Check if that's an exit flag
		for _, v := range exit_flags {
			if v == arg {
				return
			}
		}

		// Parse the command asynchronously
		go func() {
			if err := cli.parseArgument(arg); err != nil {
				fmt.Println(err)
			}
		}()
	}
}

var _cliErrorFormat string = "[CLI] expected at least %d token(s) for command %s"

func _parseIntToken(tokenIdx int, tokens []string, f func(int)) error {
	if tokenIdx >= len(tokens) {
		// Safe, because tokenIdx >= 1 && tokenIdx <= len(tokens)
		return fmt.Errorf("[CLI] exptected number value for %s", tokens[tokenIdx-1])
	}

	n, err := strconv.Atoi(tokens[tokenIdx])
	if err != nil {
		return err
	}

	// Invoke the function
	f(n)
	return nil
}

// Handle given input arguments as 1 string (should be seprated by space)
func (cli *Cli) parseArgument(arg string) error {
	if len(arg) == 0 {
		return fmt.Errorf("[CLI] Expected arguments for parsing")
	}

	// split the arguments into tokens
	tokens := strings.Split(arg, " ")

	switch tokens[0] {
	case "stop":
		cli.engine.Stop()
	case "go":
		// there should be at least 2 tokens
		if len(tokens) < 2 {
			return fmt.Errorf(_cliErrorFormat, 2, "go")
		}
		return cli.handleGo(tokens[1:])
	case "position":
		if len(tokens) < 2 {
			return fmt.Errorf(_cliErrorFormat, 1, "position")
		}
		return cli.handlePosition(tokens[1:])

	case "getpos":
		fmt.Println(cli.engine.position.Notation())
	case "makemove":
		if len(tokens) < 2 {
			return fmt.Errorf(_cliErrorFormat, 1, "makemove")
		}
		for _, token := range tokens[1:] {
			mv := MoveFromString(token)
			if mv == PosIllegal {
				return fmt.Errorf("[CLI] invalid move notation, expected [A-C][1-3][a-c][1-3]")
			}
			if !cli.engine.position.IsLegal(mv) {
				return fmt.Errorf("[CLI] Illegal move: %s", token)
			}
			cli.engine.position.MakeMove(mv)
		}
	case "undomove":
		cli.engine.position.UndoMove()
	case "test":
		if len(tokens) < 2 {
			return fmt.Errorf(_cliErrorFormat, 1, "test")
		}
		return cli.handleTest(tokens[1:])
	}

	return nil
}

type HashTestEntry struct {
	HashEntryBase
	notation string
}

func hashTest(depth int, pos *Position, tt *HashTable[HashTestEntry]) (uint, uint) {

	if depth == 0 {
		notation := pos.Notation()

		if val, ok := tt.Get(pos.hash); ok {
			// Check if that's a collision
			if notation != val.notation {
				return 1, 1
			}
		} else {
			// That's an empty key, set this new value
			tt.SetForced(pos.hash, HashTestEntry{
				HashEntryBase: HashEntryBase{Depth: depth, Hash: pos.hash},
				notation:      notation,
			})
		}

		return 1, 0
	}

	// Go through the moves
	nodes, collisions := uint(0), uint(0)
	moves := pos.GenerateMoves().Slice()

	for _, m := range moves {
		pos.MakeMove(m)
		n, c := hashTest(depth-1, pos, tt)
		pos.UndoMove()

		nodes += n
		collisions += c
	}

	return nodes, collisions
}

func (cli *Cli) handleTest(tokens []string) error {
	// Test the move generation
	if tokens[0] == "movegen" {
		return _parseIntToken(1, tokens, func(depth int) {
			now := time.Now()
			avgtime := float64(0)
			nodes := uint64(0)

			defer func() {
				fmt.Printf("\rAvg %.1f Mnps\033[K\n", float64(nodes)/avgtime)
			}()

			const Ntries = 10

			for i := 0; i < Ntries; i++ {
				nodes = Perft(cli.engine.position, depth, true, false)
				avgtime += float64(time.Since(now).Microseconds()-int64(avgtime)) / float64(i+1)
				fmt.Printf("\rProgress: %.1f (eta: %s)\033[K",
					(float32(i+1)/Ntries)*100,
					time.Duration(avgtime*float64(Ntries-(i+1))*1000).String())
				now = time.Now()
			}
		})
	}

	// Test hasing values, by performing perft test up to certain depth, and see how many collisions we get
	if tokens[0] == "hash" {
		return _parseIntToken(1, tokens, func(i int) {
			tt := NewHashTable[HashTestEntry](1 << 20)
			nodes, collisions := hashTest(i, cli.engine.position, tt)

			fmt.Printf("HashTest: %d nodes %d collisions (%.3f) load factor: %.3f\n",
				nodes, collisions, float64(collisions)/float64(nodes), tt.LoadFactor())
		})
	}

	return nil
}

// Handle the 'go' command
// Possible tokens:
// go perft|[ depth <n> | nodes <n> | movetime <n>]
func (cli *Cli) handleGo(tokens []string) error {

	// Handle 'perft' command separately
	if tokens[0] == "perft" {
		// Next token should be depth
		return _parseIntToken(1, tokens, func(depth int) {
			Perft(cli.engine.position, depth, false, true)
		})
	} else if tokens[0] == "valid-perft" {
		return _parseIntToken(1, tokens, func(depth int) {
			Perft(cli.engine.position, depth, true, true)
		})
	}

	// Parse the search commands
	limits := DefaultLimits()
	var err error
	for i := 0; i < len(tokens); i++ {

		if err != nil {
			break
		}

		switch tokens[i] {
		case "depth":
			// Next token should be an integer value
			err = _parseIntToken(i+1, tokens, func(depth int) {
				limits.SetDepth(depth)
				i++
			})
		case "nodes":
			err = _parseIntToken(i+1, tokens, func(nodes int) {
				limits.SetNodes(uint64(nodes))
				i++
			})
		case "movetime":
			err = _parseIntToken(i+1, tokens, func(movetime int) {
				limits.SetMovetime(movetime)
				i++
			})
		case "infinite":
			limits.SetInfinite(true)
		default:
			// Unsupported command
			return fmt.Errorf("[CLI] Unsupported command %s", tokens[i])
		}
	}

	// Run the engine
	if err == nil {
		cli.engine.SetLimits(*limits)
		cli.engine.Think(true)
	}

	return err
}

// Handle position command
// Possible options:
// position startpos | <notation>
// notation - is a string representing the bttt position, has 3 segments
// first one is the board position itself, then side to move,
// last one 'big index'
func (cli *Cli) handlePosition(tokens []string) error {
	switch len(tokens) {
	case 1:
		// Expecting 'startpos'
		if tokens[0] != "startpos" {
			return fmt.Errorf("[CLI] Expected 'startpos' token")
		}
	case _notationNumberOfSections:
		// Should be a valid postion
		break
	default:
		// Invalid number of sections
		return fmt.Errorf("[CLI] Invalid number of sections in given position")
	}

	// Return the parsing result of this position
	pos := NewPosition()
	err := pos.FromNotation(strings.Join(tokens, " "))

	// If we don't run into any exceptions, set this position as new one
	// Simply to preserve the position state, if user has given invalid position
	if err == nil {
		cli.engine.position = nil
		cli.engine.position = pos
	}

	// Return the error result
	return err
}
