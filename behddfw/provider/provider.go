package provider

import (
	"encoding/json"
	"time"

	"github.com/go-resty/resty/v2" // 推荐使用 resty 发送请求
)

const (
	BaseURL      = "https://hedoc.com/api"
	MerchantCode = "YOUR_MERCHANT_CODE"
	ApiKey       = "YOUR_API_KEY"
)

type HedoClient struct {
	client *resty.Client
}

// GameData 第三方游戏结构
type GameData struct {
	GameCode     string `json:"game_code"`
	GameName     string `json:"game_name"`
	Type         string `json:"type"` // slot, live, etc.
	ImgUrl       string `json:"img_url"`
	GameProvider string `json:"provider"`
	GamePlayUrl  string `json:"play_url"`
}

func NewHedocClient() *HedoClient {
	return &HedoClient{
		client: resty.New().
			SetBaseURL(BaseURL).
			SetTimeout(10*time.Second).
			SetHeader("Content-Type", "application/json").
			SetHeader("Merchant-Code", MerchantCode), // 假设鉴权方式
	}
}

func NewEdwinClientWithParams(baseurl, merchantcode string) *HedoClient {
	return &HedoClient{
		client: resty.New().
			SetBaseURL(baseurl).
			SetTimeout(10*time.Second).
			SetHeader("Content-Type", "application/json").
			SetHeader("Merchant-Code", merchantcode), // 假设鉴权方式
	}
}

// 通用响应结构 (假设第三方返回格式)
type APIResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}
