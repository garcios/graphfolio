package decimalpb_test

import (
	"testing"

	"pkg/decimalpb"

	"github.com/shopspring/decimal"
)

func TestToProtoAndFromProto(t *testing.T) {
	orig, err := decimal.NewFromString("124532.89")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p := decimalpb.ToProto(orig)
	if p.GetValue() != "124532.89" {
		t.Errorf("got %q, want %q", p.GetValue(), "124532.89")
	}

	recovered, err := decimalpb.FromProto(p)
	if err != nil {
		t.Fatalf("failed to convert from proto: %v", err)
	}
	if !recovered.Equal(orig) {
		t.Errorf("recovered %s != orig %s", recovered, orig)
	}
}

func TestNilProto(t *testing.T) {
	_, err := decimalpb.FromProto(nil)
	if err == nil {
		t.Fatal("expected error on nil proto decimal, got nil")
	}

	_, _, err = decimalpb.MoneyFromProto(nil)
	if err == nil {
		t.Fatal("expected error on nil proto money, got nil")
	}
}

func TestMoneyConversion(t *testing.T) {
	d := decimal.NewFromFloat(555.75)
	m := decimalpb.MoneyToProto(d, "USD")

	if m.GetCurrencyCode() != "USD" {
		t.Errorf("got currency %q, want USD", m.GetCurrencyCode())
	}
	if m.GetAmount().GetValue() != "555.75" {
		t.Errorf("got amount %q, want 555.75", m.GetAmount().GetValue())
	}

	amount, ccy, err := decimalpb.MoneyFromProto(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ccy != "USD" {
		t.Errorf("got currency %q, want USD", ccy)
	}
	if !amount.Equal(d) {
		t.Errorf("got amount %v, want %v", amount, d)
	}
}

func TestFromStringAndMoneyFromString(t *testing.T) {
	p, err := decimalpb.FromString("42.15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.GetValue() != "42.15" {
		t.Errorf("got %q, want 42.15", p.GetValue())
	}

	m, err := decimalpb.MoneyFromString("100.50", "AUD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.GetCurrencyCode() != "AUD" || m.GetAmount().GetValue() != "100.5" {
		t.Errorf("got %v, want 100.5 AUD", m)
	}

	_, err = decimalpb.FromString("not-a-number")
	if err == nil {
		t.Error("expected error parsing invalid string, got nil")
	}
}
