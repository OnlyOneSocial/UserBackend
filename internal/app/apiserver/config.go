package apiserver

// Selectel ...
type Selectel struct {
	User     string `toml:"user"`
	Password string `toml:"password"`
}

//Config ...
type Config struct {
	BindAddr        string   `toml:"bind_addr"`
	LogLevel        string   `toml:"log_level"`
	DatabaseURL     string   `toml:"database_url"`
	JwtSignKey      string   `toml:"jwtsignkey"`
	RecaptchaSecret string   `toml:"recaptcha_secret"`
	Selectel        Selectel `toml:"Selectel"`
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
