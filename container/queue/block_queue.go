package queue

import (
	"sync"
)

func NewBlockQueue() *blockQueue {
	var queue = &blockQueue{}
	queue.cond = sync.NewCond(new(sync.RWMutex))
	queue.status = blockQueueStatus{wait: 0, len: 0}
	queue.storage = make([]interface{}, 0)
	return queue
}

type blockQueueStatus struct {
	wait int
	len  int
}

type blockQueue struct {
	cond    *sync.Cond
	storage []interface{}
	status  blockQueueStatus
}

func (queue *blockQueue) Push(v interface{}) {

	queue.cond.L.Lock()

	queue.storage = append(queue.storage, v)
	queue.status.len++

	if queue.status.wait > 0 {
		queue.cond.Signal()
	}

	queue.cond.L.Unlock()
}

func (queue *blockQueue) Pop() interface{} {

	queue.cond.L.Lock()

	queue.status.wait++

	for {
		if len(queue.storage) > 0 {
			var r = queue.storage[0]
			queue.storage = queue.storage[1:]
			queue.status.wait--
			queue.status.len--
			queue.cond.L.Unlock()
			return r
		}
		queue.cond.Wait()
	}
}

func (queue *blockQueue) Status() blockQueueStatus {
	return queue.status
}
