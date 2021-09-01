package event

import "time"

type OrderIssued struct {
	ID             string    `json:"order_id"`
	UserID         string    `json:"user_id"`
	TotalItems     int       `json:"total_items"`
	TransactionFee float64   `json:"transaction_fee"`
	NetTotal       float64   `json:"net_total"`
	Status         string    `json:"order_status"`
	IssuedAt       time.Time `json:"issued_at"`
}

type OrderUpdated struct {
	ID        string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"order_status"`
	UpdatedAt time.Time `json:"updated_at"`
}
