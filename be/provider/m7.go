package provider

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const defaultM7BaseURL = "https://api.ddhh.work/"

type M7Config struct {
	BaseURL   string
	ClientID  string
	ClientKey string
	AgentID   string
	Currency  string
	ReturnURL string
}

type M7Client struct {
	client *resty.Client
	cfg    M7Config
}

type M7Error struct {
	Op      string
	Code    string
	Message string
}

func (e M7Error) Error() string {
	return fmt.Sprintf("m7 %s failed: code=%s message=%s", e.Op, e.Code, e.Message)
}

func newM7Error(op, code, message string) error {
	return M7Error{
		Op:      op,
		Code:    strings.TrimSpace(code),
		Message: strings.TrimSpace(message),
	}
}

type M7GameItem struct {
	ID             int64  `json:"id"`
	ThirdPartyType string `json:"thirdPartyType"`
	ThirdParty     string `json:"thirdParty"`
	GameNameEn     string `json:"gameNameEn"`
	GameNameCh     string `json:"gameNameCh"`
	GameNameTw     string `json:"gameNameTw"`
	GameNo         string `json:"gameNo"`
	Image1         string `json:"image1"`
	Image2         string `json:"image2"`
	Status         string `json:"status"`
	ProvideID      string `json:"provideId"`
	ProvideName    string `json:"provideName"`
	IsDemo         string `json:"isDemo"`
}

type m7BaseResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Msg     string `json:"msg"`
}

type m7GameListResponse struct {
	m7BaseResponse
	Result struct {
		Data []M7GameItem `json:"data"`
	} `json:"result"`
}

type m7EntryGameResponse struct {
	m7BaseResponse
	Result struct {
		URL string `json:"url"`
	} `json:"result"`
}

type m7WalletData struct {
	UserID   string  `json:"userId"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
	Amount   float64 `json:"amount"`
	TxID     string  `json:"txId"`
}

type m7WalletResult struct {
	Data     m7WalletData `json:"data"`
	UserID   string       `json:"userId"`
	Currency string       `json:"currency"`
	Balance  float64      `json:"balance"`
	Amount   float64      `json:"amount"`
	TxID     string       `json:"txId"`
}

type m7WalletResponse struct {
	m7BaseResponse
	Result m7WalletResult `json:"result"`
}

type M7GameHistoryResult struct {
	Data         any `json:"data"`
	List         any `json:"list"`
	Records      any `json:"records"`
	Total        int `json:"total"`
	TotalRecord  int `json:"totalRecord"`
	TotalRecords int `json:"totalRecords"`
	TotalPage    int `json:"totalPage"`
	TotalPages   int `json:"totalPages"`
	PageTotal    int `json:"pageTotal"`
}

type M7GameHistoryResponse struct {
	m7BaseResponse
	Result M7GameHistoryResult `json:"result"`
}

func NewM7Client(cfg M7Config) (*M7Client, error) {
	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultM7BaseURL
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/") + "/"
	cfg.Currency = strings.ToUpper(strings.TrimSpace(cfg.Currency))
	if cfg.Currency == "" {
		cfg.Currency = "USD"
	}
	cfg.ReturnURL = strings.TrimSpace(cfg.ReturnURL)
	if cfg.ReturnURL == "" {
		cfg.ReturnURL = "https://www.yourdomain.com"
	}
	if strings.TrimSpace(cfg.ClientID) == "" || strings.TrimSpace(cfg.ClientKey) == "" || strings.Contains(cfg.ClientKey, "PLACEHOLDER") {
		return nil, fmt.Errorf("m7 client_id/client_key is not configured")
	}

	return &M7Client{
		cfg: cfg,
		client: resty.New().
			SetTimeout(35*time.Second).
			SetHeader("Content-Type", "application/json"),
	}, nil
}

func (c *M7Client) FetchGameList(thirdPartyType string) ([]M7GameItem, error) {
	payload := map[string]string{
		"clientId":  c.cfg.ClientID,
		"clientKey": c.cfg.ClientKey,
	}
	if strings.TrimSpace(thirdPartyType) != "" {
		payload["thirdPartyType"] = strings.TrimSpace(thirdPartyType)
	}

	var resp m7GameListResponse
	endpoint := c.cfg.BaseURL + "busway/external/api/getGameList"
	res, err := c.client.R().
		SetBody(payload).
		SetResult(&resp).
		Post(endpoint)
	if err != nil {
		return nil, fmt.Errorf("m7 getGameList network error: %v", err)
	}
	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("m7 getGameList http status %d", res.StatusCode())
	}
	if resp.Code != "0000" {
		return nil, newM7Error("getGameList", resp.Code, m7Message(resp.m7BaseResponse))
	}

	return resp.Result.Data, nil
}

func (c *M7Client) CreatePlayer(userID string) error {
	if strings.TrimSpace(c.cfg.AgentID) == "" {
		return fmt.Errorf("m7 agent_id is not configured")
	}
	payload := map[string]string{
		"clientId":  c.cfg.ClientID,
		"clientKey": c.cfg.ClientKey,
		"agentId":   c.cfg.AgentID,
		"userId":    userID,
		"currency":  c.cfg.Currency,
	}

	var resp m7BaseResponse
	endpoint := c.cfg.BaseURL + "busway/external/api/createPlayer"
	res, err := c.client.R().
		SetBody(payload).
		SetResult(&resp).
		Post(endpoint)
	if err != nil {
		return fmt.Errorf("m7 createPlayer network error: %v", err)
	}
	if res.StatusCode() != 200 {
		return fmt.Errorf("m7 createPlayer http status %d", res.StatusCode())
	}
	if resp.Code != "0000" && !isM7PlayerAlreadyExistsMessage(m7Message(resp)) {
		return newM7Error("createPlayer", resp.Code, m7Message(resp))
	}

	return nil
}

func isM7PlayerAlreadyExistsMessage(message string) bool {
	normalized := strings.ToLower(strings.TrimSpace(message))
	if normalized == "" {
		return false
	}
	for _, token := range []string{"exist", "already", "duplicate", "\u5df2\u5b58\u5728", "\u5b58\u5728", "\u91cd\u590d"} {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func (c *M7Client) EntryGame(userID string, gameID string, isMobile bool, lang string) (string, error) {
	device := "2"
	if isMobile {
		device = "1"
	}
	payload := map[string]string{
		"lang":      normalizeM7Lang(lang),
		"gameId":    gameID,
		"clientId":  c.cfg.ClientID,
		"clientKey": c.cfg.ClientKey,
		"userId":    userID,
		"device":    device,
		"returnUrl": c.cfg.ReturnURL,
		"currency":  c.cfg.Currency,
	}

	var resp m7EntryGameResponse
	endpoint := c.cfg.BaseURL + "busway/external/api/entryGame"
	res, err := c.client.R().
		SetBody(payload).
		SetResult(&resp).
		Post(endpoint)
	if err != nil {
		return "", fmt.Errorf("m7 entryGame network error: %v", err)
	}
	if res.StatusCode() != 200 {
		return "", fmt.Errorf("m7 entryGame http status %d", res.StatusCode())
	}
	if resp.Code != "0000" {
		return "", newM7Error("entryGame", resp.Code, m7Message(resp.m7BaseResponse))
	}
	if strings.TrimSpace(resp.Result.URL) == "" {
		return "", fmt.Errorf("m7 entryGame returned empty url")
	}

	return resp.Result.URL, nil
}

func (c *M7Client) Balance(userID string) (float64, error) {
	payload := map[string]string{
		"clientId":  c.cfg.ClientID,
		"clientKey": c.cfg.ClientKey,
		"userId":    userID,
		"currency":  c.cfg.Currency,
	}

	var resp m7WalletResponse
	endpoint := c.cfg.BaseURL + "busway/external/wallet/balance"
	res, err := c.client.R().
		SetBody(payload).
		SetResult(&resp).
		Post(endpoint)
	if err != nil {
		return 0, fmt.Errorf("m7 balance network error: %v", err)
	}
	if res.StatusCode() != 200 {
		return 0, fmt.Errorf("m7 balance http status %d", res.StatusCode())
	}
	if resp.Code != "0000" {
		return 0, newM7Error("balance", resp.Code, m7Message(resp.m7BaseResponse))
	}

	return resp.Result.walletData().Balance, nil
}

func (c *M7Client) GetGameHistory(userID string, thirdPartyType string, startTime time.Time, endTime time.Time, pageIndex int, pageSize int) (*M7GameHistoryResponse, error) {
	if strings.TrimSpace(c.cfg.AgentID) == "" {
		return nil, fmt.Errorf("m7 agent_id is not configured")
	}
	if endTime.Before(startTime) {
		return nil, fmt.Errorf("invalid m7 game history range: endTime before startTime")
	}
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize <= 0 {
		pageSize = 1000
	}
	if pageSize > 5000 {
		pageSize = 5000
	}

	loc := providerSummaryLocation()
	payload := map[string]interface{}{
		"clientId":  c.cfg.ClientID,
		"clientKey": c.cfg.ClientKey,
		"agentId":   c.cfg.AgentID,
		"startTime": startTime.In(loc).Format("2006-01-02 15:04:05"),
		"endTime":   endTime.In(loc).Format("2006-01-02 15:04:05"),
		"pageIndex": pageIndex,
		"pageSize":  pageSize,
		"currency":  c.cfg.Currency,
	}
	if strings.TrimSpace(userID) != "" {
		payload["userId"] = strings.TrimSpace(userID)
	}
	if strings.TrimSpace(thirdPartyType) != "" {
		payload["thirdPartyType"] = strings.TrimSpace(thirdPartyType)
	}

	var resp M7GameHistoryResponse
	endpoint := c.cfg.BaseURL + "busway/external/api/getGameHistory"
	res, err := c.client.R().
		SetBody(payload).
		SetResult(&resp).
		Post(endpoint)
	if err != nil {
		return nil, fmt.Errorf("m7 getGameHistory network error: %v", err)
	}
	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("m7 getGameHistory http status %d", res.StatusCode())
	}
	if resp.Code != "0000" {
		return nil, newM7Error("getGameHistory", resp.Code, m7Message(resp.m7BaseResponse))
	}

	return &resp, nil
}

func (c *M7Client) Transfer(userID string, amount float64, txID string) (float64, error) {
	payload := map[string]interface{}{
		"clientId":  c.cfg.ClientID,
		"clientKey": c.cfg.ClientKey,
		"userId":    userID,
		"amount":    amount,
		"txId":      txID,
		"currency":  c.cfg.Currency,
	}

	var resp m7WalletResponse
	endpoint := c.cfg.BaseURL + "busway/external/wallet/transfer"
	res, err := c.client.R().
		SetBody(payload).
		SetResult(&resp).
		Post(endpoint)
	if err != nil {
		return 0, fmt.Errorf("m7 transfer network error: %v", err)
	}
	if res.StatusCode() != 200 {
		return 0, fmt.Errorf("m7 transfer http status %d", res.StatusCode())
	}
	if resp.Code != "0000" {
		return 0, newM7Error("transfer", resp.Code, m7Message(resp.m7BaseResponse))
	}

	return resp.Result.walletData().Balance, nil
}

func IsM7BalanceNotEnoughError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "balance not enough") ||
		strings.Contains(message, "余额不足") ||
		strings.Contains(message, "餘額不足")
}

func IsM7IPNotAllowedError(err error) bool {
	if err == nil {
		return false
	}

	var m7Err M7Error
	if errors.As(err, &m7Err) && strings.TrimSpace(m7Err.Code) == "9020" {
		return true
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "ip not allow") ||
		strings.Contains(message, "ip not allowed") ||
		strings.Contains(message, "ip não permitido") ||
		strings.Contains(message, "ip nao permitido") ||
		strings.Contains(message, "ip ไม่อนุญาต") ||
		strings.Contains(message, "区域不允许") ||
		strings.Contains(message, "區域不允許")
}

func IsM7PlayerNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	var m7Err M7Error
	if errors.As(err, &m7Err) && strings.TrimSpace(m7Err.Code) == "1001" {
		return true
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "user does not exist") ||
		strings.Contains(message, "user not exist") ||
		strings.Contains(message, "player does not exist") ||
		strings.Contains(message, "player not exist") ||
		strings.Contains(message, "用户不存在") ||
		strings.Contains(message, "使用者不存在") ||
		strings.Contains(message, "會員不存在")
}

func (c *M7Client) GetTransferHistory(txID string, startTime time.Time, endTime time.Time) (*M7GameHistoryResponse, error) {
	payload := map[string]interface{}{
		"clientId":  c.cfg.ClientID,
		"clientKey": c.cfg.ClientKey,
		"currency":  c.cfg.Currency,
	}
	if strings.TrimSpace(txID) != "" {
		payload["txId"] = strings.TrimSpace(txID)
	}
	loc := providerSummaryLocation()
	if !startTime.IsZero() {
		payload["startTime"] = startTime.In(loc).Format("2006-01-02 15:04:05")
	}
	if !endTime.IsZero() {
		payload["endTime"] = endTime.In(loc).Format("2006-01-02 15:04:05")
	}

	var resp M7GameHistoryResponse
	endpoint := c.cfg.BaseURL + "busway/external/wallet/transferHistory"
	res, err := c.client.R().
		SetBody(payload).
		SetResult(&resp).
		Post(endpoint)
	if err != nil {
		return nil, fmt.Errorf("m7 transferHistory network error: %v", err)
	}
	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("m7 transferHistory http status %d", res.StatusCode())
	}
	if resp.Code != "0000" {
		return nil, newM7Error("transferHistory", resp.Code, m7Message(resp.m7BaseResponse))
	}

	return &resp, nil
}

func (r M7GameHistoryResponse) Rows() []map[string]any {
	for _, value := range []any{r.Result.Data, r.Result.List, r.Result.Records} {
		if rows := extractM7HistoryRows(value); len(rows) > 0 {
			return rows
		}
	}
	return nil
}

func (r M7GameHistoryResponse) TotalPages() int {
	for _, value := range []int{r.Result.TotalPage, r.Result.TotalPages, r.Result.PageTotal} {
		if value > 0 {
			return value
		}
	}
	return 0
}

func (r M7GameHistoryResponse) TotalRecords() int {
	for _, value := range []int{r.Result.TotalRecord, r.Result.TotalRecords, r.Result.Total} {
		if value > 0 {
			return value
		}
	}
	return 0
}

func extractM7HistoryRows(value any) []map[string]any {
	switch v := value.(type) {
	case nil:
		return nil
	case []map[string]any:
		return v
	case []any:
		rows := make([]map[string]any, 0, len(v))
		for _, item := range v {
			if row, ok := item.(map[string]any); ok {
				rows = append(rows, row)
			}
		}
		return rows
	case map[string]any:
		for _, key := range []string{"data", "list", "records", "rows", "items"} {
			if rows := extractM7HistoryRows(v[key]); len(rows) > 0 {
				return rows
			}
		}
	}
	return nil
}

func (r m7WalletResult) walletData() m7WalletData {
	if r.Data.UserID != "" || r.Data.Currency != "" || r.Data.TxID != "" || r.Data.Balance != 0 || r.Data.Amount != 0 {
		return r.Data
	}

	return m7WalletData{
		UserID:   r.UserID,
		Currency: r.Currency,
		Balance:  r.Balance,
		Amount:   r.Amount,
		TxID:     r.TxID,
	}
}

func m7Message(resp m7BaseResponse) string {
	if strings.TrimSpace(resp.Message) != "" {
		return resp.Message
	}
	return resp.Msg
}

func normalizeM7Lang(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "cn", "zh", "zh-cn", "ch":
		return "zh"
	case "th", "thai":
		return "th"
	case "pt", "pt-br":
		return "pt"
	default:
		return "en"
	}
}
