package navigation

import "container/heap"

type AstarNode struct {
	NodeId   int
	Priority float64
	Gscore   float64
}

type PriorityQueue struct {
	items []*AstarNode
	index map[int]int // nodeId -> index in heap
}

// return number of items in the queue
func (pq PriorityQueue) Len() int { return len(pq.items) }

// check for the lowest priority
func (pq PriorityQueue) Less(i, j int) bool { return pq.items[i].Priority < pq.items[j].Priority }

func (pq PriorityQueue) Swap(i, j int) {
	pq.index[pq.items[i].NodeId] = j
	pq.index[pq.items[j].NodeId] = i
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	node := x.(*AstarNode)
	pq.items = append(pq.items, node)
	pq.index[node.NodeId] = len(pq.items) - 1
}

func (pq *PriorityQueue) Pop() interface{} {
	n := len(pq.items)
	old := pq.items[n-1]
	delete(pq.index, old.NodeId)
	pq.items = pq.items[:n-1]
	return old
}

func (pq *PriorityQueue) Update(nodeId int, newPriority float64, newGscore float64) {
	i, ok := pq.index[nodeId]
	if !ok {
		return
	}

	pq.items[i].Priority = newPriority
	pq.items[i].Gscore = newGscore

	heap.Fix(pq, i)
}

func newPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		items: []*AstarNode{},
		index: make(map[int]int),
	}
}
