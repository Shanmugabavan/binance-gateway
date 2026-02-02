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

func (snap *BinanceSnapShotService) GetSnapShotBySymbol(symbol string) (domain.SnapShot, error) {
	response, err := fetcher.FetchSnapShotBySymbol(symbol)

	if err != nil {
		return domain.SnapShot{}, err
	}

	snapshot, convError := convertSnapShotResponseToSnapShot(response)

	if convError != nil {
		return domain.SnapShot{}, convError
	}

	return snapshot, nil
}

func convertSnapShotResponseToSnapShot(response models.SnapshotResponse) (domain.SnapShot, error) {
	bids, err1 := domain.ConvertArrayToBidAsk(response.Bids)
	asks, err2 := domain.ConvertArrayToBidAsk(response.Asks)

	if err1 != nil {
		return domain.SnapShot{}, err1
	}

	if err2 != nil {
		return domain.SnapShot{}, err2
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
