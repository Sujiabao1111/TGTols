package helpers

type Config struct {
	LocalURL      string              `mapstructure:"localURL"`
	Port          int                 `mapstructure:"port"`
	DbDsn         string              `mapstructure:"dbDsn"`
	PoolIdle      int                 `mapstructure:"poolIdle"`
	PoolMax       int                 `mapstructure:"poolMax"`
	Jwt           string              `mapstructure:"jwt"`
	RedisNode     string              `mapstructure:"redisNode"`
	RedisDb       int                 `mapstructure:"redisDb"`
	Agentid       string              `mapstructure:"agentid"`
	Agentapi      string              `mapstructure:"agentapi"`
	Prefix        string              `mapstructure:"prefix"`
	InternalToken string              `mapstructure:"internal_token"`
	Payment       PaymentConfig       `mapstructure:"payment"`
	GamePlatforms GamePlatformsConfig `mapstructure:"game_platforms"`
	VocherURL     string              `mapstructure:"vocher"`
	FacebookPixel FacebookPixelConfig `mapstructure:"facebook_pixel"`
}

type GamePlatformsConfig struct {
	M7   M7GamePlatformConfig   `mapstructure:"m7"`
	M7PP M7PPGamePlatformConfig `mapstructure:"m7pp"`
}

type M7PPGamePlatformConfig struct {
	BaseURL    string `mapstructure:"base_url"`
	APIKey     string `mapstructure:"api_key"`
	APISecret  string `mapstructure:"api_secret"`
	VendorCode string `mapstructure:"vendor_code"`
	Currency   string `mapstructure:"currency"`
	ReturnURL  string `mapstructure:"return_url"`
}

type M7GamePlatformConfig struct {
	BaseURL   string `mapstructure:"base_url"`
	ClientID  string `mapstructure:"client_id"`
	ClientKey string `mapstructure:"client_key"`
	AgentID   string `mapstructure:"agent_id"`
	Currency  string `mapstructure:"currency"`
	ReturnURL string `mapstructure:"return_url"`
}

type FacebookPixelTargetConfig struct {
	PixelID        string   `mapstructure:"pixel_id"`
	AccessToken    string   `mapstructure:"access_token"`
	TestEventCode  string   `mapstructure:"test_event_code"`
	APIVersion     string   `mapstructure:"api_version"`
	InviterUserIDs []uint64 `mapstructure:"inviter_user_ids"`
}

type FacebookPixelConfig struct {
	PixelID          string                      `mapstructure:"pixel_id"`
	AccessToken      string                      `mapstructure:"access_token"`
	TestEventCode    string                      `mapstructure:"test_event_code"`
	APIVersion       string                      `mapstructure:"api_version"`
	InviterUserIDs   []uint64                    `mapstructure:"inviter_user_ids"`
	AdditionalPixels []FacebookPixelTargetConfig `mapstructure:"additional_pixels"`
}

type PaymentConfig struct {
	BaseURL     string            `mapstructure:"base_url"`
	MerchantID  string            `mapstructure:"merchant_id"`
	SecretKey   string            `mapstructure:"secret_key"`
	NotifyURL   string            `mapstructure:"notify_url"`
	ReturnURL   string            `mapstructure:"return_url"`
	QuantixCore QuantixCoreConfig `mapstructure:"quantixcore"`
	TokenPay    TokenPayConfig    `mapstructure:"tokenpay"`
	Telegram    TelegramConfig    `mapstructure:"telegram"`
	TON         TONConfig         `mapstructure:"ton"`
}

type TelegramConfig struct {
	BotToken      string `mapstructure:"bot_token"`
	WebhookSecret string `mapstructure:"webhook_secret"`
	StarsPerUSD   int64  `mapstructure:"stars_per_usd"`
	Enabled       bool   `mapstructure:"enabled"`
}

type TONConfig struct {
	WalletAddress string `mapstructure:"wallet_address"`
	// Mnemonic is the hot-wallet recovery phrase used to sign payouts.
	// Keep it out of source control and production logs.
	Mnemonic                  string  `mapstructure:"mnemonic"`
	Network                   string  `mapstructure:"network"`
	Enabled                   bool    `mapstructure:"enabled"`
	WithdrawConditionsEnabled bool    `mapstructure:"withdraw_conditions_enabled"`
	MinWithdrawAmountUSD      float64 `mapstructure:"min_withdraw_amount_usd"`
	TestMode                  bool    `mapstructure:"test_mode"`
	TestUSDPerTON             float64 `mapstructure:"test_usd_per_ton"`
	USDPerTON                 float64 `mapstructure:"usd_per_ton"`
	RateURL                   string  `mapstructure:"rate_url"`
	RateTTL                   int     `mapstructure:"rate_ttl_seconds"`
	APIURL                    string  `mapstructure:"api_url"`
	APIKey                    string  `mapstructure:"api_key"`
	GlobalConfigURL           string  `mapstructure:"global_config_url"`
	WalletVersion             string  `mapstructure:"wallet_version"`
}

type QuantixCoreConfig struct {
	BaseURL string                  `mapstructure:"base_url"`
	ID      QuantixCoreRegionConfig `mapstructure:"id"`
	PH      QuantixCoreRegionConfig `mapstructure:"ph"`
}

type QuantixCoreRegionConfig struct {
	BaseURL    string `mapstructure:"base_url"`
	MerchantID string `mapstructure:"merchant_id"`
	SecretKey  string `mapstructure:"secret_key"`
	NotifyURL  string `mapstructure:"notify_url"`
	ReturnURL  string `mapstructure:"return_url"`
}

type TokenPayConfig struct {
	BaseURL   string `mapstructure:"base_url"`
	SecretKey string `mapstructure:"secret_key"`
	NotifyURL string `mapstructure:"notify_url"`
	ReturnURL string `mapstructure:"return_url"`
}

type DbCfgMem struct {
	DbCfg struct{}
	Raw   string
}
