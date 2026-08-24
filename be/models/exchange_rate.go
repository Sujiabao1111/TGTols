package models

import "fmt"

// GetExchangeRate 获取指定货币对美元的汇率
// 返回: 1 USD = ? Target
func (db *DbWrapper) GetExchangeRate(targetCode string) (float64, error) {
	type rateResult struct {
		Rate float64
	}

	var res rateResult
	result := db.DbInstance.Table("exchange_rates").
		Select("rate").
		Where("code = ?", targetCode).
		Limit(1).
		Find(&res)

	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, fmt.Errorf("exchange rate not configured: %s", targetCode)
	}
	if res.Rate <= 0 {
		return 0, fmt.Errorf("invalid exchange rate config: %s rate is %f", targetCode, res.Rate)
	}

	return res.Rate, nil
}
