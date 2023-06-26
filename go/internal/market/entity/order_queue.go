package entity

type OrderQueue struct {
	Orders []*Order
}

func (q *OrderQueue) Len() int { return len(q.Orders) }

func (q *OrderQueue) Less(i, j int) bool {
	return q.Orders[i].Price < q.Orders[j].Price
}

func (q *OrderQueue) Swap(i, j int) {
	q.Orders[i], q.Orders[j] = q.Orders[j], q.Orders[i]
}

func (q *OrderQueue) Push(x any) {
	q.Orders = append(q.Orders, x.(*Order))
}

func (q *OrderQueue) Pop() any {
	old := q.Orders
	n := len(old)
	item := old[n-1]
	q.Orders = old[0 : n-1]
	return item
}

func NewOrderQueue() *OrderQueue {
	return &OrderQueue{}
}
