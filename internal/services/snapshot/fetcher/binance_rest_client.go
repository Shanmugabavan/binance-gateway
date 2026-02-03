package fetcher

import (
	"encoding/json"
	"io"
	"net/http"

	"binance-gateway/bootstrap"
	"binance-gateway/internal/services/snapshot/fetcher/dto"
)

func FetchSnapShotBySymbol(symbol string) (dto.SnapshotResponse, error) {
	snapshotResponse := dto.SnapshotResponse{}
	env := bootstrap.EnvironmentSingleton
	resp, err2 := http.Get(env.BinanceBaseURL + "/api/v3/depth?symbol=" + symbol)

	if err2 != nil {
		return snapshotResponse, err2
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	err := json.NewDecoder(resp.Body).Decode(&snapshotResponse)
	if err != nil {
		return snapshotResponse, err
	}
	//channel <- snapshotResponse
	return snapshotResponse, nil
}
