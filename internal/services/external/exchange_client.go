package external

import "binance-gateway/internal/domain"

type ExchangeClient interface {
	GetSnapShot(symbol string) (domain.SnapShot, error)
	Feed(symbol string, channel chan<- domain.Depth) error
}
