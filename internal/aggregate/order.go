package aggregate

import "time"

const (
	OrderStatusPreparing = "PREPARING_PACKAGE"
	OrderStatusInTransit = "IN_TRANSIT"
	OrderStatusDelivered = "DELIVERED"
)

type Order struct {
	ID             string    `json:"order_id"`
	UserID         string    `json:"user_id"`
	TotalItems     int       `json:"total_items"`
	TransactionFee float64   `json:"transaction_fee"`
	NetTotal       float64   `json:"net_total"`
	Status         string    `json:"status"`
	LastUpdate     time.Time `json:"last_update"`
}
