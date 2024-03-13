// config.go
package config

import (
	"time"

	"github.com/caarlos0/env/v6"
)

var Data *Config

type Config struct {

	// 인증을 위한 API url
	CocktailApiUrl string `env:"COCKTAIL_API_SERVER" envDefault:"http://localhost:8080/internal/log/service"`

	// 업데이트 주기
	UpdateCycle time.Duration `env:"UPDATE_CYCLE" envDefault:"15m"`
}

func init() {
	Data = &Config{}
	if err := env.Parse(Data); err != nil {
		panic(err)
	}
}
