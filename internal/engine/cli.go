package bttt

import (
	"bufio"
	"fmt"
	"os"
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

	var arg string
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		arg = scanner.Text()

		// Check if that's an exit flag
		for _, v := range exit_flags {
			if v == arg {
				return
			}
		}

		// Parse the command
		if err := cli.parseArgument(arg); err != nil {
			fmt.Println(err)
		}
	}
}

var _cliErrorFormat string = "[CLI] expected at least %d token(s) for command %s"

func _parseIntToken(tokenIdx int, tokens []string, f func(int)) error {
	if tokenIdx >= len(tokens) {
		// Safe, because tokenIdx >= 1 && tokenIdx <= len(tokens)
		return fmt.Errorf("[CLI] exptected number value for %s", tokens[tokenIdx-1])
	}

	var n int
	_, err := fmt.Sscanf(tokens[tokenIdx], "%d", &n)
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
	case "test":
		if len(tokens) < 2 {
			return fmt.Errorf(_cliErrorFormat, 1, "test")
		}
		// Test the move generation
		if tokens[1] == "movegen" {
			return _parseIntToken(2, tokens, func(depth int) {
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
	limits := Limits{}
	var err error
	for i := 0; i < len(tokens); i++ {

		if err != nil {
			break
		}

		switch tokens[i] {
		case "depth":
			// Next token should be an integer value
			err = _parseIntToken(i+1, tokens, func(depth int) {
				limits.depth = depth
				i++
			})
		case "nodes":
			err = _parseIntToken(i+1, tokens, func(nodes int) {
				limits.nodes = uint64(nodes)
				i++
			})
		case "movetime":
			err = _parseIntToken(i+1, tokens, func(movetime int) {
				limits.movetime = movetime
				i++
			})
		default:
			// Unsupported command
			return fmt.Errorf("[CLI] Unsupported command %s", tokens[i])
		}
	}

	// Run the engine
	if err == nil {
		cli.engine.SetLimits(limits)
		cli.engine.Search()
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
