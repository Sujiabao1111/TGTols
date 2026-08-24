package example

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

const (
	defaultIDRExchangeRate = 16000.0
	defaultPHPExchangeRate = 56.0
)

func normalizeExamplePaymentCurrency(dstCode string) string {
	code := strings.ToUpper(strings.TrimSpace(dstCode))
	switch {
	case code == "PHP":
		return "PHP"
	case code == "":
		return "IDR"
	case strings.HasPrefix(code, "GCASH"),
		strings.HasPrefix(code, "BDO"),
		strings.HasPrefix(code, "BPI"),
		strings.HasPrefix(code, "METROBANK"),
		strings.HasPrefix(code, "LANDBANK"),
		strings.HasPrefix(code, "PNB"),
		strings.HasPrefix(code, "SECURITY"),
		strings.HasPrefix(code, "UNIONBANK"),
		strings.HasPrefix(code, "CHINABANK"),
		strings.HasPrefix(code, "RCBC"),
		strings.HasPrefix(code, "EASTWEST"):
		return "PHP"
	default:
		return "IDR"
	}
}

func defaultExampleExchangeRate(currency string) float64 {
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "PHP":
		return defaultPHPExchangeRate
	case "IDR":
		return defaultIDRExchangeRate
	default:
		return 1
	}
}

func getExampleExchangeRate(db *gorm.DB, currency string) (float64, error) {
	type rateRow struct {
		Rate float64 `gorm:"column:rate"`
	}

	normalizedCurrency := strings.ToUpper(strings.TrimSpace(currency))
	var row rateRow
	if err := db.Table("exchange_rates").
		Select("rate").
		Where("code = ?", normalizedCurrency).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return defaultExampleExchangeRate(normalizedCurrency), nil
		}
		return 0, err
	}
	if row.Rate <= 0 {
		return defaultExampleExchangeRate(normalizedCurrency), nil
	}
	return row.Rate, nil
}

func exampleCurrencyCaseSQL(dstCodeExpr string) string {
	return fmt.Sprintf(`CASE
        WHEN UPPER(TRIM(%s)) = 'PHP'
            OR UPPER(TRIM(%s)) LIKE 'GCASH%%'
            OR UPPER(TRIM(%s)) LIKE 'BDO%%'
            OR UPPER(TRIM(%s)) LIKE 'BPI%%'
            OR UPPER(TRIM(%s)) LIKE 'METROBANK%%'
            OR UPPER(TRIM(%s)) LIKE 'LANDBANK%%'
            OR UPPER(TRIM(%s)) LIKE 'PNB%%'
            OR UPPER(TRIM(%s)) LIKE 'SECURITY%%'
            OR UPPER(TRIM(%s)) LIKE 'UNIONBANK%%'
            OR UPPER(TRIM(%s)) LIKE 'CHINABANK%%'
            OR UPPER(TRIM(%s)) LIKE 'RCBC%%'
            OR UPPER(TRIM(%s)) LIKE 'EASTWEST%%'
        THEN 'PHP'
        ELSE 'IDR'
    END`,
		dstCodeExpr,
		dstCodeExpr,
		dstCodeExpr,
		dstCodeExpr,
		dstCodeExpr,
		dstCodeExpr,
		dstCodeExpr,
		dstCodeExpr,
		dstCodeExpr,
		dstCodeExpr,
		dstCodeExpr,
		dstCodeExpr,
	)
}

func exampleRateFallbackSQL(dstCodeExpr string) string {
	return fmt.Sprintf(`CASE
        WHEN %s = 'PHP' THEN %.0f
        ELSE %.0f
    END`, exampleCurrencyCaseSQL(dstCodeExpr), defaultPHPExchangeRate, defaultIDRExchangeRate)
}
