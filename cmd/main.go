package main

import (
	"binance-gateway/internal/services/client"
	"binance-gateway/internal/services/external/binance"
	"fmt"
	"net/http"

	"binance-gateway/bootstrap"
	"binance-gateway/configs"
	"binance-gateway/internal/services/orderbook"
)

func main() {
	_ = bootstrap.App()

	configs.InitConfiguration()

	binanceClient := binance.NewExchangeClient()

	orderBookProcessor := orderbook.NewProcessor(binanceClient)
	orderBookProcessor.FetchOrderBooks()

	clientManager := client.NewManager(orderBookProcessor)

	http.HandleFunc("/ws", clientManager.Handle)

	fmt.Println("WebSocket server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}
