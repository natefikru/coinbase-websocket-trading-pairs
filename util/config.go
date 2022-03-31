package util

type Config struct {
	CoinbaseSocketURL string
}

// LoadConfig: gets config variables from os environment
func LoadConfig() (*Config, error) {
	// coinbaseSocketURL := os.Getenv("COINBASE_SOCKET_URL")

	coinbaseSocketURL := "wss://ws-feed.exchange.coinbase.com"

	return &Config{
		CoinbaseSocketURL: coinbaseSocketURL,
	}, nil
}
