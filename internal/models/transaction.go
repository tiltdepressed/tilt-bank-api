package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID            string
	FromAccountID string
	ToAccountID   string
	Amount        decimal.Decimal
	Status        string
	CreatedAt     time.Time
}
