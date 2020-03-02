package executor

type FixedExecutor struct {
	workerSize	int
	taskNotify	chan *Future
	closeNotify chan bool
	closed bool
}

func NewFixedExecutor(workerSize int) *FixedExecutor {
	if workerSize <= 0 {
		panic("workerSize <= 0")
	}

	executor := new(FixedExecutor)
	executor.workerSize	= workerSize
	executor.taskNotify = make(chan *Future, workerSize * 4)
	executor.closeNotify = make(chan bool)

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

	future = new(Future)
	future.task = task
	future.done = make(chan bool)

	e.taskNotify <- future
	return
}

func(e *FixedExecutor) Release() {
	e.closeNotify <- true
	close(e.closeNotify)
	e.closed = true
}
