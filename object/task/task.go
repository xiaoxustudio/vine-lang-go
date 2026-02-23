package task

import "sync"

type TaskObjectStatus int

const (
	TaskStateReady TaskObjectStatus = iota
	TaskStateRunning
	TaskStateDone
)

var (
	GlobalTaskArray = make([]*TaskObject, 0)
)

type TaskObject struct {
	parent *TaskObject
	next   *TaskObject
	state  TaskObjectStatus
	wg     sync.WaitGroup
	fn     func(args ...[]any) any
	result any
}

func NewTaskObject(fn func(args ...[]any) any) *TaskObject {
	instance := &TaskObject{
		state: TaskStateReady,
		fn:    fn,
	}
	GlobalTaskArray = append(GlobalTaskArray, instance)
	return instance
}

func (t *TaskObject) Run(args ...[]any) {
	t.state = TaskStateRunning

	t.wg.Add(1)

	go func() {
		defer t.wg.Done()

		t.result = t.fn(args...)

		t.Done()
	}()
}

func WaitAll() {
	for _, v := range GlobalTaskArray {
		v.Wait()
	}
}

func (t *TaskObject) Wait() any {
	t.wg.Wait()
	return t.result
}

func (t *TaskObject) Done() {
	t.state = TaskStateDone
	t.Dispose()
}

func (t *TaskObject) Dispose() {
	for i, v := range GlobalTaskArray {
		if v == t {
			GlobalTaskArray = append(GlobalTaskArray[:i], GlobalTaskArray[i+1:]...)
			break
		}
	}
}
