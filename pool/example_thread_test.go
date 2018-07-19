package pool_test

import (
	"fmt"
	"time"

	"github.com/cjysmat/golib/pool"
)

// Usage example for the thread pool.
func Example_threadPool() {
	// Create a new thread pool with 5 concurrent worker capacity
	workers := pool.NewThreadPool(5)

	// Start the pool (you could schedule tasks before starting, and they would
	// wait queued until permission is given to execute)
	workers.Start()

	// Schedule some tasks (functions with no arguments nor return values)
	for i := 0; i < 10; i++ {
		id := i // Need to copy i for the task closure
		workers.Schedule(func() {
			time.Sleep(time.Duration(id) * 50 * time.Millisecond)
			fmt.Printf("Task #%d done.\n", id)
		})
	}
	// Terminate the pool gracefully (don't clear unstarted tasks)
	workers.Terminate(false)

	// Output:
	// Task #0 done.
	// Task #1 done.
	// Task #2 done.
	// Task #3 done.
	// Task #4 done.
	// Task #5 done.
	// Task #6 done.
	// Task #7 done.
	// Task #8 done.
	// Task #9 done.
}
