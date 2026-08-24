package provider

import (
	"fmt"
	"strings"
	"time"
)

const providerSummaryTimeLayout = "2006-01-02T15:04:05.000Z07:00"

type PlayerSummaryResult struct {
	UserId   string  `json:"UserId"`
	LoginId  string  `json:"LoginId"`
	Count    int     `json:"Count"`
	Turnover float64 `json:"Turnover"`
	Bet      float64 `json:"Bet"`
	Win      float64 `json:"Win"`
	Winlose  float64 `json:"Winlose"`
	JPShare  float64 `json:"JPShare"`
	JPWin    float64 `json:"JPWin"`
}

type PlayerSummaryResponseV2 struct {
	Code    int                 `json:"Code"`
	Message string              `json:"Message"`
	Data    PlayerSummaryResult `json:"Data"`
}

type TransactionSummaryByPlayerRequest struct {
	AgentId          string `json:"AgentId"`
	LoginId          string `json:"LoginId"`
	SummaryDate      int    `json:"SummaryDate"`
	GameProviderCode int    `json:"GameProviderCode"`
	FromDate         string `json:"FromDate"`
	ToDate           string `json:"ToDate"`
}

func (c *HedoClient) GetPlayerDailySummaryV2(loginId string, dateStr string, agentid, agentapi string) (*PlayerSummaryResult, error) {
	loc := providerSummaryLocation()
	day, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return nil, err
	}

	return c.GetPlayerSummaryByRange(
		loginId,
		day,
		day.Add(24*time.Hour).Add(-1*time.Millisecond),
		agentid,
		agentapi,
	)
}

func (c *HedoClient) GetPlayerSummaryByRange(loginId string, fromDate time.Time, toDate time.Time, agentid, agentapi string) (*PlayerSummaryResult, error) {
	if toDate.Before(fromDate) {
		return nil, fmt.Errorf("invalid summary range: toDate before fromDate")
	}

	loc := providerSummaryLocation()
	req := TransactionSummaryByPlayerRequest{
		AgentId:          agentid,
		LoginId:          loginId,
		SummaryDate:      1,
		GameProviderCode: 0,
		FromDate:         fromDate.In(loc).Format(providerSummaryTimeLayout),
		ToDate:           toDate.In(loc).Format(providerSummaryTimeLayout),
	}

	var resp PlayerSummaryResponseV2
	res, err := c.client.R().
		SetBody(req).
		SetResult(&resp).
		Post(buildProviderReportURL(agentapi, "/Report/TransactionSummaryByPlayer"))
	if err != nil {
		return nil, fmt.Errorf("network error: %v", err)
	}
	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("provider http %d: %s", res.StatusCode(), strings.TrimSpace(res.String()))
	}
	if !isProviderReportSuccess(resp.Code, resp.Message) {
		message := strings.TrimSpace(resp.Message)
		if message == "" {
			message = strings.TrimSpace(res.String())
		}
		return nil, fmt.Errorf("provider error: code=%d message=%s", resp.Code, message)
	}

	summary := resp.Data
	if strings.TrimSpace(summary.LoginId) == "" {
		summary.LoginId = loginId
	}

	return &summary, nil
}

func providerSummaryLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("GMT+8", 8*60*60)
	}

	return loc
}
