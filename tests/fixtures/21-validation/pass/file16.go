// Go defines 5 statuses
type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusConfirmed  OrderStatus = "confirmed"
	StatusProcessing OrderStatus = "processing"   // TS MISSING
	StatusShipped    OrderStatus = "shipped"
	StatusDelivered  OrderStatus = "delivered"    // TS MISSING
)
