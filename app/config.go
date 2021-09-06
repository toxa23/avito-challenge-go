package app

import (
	"github.com/caarlos0/env/v6"
)

type configType struct {
	DbUrl      string `env:"DB_URL,required"`
	RedisUrl   string `env:"REDIS_URL" envDefault:""`
	HttpPort   int    `env:"HTTP_PORT" envDefault:"5000"`
	Currency   string `env:"DEFAULT_CURRENCY" envDefault:"RUB"`
	CorsOrigin string `env:"CORS_ORIGIN" envDefault:"*"`
	PerPage    int    `envDefault:"25"`
}

// Config Global configuration object
var Config configType

// InitConfig Read config settings from environment
func InitConfig() error {
	return env.Parse(&Config)
}
