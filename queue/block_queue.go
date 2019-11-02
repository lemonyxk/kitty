package queue

import (
	"sync"
)

type BlockQueueConfig struct {
	New func() interface{}
}

type BlockQueueStatus struct {
	Wait int
	Len  int
}

type blockQueue struct {
	cond    *sync.Cond
	storage []interface{}
	config  BlockQueueConfig
	status  BlockQueueStatus
}

func NewBlockQueue(config BlockQueueConfig) *blockQueue {
	var queue = &blockQueue{}
	queue.cond = sync.NewCond(new(sync.RWMutex))
	queue.config = config
	queue.status = BlockQueueStatus{Wait: 0, Len: 0}
	return queue
}

func (queue *blockQueue) Put(v interface{}) {

	queue.cond.L.Lock()
	defer queue.cond.L.Unlock()

	queue.storage = append(queue.storage, v)

	if queue.status.Wait > 0 {
		queue.cond.Signal()
	}
}

func (queue *blockQueue) Get() interface{} {

	queue.cond.L.Lock()
	defer queue.cond.L.Unlock()

	queue.status.Wait++

	for {
		if len(queue.storage) > 0 {
			var r = queue.storage[0]
			queue.storage = queue.storage[1:]
			queue.status.Wait--
			return r
		}
		queue.cond.Wait()
	}
}

func (queue *blockQueue) Status() BlockQueueStatus {
	queue.cond.L.Lock()
	defer queue.cond.L.Unlock()
	queue.status.Len = len(queue.storage)
	return queue.status
}
