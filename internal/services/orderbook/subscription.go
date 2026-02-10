package orderbook

import "binance-gateway/internal/domain"

type Subscription interface {
	Subscribe(symbol string, channel *chan domain.Depth) (domain.SnapShot, error)
	Unsubscribe(symbol string, channel *chan domain.Depth) error
}
