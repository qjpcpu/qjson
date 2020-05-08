package qjson

import (
	"runtime"
	"sync/atomic"
)

const (
	notReady uint32 = 0
	isReady  uint32 = 1
)

type Element struct {
	value interface{}
	ready uint32
}

// lfQueue is lock free queue
type lfQueue struct {
	capaciity uint32
	capMod    uint32
	putPos    uint32
	getPos    uint32
	cache     []Element
}

func newQueue(capaciity uint32) *lfQueue {
	q := new(lfQueue)
	q.capaciity = minQuantity(capaciity)
	q.capMod = q.capaciity - 1
	q.cache = make([]Element, q.capaciity)
	return q
}

func (q *lfQueue) Capaciity() uint32 {
	return q.capaciity
}

func (q *lfQueue) Quantity() uint32 {
	var putPos, getPos uint32
	var quantity uint32
	getPos = q.getPos
	putPos = q.putPos

	if putPos >= getPos {
		quantity = putPos - getPos
	} else {
		quantity = q.capMod + putPos - getPos
	}

	return quantity
}

// put queue functions
func (q *lfQueue) Put(val interface{}) bool {
	var putPos, putPosNew, getPos, posCnt uint32
	var cache *Element
	capMod := q.capMod
	for {
		putPos = atomic.LoadUint32(&q.putPos)
		getPos = atomic.LoadUint32(&q.getPos)

		if putPos >= getPos {
			posCnt = putPos - getPos
		} else {
			posCnt = capMod + putPos - getPos
		}

		if posCnt >= capMod {
			return false
		}

		putPosNew = putPos + 1
		if atomic.CompareAndSwapUint32(&q.putPos, putPos, putPosNew) {
			break
		} else {
			return false
		}
	}

	cache = &q.cache[putPosNew&capMod]

	for {
		if atomic.LoadUint32(&cache.ready) == notReady {
			cache.value = val
			atomic.StoreUint32(&cache.ready, isReady)
			return true
		} else {
			runtime.Gosched()
		}
	}
}

// get queue functions
func (q *lfQueue) Get() (val interface{}, ok bool) {
	var putPos, getPos, getPosNew, posCnt uint32
	var cache *Element
	capMod := q.capMod
	for {
		putPos = atomic.LoadUint32(&q.putPos)
		getPos = atomic.LoadUint32(&q.getPos)

		if putPos >= getPos {
			posCnt = putPos - getPos
		} else {
			posCnt = capMod + putPos - getPos
		}

		if posCnt < 1 {
			return nil, false
		}

		getPosNew = getPos + 1
		if atomic.CompareAndSwapUint32(&q.getPos, getPos, getPosNew) {
			break
		} else {
			return nil, false
		}
	}

	cache = &q.cache[getPosNew&capMod]

	for {
		if atomic.LoadUint32(&cache.ready) == isReady {
			val = cache.value
			atomic.StoreUint32(&cache.ready, notReady)
			/* clear reference */
			cache.value = nil
			return val, true
		} else {
			runtime.Gosched()
		}
	}
}

func minQuantity(v uint32) uint32 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}
