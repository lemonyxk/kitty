/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-11-14 12:14
**/

package queue

func NewQueue(list ...interface{}) *queue {
	return &queue{list: list}
}

type queue struct {
	list []interface{}
}

func (q *queue) Push(v interface{}) {
	q.list = append(q.list, v)
}

func (q *queue) Pop() (interface{}, bool) {
	if len(q.list) == 0 {
		return nil, false
	}
	var v = q.list[0]
	q.list = q.list[:len(q.list)-1]
	return v, true
}

func (q *queue) Size() int {
	return len(q.list)
}
