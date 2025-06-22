Big Tic Tac Toe in Go:
## TESTS
  - [x] Add tests for terminations (win, draw, resign)
  - [x] Add some evaluation tests (in a clearly winning position, in a clearly losing position, etc.)

## FEATURES
- [x] Make movelist struct:
  - [x] Holds all moves (`[]Move`)
  - [x] can return a range of moves
- [x] Make a move generator
- [x] Add to small square position termination flags (because we can have a both win and lose in the same small square, but it depends on who first achieved it)
- [ ] Notation:
  - [ ] Add to each 'small square' the state (O, X won, Draw, None)
- [ ] Optimization:
  - [ ] Evaluation: 
    - [ ] Make the board hold also bitboards for each square (so we don't have to calculate them every time)
  - [ ] Search:
    - [ ] Use transposition tables:
      - [ ] Instead of creating a perftect hash (using 9 * 18 bits to store the whole position), find such 'magic' number that will 
    - [ ] Add move ordering
    - [ ] Use killer moves
    - [ ] Use history heuristic
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