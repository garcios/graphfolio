package domain

import (
	"github.com/shopspring/decimal"
)

// Money represents an exact monetary quantity denominated in a specific ISO 4217 currency.
type Money struct {
	Amount       decimal.Decimal
	CurrencyCode string
}

func NewMoney(amount decimal.Decimal, currency string) Money {
	return Money{
		Amount:       amount,
		CurrencyCode: currency,
	}
}

func ZeroMoney(currency string) Money {
	return Money{
		Amount:       decimal.Zero,
		CurrencyCode: currency,
	}
}
