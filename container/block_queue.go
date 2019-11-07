package container

import (
	"sync"
)

type BlockQueueStatus struct {
	Wait int
	Len  int
}

type blockQueue struct {
	cond    *sync.Cond
	storage []interface{}
	status  BlockQueueStatus
}

func NewBlockQueue() *blockQueue {
	var queue = &blockQueue{}
	queue.cond = sync.NewCond(new(sync.RWMutex))
	queue.status = BlockQueueStatus{Wait: 0, Len: 0}
	queue.storage = make([]interface{}, 0)
	return queue
}

func (queue *blockQueue) Put(v interface{}) {

	queue.cond.L.Lock()

	queue.storage = append(queue.storage, v)

	if queue.status.Wait > 0 {
		queue.cond.Signal()
	}

	queue.cond.L.Unlock()
}

func (queue *blockQueue) Get() interface{} {

	queue.cond.L.Lock()

	queue.status.Wait++

	for {
		if len(queue.storage) > 0 {
			var r = queue.storage[0]
			queue.storage = queue.storage[1:]
			queue.status.Wait--
			queue.cond.L.Unlock()
			return r
		}
		queue.cond.Wait()
	}
}

func (queue *blockQueue) Status() BlockQueueStatus {
	queue.cond.L.Lock()
	queue.status.Len = len(queue.storage)
	queue.cond.L.Unlock()
	return queue.status
}