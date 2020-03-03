package executor

import (
	"sync"
)

type FixedExecutor struct {
	buf []*Future
	taskNotify	chan *Future
	closeNotify chan bool
	closed bool

	head 	int
	tail 	int
	count int

	cond *sync.Cond
}

func NewFixedExecutor(workerSize int) *FixedExecutor {
	if workerSize <= 0 {
		panic("workerSize <= 0")
	}

	executor := new(FixedExecutor)
	executor.buf = make([]*Future, 32, 32)
	executor.taskNotify = make(chan *Future, workerSize * 4)
	executor.closeNotify = make(chan bool)
	executor.cond = sync.NewCond(&sync.Mutex{})

	go func(executor *FixedExecutor) {
		for {
			select {
			case <-executor.closeNotify:
				return
			default:
				executor.taskNotify <- executor.shift()
			}
		}
	}(executor)

	for i := 0; i < workerSize; i++ {
		go func(taskNotify chan *Future, closeNotify chan bool, id int) {
			for {
				select {
				case future := <-taskNotify:
					future.run()

				case <-closeNotify:
					return
				}
			}
		}(executor.taskNotify, executor.closeNotify, i)
	}

	return executor
}

func(e *FixedExecutor) Execute(task Task) (future *Future) {
	if e.closed {
		panic("executor is closed")
	}

	if task == nil {
		panic("task is nil")
	}

	future = new(Future)
	future.task = task
	future.done = make(chan bool)

	e.add(future)
	return
}

func(e *FixedExecutor) add(future *Future) {
	e.cond.L.Lock()
	defer e.cond.L.Unlock()
	if e.count == len(e.buf) {
		e.resize()
	}

	e.buf[e.tail] = future
	// bitwise modulus
	e.tail = (e.tail + 1) & (len(e.buf) - 1)

	if e.count <= 0 {
		e.cond.Signal()
	}

	e.count++
}

func(e *FixedExecutor) shift() *Future {
	e.cond.L.Lock()
	defer e.cond.L.Unlock()

	if e.count <= 0 {
		e.cond.Wait()
	}
	ret := e.buf[e.head]
	e.buf[e.head] = nil

	e.head = (e.head + 1) & (len(e.buf) - 1)
	e.count--

	if len(e.buf) > 32 && (e.count<<2) == len(e.buf) {
		e.resize()
	}
	return ret
}

func (e *FixedExecutor) resize() {
	newBuf := make([]*Future, e.count<<1)

	if e.tail > e.head {
		copy(newBuf, e.buf[e.head:e.tail])
	} else {
		n := copy(newBuf, e.buf[e.head:])
		copy(newBuf[n:], e.buf[:e.tail])
	}

	e.head = 0
	e.tail = e.count
	e.buf = newBuf
}

func(e *FixedExecutor) Release() {
	e.closeNotify <- true
	close(e.closeNotify)
	e.cond.Signal()
	e.closed = true
}
