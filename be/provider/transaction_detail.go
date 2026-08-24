package provider

import (
	"fmt"
	"strings"
	"time"
)

type TransactionDetailRequest struct {
	AgentId          string `json:"AgentId"`
	LoginId          string `json:"LoginId"`
	GameProviderCode int    `json:"GameProviderCode"`
	FromDate         string `json:"FromDate"`
	ToDate           string `json:"ToDate"`
	TimeZone         int    `json:"TimeZone"`
	PageIndex        int    `json:"PageIndex"`
	RowPerPage       int    `json:"RowPerPage"`
}

type TransactionDetailResponse struct {
	TotalRecord int              `json:"TotalRecord"`
	TotalPage   int              `json:"TotalPage"`
	Data        []map[string]any `json:"Data"`
	Code        int              `json:"Code"`
	Message     string           `json:"Message"`
}

func (c *HedoClient) GetPlayerTransactionDetails(
	loginId string,
	fromDate time.Time,
	toDate time.Time,
	gameProviderCode int,
	pageIndex int,
	rowPerPage int,
	agentid string,
	agentapi string,
) (*TransactionDetailResponse, error) {
	if toDate.Before(fromDate) {
		return nil, fmt.Errorf("invalid detail range: toDate before fromDate")
	}
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if rowPerPage <= 0 {
		rowPerPage = 200
	}

	loc := providerSummaryLocation()
	req := TransactionDetailRequest{
		AgentId:          agentid,
		LoginId:          loginId,
		GameProviderCode: gameProviderCode,
		FromDate:         fromDate.In(loc).Format(providerSummaryTimeLayout),
		ToDate:           toDate.In(loc).Format(providerSummaryTimeLayout),
		TimeZone:         8,
		PageIndex:        pageIndex,
		RowPerPage:       rowPerPage,
	}

	var resp TransactionDetailResponse
	res, err := c.client.R().
		SetBody(req).
		SetResult(&resp).
		Post(buildProviderReportURL(agentapi, "/Report/TransactionDetail"))
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

	return &resp, nil
}
