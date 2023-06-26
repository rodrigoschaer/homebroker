package entity

type OrderType string

const (
	Buy  OrderType = "BUY"
	Sell OrderType = "SELL"
)

type OrderStatus string

const (
	Closed    OrderStatus = "CLOSED"
	Open      OrderStatus = "OPEN"
	Cancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID            string
	Investor      *Investor
	Asset         *Asset
	Shares        int
	PendingShares int
	Price         float64
	Type          OrderType
	Status        OrderStatus
	Transactions  []*Transaction
}

func NewOrder(orderID string, investor *Investor, asset *Asset, shares int, price float64, orderType OrderType, status OrderStatus) *Order {
	return &Order{
		ID:            orderID,
		Investor:      investor,
		Asset:         asset,
		Shares:        shares,
		PendingShares: shares,
		Price:         price,
		Type:          orderType,
		Status:        "OPEN",
		Transactions:  []*Transaction{},
	}
}
