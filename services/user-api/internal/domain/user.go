package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user identity and personal preferences.
type User struct {
	ID                 uuid.UUID
	Email              string
	PasswordHash       string
	IsActive           bool
	DisplayName        string
	DisplayCurrency    string
	Theme              string
	Role                   string // "USER" or "ADMIN"
	MustChangePassword     bool   // true if bootstrap password change required
	FinancialYearEndMonth  int    // 1..12 (1 = Jan, 12 = Dec)
	FinancialYearEndDay    int    // 1..31
	EmailVerifiedAt        *time.Time
	LastLoginAt            *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

const (
	RoleUser  = "USER"
	RoleAdmin = "ADMIN"
)

// IsValidRole returns true if the role is USER or ADMIN.
func IsValidRole(role string) bool {
	return role == RoleUser || role == RoleAdmin
}

// Session represents an active refresh token session for a user.
type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash string
	FamilyID         uuid.UUID
	IsRevoked        bool
	UserAgent        string
	IPAddress        string
	ExpiresAt        time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
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

// IsValidFinancialYearEnd validates that the given month and day represent a viable calendar date.
// February allows day 29 to accommodate leap-year enders (e.g. South Africa).
func IsValidFinancialYearEnd(month, day int) bool {
	if month < 1 || month > 12 || day < 1 || day > 31 {
		return false
	}
	// Max days per month in a leap year
	maxDays := []int{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	return day <= maxDays[month-1]
}
