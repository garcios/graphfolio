package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user identity and personal preferences.
type User struct {
	ID              uuid.UUID
	Email           string
	DisplayName     string
	DisplayCurrency string
	Theme           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// CurrencyInfo represents metadata about a supported currency.
type CurrencyInfo struct {
	Code   string
	Name   string
	Symbol string
}

// SupportedCurrencies lists the officially supported display currencies in GraphFolio.
var SupportedCurrencies = []CurrencyInfo{
	{Code: "USD", Name: "US Dollar", Symbol: "$"},
	{Code: "EUR", Name: "Euro", Symbol: "€"},
	{Code: "GBP", Name: "British Pound", Symbol: "£"},
	{Code: "AUD", Name: "Australian Dollar", Symbol: "A$"},
	{Code: "CAD", Name: "Canadian Dollar", Symbol: "C$"},
	{Code: "JPY", Name: "Japanese Yen", Symbol: "¥"},
	{Code: "CHF", Name: "Swiss Franc", Symbol: "CHF"},
}

// IsSupportedCurrency returns true if the 3-letter currency code is in SupportedCurrencies.
func IsSupportedCurrency(code string) bool {
	for _, c := range SupportedCurrencies {
		if c.Code == code {
			return true
		}
	}
	return false
}

// IsValidTheme returns true if the theme is one of DARK, LIGHT, or SYSTEM.
func IsValidTheme(theme string) bool {
	switch theme {
	case "DARK", "LIGHT", "SYSTEM":
		return true
	default:
		return false
	}
}
