package main

import (
	"binance-gateway/api"
	"binance-gateway/bootstrap"
	"binance-gateway/configs"
	"binance-gateway/internal/services/orderbook"
	"fmt"
	"net/http"
)

func main() {
	_ = bootstrap.App()

	configs.InitConfiguration()

	orderBookService := orderbook.CreateOrderBookService()
	orderBookService.FetchOrderBooks()

	service := api.CreateWebSocketService(orderBookService)

	http.HandleFunc("/ws", service.Handle)

	fmt.Println("WebSocket server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

	//select {}

}
