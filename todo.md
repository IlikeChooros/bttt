Big Tic Tac Toe in Go:
# Engine

## TESTS
  - [x] Add tests for terminations (win, draw, resign)
  - [x] Add some evaluation tests (in a clearly winning position, in a clearly losing position, etc.)

## FEATURES
- [x] Make movelist struct:
  - [x] Holds all moves (`[]Move`)
  - [x] can return a range of moves
- [x] Make a move generator
- [x] Add to small square position termination flags (because we can have a both win and lose in the same small square, but it depends on who first achieved it)
- [x] Notation:
  - [x] Add to each 'small square' the state (O, X won, Draw, None)
- [ ] UI:
  - [ ] After termination, show the board with the winning line
  - [ ] *Fix bug: after few moves, pieces start disappearing*
- [x] Optimization:
  - [x] Evaluation: 
    - [x] Make the board hold also bitboards for each square (so we don't have to calculate them every time)
  - [x] Search:
    - [x] Use transposition tables:
      - [x] Zobrist hashing
      - [x] Use better seed for hashing
- [ ] OTHER:
  - [x] Refactor the file structure, use folders, and if possible use different folder for tests
  
## IDEAS

### Hashing:
```
In the big tic tac toe, we can get an 'illegal' position on the small square, meaning number of 
all positions is 3^9 = 19683. But on the 'big board' we can have only 5478 positions.

```
Sources:
- [Number of valid positions in TTT](https://math.stackexchange.com/questions/469371/determining-the-number-of-valid-tictactoe-board-states-in-terms-of-board-dimensi)

### Move generation:
```
Use bitboards to generate moves, this is as simple as xoring the occupancy bitboard with 511,
and then counting the number of bits in the result.
```


### Tester

```
Program that will run asynchronously many games, and will report the results.
Also allowing user to lookup one of the games and see the moves.
```

### UI

```
Add simple cmd line interface, that will allow user to play against the engine.
```



### Optimization

Benchmarks:

*Before optimization:*

```
goos: linux
goarch: amd64
pkg: uttt/internal/engine
cpu: AMD Ryzen 9 6900HX with Radeon Graphics        
BenchmarkMakeUndo-16            23561091                23.27 ns/op
BenchmarkGenerateMoves-16        1279252               472.9 ns/op
PASS
ok      uttt/internal/engine    1.669s
```

**Using bitboards:**

```
goos: linux
goarch: amd64
pkg: uttt/internal/engine
cpu: AMD Ryzen 9 6900HX with Radeon Graphics        
BenchmarkMakeUndo-16            29519068                19.73 ns/op            0 B/op          0 allocs/op
BenchmarkGenerateMoves-16        1929310               316.9 ns/op           128 B/op          2 allocs/op
BenchmarkNotationLoad-16         2597977               224.5 ns/op            32 B/op          1 allocs/op
PASS
ok      uttt/internal/engine    2.368s
```

*(Added BenchmarkNotationLoad, to see how time in the 'GeneratMoves' we use to load the position)*
Current perft results:
perft 10: `Nodes: 18466787808 (112.0 Mnps)`


### Killer moves, history heuristic, etc.

```
Apparently, history heuristic is not that useful in this game.
It literally doesn't matter if we use move ordering or not.
I have tried several different approaches, and the results are always the same.
I have tried:
- History heuristic:
  - simple formula, if beta cutoff occurs, set the history value to depth*depth at [side][bigindex][smallindex]
  - using gravity formula
- Move:
  - Added piece square tables, pattern evaluation, etc.
```

# New approach
- [x] Use monte carlo tree search (MCTS) instead of alpha-beta pruning
- [x] Support multi-threading
- [ ] Proper pv support:
  - [ ] Working for maximizing player (or for the player whose turn is set to 1)
- [ ] Add channels to set the engine results
- [ ] Add turn to the nodes, since this algorithm is used for zero-sum games anyway

# Server

TODO:
- [x] Read about `context` package, and how to use it with the server
- [x] Add actual timeout logic:
  - Maybe try using `context` package, or implement a use simply a slice with the engine structs, and then use engine.Stop() after timeout.
  - Simply used a async function that waits for the timeout, and then calls engine.Stop()
- [ ] Add prometheus metrics (whatever it is)
  - Read about docker, basic usage, and how to use it with the server,
  - Use later in the project, right now it is not needed
- [ ] Add rate limiting
- [x] Make frontend for backend proxy in Next.js app
  
Suggested by AI:
- Docker containerization with multi-stage builds for cleaner deployment.
- Environment variable configuration for flexibility in different environments.
- Circut breaker pattern to handle engine failures gracefully.
- TLS/HTTPS support for secure communication.
- Database pooling for efficient resource management.
- Load testing with tools like `hey` or `wrk` to ensure the server can handle high traffic.
- API versioning to manage changes without breaking existing clients.
