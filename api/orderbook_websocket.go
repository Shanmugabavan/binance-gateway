package api

import (
	"binance-gateway/internal/domain"
	"binance-gateway/internal/services/orderbook"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

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
	OrderBookService     orderbook.OrderBookService
	Lock                 sync.Mutex
	SymbolMappedChannels map[string]chan domain.Depth
}

func (orderBookWebSocketService *OrderBookWebSocketService) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	defer conn.Close()

	channel := make(chan domain.Depth)

	if err != nil {
		log.Println(err)
	}

	for {
		_, readMessageBuffer, err := conn.ReadMessage()
		readMessage := string(readMessageBuffer)
		readMessageArray := strings.Fields(readMessage)

		subsUnsubscribe := readMessageArray[0]
		symbol := readMessageArray[1]

		if err != nil {
			log.Println(err)
		}

		switch subsUnsubscribe {
		case "subscribe":
			if orderBookWebSocketService.SymbolMappedChannels[symbol] == nil {
				orderBookWebSocketService.Lock.Lock()
				orderBookWebSocketService.SymbolMappedChannels[symbol] = channel
				orderBookWebSocketService.Lock.Unlock()
				err := orderBookWebSocketService.SubscribeWebSocket(conn, symbol, &channel)
				if err != nil {
					log.Println(err)
					break
				}
			}
		case "unsubscribe":
			err := orderBookWebSocketService.OrderBookService.RemoveClientWebSocketSubscription(symbol, &channel)
			if err != nil {
				log.Println(err)
				break
			}
			orderBookWebSocketService.Lock.Lock()
			delete(orderBookWebSocketService.SymbolMappedChannels, symbol)
			orderBookWebSocketService.Lock.Unlock()
		default:
			log.Println("Unsupported command")
		}
	}
}

// websocket will not allow concurrent right, so mutex lock added.
func (orderBookWebSocketService *OrderBookWebSocketService) writeToWebsocket(channel <-chan domain.Depth, conn *websocket.Conn) {
	for event := range channel {
		eventBytes, _ := json.Marshal(event)
		orderBookWebSocketService.Lock.Lock()
		err := conn.WriteMessage(websocket.TextMessage, eventBytes)
		orderBookWebSocketService.Lock.Unlock()
		if err != nil {
			return
		}
	}
}

func (orderBookWebSocketService *OrderBookWebSocketService) SubscribeWebSocket(conn *websocket.Conn, symbol string, channel *chan domain.Depth) error {
	snapshot, err := orderBookWebSocketService.OrderBookService.InitiateClientWebSocketSubscription(symbol, channel)
	if err != nil {
		log.Println(err)
	}

	snapshotBytes, _ := json.Marshal(snapshot)

	orderBookWebSocketService.Lock.Lock()
	err = conn.WriteMessage(websocket.TextMessage, snapshotBytes)
	orderBookWebSocketService.Lock.Unlock()

	if err != nil {
		return err
	}

	go orderBookWebSocketService.writeToWebsocket(*channel, conn)
	return nil
}

func CreateWebSocketService(orderBookService orderbook.OrderBookService) *OrderBookWebSocketService {
	return &OrderBookWebSocketService{
		OrderBookService:     orderBookService,
		Lock:                 sync.Mutex{},
		SymbolMappedChannels: make(map[string]chan domain.Depth),
	}
}
