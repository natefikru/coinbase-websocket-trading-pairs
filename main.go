package main

import (
	"coinbase-websocket-trading-pairs/service"
	"coinbase-websocket-trading-pairs/util"
	"coinbase-websocket-trading-pairs/websocketClient"
	"fmt"
	"log"
)

func main() {
	log.Println("Running Coinbase Trading Pairs Web Socket Project")

	// initialize configuration
	log.Println("Initializing Configuration")
	config, err := util.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("cannot load config: %v", err))
	}

	wsClient := websocketClient.NewWebSocketClient(config)
	service := service.NewService(wsClient, config)

	log.Println("Starting Service")
	err = service.Run()
	if err != nil {
		panic(fmt.Sprintf("Error running the coinbase websocket service: %v", err))
	}
}
