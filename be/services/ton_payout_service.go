package services

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"

	"gogogo/helpers"
	"gogogo/models/dtos"
)

const defaultTONMainnetConfigURL = "https://ton-blockchain.github.io/global.config.json"

type tonPayoutService struct {
	wallet *wallet.Wallet
	client *liteclient.ConnectionPool
}

func newTONPayoutService(ctx context.Context, cfg helpers.TONConfig) (*tonPayoutService, error) {
	if !cfg.Enabled {
		return nil, errors.New("TON payouts are disabled")
	}
	if !strings.EqualFold(strings.TrimSpace(cfg.Network), "mainnet") {
		return nil, errors.New("TON payout network must be mainnet")
	}
	words := strings.Fields(cfg.Mnemonic)
	if len(words) != 24 {
		return nil, errors.New("TON mnemonic must contain exactly 24 words")
	}
	expected, err := address.ParseAddr(strings.TrimSpace(cfg.WalletAddress))
	if err != nil {
		return nil, errors.New("configured TON payout wallet address is invalid")
	}

	configURL := strings.TrimSpace(cfg.GlobalConfigURL)
	if configURL == "" {
		configURL = defaultTONMainnetConfigURL
	}
	client := liteclient.NewConnectionPool()
	globalCfg, err := liteclient.GetConfigFromUrl(ctx, configURL)
	if err != nil {
		return nil, fmt.Errorf("load TON mainnet config: %w", err)
	}
	if err := client.AddConnectionsFromConfig(ctx, globalCfg); err != nil {
		return nil, fmt.Errorf("connect TON mainnet: %w", err)
	}
	api := ton.NewAPIClient(client, ton.ProofCheckPolicyFast).WithRetryTimeout(3, 5*time.Second)
	api.SetTrustedBlockFromConfig(globalCfg)

	walletVersion := strings.ToLower(strings.TrimSpace(cfg.WalletVersion))
	var hotWallet *wallet.Wallet
	switch walletVersion {
	case "", "v4r2":
		hotWallet, err = wallet.FromSeed(api, words, wallet.V4R2)
	case "v5r1", "w5", "w5r1":
		hotWallet, err = wallet.FromSeed(api, words, wallet.ConfigV5R1Final{NetworkGlobalID: wallet.MainnetGlobalID})
	default:
		return nil, fmt.Errorf("unsupported TON wallet version %q", cfg.WalletVersion)
	}
	if err != nil {
		return nil, fmt.Errorf("derive TON payout wallet: %w", err)
	}
	if hotWallet.WalletAddress().StringRaw() != expected.StringRaw() {
		client.Stop()
		return nil, errors.New("TON mnemonic does not match configured payout wallet address")
	}
	return &tonPayoutService{wallet: hotWallet, client: client}, nil
}

func (s *tonPayoutService) close() {
	if s != nil && s.client != nil {
		s.client.Stop()
	}
}

func (s *tonPayoutService) transfer(ctx context.Context, destination string, amountTON float64, orderID string) (string, error) {
	to, err := address.ParseAddr(strings.TrimSpace(destination))
	if err != nil {
		return "", errors.New("invalid destination TON address")
	}
	if amountTON <= 0 {
		return "", errors.New("TON payout amount must be greater than zero")
	}
	coins, err := tlb.FromTON(fmt.Sprintf("%.9f", amountTON))
	if err != nil {
		return "", fmt.Errorf("convert TON payout amount: %w", err)
	}
	transfer, err := s.wallet.BuildTransfer(to, coins, false, orderID)
	if err != nil {
		return "", fmt.Errorf("build TON payout transfer: %w", err)
	}
	stickyCtx := s.client.StickyContext(ctx)
	tx, _, err := s.wallet.SendWaitTransaction(stickyCtx, transfer)
	if err != nil {
		return "", fmt.Errorf("broadcast or confirm TON payout: %w", err)
	}
	if tx == nil || len(tx.Hash) == 0 {
		return "", errors.New("TON payout confirmed without transaction hash")
	}
	return base64.RawURLEncoding.EncodeToString(tx.Hash), nil
}

func (s *PaymentService) submitTONWithdraw(ctx context.Context, order *dtos.WithdrawOrder) (string, string, error) {
	cfg := tonWithdrawConfig()
	if _, err := canonicalTONAddress(order.Account); err != nil {
		return "", "", errors.New("invalid destination TON address")
	}
	rate := s.CurrentTONRate(ctx)
	if rate <= 0 {
		return "", "", errors.New("invalid TON/USD exchange rate")
	}
	amountUSD := order.Cost
	if amountUSD <= 0 {
		amountUSD = order.Amount
	}
	// TON withdrawals charge a fixed 2 USD fee. The full requested amount
	// remains recorded in the order so failed payouts can refund it in full.
	const tonWithdrawFeeUSD = 2.0
	netAmountUSD := amountUSD - tonWithdrawFeeUSD
	if netAmountUSD <= 0 {
		return "", "", errors.New("TON withdrawal amount must exceed the 2 USD fee")
	}
	amountTON := netAmountUSD / rate
	payout, err := newTONPayoutService(ctx, cfg)
	if err != nil {
		return "", "", err
	}
	defer payout.close()
	txHash, err := payout.transfer(ctx, order.Account, amountTON, order.OrderID)
	if err != nil {
		return "", "", err
	}
	return txHash, fmt.Sprintf("TON payout confirmed: %.9f TON", amountTON), nil
}
