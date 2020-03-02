package executor

import (
	"errors"
)

type Task func() interface{}

type Future struct {
	task 	Task
	value interface{}
	done	chan bool
	err	error
}

func(f *Future) run() {
	defer func() {
		if r := recover(); r != nil {
			f.err = errors.New(r.(string))
			f.value = nil
		}

		f.done <- true
	}()

	f.value = f.task()
}

func(f *Future) Done() <-chan bool {
	return f.done
}
