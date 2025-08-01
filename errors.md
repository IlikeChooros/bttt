Errors I got when testing the code:


### On high depth search:
*Fixed by increasing the channel buffer size to DefaultConfig.Engine.MaxDepth*


`x1xoo1o2/x5xo1/2o1x1oox/2ox1ox2/oxx6/o2o3x1/2x2x1oo/xo1x1x3/1oo5x x 2`

I was just playing the game in analysis mode, making the recommended moves, maybe I didn't wait long enough for the analysis to finish, but I got this panic. Looks like there is an issue with the search itself, since it this is reproducible with the position
attached above.

```bash
time=2025-08-01T19:44:29.347+02:00 level=INFO msg="Sending response" data="{Depth:2 Pv:[C3a3] Nps:17000 Eval:0.00 Final:false Error:}"
time=2025-08-01T19:44:29.347+02:00 level=INFO msg="Sending response" data="{Depth:3 Pv:[C3c3 C3a3] Nps:58000 Eval:0.04 Final:false Error:}"
time=2025-08-01T19:44:29.347+02:00 level=INFO msg="Sending response" data="{Depth:4 Pv:[C3a3 A3c1 C1a3] Nps:485000 Eval:0.16 Final:false Error:}"
time=2025-08-01T19:44:29.347+02:00 level=INFO msg="Sending response" data="{Depth:5 Pv:[C3a3 A3b1 B1b2 B2a2] Nps:1278000 Eval:0.30 Final:false Error:}"
time=2025-08-01T19:44:29.348+02:00 level=INFO msg="Sending response" data="{Depth:6 Pv:[C3a2 A2a3 A3b3 B3b3 B3c3] Nps:2917000 Eval:0.33 Final:false Error:}"
time=2025-08-01T19:44:29.351+02:00 level=INFO msg="Sending response" data="{Depth:7 Pv:[C3a2 A2c1 C1a2 A2b3 B3b3 B3c3] Nps:4143000 Eval:0.44 Final:false Error:}"
time=2025-08-01T19:44:29.358+02:00 level=INFO msg="Sending response" data="{Depth:8 Pv:[C3a2 A2c1 C1a2 A2a3 A3b3 B3a2 A3c2] Nps:4316400 Eval:0.44 Final:false Error:}"
time=2025-08-01T19:44:29.362+02:00 level=INFO msg="Sending response" data="{Depth:9 Pv:[C3a2 A2c1 C1a2 A2a3 A3b3 B3a2 A3c2] Nps:4156066 Eval:0.42 Final:false Error:}"
time=2025-08-01T19:44:29.389+02:00 level=INFO msg="Sending response" data="{Depth:10 Pv:[C3a2 A2c1 C1b1 B1c1 C1a2 A2b2 B2a2 A2a3 A3b3] Nps:4327365 Eval:0.44 Final:false Error:}"
time=2025-08-01T19:44:29.412+02:00 level=INFO msg="Sending response" data="{Depth:11 Pv:[C3c3 C3b3 B3c3 C3c2 C2a1 A1b3 B3c1 C1b1 B1b2 B2a2] Nps:4900562 Eval:0.45 Final:false Error:}"
time=2025-08-01T19:44:29.485+02:00 level=INFO msg="Sending response" data="{Depth:12 Pv:[C3c3 C3b3 B3c3 C3c2 C2a1 A1a1 A1b3 B3a2 A2c1 C1a2 A3b3] Nps:4936536 Eval:0.45 Final:false Error:}"
time=2025-08-01T19:44:29.521+02:00 level=INFO msg="Sending response" data="{Depth:13 Pv:[C3c3 C3b3 B3c3 C3c2 C2a1 A1a1 A1b2 B2b2 B2c1 C1a3 A3b3 B3b3] Nps:5072310 Eval:0.45 Final:false Error:}"
panic: send on closed channel

goroutine 91 [running]:
uttt/internal/server.rtAnalysis.func1({0x22, 0x3fdd635234ec9c30, 0xe, 0x10e85, 0xdc, 0x5156a3, {0xc00664a0c0, 0xc, 0xc}, 0x0, ...})
        /home/minis/Desktop/bttt/internal/server/analysis.go:60 +0x33a
uttt/internal/mcts.listenerInvoke[...](0xc0002a6060?, 0x3d0000003d)
        /home/minis/Desktop/bttt/internal/mcts/stats_listener.go:66 +0x7b
uttt/internal/mcts.(*MCTS[...]).invokeListener(...)
        /home/minis/Desktop/bttt/internal/mcts/mcts.go:259
uttt/internal/mcts.(*MCTS[...]).Selection(0x80ca60, {0x808590, 0xc000212b40}, 0xc000202f48)
        /home/minis/Desktop/bttt/internal/mcts/search.go:180 +0x1ff
uttt/internal/mcts.(*MCTS[...]).Search(0x80ca60, {0x808590, 0xc000212b40}, 0x3)
        /home/minis/Desktop/bttt/internal/mcts/search.go:116 +0x24c
created by uttt/internal/mcts.(*MCTS[...]).SearchMultiThreaded in goroutine 7
        /home/minis/Desktop/bttt/internal/mcts/search.go:76 +0x8e
exit status 2
```