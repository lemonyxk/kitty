/**
* @program: lemo
*
* @description:
*
* @author: lemo
*
* @create: 2019-11-14 11:35
**/

package head

func NewMaxHead(list ...Interface) *maxHead {
	var head = &maxHead{}
	head.data = list
	head.len = len(list)
	head.create()
	return head
}

type maxHead struct {
	data []Interface
	len  int
}

func (h *maxHead) Pop() (Interface, bool) {
	if h.len == 0 {
		return nil, false
	}

	var value = h.data[0]
	h.data[0] = h.data[h.len-1]
	h.data = h.data[:h.len-1]
	h.len--

	if h.len > 1 {
		h.down(0)
	}
	return value, true
}

func (h *maxHead) Push(v Interface) {
	h.data = append(h.data, v)
	h.len++
	h.up(h.len - 1)
}

func (h *maxHead) Size() int {
	return h.len
}

func (h *maxHead) create() {
	for i := (h.len - 2) / 2; i >= 0; i-- {
		h.down(i)
	}
}

func (h *maxHead) down(parentIndex int) {
	// 暂存父节点
	var temp = h.data[parentIndex]
	// 子节点 默认为左节点
	var childIndex = parentIndex*2 + 1
	for {
		if childIndex >= h.len {
			break
		}

		// 如果有右节点 则 一定有左节点
		// 有右节点 且 右节点大于左节点 则 定位至右节点
		if childIndex+1 < h.len && h.data[childIndex+1].Value() > h.data[childIndex].Value() {
			childIndex++
		}

		// 如果父节点大于等于孩子值 则 退出
		if temp.Value() >= h.data[childIndex].Value() {
			break
		}

		// 否则 交换值
		h.data[parentIndex] = h.data[childIndex]

		// 孩子节点继续下沉
		parentIndex = childIndex

		// 查找下一个子节点 默认为左节点
		childIndex = childIndex*2 + 1

	}
	// 最后定位的子节点交换父节点
	h.data[parentIndex] = temp
}

func (h *maxHead) up(index int) {
	var childIndex = index

	var parentIndex = (index - 1) / 2

	var temp = h.data[childIndex]

	for {
		if childIndex <= 0 {
			break
		}

		if temp.Value() <= h.data[parentIndex].Value() {
			break
		}

		h.data[childIndex] = h.data[parentIndex]

		childIndex = parentIndex

		parentIndex = (parentIndex - 1) / 2
	}

	h.data[childIndex] = temp
}
