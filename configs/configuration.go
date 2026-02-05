package configs

import (
	"log"

	"github.com/spf13/viper"
)

var Config *Configuration

type Configuration struct {
	Defaults defaults `mapstructure:"defaults"`
}

type defaults struct {
	Symbols []string `mapstructure:"symbols"`
}

func (conf *Configuration) GetSymbols() []string {
	return conf.Defaults.Symbols
}

func InitConfiguration() {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs/")

	err := v.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	err = v.Unmarshal(&Config)
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
}
