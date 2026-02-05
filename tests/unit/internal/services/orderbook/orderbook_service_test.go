package orderbookTest

import (
	"binance-gateway/internal/domain"
	"binance-gateway/internal/services/orderbook"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type exchangeClientMock struct{}

func (exchangeClientMock *exchangeClientMock) ConnectDepthWebsocketForSymbol(symbol string, channel chan<- domain.Depth) {
	channel <- domain.Depth{
		Symbol:        symbol,
		FinalUpdateId: int64(200),
		FirstUpdateId: int64(100),
	}
}

func (exchangeClientMock *exchangeClientMock) GetSnapShotBySymbol(symbol string) (domain.SnapShot, error) {
	return domain.SnapShot{
		LastUpdateId: int64(102),
	}, nil
}

func TestInitiateOrderBookWithSuccess(t *testing.T) {
	exchangeClient := exchangeClientMock{}
	orderBooksDic := make(map[string]*domain.OrderBook)
	orderBookProcessor := orderbook.BinanceOrderBookProcessor{
		&exchangeClient,
		orderBooksDic,
		sync.Mutex{},
	}

	book, err := orderBookProcessor.InitiateOrderBook("BTCUSDT")
	if err != nil {
		return
	}

	assert.NotNil(t, book)
}
