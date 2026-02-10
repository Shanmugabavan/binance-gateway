package client

import (
	"binance-gateway/internal/services/orderbook"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"

	"binance-gateway/internal/domain"
)

type Manager struct {
	orderBookSubscription orderbook.Subscription
	mu                    sync.Mutex
	symbolsUpdates        map[string]chan domain.Depth
}

func (manager *Manager) Handle(w http.ResponseWriter, r *http.Request) {
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
			if manager.symbolsUpdates[symbol] == nil {
				manager.mu.Lock()
				manager.symbolsUpdates[symbol] = channel
				manager.mu.Unlock()
				err := manager.subscribe(conn, symbol, &channel)
				if err != nil {
					log.Println(err)
					break
				}
			}
		case "unsubscribe":
			err := manager.orderBookSubscription.Unsubscribe(symbol, &channel)
			if err != nil {
				log.Println(err)
				break
			}
			manager.mu.Lock()
			delete(manager.symbolsUpdates, symbol)
			manager.mu.Unlock()
		default:
			log.Println("Unsupported command")
		}
	}
}

// websocket will not allow concurrent right, so mutex lock added.
func (manager *Manager) write(channel <-chan domain.Depth, conn *websocket.Conn) {
	for event := range channel {
		eventBytes, _ := json.Marshal(event)
		manager.mu.Lock()
		err := conn.WriteMessage(websocket.TextMessage, eventBytes)
		manager.mu.Unlock()
		if err != nil {
			return
		}
	}
}

func (manager *Manager) subscribe(conn *websocket.Conn, symbol string, channel *chan domain.Depth) error {
	snapshot, err := manager.orderBookSubscription.Subscribe(symbol, channel)
	if err != nil {
		log.Println(err)
	}

	snapshotBytes, _ := json.Marshal(snapshot)

	manager.mu.Lock()
	err = conn.WriteMessage(websocket.TextMessage, snapshotBytes)
	manager.mu.Unlock()

	if err != nil {
		return err
	}

	go manager.write(*channel, conn)
	return nil
}

func NewManager(orderBookProcessor orderbook.Subscription) *Manager {
	return &Manager{
		orderBookSubscription: orderBookProcessor,
		mu:                    sync.Mutex{},
		symbolsUpdates:        make(map[string]chan domain.Depth),
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
