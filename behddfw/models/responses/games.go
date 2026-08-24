package responses

import "gorm.io/datatypes"

// ExternalGameItem 第三方返回的单个游戏对象
type ExternalGameItem struct {
	Id               string            `json:"id"`
	GameProviderCode string            `json:"gameProviderCode"` // 返回的是字符串 Code (如 "PRG")
	Code             string            `json:"code"`             // 游戏代码
	Name             map[string]string `json:"name"`             // 多语言名称
	Type             int               `json:"type"`             // 对应 GameType 的 Sort
	GameIcon         string            `json:"gameIcon"`
	Order            int               `json:"order"`
}

// GameListResponse 响应外层
type GameListResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    []ExternalGameItem `json:"data"`
}

// OpenLobbyResponse 响应结构
type OpenLobbyResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Url     string `json:"url"` // 大厅跳转链接
}

// UpdateBalanceResponse 响应结构
type UpdateBalanceResponse struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Balance float64 `json:"balance"` // 交易后的最新余额
	// UserId string `json:"userId"` //
}

// GameUrlResponse 统一的URL响应结构
type GameUrlResponse struct {
	Code    int    `json:"Code"`
	Message string `json:"Message"`
	Url     string `json:"Url"` // 游戏跳转链接
}

// GameLiteResponse 精简版游戏信息 (用于大列表传输)
type GameLiteResponse struct {
	ID   uint64         `json:"id"`
	Code string         `json:"c"` // 缩写 key: code -> c
	Name datatypes.JSON `json:"n"` // 缩写 key: name -> n
	Img  string         `json:"i"` // 缩写 key: img_url -> i
	// 如果前端需要厂商ID进行本地筛选，可以加上
	Pid uint `json:"p"` // 缩写 key: provider_id -> p
}
