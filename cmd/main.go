package main

import (
	"binance-gateway/internal/ws"
	"fmt"
	"net/http"

	"binance-gateway/bootstrap"
	"binance-gateway/configs"
	"binance-gateway/internal/services/orderbook"
)

func main() {
	_ = bootstrap.App()

	configs.InitConfiguration()

	orderBookProcessor := orderbook.CreateOrderBookProcessor()
	orderBookProcessor.FetchOrderBooks()

	wsProcessor := ws.CreateWebSocketProcessor(orderBookProcessor)

	http.HandleFunc("/ws", wsProcessor.Handle)

	fmt.Println("WebSocket server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}
