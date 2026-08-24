package services

import (
	"math"
	"testing"
)

func TestNormalizeRatesToIDRBaseFromUSDBasedRates(t *testing.T) {
	rates := map[string]float64{
		"IDR": 15500,
		"PHP": 56,
		"RUB": 92.5,
	}

	normalized := normalizeRatesToIDRBase(rates)

	assertAlmostEqual(t, normalized["IDR"], 1)
	assertAlmostEqual(t, normalized["USD"], 1.0/15500.0)
	assertAlmostEqual(t, normalized["PHP"], 56.0/15500.0)
	assertAlmostEqual(t, normalized["RUB"], 92.5/15500.0)
}

func TestNormalizeRatesToIDRBaseKeepsIDRBasedRates(t *testing.T) {
	rates := map[string]float64{
		"IDR": 1,
		"USD": 0.000061,
		"PHP": 0.0035,
	}

	normalized := normalizeRatesToIDRBase(rates)

	assertAlmostEqual(t, normalized["IDR"], 1)
	assertAlmostEqual(t, normalized["USD"], 0.000061)
	assertAlmostEqual(t, normalized["PHP"], 0.0035)
}

func assertAlmostEqual(t *testing.T, got, want float64) {
	t.Helper()

	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %.12f, want %.12f", got, want)
	}
}
