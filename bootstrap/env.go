package bootstrap

import (
	"log"

	"github.com/spf13/viper"
)

var EnvironmentSingleton *Env

type Env struct {
	BinanceBaseURL          string `mapstructure:"BINANCE_BASE_URL"`
	BinanceBaseWebsocketURL string `mapstructure:"BINANCE_BASE_WEBSOCKET_URL"`
}

func NewEnv() *Env {
	EnvironmentSingleton = &Env{}
	viper.SetConfigFile(".env")

	_ = viper.ReadInConfig()
	err := viper.Unmarshal(EnvironmentSingleton)
	if err != nil {
		log.Fatal("Can't find the file .env : ", err)
	}

	return EnvironmentSingleton
}
