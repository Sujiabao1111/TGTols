package config

type Custom struct {
	Commonsrv      string        `mapstructure:"commonsrv" json:"commonsrv" yaml:"commonsrv"`
	CommonsrvToken string        `mapstructure:"commonsrv_token" json:"commonsrv_token" yaml:"commonsrv_token"`
	GeoIPURL       string        `mapstructure:"geoip_url" json:"geoip_url" yaml:"geoip_url"`
	GeoIPTimeoutMs int           `mapstructure:"geoip_timeout_ms" json:"geoip_timeout_ms" yaml:"geoip_timeout_ms"`
	Payment        PaymentConfig `mapstructure:"payment" json:"payment" yaml:"payment"`
}

type PaymentConfig struct {
	BaseURL     string            `mapstructure:"base_url" json:"base_url" yaml:"base_url"`
	MerchantID  string            `mapstructure:"merchant_id" json:"merchant_id" yaml:"merchant_id"`
	SecretKey   string            `mapstructure:"secret_key" json:"secret_key" yaml:"secret_key"`
	NotifyURL   string            `mapstructure:"notify_url" json:"notify_url" yaml:"notify_url"`
	QuantixCore QuantixCoreConfig `mapstructure:"quantixcore" json:"quantixcore" yaml:"quantixcore"`
}

type QuantixCoreConfig struct {
	BaseURL string                  `mapstructure:"base_url" json:"base_url" yaml:"base_url"`
	ID      QuantixCoreRegionConfig `mapstructure:"id" json:"id" yaml:"id"`
	PH      QuantixCoreRegionConfig `mapstructure:"ph" json:"ph" yaml:"ph"`
}

type QuantixCoreRegionConfig struct {
	BaseURL    string `mapstructure:"base_url" json:"base_url" yaml:"base_url"`
	MerchantID string `mapstructure:"merchant_id" json:"merchant_id" yaml:"merchant_id"`
	SecretKey  string `mapstructure:"secret_key" json:"secret_key" yaml:"secret_key"`
	NotifyURL  string `mapstructure:"notify_url" json:"notify_url" yaml:"notify_url"`
}
