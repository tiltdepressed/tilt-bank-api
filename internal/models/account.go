package models

import "github.com/shopspring/decimal"

type Account struct {
	ID       string
	UserID   string
	Balance  decimal.Decimal
	Currency string
}
