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
		,OrdersChanOut:   orderChanOut,
		Wg:              wg,
		BuyOrderQueues:  make(map[string]*OrderQueue),
		SellOrderQueues: make(map[string]*OrderQueue),
	}
}

func (b *Book) Trade() {
	for order := range b.OrdersChan {
		b.Wg.Add(1)
		go func(o *Order) {
			defer b.Wg.Done()
			b.processOrder(o)
		}(order)
	}
	b.Wg.Wait()
}

func (b *Book) processOrder(order *Order) {
	asset := order.Asset.ID

	buyOrders := b.getOrCreateBuyOrderQueue(asset)
	sellOrders := b.getOrCreateSellOrderQueue(asset)

	if order.OrderType == "BUY" {
		buyOrders.Push(order)
		go b.checkSellOrders(asset, buyOrders, sellOrders, order)
	} else if order.OrderType == "SELL" {
		sellOrders.Push(order)
		go b.checkBuyOrders(asset, buyOrders, sellOrders, order)
	}
}

func (b *Book) checkSellOrders(asset string, buyOrders, sellOrders *OrderQueue, buyOrder *Order) {
	if sellOrders.Len() > 0 && sellOrders.Orders[0].Price <= buyOrder.Price {
		sellOrder := sellOrders.Pop().(*Order)
		if sellOrder.PendingShares > 0 {
			transaction := NewTransaction(sellOrder, buyOrder, buyOrder.Shares, sellOrder.Price)
			b.AddTransaction(transaction)
			sellOrder.Transactions = append(sellOrder.Transactions, transaction)
			buyOrder.Transactions = append(buyOrder.Transactions, transaction)
			b.OrdersChanOut <- sellOrder
			b.OrdersChanOut <- buyOrder
			if sellOrder.PendingShares > 0 {
				sellOrders.Push(sellOrder)
			}
		}
	}
}

func (b *Book) checkBuyOrders(asset string, buyOrders, sellOrders *OrderQueue, sellOrder *Order) {
	if buyOrders.Len() > 0 && buyOrders.Orders[0].Price >= sellOrder.Price {
		buyOrder := buyOrders.Pop().(*Order)
		if buyOrder.PendingShares > 0 {
			transaction := NewTransaction(sellOrder, buyOrder, sellOrder.Shares, buyOrder.Price)
			b.AddTransaction(transaction)
			sellOrder.Transactions = append(sellOrder.Transactions, transaction)
			buyOrder.Transactions = append(buyOrder.Transactions, transaction)
			b.OrdersChanOut <- sellOrder
			b.OrdersChanOut <- buyOrder
			if buyOrder.PendingShares > 0 {
				buyOrders.Push(buyOrder)
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

	transaction.CalculateTotal(minShares, transaction.BuyingOrder.Price)
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
