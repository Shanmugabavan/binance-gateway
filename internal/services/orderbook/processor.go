package orderbook

import (
	"binance-gateway/internal/services/external"
	"errors"
	"fmt"
	"log"
	"sync"

	"binance-gateway/configs"
	"binance-gateway/internal/domain"
)

type exchangeClient interface {
	ConnectDepthWebsocketForSymbol(symbol string, channel chan<- domain.Depth)
	GetSnapShotBySymbol(symbol string) (domain.SnapShot, error)
}

type BinanceOrderBookProcessor struct {
	ExchangeClient exchangeClient
	OrderBooks     map[string]*domain.OrderBook
	mutex          sync.Mutex
}

func (orderBookProcessor *BinanceOrderBookProcessor) InitiateOrderBook(symbol string) (*domain.OrderBook, error) {
	orderBook := domain.OrderBook{}

	orderBookProcessor.mutex.Lock()
	orderBookProcessor.OrderBooks[symbol] = &orderBook
	orderBookProcessor.mutex.Unlock()

	orderBook.DepthChan = make(chan domain.Depth)
	orderBook.Symbol = symbol

	go orderBookProcessor.ExchangeClient.ConnectDepthWebsocketForSymbol(orderBook.Symbol, orderBook.DepthChan)

	snapShot, err := orderBookProcessor.ExchangeClient.GetSnapShotBySymbol(symbol)

	if err != nil {
		return &orderBook, err
	}

	orderBook.Snapshot = snapShot

	go func() {
		err := orderBookProcessor.applyBufferedEvents(&orderBook)
		if err != nil {
			log.Println(err)
		}
	}()

	return &orderBook, nil

}

func (orderBookProcessor *BinanceOrderBookProcessor) applyBufferedEvents(orderbook *domain.OrderBook) error {
	for depthUpdate := range orderbook.DepthChan {
		// skipping old events
		if depthUpdate.FinalUpdateId < orderbook.Snapshot.LastUpdateId {
			continue
		}

		//  get new snapshot if all the events are newer than current snapshot
		if depthUpdate.FirstUpdateId > orderbook.Snapshot.LastUpdateId+1 {
			newSnapshot, err := orderBookProcessor.ExchangeClient.GetSnapShotBySymbol(orderbook.Symbol)

			if err != nil {
				return err
			}
			orderbook.Snapshot = newSnapshot
			fmt.Printf("fetching new order book snapshot: %+v\n", orderbook.Snapshot)
			continue
		}

		// update the order book
		orderBookProcessor.applyUpdateEvent(orderbook, depthUpdate)
		fmt.Printf("updated order book snapshot: %+v\n", orderbook.Snapshot)

		// update subscribed channels
		orderBookProcessor.updateSubscriptionChannels(orderbook, depthUpdate)

	}
	return nil
}

func (orderBookProcessor *BinanceOrderBookProcessor) applyUpdateEvent(orderBook *domain.OrderBook, update domain.Depth) {
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
func (orderBookProcessor *BinanceOrderBookProcessor) InitiateClientWebSocketSubscription(symbol string, channel *chan domain.Depth) (domain.SnapShot, error) {
	orderBook, found := orderBookProcessor.OrderBooks[symbol]

	if !found {
		return domain.SnapShot{}, errors.New("order book not found")
	}

	orderBook.Mutex.Lock()
	currentSnapshot := orderBook.Snapshot
	orderBook.Subscribers = append(orderBook.Subscribers, channel)
	orderBook.Mutex.Unlock()

	return currentSnapshot, nil
}

func (orderBookProcessor *BinanceOrderBookProcessor) RemoveClientWebSocketSubscription(symbol string, channel *chan domain.Depth) error {
	orderBook, found := orderBookProcessor.OrderBooks[symbol]
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

func (orderBookProcessor *BinanceOrderBookProcessor) updateSubscriptionChannels(orderBook *domain.OrderBook, depth domain.Depth) {
	for _, sub := range orderBook.Subscribers {
		*sub <- depth
	}
}

func (orderBookProcessor *BinanceOrderBookProcessor) FetchOrderBooks() {
	for _, symbol := range configs.Config.GetSymbols() {
		go func() {
			_, err := orderBookProcessor.InitiateOrderBook(symbol)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
}

func CreateOrderBookProcessor() *BinanceOrderBookProcessor {
	exchangeClient := external.BinanceExchangeClient{}
	orderBooksDic := make(map[string]*domain.OrderBook)

	orderBookService := BinanceOrderBookProcessor{
		&exchangeClient,
		orderBooksDic,
		sync.Mutex{},
	}

	return &orderBookService
}
