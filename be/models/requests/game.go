package requests

type TransferRequest struct {
	Amount float64 `json:"amount"`
	Type   int     `json:"type"` // 1:转入游戏, 2:转出游戏
}

// DepositRequest 充值请求
type DepositRequest struct {
	Amount  float64 `json:"amount"`
	Channel string  `json:"channel"` // alipay, usdt, bank
}

// WithdrawRequest 提现申请
type WithdrawRequest struct {
	Amount   float64 `json:"amount"`
	BankInfo string  `json:"bank_info"`
}

// GameList3rdPartyRequest 从第三方获取游戏列表请求结构
type GameList3rdPartyRequest struct {
	AgentId          string      `json:"AgentId"`
	LoginId          string      `json:"LoginId"`
	GameProviderCode int         `json:"GameProviderCode"` // 对应数据库 ID (int)
	GameCode         string      `json:"GameCode"`
	IsMobile         interface{} `json:"IsMobile"` // null, true, false
}

// GameListClientRequest 客户端游戏列表请求参数
type GameListClientRequest struct {
	Page         int    `json:"page"`          // 页码
	PageSize     int    `json:"page_size"`     // 每页数量
	PlatformCode string `json:"platform_code"` // 游戏平台代码, 默认 HEDOC
	GameTypeCode string `json:"game_type_id"`  // 游戏类型代码 (如 "SLOT", "LIVE", "SPORTS", "ALL")
	ProviderCode string `json:"game_provider"` // 厂商代码 (如 "PG", "JILI", "ALL")
	GameName     string `json:"game_name"`     // 游戏名字
}

// UpdateBalanceRequest 上下分请求参数
type UpdateBalanceTo3rdPartyRequest struct {
	AgentId   string  `json:"AgentId"`
	LoginId   string  `json:"LoginId"`
	Amount    float64 `json:"Amount"`    // 正数通常为存入，负数可能为取出（具体看文档定义，有的接口分Type）
	Reference string  `json:"Reference"` // unique identifier for every request
}

// OpenLobbyRequest 打开大厅请求参数
type OpenLobbyRequest struct {
	AgentId          string      `json:"AgentId"`
	LoginId          string      `json:"LoginId"`
	GameProviderCode int         `json:"GameProviderCode"` // 厂商ID，用于打开特定厂商的大厅
	IsMobile         interface{} `json:"IsMobile"`         // true/false 或 null
	Language         int         `json:"Language"`         // e.g. "zh_CN"
	Password         string      `json:"Password"`         //
}

// OpenGameRequest 打开游戏请求参数
type OpenGameRequest struct {
	AgentId          string      `json:"AgentId"`
	LoginId          string      `json:"LoginId"`
	GameProviderCode int         `json:"GameProviderCode"`
	GameCode         string      `json:"GameCode"`
	IsMobile         interface{} `json:"IsMobile"`
	Language         int         `json:"Language"`
	Password         string      `json:"Password"`
}

// OpenLobbyRequest API请求参数
type OpenLobbyPayload struct {
	PlatformCode string `json:"platform_code"`
	ProviderCode string `json:"provider_code"` // 厂商代码 (如 PG)
	IsMobile     bool   `json:"is_mobile"`
	Language     string `json:"language"`
	LaunchSource string `json:"launch_source"`
}

// OpenGamePayload API请求参数
type OpenGamePayload struct {
	GameID       uint64 `json:"game_id"`
	PlatformCode string `json:"platform_code"`
	GameCode     string `json:"game_code"` // 游戏代码 (如 vs20olympgate)
	IsMobile     bool   `json:"is_mobile"`
	Language     string `json:"language"`
	LaunchSource string `json:"launch_source"`
}

// FavoriteRequest 收藏操作请求
type FavoriteRequest struct {
	GameID     uint64 `json:"game_id"`
	IsFavorite bool   `json:"is_favorite"` // true: 添加收藏, false: 取消收藏
}

// FavoriteListRequest 获取收藏列表请求参数
type FavoriteListRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// TolsGameListRequest 游戏列表请求参数
type TolsGameListRequest struct {
	Page         int    `json:"page"`
	PageSize     int    `json:"page_size"`
	PlatformCode string `json:"platform_code"`

	// 排序类型: "POPULAR" | "NEW" | "RECOMMEND"
	SortType string `json:"sort_type"`
}
