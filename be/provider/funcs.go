package provider

import (
	"encoding/json"
	"fmt"
	"gogogo/models/requests"
	"gogogo/models/responses"
	"log"
)

const debugGameListProviderID = 23

func (c *HedoClient) GetGameList() ([]GameData, error) {
	var resp APIResponse
	_, err := c.client.R().SetResult(&resp).Get("/games")
	if err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf("provider error: %s", resp.Msg)
	}

	var games []GameData
	json.Unmarshal(resp.Data, &games)
	return games, nil
}

func (c *HedoClient) LaunchGame(userid int, gameCode, apibase string) (string, error) {
	// 1. 请求载荷
	payload := map[string]any{
		"userid":    userid,
		"game_code": gameCode,
		"language":  "zh_CN",
	}

	var resp APIResponse
	// 假设接口是 /launch
	endpoint := apibase + "Game/OpenGameByCode"
	_, err := c.client.R().SetBody(payload).SetResult(&resp).Post(endpoint)
	if err != nil {
		return "", err
	}
	if resp.Code != 200 {
		return "", fmt.Errorf("launch failed: %s", resp.Msg)
	}

	// 解析返回的 URL
	var result struct {
		Url string `json:"url"`
	}
	json.Unmarshal(resp.Data, &result)
	return result.Url, nil
}

// update balance
func (c *HedoClient) Transfer(userid int, amount float64, transferType int, refId string) error {
	// transferType: 1=Deposit(IN), 2=Withdraw(OUT)
	// amount > 0 in, amount < 0 out
	payload := map[string]interface{}{
		"userid": userid,
		"amount": amount,
		"type":   transferType,
		"ref_id": refId, // 传递我们系统的订单号，用于对账
	}

	var resp APIResponse
	_, err := c.client.R().SetBody(payload).SetResult(&resp).Post("/transfer")
	if err != nil {
		// 网络错误，状态未知，需人工介入或自动补单
		return fmt.Errorf("network_error")
	}
	if resp.Code != 200 {
		return fmt.Errorf("%s", resp.Msg)
	}
	return nil
}

func (c *HedoClient) ChangePassword(userid int, newPassword, apibase string) error {
	payload := map[string]any{
		"userid":       userid,
		"new_password": newPassword,
	}
	var resp APIResponse
	_, err := c.client.R().SetBody(payload).SetResult(&resp).Post(apibase + "User/EditPassword")
	if err != nil || resp.Code != 200 {
		return fmt.Errorf("failed to change password")
	}
	return nil
}

type RegisterPayload struct {
	AgentId  string `json:"AgentId"`
	LoginId  string `json:"LoginId"`
	Password string `json:"Password"`
}

// RegisterResponse 注册接口响应结构
type RegisterResponse struct {
	Code    int    `json:"Code"`
	Message string `json:"Message"`
	UserId  string `json:"UserId"` // 第三方返回的用户ID
}

// RegisterUser 在第三方平台注册用户
func (c *HedoClient) RegisterUser(agentid, loginId, password, apibase string) error {
	// 1. 构造请求参数
	payload := RegisterPayload{
		AgentId:  agentid,
		LoginId:  loginId,
		Password: password,
	}

	// 2. 定义响应接收结构 (假设第三方标准响应)
	var resp RegisterResponse

	// 3. 发送 POST 请求
	// 注册接口路径为 User/Register
	res, err := c.client.R().
		SetBody(payload).
		SetResult(&resp).
		Post(apibase + "User/Register")

	// 4. 网络错误处理
	if err != nil {
		return fmt.Errorf("provider network error: %v", err)
	}
	if res.StatusCode() != 200 {
		return fmt.Errorf("http status error: %d", res.StatusCode())
	}

	// 5. 业务层面逻辑处理
	// Case A: 注册成功 (假设 1 是成功代码，根据之前的 API 上下文)
	if resp.Code == 1 {
		return nil
	}

	// Case B: 用户已存在 (Code 15)
	// 需求：这种情况视为“成功”，允许后续流程继续
	if resp.Code == 15 {
		// 可以选择在这里记录一条 debug 日志
		// fmt.Printf("User %s already exists (UserId: %s), proceeding...\n", loginId, resp.UserId)
		return nil
	}

	// Case C: 其他真正的错误 (如参数错误、IP限制等)
	return fmt.Errorf("provider register failed: [%d] %s", resp.Code, resp.Message)
}

// GetBalanceRequest 查询余额请求参数
type GetBalanceRequest struct {
	AgentId string `json:"AgentId"`
	LoginId string `json:"LoginId"`
}

// GetBalanceResponse 查询余额响应结构
type GetBalanceResponse struct {
	Code    int     `json:"Code"`
	Message string  `json:"Message"`
	UserId  string  `json:"UserId"`  // 第三方返回的用户唯一标识
	Balance float64 `json:"Balance"` // 余额
}

// GetBalance 仅执行查询余额的原始调用
func (c *HedoClient) GetBalance(agentid, loginId, agentapi string) (*GetBalanceResponse, error) {
	req := GetBalanceRequest{
		AgentId: agentid, // 从常量或配置中获取
		LoginId: loginId,
	}

	var resp GetBalanceResponse

	// 发起 POST 请求
	res, err := c.client.R().
		SetBody(req).
		SetResult(&resp).
		Post(agentapi + "User/GetBalance")

	if err != nil {
		return nil, fmt.Errorf("network error: %v", err)
	}

	// 即使 HTTP 200，业务 Code 也可能不为 0 (视第三方文档而定)
	// 但根据描述，我们需要通过 UserId 字段判断逻辑，所以这里先返回 resp
	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("http status %d", res.StatusCode())
	}

	return &resp, nil
}

// GetBalanceAndRegister 封装业务逻辑：查询余额 -> 如果不存在 -> 注册 -> 返回余额(0)
func (c *HedoClient) GetBalanceAndRegister(MerchantCode, loginId, password, apibase string) (float64, error) {
	// 1. 尝试获取余额
	resp, err := c.GetBalance(MerchantCode, loginId, apibase)
	if err != nil {
		return 0, err
	}

	// 2. 核心判断逻辑：如果 UserId 为空字符串，说明用户未注册
	// 注意：这里假设 Code=0 时也可能 UserId 为空，或者第三方明确返回特定 Code
	if resp.UserId == "" {
		fmt.Printf("[Hedo] User %s not found (UserId is empty), registering...\n", loginId)

		// 3. 调用之前写好的注册方法
		err := c.RegisterUser(MerchantCode, loginId, password, apibase)
		if err != nil {
			return 0, fmt.Errorf("auto register failed: %v", err)
		}

		// 注册成功后，余额肯定是 0
		return 0.0, nil
	}

	// 4. 用户存在，直接返回余额
	return resp.Balance, nil
}

// FetchGameList 调用第三方接口获取列表
func (c *HedoClient) FetchGameList(providerId int, agentid string, apibase string) ([]responses.ExternalGameItem, error) {
	// 构造请求，使用固定的 AgentId (或者从配置读取)
	req := requests.GameList3rdPartyRequest{
		AgentId:          agentid,
		LoginId:          "",
		GameProviderCode: providerId,
		GameCode:         "",
		IsMobile:         nil,
	}

	if providerId == debugGameListProviderID {
		log.Printf("[FetchGameList][Provider %d] request payload: AgentId=%s LoginId=%q GameProviderCode=%d GameCode=%q IsMobile=%v endpoint=%s\n",
			providerId, req.AgentId, req.LoginId, req.GameProviderCode, req.GameCode, req.IsMobile, apibase+"Game/GetList")
	}

	var resp responses.GameListResponse
	// 假设接口路径为 /Game/GetList
	res, err := c.client.R().
		SetBody(req).
		SetResult(&resp).
		Post(apibase + "Game/GetList")

	if err != nil {
		if providerId == debugGameListProviderID {
			log.Printf("[FetchGameList][Provider %d] request failed: %v\n", providerId, err)
		}
		return nil, fmt.Errorf("request failed: %v", err)
	}

	if providerId == debugGameListProviderID {
		log.Printf("[FetchGameList][Provider %d] response: http_status=%d code=%d message=%q data_len=%d\n",
			providerId, res.StatusCode(), resp.Code, resp.Message, len(resp.Data))
		if len(resp.Data) > 0 {
			first := resp.Data[0]
			log.Printf("[FetchGameList][Provider %d] first game sample: provider_code=%q game_code=%q type=%d icon=%q\n",
				providerId, first.GameProviderCode, first.Code, first.Type, first.GameIcon)
		} else {
			log.Printf("[FetchGameList][Provider %d] third-party returned empty data array\n", providerId)
		}
	}

	if resp.Code != 1 { // 假设 1 是成功 (根据提供的JSON Code:1)
		return nil, fmt.Errorf("provider api error: %s", resp.Message)
	}

	return resp.Data, nil
}

// OpenLobby 调用打开大厅接口
func (c *HedoClient) OpenLobby(agentid, loginId, password string, isMobile bool, lang string, providerId int, apibase string) (string, error) {
	// 构造请求
	req := requests.OpenLobbyRequest{
		AgentId:          agentid, // 你的商户ID
		LoginId:          loginId,
		Password:         password,
		IsMobile:         isMobile,
		Language:         0,
		GameProviderCode: providerId,
	}

	// 默认语言处理
	if lang == "CN" {
		req.Language = 1
	}

	var resp responses.OpenLobbyResponse

	// 假设接口路径为 /Game/OpenLobby (根据常规 Nexus API 推断)
	// 如果实际文档路径不同，请修改此处字符串，例如 "/Lobby/GetUrl" 或 "/Player/OpenLobby"
	endpoint := apibase + "Game/OpenGame"
	res, err := c.client.R().
		SetBody(req).
		SetResult(&resp).
		Post(endpoint)

	if err != nil {
		return "", fmt.Errorf("network error: %v", err)
	}

	// 检查 HTTP 状态码
	if res.StatusCode() != 200 {
		return "", fmt.Errorf("http status %d", res.StatusCode())
	}

	// 检查业务状态码 (假设 1 是成功，依据是你之前提供的 GetList 返回结构)
	if resp.Code != 1 {
		return "", fmt.Errorf("provider error: [%d] %s", resp.Code, resp.Message)
	}

	return resp.Url, nil
}

// UpdateBalance 调用第三方进行转账 (上下分)
// amount: 变动金额 (正数)
// transType: 1:转入(Deposit), 2:转出(Withdraw) - 具体根据文档调整
// billNo: 唯一订单号
func (c *HedoClient) UpdateBalance(agentid, loginId string, amount float64, transType int, billNo, apibase string) (float64, error) {
	// 构造请求
	// 注意：有些接口要求取款传负数，有些要求传正数配合 Type。
	// 这里假设：需要 Type 字段，且金额始终为正数。
	// 如果文档不需要 Type，仅靠金额正负判断，请去掉 Type 字段并将逻辑改为 check amount symbol。
	req := requests.UpdateBalanceTo3rdPartyRequest{
		AgentId:   agentid,
		LoginId:   loginId,
		Amount:    amount,
		Reference: billNo,
	}
	if transType == 2 {
		req.Amount = req.Amount * -1
	}

	var resp responses.UpdateBalanceResponse

	// 假设接口路径为 /User/UpdateBalance
	res, err := c.client.R().
		SetBody(req).
		SetResult(&resp).
		Post(apibase + "User/UpdateBalance")

	if err != nil {
		return 0, fmt.Errorf("network error: %v", err)
	}

	if res.StatusCode() != 200 {
		return 0, fmt.Errorf("http status %d", res.StatusCode())
	}

	// 假设 1 或 0 为成功，根据之前的 GetBalance 来看，Message="SUCCESS" 且 Code=1 可能是成功
	// 请根据实际文档调整成功 Code 判断
	if resp.Code != 1 {
		return 0, fmt.Errorf("provider error: [%d] %s", resp.Code, resp.Message)
	}

	return resp.Balance, nil
}

// OpenGame 调用打开游戏接口
func (c *HedoClient) OpenGame(loginId, password string, providerId int, gameCode string, isMobile bool, lang, apibase, agentid string) (string, error) {
	req := requests.OpenGameRequest{
		AgentId:          agentid,
		LoginId:          loginId,
		Password:         password,
		GameProviderCode: providerId,
		GameCode:         gameCode,
		IsMobile:         isMobile,
		Language:         0,
	}
	if lang == "CN" {
		req.Language = 1
	}

	var resp responses.GameUrlResponse
	// 接口路径为 /Game/OpenGame
	endpoint := apibase + "Game/OpenGameByCode"
	res, err := c.client.R().SetBody(req).SetResult(&resp).Post(endpoint)

	if err != nil {
		return "", fmt.Errorf("network error: %v", err)
	}
	if res.StatusCode() != 200 || resp.Code != 1 {
		return "", fmt.Errorf("provider error: [%d] %s", resp.Code, resp.Message)
	}

	return resp.Url, nil
}

// PlayerSummaryRequest 请求参数
type PlayerSummaryRequest struct {
	AgentId          string `json:"AgentId"`
	LoginId          string `json:"LoginId"`
	SummaryDate      int    `json:"SummaryDate"`      // 1: Daily
	GameProviderCode int    `json:"GameProviderCode"` // 0: All
	FromDate         string `json:"FromDate"`         // 格式: 2025-12-29T00:00:00+08:00
	ToDate           string `json:"ToDate"`
}

// PlayerSummaryData 返回的数据结构
type PlayerSummaryData struct {
	UserId   string  `json:"UserId"`
	LoginId  string  `json:"LoginId"`
	Count    int     `json:"Count"`    // 注单数
	Turnover float64 `json:"Turnover"` // 有效投注 (返佣依据)
	Bet      float64 `json:"Bet"`      // 总投注
	Win      float64 `json:"Win"`      // 总派彩
	Winlose  float64 `json:"Winlose"`  // 输赢
}

type PlayerSummaryResponse struct {
	Code    int               `json:"Code"`
	Message string            `json:"Message"`
	Data    PlayerSummaryData `json:"Data"`
}

// GetPlayerDailySummary 获取指定用户、指定日期的报表
// dateStr: 格式 "2006-01-02"
func (c *HedoClient) GetPlayerDailySummary(loginId string, dateStr string, agentid, agentapi string) (*PlayerSummaryData, error) {
	summary, err := c.GetPlayerDailySummaryV2(loginId, dateStr, agentid, agentapi)
	if err != nil {
		return nil, err
	}

	return &PlayerSummaryData{
		UserId:   summary.UserId,
		LoginId:  summary.LoginId,
		Count:    summary.Count,
		Turnover: summary.Turnover,
		Bet:      summary.Bet,
		Win:      summary.Win,
		Winlose:  summary.Winlose,
	}, nil
}
