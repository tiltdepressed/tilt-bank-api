package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type TransactionStatus string

const (
	StatusPending TransactionStatus = "pending"
	StatusSuccess TransactionStatus = "success"
	StatusFailed  TransactionStatus = "failed"
)

type Transaction struct {
	ID            string
	FromAccountID string
	ToAccountID   string
	Amount        decimal.Decimal
	Status        TransactionStatus
	CreatedAt     time.Time
}
