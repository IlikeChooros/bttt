# uttt
Ultimate Tic Tac Toe algorithm written in Go, with command line interface (much like `UCI` for chess engines).


### Usage

Run as cli engine:

```bash
go run cmd/enigne/main.go
```

Run with a simple UI:

```bash
go run cmd/ui/main.go
```

### Tests
```bash
go test -v -cover ./internal/engine
```