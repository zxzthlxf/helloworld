package pq

import (
	"errors"
	"sync"
	"unsafe"
)

var pointersMapLock sync.RWMutex
var pointersMap = map[unsafe.Pointer]interface{}{}

func addPointer(pointer interface{}) (unsafe.Pointer, error) {
	if pointer == nil {
		return nil, nil
	}

	uniqueIndex := Cmalloc(Csize_t(1)) // asking OS for unique index

	if uniqueIndex != nil {
		pointersMapLock.Lock()
		pointersMap[uniqueIndex] = pointer
		pointersMapLock.Unlock()
		return uniqueIndex, nil
	}
	return nil, errors.New("unique index allocation failed")
}

func getPointer(ptr unsafe.Pointer) (resultPtr interface{}) {
	if ptr != nil {
		pointersMapLock.RLock()
		resultPtr = pointersMap[ptr]
		pointersMapLock.RUnlock()
		return
	}
	return nil
}

func deletePointer(ptr unsafe.Pointer) {
	if ptr != nil {
		pointersMapLock.Lock()
		_, ok := pointersMap[ptr]
		if ok {
			delete(pointersMap, ptr)
		}
		pointersMapLock.Unlock()
		Cfree(ptr)
	}
}
