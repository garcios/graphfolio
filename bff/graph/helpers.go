package graph

import (
	"bff/graph/model"

	commonpb "graphfolio/proto/common/v1"

	"github.com/shopspring/decimal"
)

func toModelMoney(m *commonpb.Money) *model.Money {
	if m == nil {
		return &model.Money{
			Amount:       model.Decimal(decimal.Zero),
			CurrencyCode: "USD",
		}
	}
	d, err := decimal.NewFromString(m.GetAmount().GetValue())
	if err != nil {
		d = decimal.Zero
	}
	return &model.Money{
		Amount:       model.Decimal(d),
		CurrencyCode: m.GetCurrencyCode(),
	}
}

func toModelDecimal(p *commonpb.Decimal) model.Decimal {
	if p == nil {
		return model.Decimal(decimal.Zero)
	}
	d, err := decimal.NewFromString(p.GetValue())
	if err != nil {
		return model.Decimal(decimal.Zero)
	}
	return model.Decimal(d)
}
