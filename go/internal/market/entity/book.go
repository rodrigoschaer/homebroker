package entity

import (
	"container/heap"
	"sync"
)

type Book struct {
	Orders          []*Order
	Transactions    []*Transaction
	OrdersChan      chan *Order // input
	OrdersChanOut   chan *Order
	Wg              *sync.WaitGroup
	BuyOrderQueues  map[string]*OrderQueue
	SellOrderQueues map[string]*OrderQueue
}

func NewBook(orderChan chan *Order, orderChanOut chan *Order, wg *sync.WaitGroup) *Book {
	return &Book{
		Orders:          []*Order{},
		Transactions:    []*Transaction{},
		OrdersChan:      orderChan,
		OrdersChanOut:   orderChanOut,
		Wg:              wg,
		BuyOrderQueues:  make(map[string]*OrderQueue),
		SellOrderQueues: make(map[string]*OrderQueue),
	}
}

func (b *Book) Trade() {
	for order := range b.OrdersChan {
		asset := order.Asset.ID

		buyOrders := b.getOrCreateBuyOrderQueue(asset)
		sellOrders := b.getOrCreateSellOrderQueue(asset)

		switch order.OrderType {
		case "BUY":
			buyOrders.Push(order)
			if sellOrders.Len() > 0 && sellOrders.Orders[0].Price <= order.Price {
				sellOrder := sellOrders.Pop().(*Order)
				if sellOrder.PendingShares > 0 {
					transaction := NewTransaction(sellOrder, order, order.Shares, sellOrder.Price)
					b.AddTransaction(transaction)
					sellOrder.Transactions = append(sellOrder.Transactions, transaction)
					order.Transactions = append(order.Transactions, transaction)
					b.OrdersChanOut <- sellOrder
					b.OrdersChanOut <- order
					if sellOrder.PendingShares > 0 {
						sellOrders.Push(sellOrder)
					}
				}
			}
		case "SELL":
			sellOrders.Push(order)
			if buyOrders.Len() > 0 && buyOrders.Orders[0].Price >= order.Price {
				buyOrder := buyOrders.Pop().(*Order)
				if buyOrder.PendingShares > 0 {
					transaction := NewTransaction(order, buyOrder, order.Shares, buyOrder.Price)
					b.AddTransaction(transaction)
					buyOrder.Transactions = append(buyOrder.Transactions, transaction)
					order.Transactions = append(order.Transactions, transaction)
					b.OrdersChanOut <- buyOrder
					b.OrdersChanOut <- order
					if buyOrder.PendingShares > 0 {
						buyOrders.Push(buyOrder)
					}
				}
			}
		}
	}
}

func (b *Book) AddTransaction(transaction *Transaction) {
	defer b.Wg.Done()

	sellingShares := transaction.SellingOrder.PendingShares
	buyingShares := transaction.BuyingOrder.PendingShares

	minShares := sellingShares
	if buyingShares < minShares {
		minShares = buyingShares
	}

	transaction.SellingOrder.Investor.UpdateAssetPosition(transaction.SellingOrder.Asset.ID, -minShares)
	transaction.AddSellOrderPendingShares(-minShares)

	transaction.BuyingOrder.Investor.UpdateAssetPosition(transaction.BuyingOrder.Asset.ID, minShares)
	transaction.AddBuyOrderPendingShares(-minShares)

	transaction.CalculateTotal(transaction.Shares, transaction.BuyingOrder.Price)
	transaction.CloseBuyOrder()
	transaction.CloseSellOrder()
	b.Transactions = append(b.Transactions, transaction)
}

func (b *Book) getOrCreateBuyOrderQueue(asset string) *OrderQueue {
	queue, exists := b.BuyOrderQueues[asset]
	if !exists {
		queue = NewOrderQueue()
		heap.Init(queue)
		b.BuyOrderQueues[asset] = queue
	}
	return queue
}

func (b *Book) getOrCreateSellOrderQueue(asset string) *OrderQueue {
	queue, exists := b.SellOrderQueues[asset]
	if !exists {
		queue = NewOrderQueue()
		heap.Init(queue)
		b.SellOrderQueues[asset] = queue
	}
	return queue
}
