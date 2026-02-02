package depth

import (
	"binance-gateway/internal/domain"
	"binance-gateway/internal/services/depth/fetcher"
	"strings"
)

type DepthService interface {
	ConnectDepthWebsocketForSymbol(symbol string, channel chan<- domain.Depth)
}

type BinanceDepthService struct {
}

func (depthService *BinanceDepthService) ConnectDepthWebsocketForSymbol(symbol string, channel chan<- domain.Depth) {
	lowerSymbol := strings.ToLower(symbol)
	fetcher.FetchDepthStream(lowerSymbol, channel)
}
