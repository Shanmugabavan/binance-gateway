package configs

var Config *Configuration

type Configuration struct {
	Symbols []string
}

func (conf *Configuration) GetSymbols() []string {
	return conf.Symbols
}

func (conf *Configuration) SetDefaultSymbols() {
	// TODO Need to fetch from configuration file
	conf.Symbols = append(conf.Symbols, "BTCUSDT")
	conf.Symbols = append(conf.Symbols, "SOLUSDT")

}

func InitConfiguration() {
	Config = &Configuration{}
	Config.SetDefaultSymbols()
}
