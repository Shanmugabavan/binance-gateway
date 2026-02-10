package orderbookTest

import (
	"binance-gateway/internal/domain"
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
