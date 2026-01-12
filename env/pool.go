package env

import "sync"

var envPool = sync.Pool{
	New: func() any {
		return &Environment{
			store:   make(map[Token]any, 8),
			nameMap: make(map[string]Token, 8),
		}
	},
}

func NewPooled(fileName string) *Environment {
	e := envPool.Get().(*Environment)
	e.FileName = fileName
	e.parent = nil
	for k := range e.store {
		delete(e.store, k)
	}
	for k := range e.nameMap {
		delete(e.nameMap, k)
	}
	return e
}

func (e *Environment) Release() {
	if e != nil {
		e.parent = nil
		envPool.Put(e)
	}
}
