package util

type Config struct {
	CoinbaseSocketURL string
	FileName          string
}

// LoadConfig: gets config variables from os environment
func LoadConfig() (*Config, error) {
	coinbaseSocketURL := "wss://ws-feed.exchange.coinbase.com"
	fileName := "messages.log"
	return &Config{
		CoinbaseSocketURL: coinbaseSocketURL,
		FileName:          fileName,
	}, nil
}
