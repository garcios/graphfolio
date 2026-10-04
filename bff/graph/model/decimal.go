package model

import (
	"fmt"
	"io"
	"strconv"

	"github.com/shopspring/decimal"
)

// Decimal wraps shopspring decimal.Decimal and implements graphql.Marshaler and graphql.Unmarshaler.
type Decimal decimal.Decimal

// AsDecimal unwraps to shopspring decimal.Decimal.
func (d Decimal) AsDecimal() decimal.Decimal {
	return decimal.Decimal(d)
}

// FromDecimal wraps a shopspring decimal.Decimal.
func FromDecimal(d decimal.Decimal) Decimal {
	return Decimal(d)
}

// String returns the string representation.
func (d Decimal) String() string {
	return decimal.Decimal(d).String()
}

// MarshalGQL implements graphql.Marshaler.
func (d Decimal) MarshalGQL(w io.Writer) {
	io.WriteString(w, strconv.Quote(decimal.Decimal(d).String()))
}

// UnmarshalGQL implements graphql.Unmarshaler.
func (d *Decimal) UnmarshalGQL(v any) error {
	switch val := v.(type) {
	case string:
		dec, err := decimal.NewFromString(val)
		if err != nil {
			return err
		}
		*d = Decimal(dec)
		return nil
	case float64:
		*d = Decimal(decimal.NewFromFloat(val))
		return nil
	case int64:
		*d = Decimal(decimal.NewFromInt(val))
		return nil
	case int:
		*d = Decimal(decimal.NewFromInt(int64(val)))
		return nil
	default:
		return fmt.Errorf("model: cannot unmarshal %T into Decimal", v)
	}
}
