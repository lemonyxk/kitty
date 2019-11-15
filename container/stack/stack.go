/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-11-14 12:34
**/

package stack

type stack struct {
	list []interface{}
}

func NewStack(list ...interface{}) *stack {
	return &stack{list: list}
}

func (s *stack) Push(v interface{}) {
	s.list = append(s.list, v)
}

func (s *stack) Pop() (interface{}, bool) {
	if len(s.list) == 0 {
		return nil, false
	}
	var v = s.list[len(s.list)-1]
	s.list = s.list[:len(s.list)-1]
	return v, true
}

func (s *stack) Top() (interface{}, bool) {
	if len(s.list) == 0 {
		return nil, false
	}
	return s.list[len(s.list)-1], true
}

func (s *stack) Size() int {
	return len(s.list)
}
