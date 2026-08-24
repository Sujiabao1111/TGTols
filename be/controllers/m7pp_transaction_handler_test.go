package controllers

import (
	"gogogo/provider"
	"testing"
	"time"
)

func TestBuildM7PPTransactionRecordKeyUsesExternalTransactionID(t *testing.T) {
	tx := provider.M7PPTransaction{BetID: "bet-1", ExternalTransactionID: "transaction-2"}
	if got, want := buildM7PPTransactionRecordKey(tx), "m7pp|tx|transaction-2"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestSplitM7PPTransactionWindowsKeepsProviderRangesSmall(t *testing.T) {
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(25 * time.Hour)
	windows := splitM7PPTransactionWindows(from, to)
	if len(windows) != 5 {
		t.Fatalf("got %d windows want 5", len(windows))
	}
	if windows[0][0] != from || windows[len(windows)-1][1] != to {
		t.Fatalf("windows do not cover requested range: %+v", windows)
	}
	for _, window := range windows {
		if duration := window[1].Sub(window[0]); duration <= 0 || duration > m7ppTransactionMaxWindow {
			t.Fatalf("invalid window duration %s", duration)
		}
	}
}

func TestBuildM7PPTransactionRecordKeyKeepsMultiplePayoutsForBet(t *testing.T) {
	first := provider.M7PPTransaction{Username: "platu1", BetID: "bet-1", RoundID: "round-1", UpdateTime: 10, WinAmount: 1}
	second := first
	second.UpdateTime = 11
	second.WinAmount = 2
	if buildM7PPTransactionRecordKey(first) == buildM7PPTransactionRecordKey(second) {
		t.Fatal("multiple payout rows for one bet must have different record keys")
	}
}
