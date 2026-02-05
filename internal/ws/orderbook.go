package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"

	"binance-gateway/internal/domain"
)

type orderBookProcessor interface {
	InitiateOrderBook(symbol string) (*domain.OrderBook, error)
	FetchOrderBooks()
	InitiateClientWebSocketSubscription(symbol string, channel *chan domain.Depth) (domain.SnapShot, error)
	RemoveClientWebSocketSubscription(symbol string, channel *chan domain.Depth) error
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type OrderBookWebSocketProcessor struct {
	OrderBookProcessor   orderBookProcessor
	Lock                 sync.Mutex
	SymbolMappedChannels map[string]chan domain.Depth
}

func (orderBookWebSocketProcessor *OrderBookWebSocketProcessor) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Error closing websocket connection")
		}
	}(conn)

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
			if orderBookWebSocketProcessor.SymbolMappedChannels[symbol] == nil {
				orderBookWebSocketProcessor.Lock.Lock()
				orderBookWebSocketProcessor.SymbolMappedChannels[symbol] = channel
				orderBookWebSocketProcessor.Lock.Unlock()
				err := orderBookWebSocketProcessor.SubscribeWebSocket(conn, symbol, &channel)
				if err != nil {
					log.Println(err)
					break
				}
			}
		case "unsubscribe":
			err := orderBookWebSocketProcessor.OrderBookProcessor.RemoveClientWebSocketSubscription(symbol, &channel)
			if err != nil {
				log.Println(err)
				break
			}
			orderBookWebSocketProcessor.Lock.Lock()
			delete(orderBookWebSocketProcessor.SymbolMappedChannels, symbol)
			orderBookWebSocketProcessor.Lock.Unlock()
		default:
			log.Println("Unsupported command")
		}
	}
}

// websocket will not allow concurrent right, so mutex lock added.
func (orderBookWebSocketProcessor *OrderBookWebSocketProcessor) writeToWebsocket(channel <-chan domain.Depth, conn *websocket.Conn) {
	for event := range channel {
		eventBytes, _ := json.Marshal(event)
		orderBookWebSocketProcessor.Lock.Lock()
		err := conn.WriteMessage(websocket.TextMessage, eventBytes)
		orderBookWebSocketProcessor.Lock.Unlock()
		if err != nil {
			return
		}
	}
}

func (orderBookWebSocketProcessor *OrderBookWebSocketProcessor) SubscribeWebSocket(conn *websocket.Conn, symbol string, channel *chan domain.Depth) error {
	snapshot, err := orderBookWebSocketProcessor.OrderBookProcessor.InitiateClientWebSocketSubscription(symbol, channel)
	if err != nil {
		log.Println(err)
	}

	snapshotBytes, _ := json.Marshal(snapshot)

	orderBookWebSocketProcessor.Lock.Lock()
	err = conn.WriteMessage(websocket.TextMessage, snapshotBytes)
	orderBookWebSocketProcessor.Lock.Unlock()

	if err != nil {
		return err
	}

	go orderBookWebSocketProcessor.writeToWebsocket(*channel, conn)
	return nil
}

func CreateWebSocketProcessor(orderBookProcessor orderBookProcessor) *OrderBookWebSocketProcessor {
	return &OrderBookWebSocketProcessor{
		OrderBookProcessor:   orderBookProcessor,
		Lock:                 sync.Mutex{},
		SymbolMappedChannels: make(map[string]chan domain.Depth),
	}
}
