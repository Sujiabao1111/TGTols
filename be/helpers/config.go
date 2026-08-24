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
