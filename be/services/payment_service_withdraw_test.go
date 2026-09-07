package services

import (
	"testing"

	"gogogo/helpers"
	"gogogo/models/dtos"
)

func TestBuildLegacyGatewayWithdrawParams(t *testing.T) {
	cfg := &helpers.PaymentConfig{
		MerchantID: " 15700 ",
		NotifyURL:  "https://example.com/api/payments/notify",
	}
	order := &dtos.WithdrawOrder{
		OrderID:     " WDR123 ",
		Amount:      50000,
		Type:        "ewallet",
		DstCode:     "DANA",
		Account:     " 08123456789 ",
		AccountName: " Player One ",
		Phone:       " 08123456789 ",
		Email:       " user@example.com ",
		Address:     " Jakarta ",
	}

	params := buildLegacyGatewayWithdrawParams(order, cfg, withdrawFeeConfigValues{Enabled: true, FeeRate: defaultWithdrawPlatformFeeRate})

	assertEqual(t, params["memberId"], "15700")
	assertEqual(t, params["orderId"], "WDR123")
	assertEqual(t, params["amount"], "49850")
	assertEqual(t, params["type"], "ewallet")
	assertEqual(t, params["dstCode"], "DANA")
	assertEqual(t, params["account"], "08123456789")
	assertEqual(t, params["name"], "Player One")
	assertEqual(t, params["phone"], "08123456789")
	assertEqual(t, params["email"], "user@example.com")
	assertEqual(t, params["address"], "Jakarta")
	assertEqual(t, params["notifyUrl"], "https://example.com/api/withdraw/notify")
}

func TestBuildQuantixCoreWithdrawPayloadForGcash(t *testing.T) {
	cfg := &helpers.PaymentConfig{
		NotifyURL: "https://example.com/api/payments/notify",
	}
	quantixCfg := helpers.QuantixCoreRegionConfig{
		MerchantID: "MO0CKQIH",
	}
	order := &dtos.WithdrawOrder{
		UserID:      1853,
		OrderID:     "WDR2026061501",
		Amount:      300,
		Type:        "ewallet",
		DstCode:     "GCASH_QR",
		Account:     "09171234567",
		AccountName: "PLAYER ONE",
		Phone:       "09171234567",
		Email:       "user@example.com",
	}

	payload, err := buildQuantixCoreWithdrawPayload(order, paymentMethodConfig{}, "PHP", quantixCfg, cfg, withdrawFeeConfigValues{Enabled: true, FeeRate: defaultWithdrawPlatformFeeRate})
	if err != nil {
		t.Fatalf("buildQuantixCoreWithdrawPayload() error = %v", err)
	}

	assertEqual(t, payload["merchantNo"], "MO0CKQIH")
	assertEqual(t, payload["merchantOrderNo"], "WDR2026061501")
	assertEqual(t, payload["uid"], "1853")
	assertEqual(t, payload["amount"], "29910")
	assertEqual(t, payload["currency"], "PHP")
	assertEqual(t, payload["accountType"], "GCASH")
	assertEqual(t, payload["account"], "09171234567")
	assertEqual(t, payload["accountName"], "PLAYER ONE")
	assertEqual(t, payload["phone"], "09171234567")
	assertEqual(t, payload["email"], "user@example.com")
	assertEqual(t, payload["callback"], "https://example.com/api/withdraw/notify")

	if _, ok := payload["bankCode"]; ok {
		t.Fatal("GCASH wallet payout must not send bankCode")
	}
}

func TestBuildQuantixCoreWithdrawPayloadForIndonesiaBank(t *testing.T) {
	cfg := &helpers.PaymentConfig{}
	quantixCfg := helpers.QuantixCoreRegionConfig{
		MerchantID: "15700",
		NotifyURL:  "https://example.com/withdraw-callback",
	}
	order := &dtos.WithdrawOrder{
		UserID:      77,
		OrderID:     "WDR77",
		Amount:      100000,
		Type:        "bankcard",
		DstCode:     "BCA",
		Account:     "1234567890",
		AccountName: "PLAYER TWO",
	}

	payload, err := buildQuantixCoreWithdrawPayload(order, paymentMethodConfig{}, "IDR", quantixCfg, cfg, withdrawFeeConfigValues{Enabled: true, FeeRate: defaultWithdrawPlatformFeeRate})
	if err != nil {
		t.Fatalf("buildQuantixCoreWithdrawPayload() error = %v", err)
	}

	assertEqual(t, payload["amount"], "99700")
	assertEqual(t, payload["currency"], "IDR")
	assertEqual(t, payload["accountType"], "PERSONAL_BANK")
	assertEqual(t, payload["bankCode"], "BCA")
	assertEqual(t, payload["callback"], "https://example.com/withdraw-callback/withdraw/notify")
}

func TestWithdrawGatewayAmountDeductsPlatformFee(t *testing.T) {
	feeConfig := withdrawFeeConfigValues{Enabled: true, FeeRate: defaultWithdrawPlatformFeeRate}
	if got := withdrawGatewayAmount(100000, "IDR", feeConfig); got != 99700 {
		t.Fatalf("expected IDR gateway amount 99700, got %v", got)
	}
	if got := withdrawGatewayAmount(300, "PHP", feeConfig); got != 299.1 {
		t.Fatalf("expected PHP gateway amount 299.10, got %v", got)
	}
}

func TestWithdrawGatewayAmountCanDisablePlatformFee(t *testing.T) {
	if got := withdrawGatewayAmount(100000, "IDR", withdrawFeeConfigValues{Enabled: false, FeeRate: defaultWithdrawPlatformFeeRate}); got != 100000 {
		t.Fatalf("expected disabled fee to keep amount 100000, got %v", got)
	}
}

func TestShouldAutoSubmitWithdraw(t *testing.T) {
	tests := []struct {
		name      string
		amountUSD float64
		want      bool
	}{
		{name: "below limit", amountUSD: 19.99, want: true},
		{name: "equal to limit requires review", amountUSD: 20, want: false},
		{name: "above limit requires review", amountUSD: 20.01, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := shouldAutoSubmitWithdraw(test.amountUSD); got != test.want {
				t.Fatalf("shouldAutoSubmitWithdraw(%v) = %v, want %v", test.amountUSD, got, test.want)
			}
		})
	}
}

func assertEqual(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
