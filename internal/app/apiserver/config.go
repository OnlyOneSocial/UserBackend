package apiserver

// Selectel ...
type Selectel struct {
	User     string `toml:"user"`
	Password string `toml:"password"`
}

// KeyclockClient ...
type KeyclockClient struct {
	ClientID string `toml:"client_id"`
	SecretID string `toml:"secret_id"`
}

//Config ...
type Config struct {
	BindAddr               string         `toml:"bind_addr"`
	LogLevel               string         `toml:"log_level"`
	DatabaseURL            string         `toml:"database_url"`
	JwtSignKey             string         `toml:"jwtsignkey"`
	RecaptchaSecret        string         `toml:"recaptcha_secret"`
	RecaptchaSecretAndroid string         `toml:"recaptcha_secret_android"`
	Selectel               Selectel       `toml:"Selectel"`
	KeyclockClient         KeyclockClient `toml:"KeycloakClient"`
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		BindAddr:        ":8080",
		LogLevel:        "debug",
		JwtSignKey:      "",
		RecaptchaSecret: "",
	}
}
