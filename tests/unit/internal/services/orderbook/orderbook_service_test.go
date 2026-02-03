package orderbook

import "binance-gateway/internal/domain"

type DepthServiceMock struct {
}

func (depthServiceMock *DepthServiceMock) ConnectDepthWebsocketForSymbol(symbol string, channel chan<- domain.Depth) {
	channel <- domain.Depth{
		Symbol:        symbol,
		FinalUpdateId: int64(200),
		FirstUpdateId: int64(100),
	}
}

type SnapShotServiceMock struct {
}

func (SnapShotServiceMock *SnapShotServiceMock) GetSnapShotBySymbol(symbol string) (domain.SnapShot, error) {
	return domain.SnapShot{
		LastUpdateId: int64(102),
	}, nil
}
