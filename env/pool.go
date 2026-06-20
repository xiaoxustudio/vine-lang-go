package env

import "sync"

var envPool = sync.Pool{
	New: func() any {
		return &Environment{
			store:   make(map[string]any, 8),
			nameMap: make(map[string]Token, 8),
			consts:  make(map[string]struct{}, 8),
		}
	},
}

func NewPooled(fileName string) *Environment {
	e := envPool.Get().(*Environment)
	e.FileName = fileName
	e.parent = nil
	e.MountScope = nil
	e.Exports = nil
	e.isPassing = false
	for k := range e.consts {
		delete(e.consts, k)
	}
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
		e.MountScope = nil
		e.Exports = nil
		e.isPassing = false
		for k := range e.consts {
			delete(e.consts, k)
		}
		for k := range e.store {
			delete(e.store, k)
		}
		for k := range e.nameMap {
			delete(e.nameMap, k)
		}
		envPool.Put(e)
	}
}
