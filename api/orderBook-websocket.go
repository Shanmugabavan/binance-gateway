package api

import (
	"binance-gateway/internal/services/orderbook"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type IOrderBookWebSocketService interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type OrderBookWebSocketService struct {
	OrderBookService orderbook.OrderBookService
}

func (orderBookWebSocketService *OrderBookWebSocketService) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	defer conn.Close()

	if err != nil {
		log.Println(err)
	}

	symbol := r.URL.Query().Get("symbol")

	snapshot, channel, err := orderBookWebSocketService.OrderBookService.InitiateClientWebSocketSubscription(symbol)
	if err != nil {
		log.Println(err)
	}

	snapshotBytes, _ := json.Marshal(snapshot)
	err = conn.WriteMessage(websocket.TextMessage, snapshotBytes)
	if err != nil {
		return
	}

	for event := range channel {
		eventBytes, _ := json.Marshal(event)
		err = conn.WriteMessage(websocket.TextMessage, eventBytes)
		if err != nil {
			return
		}
	}

	defer orderBookWebSocketService.OrderBookService.RemoveClientWebSocketSubscription(symbol, &channel)

}
