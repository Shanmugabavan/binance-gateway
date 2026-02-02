package orderbook

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"binance-gateway/configs"
	"binance-gateway/internal/domain"
	"binance-gateway/internal/services/depth"
	"binance-gateway/internal/services/snapshot"
)

type OrderBookService interface {
	InitiateOrderBook(symbol string) (*domain.OrderBook, error)
	FetchOrderBooks()
	InitiateClientWebSocketSubscription(symbol string, channel *chan domain.Depth) (domain.SnapShot, error)
	RemoveClientWebSocketSubscription(symbol string, channel *chan domain.Depth) error
}

type BinanceOrderBookService struct {
	DepthService    depth.DepthService
	SnapShotService snapshot.SnapShotService
	OrderBooks      map[string]*domain.OrderBook
	mutex           sync.Mutex
}

func (orderBookService *BinanceOrderBookService) InitiateOrderBook(symbol string) (*domain.OrderBook, error) {
	orderBook := domain.OrderBook{}

	orderBookService.mutex.Lock()
	orderBookService.OrderBooks[symbol] = &orderBook
	orderBookService.mutex.Unlock()

	orderBook.DepthChan = make(chan domain.Depth)
	orderBook.Symbol = symbol

	go orderBookService.DepthService.ConnectDepthWebsocketForSymbol(orderBook.Symbol, orderBook.DepthChan)

	snapShot, err := orderBookService.SnapShotService.GetSnapShotBySymbol(symbol)

	if err != nil {
		return &orderBook, err
	}

	orderBook.Snapshot = snapShot

	go func() {
		err := orderBookService.applyBufferedEvents(&orderBook)
		if err != nil {
			log.Println(err)
		}
	}()

	return &orderBook, nil

}

func (orderBookService *BinanceOrderBookService) applyBufferedEvents(orderbook *domain.OrderBook) error {
	for depthUpdate := range orderbook.DepthChan {
		// skipping old events
		if depthUpdate.FinalUpdateId < orderbook.Snapshot.LastUpdateId {
			continue
		}

		//  get new snapshot if all the events are newer than current snapshot
		if depthUpdate.FirstUpdateId > orderbook.Snapshot.LastUpdateId+1 {
			newSnapshot, err := orderBookService.SnapShotService.GetSnapShotBySymbol(orderbook.Symbol)

			if err != nil {
				return err
			}
			orderbook.Snapshot = newSnapshot
			fmt.Printf("fetching new order book snapshot: %+v\n", orderbook.Snapshot)
			continue
		}

		// update the order book
		orderBookService.applyUpdateEvent(orderbook, depthUpdate)
		fmt.Printf("updated order book snapshot: %+v\n", orderbook.Snapshot)

		// update subscribed channels
		orderBookService.updateSubscriptionChannels(orderbook, depthUpdate)

	}
	return nil
}

func (orderBookService *BinanceOrderBookService) applyUpdateEvent(orderBook *domain.OrderBook, update domain.Depth) {
	for i, bid := range update.Bids {
		if bid.Quantity == 0.0 {
			orderBook.Snapshot.Bids = append(orderBook.Snapshot.Bids[:i], update.Bids[i+1:]...)
		} else {
			orderBook.Snapshot.Bids = append(orderBook.Snapshot.Bids, bid)
		}
	}

	for i, ask := range update.Asks {
		if ask.Quantity == 0.0 {
			orderBook.Snapshot.Asks = append(orderBook.Snapshot.Asks[:i], update.Asks[i+1:]...)
		} else {
			orderBook.Snapshot.Asks = append(orderBook.Snapshot.Asks, ask)
		}
	}

	orderBook.Snapshot.LastUpdateId = update.FinalUpdateId
}

// Initiate Websocket connection
// Lock the order book and give the snapshot to client and unlock the order book
// With this client can get the latest orderbook then with the channel can get the updated events
func (orderBookService *BinanceOrderBookService) InitiateClientWebSocketSubscription(symbol string, channel *chan domain.Depth) (domain.SnapShot, error) {
	orderBook, found := orderBookService.OrderBooks[symbol]

	if !found {
		return domain.SnapShot{}, errors.New("order book not found")
	}

	orderBook.Mutex.Lock()
	currentSnapshot := orderBook.Snapshot
	orderBook.Subscribers = append(orderBook.Subscribers, channel)
	orderBook.Mutex.Unlock()

	return currentSnapshot, nil
}

func (orderBookService *BinanceOrderBookService) RemoveClientWebSocketSubscription(symbol string, channel *chan domain.Depth) error {
	orderBook, found := orderBookService.OrderBooks[symbol]
	if !found {
		return errors.New("order book not found")
	}

	orderBook.Mutex.Lock()

	var newSubscribers = make([]*chan domain.Depth, 0)

	for _, subscriberChannel := range orderBook.Subscribers {
		if channel != subscriberChannel {
			newSubscribers = append(newSubscribers, subscriberChannel)
		}
	}

	orderBook.Subscribers = newSubscribers

	orderBook.Mutex.Unlock()

	return nil
}

func (orderBookService *BinanceOrderBookService) updateSubscriptionChannels(orderBook *domain.OrderBook, depth domain.Depth) {
	for _, sub := range orderBook.Subscribers {
		*sub <- depth
	}
}

func (orderBookService *BinanceOrderBookService) FetchOrderBooks() {
	for _, symbol := range configs.Config.GetSymbols() {
		go func() {
			_, err := orderBookService.InitiateOrderBook(symbol)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
}

func CreateOrderBookService() *BinanceOrderBookService {
	var snapshotService snapshot.SnapShotService = &snapshot.BinanceSnapShotService{}

	var depthService depth.DepthService = &depth.BinanceDepthService{}

	orderBooksDic := make(map[string]*domain.OrderBook)

	orderBookService := BinanceOrderBookService{
		depthService,
		snapshotService,
		orderBooksDic,
		sync.Mutex{},
	}

	return &orderBookService
}
