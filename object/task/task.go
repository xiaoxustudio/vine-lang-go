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
	mu              sync.Mutex
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
	mu.Lock()
	GlobalTaskArray = append(GlobalTaskArray, instance)
	mu.Unlock()

	return instance
}

func (t *TaskObject) GetResult() any {
	return t.Wait()
}

func (t *TaskObject) IsRunning() bool {
	return t.state == TaskStateRunning
}

func (t *TaskObject) IsReady() bool {
	return t.state == TaskStateReady
}

func (t *TaskObject) IsDone() bool {
	return t.state == TaskStateDone
}

func (t *TaskObject) GetParent() *TaskObject {
	return t.parent
}

func (t *TaskObject) Run(args ...[]any) {
	t.state = TaskStateRunning

	t.wg.Go(func() {
		t.result = t.fn(args...)
		t.Done()
	})
}

func (t *TaskObject) Next(fn func(args ...[]any) any) *TaskObject {
	next := NewTaskObject(fn)
	next.parent = t
	t.next = next
	return next
}

func WaitAll() {
	mu.Lock()
	tasks := make([]*TaskObject, len(GlobalTaskArray))
	copy(tasks, GlobalTaskArray)
	mu.Unlock()

	for _, v := range tasks {
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
	mu.Lock()
	for i, v := range GlobalTaskArray {
		if v == t {
			GlobalTaskArray = append(GlobalTaskArray[:i], GlobalTaskArray[i+1:]...)
			break
		}
	}
	mu.Unlock()
}

func ClearAll() {
	mu.Lock()
	GlobalTaskArray = make([]*TaskObject, 0)
	mu.Unlock()
}
