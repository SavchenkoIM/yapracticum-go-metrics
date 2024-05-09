package dbstore

import (
	"golang.org/x/exp/constraints"
	"maps"
	"sync"
)

type ThreadSafeMap[S constraints.Float | constraints.Integer] struct {
	mutex sync.RWMutex
	data  map[string]S
}

func (tsm *ThreadSafeMap[S]) Set(key string, value S) {
	tsm.mutex.Lock()
	defer tsm.mutex.Unlock()
	tsm.data[key] = value
}

func (tsm *ThreadSafeMap[S]) Inc(key string, value S) {
	tsm.mutex.Lock()
	defer tsm.mutex.Unlock()
	tsm.data[key] += value
}

func (tsm *ThreadSafeMap[S]) Get(key string, value S) (S, bool) {
	tsm.mutex.RLock()
	defer tsm.mutex.RUnlock()
	val, ok := tsm.data[key]
	return val, ok
}

func (tsm *ThreadSafeMap[S]) GetData() map[string]S {
	tsm.mutex.RLock()
	defer tsm.mutex.RUnlock()
	return maps.Clone(tsm.data)
}

func (tsm *ThreadSafeMap[S]) Clear() {
	tsm.mutex.Lock()
	defer tsm.mutex.Unlock()
	tsm.data = make(map[string]S)
}
