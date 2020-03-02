package executor

import (
	"time"
	"sync"
	"testing"
)

func TestFixedExecutor(t *testing.T)  {
	executor := NewFixedExecutor(50)

	var wg sync.WaitGroup

	for i := 0; i < 5000; i++ {
		wg.Add(1)
		go func(i int) {
			future := executor.Execute(func() interface{} {
				time.Sleep(1 * time.Second)
				return i
			})

			t.Log(i, <-future.Done())
			wg.Done()
		}(i)
	}

	wg.Wait()
	executor.Release()
}
