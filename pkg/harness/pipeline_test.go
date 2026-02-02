package harness

import (
	"sync"
	"testing"
)

func TestExecutionState_Concurrency(t *testing.T) {
	state := NewExecutionState()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(2)
		val := i
		go func() {
			defer wg.Done()
			state.Set("counter", val)
		}()
		go func() {
			defer wg.Done()
			_, _ = state.Get("counter")
		}()
	}

	wg.Wait()
}
