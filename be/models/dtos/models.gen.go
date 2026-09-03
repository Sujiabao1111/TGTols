package dtos

import (
	"time"

	"gorm.io/datatypes"
)

// User 用户表
type User struct {
	ID       uint64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string  `gorm:"size:32;not null;uniqueIndex:idx_username" json:"username"`
	Password string  `gorm:"size:255;not null" json:"-"` // JSON 序列化时隐藏密码
	Balance  float64 `gorm:"type:decimal(15,2);default:0.00" json:"balance"`
	VipLevel int     `gorm:"default:0" json:"vip_level"`

	// 分销核心字段
	ParentID   uint64 `gorm:"index:idx_parent_id;default:0" json:"parent_id"`
	Path       string `gorm:"size:255;index:idx_path;default:''" json:"path"` // 物化路径: "0/1/5/"
	Level      int    `gorm:"default:1" json:"level"`
	InviteCode string `gorm:"size:10;uniqueIndex:idx_invite_code" json:"invite_code"`

	// 资金统计
	TotalDeposit  float64 `gorm:"type:decimal(15,2);default:0.00" json:"total_deposit"`  // 累计充值
	TotalWithdraw float64 `gorm:"type:decimal(15,2);default:0.00" json:"total_withdraw"` // 累计提现

	Status            int        `gorm:"default:1;comment:1:正常 0:禁用" json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	RegisterIP        string     `gorm:"size:64;default:''" json:"register_ip"`
	RegisterDomain    string     `gorm:"size:255;default:'';index:idx_register_domain" json:"register_domain"`
	TelegramUserID    string     `gorm:"size:32;default:null;uniqueIndex:idx_user_telegram_id" json:"telegram_user_id,omitempty"`
	TelegramUsername  string     `gorm:"size:64;default:''" json:"telegram_username,omitempty"`
	TelegramFirstName string     `gorm:"size:128;default:''" json:"telegram_first_name,omitempty"`
	TelegramLastName  string     `gorm:"size:128;default:''" json:"telegram_last_name,omitempty"`
	TelegramPhotoURL  string     `gorm:"size:512;default:''" json:"telegram_photo_url,omitempty"`
	TelegramBoundAt   *time.Time `json:"telegram_bound_at,omitempty"`
}

// VipConfig VIP等级配置
type VipConfig struct {
	Level      int       `gorm:"primaryKey" json:"level"`
	Title      string    `gorm:"size:50;not null" json:"title"`
	MinDeposit float64   `gorm:"type:decimal(15,2);default:0.00" json:"min_deposit"`
	RebateRate float64   `gorm:"type:decimal(5,4);default:0.0000" json:"rebate_rate"`
	CreatedAt  time.Time `json:"created_at"`
}

// GameProvider 游戏供应商
type GameProvider struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	PlatformCode string         `gorm:"size:32;not null;default:HEDOC;uniqueIndex:idx_provider_platform_code" json:"platform_code"`
	Code         string         `gorm:"size:50;not null;uniqueIndex:idx_provider_platform_code" json:"code"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	ApiConfig    datatypes.JSON `gorm:"type:json" json:"-"` // 敏感配置不直接返回给前端
	Status       int            `gorm:"default:1;comment:1:开启 0:维护" json:"status"`
	TypeId       int            `gorm:"default:1;comment:类型Id" json:"type_id"`
	CreatedAt    time.Time      `json:"created_at"`
}

// GameType 游戏分类
type GameType struct {
	ID     uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name   string `gorm:"size:50;not null" json:"name"`
	Sort   int    `gorm:"default:0" json:"sort"`
	Code   string `gorm:"size:32" json:"code"`
	Status int    `gorm:"default:1;comment:1:启用 0:禁用" json:"status"`
}

// Game 游戏库
type Game struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	// PlatformCode 区分原游戏平台和新增游戏平台，避免同厂商/同 game_code 串数据。
	PlatformCode string `gorm:"size:32;not null;default:HEDOC;uniqueIndex:idx_platform_provider_game_code;index:idx_platform_provider_type" json:"platform_code"`
	ProviderID   uint   `gorm:"not null;uniqueIndex:idx_platform_provider_game_code;index:idx_platform_provider_type" json:"provider_id"`
	GameTypeID   uint   `gorm:"index:idx_platform_provider_type" json:"game_type_id"`
	GameCode     string `gorm:"size:100;not null;uniqueIndex:idx_platform_provider_game_code" json:"game_code"`

	// 修改: Name 改为 JSON 存储多语言 {"CN": "...", "EN": "..."}
	Name datatypes.JSON `gorm:"type:json" json:"name"`

	ImgUrl    string    `gorm:"size:255;default:''" json:"img_url"`
	Views     int64     `gorm:"default:0;index:idx_sort_views" json:"views"`
	Sort      int       `gorm:"default:0;index:idx_sort_views" json:"sort"`
	Status    int       `gorm:"default:1" json:"status"`
	CreatedAt time.Time `json:"created_at"`

	// 关联
	Provider *GameProvider `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
	GameType *GameType     `gorm:"foreignKey:GameTypeID" json:"game_type,omitempty"`
}

// Banner 轮播图模型
type Banner struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Title    string `gorm:"size:255;not null" json:"title"`
	Subtitle string `gorm:"size:255" json:"subtitle"`
	Tag      string `gorm:"size:100" json:"tag"`
	Image    string `gorm:"size:500;not null" json:"image"`
	Color    string `gorm:"size:255" json:"color"`

	// 差异化字段处理
	// 使用 omitempty: 如果字符串为空，不返回该字段
	JumpLink string `gorm:"size:255" json:"jumpLink,omitempty"`

	// 使用 datatypes.JSON: GORM 会自动处理 JSON 的序列化和反序列化
	// 如果数据库存的是 NULL，这里通过指针或 GORM 逻辑处理，通常如果是 nil 也会被 omitempty 忽略
	Buttons datatypes.JSON `gorm:"type:json" json:"buttons,omitempty"`

	Sort      int       `gorm:"default:0" json:"-"` // 前端不需要看到排序值，隐藏
	Status    int       `gorm:"default:1" json:"-"` // 前端不需要看到状态，隐藏
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// 辅助结构体：用于明确 Buttons 里的结构 (虽然存取时用 datatypes.JSON 已经够了，但定义出来方便文档查看)
type BannerButton struct {
	Label   string `json:"label"`
	Action  string `json:"action"`
	Primary bool   `json:"primary"`
}

// Activity 运营活动
type Activity struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Type        string         `gorm:"size:50;not null;index:idx_type" json:"type"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description string         `gorm:"size:255" json:"description"`
	Config      datatypes.JSON `gorm:"type:json" json:"config"` // 活动具体规则
	StartTime   time.Time      `json:"start_time"`
	EndTime     time.Time      `json:"end_time"`
	Status      int            `gorm:"default:1;comment:1:启用 0:禁用" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// TableName 指定表名
func (Activity) TableName() string {
	return "activities"
}

// UserSignIn 签到记录
type UserSignIn struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"not null;index:idx_user_date,unique" json:"user_id"`
	SignDate  time.Time `gorm:"type:date;not null;index:idx_user_date,unique" json:"sign_date"`
	Reward    float64   `gorm:"type:decimal(10,2);default:0.00" json:"reward"`
	CreatedAt time.Time `json:"created_at"`
}

// Transaction 资金流水
type Transaction struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint64    `gorm:"not null;index:idx_user_created" json:"user_id"`
	Type          int       `gorm:"not null;comment:1:充值 2:提现 3:下注 4:派彩 5:返点 6:活动" json:"type"`
	Amount        float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	BeforeBalance float64   `gorm:"type:decimal(15,2);not null" json:"before_balance"`
	AfterBalance  float64   `gorm:"type:decimal(15,2);not null" json:"after_balance"`
	ReferenceID   string    `gorm:"size:100;default:''" json:"reference_id"`
	Remark        string    `gorm:"size:255;default:''" json:"remark"`
	CreatedAt     time.Time `gorm:"index:idx_user_created" json:"created_at"`
}

// DailyUserStats 用户每日报表 (OLAP聚合)
type DailyUserStats struct {
	ID       uint64 `gorm:"primaryKey" json:"id"`
	UserID   uint64 `gorm:"uniqueIndex:idx_user_date" json:"user_id"`
	StatDate string `gorm:"type:date;uniqueIndex:idx_user_date" json:"stat_date"`

	BetAmount      float64 `gorm:"type:decimal(15,2);default:0" json:"bet_amount"`
	WinAmount      float64 `gorm:"type:decimal(15,2);default:0" json:"win_amount"`
	DepositAmount  float64 `gorm:"type:decimal(15,2);default:0" json:"deposit_amount"`
	WithdrawAmount float64 `gorm:"type:decimal(15,2);default:0" json:"withdraw_amount"`

	LoginCount int `gorm:"default:0" json:"login_count"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联用户 (方便返佣时查询用户的 Path 和 ParentID)
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// 指定表名（可选，如果不想让 GORM 自动转复数）
func (User) TableName() string {
	return "users"
}

// ... 可以为其他结构体添加 TableName 方法，虽然 GORM 默认已经会处理为 snake_case 复数

// StatsRetention 留存统计
type StatsRetention struct {
	ID               uint64    `gorm:"primaryKey" json:"id"`
	StatDate         string    `gorm:"type:date;uniqueIndex" json:"stat_date"` // 注册日期 YYYY-MM-DD
	NewUsers         int       `json:"new_users"`
	DepositUsers     int       `gorm:"column:deposit_users" json:"deposit_users"`
	DepositAmount    float64   `gorm:"type:decimal(15,2);default:0" json:"deposit_amount"`
	GameWinloseUsers int       `gorm:"column:game_winlose_users" json:"game_winlose_users"`
	Retention1       int       `gorm:"column:retention_1" json:"retention_1"`
	Retention3       int       `gorm:"column:retention_3" json:"retention_3"`
	Retention7       int       `gorm:"column:retention_7" json:"retention_7"`
	Retention30      int       `gorm:"column:retention_30" json:"retention_30"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (StatsRetention) TableName() string {
	return "stats_retention"
}

// AgentRebateConfig 返点配置
type AgentRebateConfig struct {
	Level int     `gorm:"primaryKey" json:"level"` // 1-9
	Rate  float64 `gorm:"type:decimal(5,4)" json:"rate"`
}

// UserFavorite 用户收藏关联
type UserFavorite struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"not null;uniqueIndex:idx_user_game" json:"user_id"`
	GameID    uint64    `gorm:"not null;uniqueIndex:idx_user_game" json:"game_id"`
	CreatedAt time.Time `json:"created_at"`

	// 关联方便查询
	Game *Game `gorm:"foreignKey:GameID" json:"game,omitempty"`
}
