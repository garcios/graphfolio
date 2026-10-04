package decimalpb

import (
	"errors"
	"fmt"

	commonpb "graphfolio/proto/common/v1"

	"github.com/shopspring/decimal"
)

var (
	ErrNilDecimal = errors.New("decimalpb: nil proto decimal")
	ErrNilMoney   = errors.New("decimalpb: nil proto money")
)

// ToProto converts a shopspring decimal.Decimal to a *commonpb.Decimal.
func ToProto(d decimal.Decimal) *commonpb.Decimal {
	return &commonpb.Decimal{
		Value: d.String(),
	}
}

// FromProto converts a *commonpb.Decimal to a shopspring decimal.Decimal.
func FromProto(p *commonpb.Decimal) (decimal.Decimal, error) {
	if p == nil {
		return decimal.Zero, ErrNilDecimal
	}
	d, err := decimal.NewFromString(p.GetValue())
	if err != nil {
		return decimal.Zero, fmt.Errorf("decimalpb: invalid decimal string %q: %w", p.GetValue(), err)
	}
	return d, nil
}

// MustFromProto converts a *commonpb.Decimal to a decimal.Decimal, panicking on error.
func MustFromProto(p *commonpb.Decimal) decimal.Decimal {
	d, err := FromProto(p)
	if err != nil {
		panic(err)
	}
	return d
}

// MoneyToProto converts an amount and currency to a *commonpb.Money.
func MoneyToProto(d decimal.Decimal, currencyCode string) *commonpb.Money {
	return &commonpb.Money{
		Amount:       ToProto(d),
		CurrencyCode: currencyCode,
	}
}

// MoneyFromProto converts a *commonpb.Money to an amount decimal.Decimal and currency string.
func MoneyFromProto(m *commonpb.Money) (decimal.Decimal, string, error) {
	if m == nil {
		return decimal.Zero, "", ErrNilMoney
	}
	d, err := FromProto(m.GetAmount())
	if err != nil {
		return decimal.Zero, "", err
	}
	return d, m.GetCurrencyCode(), nil
}

// FromString creates a *commonpb.Decimal from a string representation.
func FromString(s string) (*commonpb.Decimal, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return nil, fmt.Errorf("decimalpb: invalid decimal string %q: %w", s, err)
	}
	return ToProto(d), nil
}

// MoneyFromString creates a *commonpb.Money from an amount string and currency string.
func MoneyFromString(s string, currencyCode string) (*commonpb.Money, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return nil, fmt.Errorf("decimalpb: invalid money amount string %q: %w", s, err)
	}
	return MoneyToProto(d, currencyCode), nil
}
