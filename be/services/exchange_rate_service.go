package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"gogogo/models"
	"gogogo/redis"
)

const (
	ExchangeRateAPI = "https://api.exchangerate-api.com/v4/latest/IDR"
	CacheKey        = "exchange_rates"
	CacheDuration   = 2 * time.Hour // 缓存2小时
)

// ExchangeRateResponse 汇率 API 响应
type ExchangeRateResponse struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

// ExchangeRateService 汇率服务
type ExchangeRateService struct {
	// 内存缓存作为备用
	memoryCache map[string]float64
	cacheTime   time.Time
}

// NewExchangeRateService 创建汇率服务
func NewExchangeRateService() *ExchangeRateService {
	return &ExchangeRateService{
		memoryCache: make(map[string]float64),
	}
}

// FetchAndCache 从 API 获取汇率并缓存
func (s *ExchangeRateService) FetchAndCache() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_ = ctx // suppress unused warning
	rates, err := s.fetchFromAPI()
	if err != nil {
		return fmt.Errorf("fetch from api failed: %w", err)
	}

	rates = normalizeRatesToIDRBase(rates)

	// 更新内存缓存
	s.memoryCache = rates
	s.cacheTime = time.Now()

	// 写入 Redis 缓存
	data, err := json.Marshal(rates)
	if err != nil {
		return fmt.Errorf("marshal rates failed: %w", err)
	}

	rdb := redis.GetRDbInstance()
	if rdb != nil && rdb.C != nil {
		if err := rdb.StoreTicket(CacheKey, string(data), CacheDuration); err != nil {
			fmt.Printf("[ExchangeRate] Failed to cache to redis: %v\n", err)
			// Redis 失败不影响主流程，内存缓存已更新
		}
	}

	fmt.Println("[ExchangeRate] Exchange rates updated successfully")
	return nil
}

// GetRates 获取所有汇率
func (s *ExchangeRateService) GetRates(ctx context.Context) (map[string]float64, error) {
	// 优先从 Redis 获取
	rdb := redis.GetRDbInstance()
	if rdb != nil && rdb.C != nil {
		data, err := rdb.FindTicket(CacheKey)
		if err == nil && data != "" {
			var rates map[string]float64
			if err := json.Unmarshal([]byte(data), &rates); err == nil {
				return normalizeRatesToIDRBase(rates), nil
			}
		}
	}

	// 从内存缓存获取
	if len(s.memoryCache) > 0 && time.Since(s.cacheTime) < CacheDuration {
		return normalizeRatesToIDRBase(s.memoryCache), nil
	}

	// 从数据库获取
	db := models.GetInstance()
	if db != nil && db.DbInstance != nil {
		rates, err := s.getRatesFromDB()
		if err == nil && len(rates) > 0 {
			return normalizeRatesToIDRBase(rates), nil
		}
	}

	// 返回默认汇率
	return normalizeRatesToIDRBase(s.getDefaultRates()), nil
}

// GetRate 获取指定货币对 IDR 的汇率
// 返回: 1 IDR = ? Target
func (s *ExchangeRateService) GetRate(ctx context.Context, targetCode string) (float64, error) {
	rates, err := s.GetRates(ctx)
	if err != nil {
		return 0, err
	}

	rate, ok := rates[targetCode]
	if !ok {
		return 0, fmt.Errorf("currency not found: %s", targetCode)
	}

	return rate, nil
}

// Convert 转换金额
// amount: 原金额
// from: 原货币代码
// to: 目标货币代码
func (s *ExchangeRateService) Convert(ctx context.Context, amount float64, from, to string) (float64, error) {
	if from == to {
		return amount, nil
	}

	rates, err := s.GetRates(ctx)
	if err != nil {
		return 0, err
	}

	// 获取汇率，基于 IDR
	fromRate, fromOk := rates[from]
	toRate, toOk := rates[to]

	if !fromOk {
		return 0, fmt.Errorf("source currency not found: %s", from)
	}
	if !toOk {
		return 0, fmt.Errorf("target currency not found: %s", to)
	}

	// amount / fromRate = IDR 金额
	// IDR 金额 * toRate = 目标货币金额
	idrAmount := amount / fromRate
	return idrAmount * toRate, nil
}

// ConvertFromIDR 从 IDR 转换到目标货币
func (s *ExchangeRateService) ConvertFromIDR(ctx context.Context, amountIDR float64, targetCode string) (float64, error) {
	rate, err := s.GetRate(ctx, targetCode)
	if err != nil {
		return 0, err
	}

	return amountIDR * rate, nil
}

// fetchFromAPI 从 API 获取汇率
func (s *ExchangeRateService) fetchFromAPI() (map[string]float64, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(ExchangeRateAPI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result ExchangeRateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return normalizeRatesToIDRBase(result.Rates), nil
}

// getRatesFromDB 从数据库获取汇率
func (s *ExchangeRateService) getRatesFromDB() (map[string]float64, error) {
	db := models.GetInstance()
	if db == nil || db.DbInstance == nil {
		return nil, fmt.Errorf("db not initialized")
	}

	type RateRow struct {
		Code string  `gorm:"column:code"`
		Rate float64 `gorm:"column:rate"`
	}

	var rows []RateRow
	if err := db.DbInstance.Table("exchange_rates").Select("code, rate").Find(&rows).Error; err != nil {
		return nil, err
	}

	rates := make(map[string]float64, len(rows))
	for _, row := range rows {
		rates[row.Code] = row.Rate
	}

	return rates, nil
}

// getDefaultRates 获取默认汇率
func (s *ExchangeRateService) getDefaultRates() map[string]float64 {
	return map[string]float64{
		"IDR": 1,
		"USD": 0.000061, // 1 IDR = 0.000061 USD
		"CNY": 0.00044,  // 1 IDR = 0.00044 CNY
		"RUB": 0.0052,   // 1 IDR = 0.0052 RUB
		"EUR": 0.000057, // 1 IDR = 0.000057 EUR
		"PHP": 0.0035,   // 1 IDR = 0.0035 PHP
	}
}

// GetSupportedCurrencies 获取支持的货币列表
func (s *ExchangeRateService) GetSupportedCurrencies() []string {
	return []string{"IDR", "USD", "CNY", "RUB", "EUR", "PHP"}
}

func normalizeRatesToIDRBase(rates map[string]float64) map[string]float64 {
	if len(rates) == 0 {
		return map[string]float64{}
	}

	normalized := make(map[string]float64, len(rates)+1)
	for code, rate := range rates {
		if rate > 0 {
			normalized[code] = rate
		}
	}

	idrRate, hasIDR := normalized["IDR"]
	if !hasIDR || idrRate <= 0 {
		return normalized
	}

	// 数据库中的汇率存的是 "1 USD = ? 目标货币"，
	// 但前端和该 API 需要 "1 IDR = ? 目标货币"。
	if idrRate > 10 && !almostEqual(idrRate, 1) {
		for code, rate := range normalized {
			normalized[code] = rate / idrRate
		}
		normalized["IDR"] = 1
		normalized["USD"] = 1 / idrRate
		return normalized
	}

	normalized["IDR"] = 1
	return normalized
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}
