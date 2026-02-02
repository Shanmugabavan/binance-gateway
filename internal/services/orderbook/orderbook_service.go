package orderbook

import (
	"binance-gateway/configs"
	"binance-gateway/internal/domain"
	"binance-gateway/internal/services/depth"
	"binance-gateway/internal/services/snapshot"
	"errors"
	"fmt"
	"log"
	"sync"
)

type OrderBook struct {
	snapshot    domain.SnapShot
	subscribers []chan domain.Depth
	depthChan   chan domain.Depth
	symbol      string
	mutex       sync.Mutex
}

type OrderBookService interface {
	InitiateOrderBook(symbol string) (*OrderBook, error)
	FetchOrderBooks()
	InitiateClientWebSocketSubscription(symbol string) (domain.SnapShot, chan domain.Depth, error)
	RemoveClientWebSocketSubscription(symbol string, channel *chan domain.Depth) error
}

type BinanceOrderBookService struct {
	DepthService    depth.DepthService
	SnapShotService snapshot.SnapShotService
	OrderBooks      map[string]*OrderBook
	mutex           sync.Mutex
}

func (orderBookService *BinanceOrderBookService) InitiateOrderBook(symbol string) (*OrderBook, error) {
	orderBook := OrderBook{}

	orderBookService.mutex.Lock()
	orderBookService.OrderBooks[symbol] = &orderBook
	orderBookService.mutex.Unlock()

	orderBook.depthChan = make(chan domain.Depth)
	orderBook.symbol = symbol

	go orderBookService.DepthService.ConnectDepthWebsocketForSymbol(orderBook.symbol, orderBook.depthChan)

	snapShot, err := orderBookService.SnapShotService.GetSnapShotBySymbol(symbol)

	if err != nil {
		return &orderBook, err
	}

	orderBook.snapshot = snapShot

	go orderBookService.applyBufferedEvents(&orderBook)

	return &orderBook, nil

}

func (orderBookService *BinanceOrderBookService) applyBufferedEvents(orderbook *OrderBook) error {
	for depthUpdate := range orderbook.depthChan {
		// skipping old events
		if depthUpdate.FinalUpdateId < orderbook.snapshot.LastUpdateId {
			continue
		}

		//  get new snapshot if all the events are newer than current snapshot
		if depthUpdate.FirstUpdateId > orderbook.snapshot.LastUpdateId+1 {
			newSnapshot, err := orderBookService.SnapShotService.GetSnapShotBySymbol(orderbook.symbol)

			if err != nil {
				return err
			}
			orderbook.snapshot = newSnapshot
			fmt.Printf("fetching new order book snapshot: %+v\n", orderbook.snapshot)
			continue
		}

		// update the order book
		orderBookService.applyUpdateEvent(orderbook, depthUpdate)
		fmt.Printf("updated order book snapshot: %+v\n", orderbook.snapshot)

		// update subscribed channels
		orderBookService.updateSubscriptionChannels(orderbook, depthUpdate)

	}
	return nil
}

func (orderBookService *BinanceOrderBookService) applyUpdateEvent(orderBook *OrderBook, update domain.Depth) {
	for i, bid := range update.Bids {
		if bid.Quantity == 0.0 {
			orderBook.snapshot.Bids = append(orderBook.snapshot.Bids[:i], update.Bids[i+1:]...)
		} else {
			orderBook.snapshot.Bids = append(orderBook.snapshot.Bids, bid)
		}
	}

	for i, ask := range update.Asks {
		if ask.Quantity == 0.0 {
			orderBook.snapshot.Asks = append(orderBook.snapshot.Asks[:i], update.Asks[i+1:]...)
		} else {
			orderBook.snapshot.Asks = append(orderBook.snapshot.Asks, ask)
		}
	}

	orderBook.snapshot.LastUpdateId = update.FinalUpdateId
}

// Initiate Websocket connection
// Lock the order book and give the snapshot to client and unlock the order book
// With this client can get the latest orderbook then with the channel can get the updated events
func (orderBookService *BinanceOrderBookService) InitiateClientWebSocketSubscription(symbol string) (domain.SnapShot, chan domain.Depth, error) {
	orderBook, found := orderBookService.OrderBooks[symbol]

	if !found {
		return domain.SnapShot{}, nil, errors.New("order book not found")
	}

	orderBook.mutex.Lock()

	channel := make(chan domain.Depth, 100)
	currentSnapshot := orderBook.snapshot
	orderBook.subscribers = append(orderBook.subscribers, channel)

	orderBook.mutex.Unlock()

	return currentSnapshot, channel, nil
}

func (orderBookService *BinanceOrderBookService) RemoveClientWebSocketSubscription(symbol string, channel *chan domain.Depth) error {
	orderBook, found := orderBookService.OrderBooks[symbol]
	if !found {
		return errors.New("order book not found")
	}

	orderBook.mutex.Lock()

	channelIndex := -1

	for i, subscriberChannel := range orderBook.subscribers {
		if &subscriberChannel == channel {
			channelIndex = i
			break
		}
	}
	if channelIndex == -1 {
		orderBook.subscribers = append(orderBook.subscribers[:channelIndex], orderBook.subscribers[channelIndex+1:]...)
	}

	orderBook.mutex.Unlock()

	return nil
}

func (orderBookService *BinanceOrderBookService) updateSubscriptionChannels(orderBook *OrderBook, depth domain.Depth) {
	for _, sub := range orderBook.subscribers {
		sub <- depth
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
	var snapshotService snapshot.SnapShotService
	snapshotService = &snapshot.BinanceSnapShotService{}

	var depthService depth.DepthService
	depthService = &depth.BinanceDepthService{}

	orderBooksDic := make(map[string]*OrderBook)

	orderBookService := BinanceOrderBookService{
		depthService,
		snapshotService,
		orderBooksDic,
		sync.Mutex{},
	}

	return &orderBookService
}
