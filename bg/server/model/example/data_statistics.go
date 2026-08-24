package example

type DataStatisticsOverview struct {
	TotalRegisteredUsers int64   `json:"totalRegisteredUsers"`
	TotalRechargeUsers   int64   `json:"totalRechargeUsers"`
	TotalRechargeAmount  float64 `json:"totalRechargeAmount"`
	TotalWithdrawAmount  float64 `json:"totalWithdrawAmount"`
	TotalPaymentGap      float64 `json:"totalPaymentGap"`
	TodayWinLoss         float64 `json:"todayWinLoss"`
	PlatformWinLoss      float64 `json:"platformWinLoss"`
}

type DataStatisticsDaily struct {
	Date           string  `json:"date" gorm:"column:stat_date"`
	RegisterUsers  int64   `json:"registerUsers"`
	ActiveUsers    int64   `json:"activeUsers"`
	RechargeUsers  int64   `json:"rechargeUsers"`
	RechargeAmount float64 `json:"rechargeAmount"`
	WithdrawUsers  int64   `json:"withdrawUsers"`
	WithdrawAmount float64 `json:"withdrawAmount"`
	PaymentGap     float64 `json:"paymentGap"`
	DailyWinLoss   float64 `json:"dailyWinLoss"`
}

type DataStatisticsResult struct {
	Overview  DataStatisticsOverview `json:"overview"`
	Daily     []DataStatisticsDaily  `json:"daily"`
	StartDate string                 `json:"startDate"`
	EndDate   string                 `json:"endDate"`
}
