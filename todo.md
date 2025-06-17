Big Tic Tac Toe in Go:
## TESTS
  - [x] Add tests for terminations (win, draw, resign)
  - [ ] Add some evaluation tests (in a clearly winning position, in a clearly losing position, etc.)

## FEATURES
- [x] Make movelist struct:
  - [x] Holds all moves (`[]Move`)
  - [x] can return a range of moves
- [x] Make a move generator
- [ ] Add to small square position termination flags (because we can have a both win and lose in the same small square, but it depends on who first achieved it)
- [ ] Optimization:
  - [ ] Evaluation: 
    - [ ] Make the board hold also bitboards for each square (so we don't have to calculate them every time)
  - [ ] Search:
    - [ ] Use transposition tables
    - [ ] Add move ordering
    - [ ] Use killer moves
    - [ ] Use history heuristic
- [ ] OTHER:
  - [x] Refactor the file structure, use folders, and if possible use different folder for tests