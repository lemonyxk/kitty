package pool

import (
	"runtime"
	"sync"
)

type LastPoolConfig struct {
	Max int
	Min int
	New func() interface{}
}

func NewLastPool(config LastPoolConfig) *lastPool {

	if config.Max <= 0 {
		config.Max = runtime.NumCPU() * 2
	}

	if config.Min <= 0 {
		config.Min = runtime.NumCPU()
	}

	if config.Min >= config.Max {
		config.Min = config.Max
	}

	var pool = &lastPool{}
	pool.config = config
	pool.status = lastPoolStatus{max: config.Max, min: config.Min, len: 0}

	if len(pool.storage) < pool.config.Min {
		pool.storage = append(pool.storage, config.New())
	}

	return pool
}

type lastPoolStatus struct {
	max int
	min int
	len int
}

type lastPool struct {
	mux     sync.RWMutex
	storage []interface{}
	config  LastPoolConfig
	status  lastPoolStatus
}

func (pool *lastPool) Put(v interface{}) {

	if len(pool.storage) >= pool.config.Max {
		return
	}

	pool.mux.Lock()

	// if put too fast and get slowly that you will lose some put things
	// pool do not need worry
	if len(pool.storage) < pool.config.Max {
		pool.storage = append(pool.storage, v)
		pool.status.len++
	}

	pool.mux.Unlock()
}

func (pool *lastPool) Get() interface{} {

	pool.mux.Lock()

	if len(pool.storage) > 0 {
		var r = pool.storage[0]
		pool.storage = pool.storage[1:]
		pool.status.len--
		pool.mux.Unlock()
		return r
	}

	pool.mux.Unlock()
	return pool.config.New()

}

func (pool *lastPool) Status() lastPoolStatus {
	return pool.status
}
