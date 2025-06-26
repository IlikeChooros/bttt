package uttt

import (
	"fmt"
	"math/rand"
	"time"
)

func OptimizeHash(nTries, timeoutSec int) {
	notations := []string{
		StartingPosition,
		"1x7/2o6/x8/xoxoxo3/9/9/9/9/oo7 x 3",
		"3x5/o8/9/xox1x4/o3o4/8x/9/4o4/9 o 4",
		"9/2o6/1xo1x2x1/9/2x4o1/9/9/2o4x1/9 o 7",
	}

	pos := NewPosition()
	tt := NewHashTable[HashTestEntry](1 << 22)
	collisions := make([][]int64, nTries)
	timer := time.NewTimer(time.Second * time.Duration(timeoutSec))
	bestSeed := int64(-1)
	bestCollisions := make([]int64, len(notations))
	currentSeed := _seedHash
	stop := false

	for i := 0; i < nTries && !stop; i++ {

		nodes := uint(0)
		collisions[i] = make([]int64, len(notations))

		fmt.Printf("Using seed=%d\n", currentSeed)
		_seedHash = currentSeed
		_InitHashing()

		for j, notation := range notations {
			select {
			case <-timer.C:
				stop = true
			default:
			}

			if err := pos.FromNotation(notation); err != nil {
				fmt.Print(err)
				continue
			}

			n, c := hashTest(6, pos, tt)
			tt.Clear()
			collisions[i][j] += int64(c)
			nodes += n

			fmt.Printf("%s: %d nodes %d collisions (%.3f) load factor: %.3f\n",
				notation, nodes, collisions[i][j],
				float64(collisions[i][j])/float64(nodes), tt.LoadFactor())

			if stop {
				break
			}
		}

		if stop {
			break
		}

		// Set the 'bestCollisons' slice
		if bestSeed == -1 {
			bestSeed = currentSeed
			copy(bestCollisions, collisions[i])
		} else {
			// Compare the collisions
			totaldelta := int64(0)

			for k, val := range bestCollisions {
				totaldelta += collisions[i][k] - val
			}

			fmt.Printf("delta: %d\n", totaldelta)

			// Means, on average we got less collisions
			if totaldelta < 0 {
				copy(bestCollisions, collisions[i])
				bestSeed = currentSeed
				fmt.Printf("------ NEW BEST SEED: %d\n", bestSeed)
			}
		}

		currentSeed = rand.Int63()
	}

	// Print the best hash found
	fmt.Printf("Best seed %d\n", bestSeed)
	for i, notataion := range notations {
		fmt.Printf("%s: %d collisions\n", notataion, bestCollisions[i])
	}
}
