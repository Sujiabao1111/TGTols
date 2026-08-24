package services

import (
	"testing"

	"gogogo/models/dtos"
)

func TestPickClosestPositiveAmountPrefersExpectedCryptoAmount(t *testing.T) {
	got := pickClosestPositiveAmount(1.12, "1.12", "17360")
	if got != 1.12 {
		t.Fatalf("expected 1.12, got %v", got)
	}
}

func TestPickClosestPositiveAmountFallsBackToOnlyValidAmount(t *testing.T) {
	got := pickClosestPositiveAmount(1.12, "", "1.12")
	if got != 1.12 {
		t.Fatalf("expected 1.12, got %v", got)
	}
}

func TestPickClosestPositiveAmountReturnsZeroWhenMissing(t *testing.T) {
	got := pickClosestPositiveAmount(1.12, "", "abc")
	if got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestFormatLegacyGatewayAmount(t *testing.T) {
	if got := formatLegacyGatewayAmount(10000); got != "10000" {
		t.Fatalf("expected integer amount, got %q", got)
	}
	if got := formatLegacyGatewayAmount(10000.25); got != "10000.25" {
		t.Fatalf("expected decimal amount, got %q", got)
	}
}

func TestNormalizePaymentNotifyQuantixCoreFields(t *testing.T) {
	notify := normalizePaymentNotify(dtos.PaymentNotifyRequest{
		MerchantNo:      "10036",
		MerchantOrderNo: "ORD202606031341270400706",
		OrderNo:         "QC123456789",
		Amount:          15000,
		Status:          "PAID",
		Currency:        "php",
		Code:            "gcashh5",
	})

	if notify.MerchantID != "10036" {
		t.Fatalf("expected merchant id to normalize from merchantNo, got %q", notify.MerchantID)
	}
	if notify.OrderID != "ORD202606031341270400706" {
		t.Fatalf("expected order id to normalize from merchantOrderNo, got %q", notify.OrderID)
	}
	if notify.PlatOrderID != "QC123456789" {
		t.Fatalf("expected plat order id to normalize from orderNo, got %q", notify.PlatOrderID)
	}
	if notify.RawAmount != "15000" {
		t.Fatalf("expected raw amount to preserve minor-unit amount, got %q", notify.RawAmount)
	}
	if notify.Currency != "PHP" {
		t.Fatalf("expected currency to normalize to PHP, got %q", notify.Currency)
	}
	if notify.Code != "GCASHH5" {
		t.Fatalf("expected code to normalize to uppercase, got %q", notify.Code)
	}
}
