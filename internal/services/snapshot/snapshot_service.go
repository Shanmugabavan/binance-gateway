package snapshot

import (
	"binance-gateway/internal/domain"
	"binance-gateway/internal/services/snapshot/fetcher"
	"binance-gateway/internal/services/snapshot/fetcher/models"
)

type SnapShotService interface {
	GetSnapShotBySymbol(symbol string) (domain.SnapShot, error)
}

type BinanceSnapShotService struct {
}

func ConvertSnapShotResponseToSnapShot(response models.SnapshotResponse) (domain.SnapShot, error) {
	bids, err := domain.ConvertArrayToBidAsk(response.Bids)
	asks, err := domain.ConvertArrayToBidAsk(response.Asks)

	if err != nil {
		return domain.SnapShot{}, err
	}

	snapshot := domain.SnapShot{
		LastUpdateId:      response.LastUpdateId,
		MessageOutputTime: response.MessageOutputTime,
		TransactionTime:   response.TransactionTime,
		Bids:              bids,
		Asks:              asks,
	}
	return snapshot, nil
}

func (snap *BinanceSnapShotService) GetSnapShotBySymbol(symbol string) (domain.SnapShot, error) {
	response, err := fetcher.FetchSnapShotBySymbol(symbol)

	if err != nil {
		return domain.SnapShot{}, err
	}

	snapshot, convError := ConvertSnapShotResponseToSnapShot(response)

	if convError != nil {
		return domain.SnapShot{}, convError
	}

	return snapshot, nil
}
