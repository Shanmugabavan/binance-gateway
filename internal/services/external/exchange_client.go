package external

import (
	"binance-gateway/bootstrap"
	"binance-gateway/internal/domain"
	external "binance-gateway/internal/services/external/dto"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

type BinanceExchangeClient struct {
}

func (binanceExchangeClient *BinanceExchangeClient) ConnectDepthWebsocketForSymbol(symbol string, channel chan<- domain.Depth) {
	lowerSymbol := strings.ToLower(symbol)
	binanceExchangeClient.fetchDepthStream(lowerSymbol, channel)
}

func (binanceExchangeClient *BinanceExchangeClient) GetSnapShotBySymbol(symbol string) (domain.SnapShot, error) {
	response, err := binanceExchangeClient.fetchSnapShotBySymbol(symbol)

	if err != nil {
		return domain.SnapShot{}, err
	}

	snapshot, convError := binanceExchangeClient.convertSnapShotResponseToSnapShot(response)

	if convError != nil {
		return domain.SnapShot{}, convError
	}

	return snapshot, nil
}

func (binanceExchangeClient *BinanceExchangeClient) convertSnapShotResponseToSnapShot(response external.SnapshotResponse) (domain.SnapShot, error) {
	bids, err1 := domain.ConvertArrayToSide(response.Bids)
	asks, err2 := domain.ConvertArrayToSide(response.Asks)

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

func (binanceExchangeClient *BinanceExchangeClient) fetchSnapShotBySymbol(symbol string) (external.SnapshotResponse, error) {
	snapshotResponse := external.SnapshotResponse{}
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

func (binanceExchangeClient *BinanceExchangeClient) fetchDepthStream(symbol string, channel chan<- domain.Depth) {
	env := bootstrap.EnvironmentSingleton
	u := url.URL{Scheme: "wss", Host: env.BinanceBaseWebsocketURL, Path: "/ws/" + strings.ToLower(symbol) + "@depth"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

	if err != nil {
		log.Println("dial:", err)
	}

	defer func(c *websocket.Conn) {
		err := c.Close()
		if err != nil {
			log.Println("close:", err)
		}
	}(c)

	binanceExchangeClient.writeToChannel(c, channel)
}

func (binanceExchangeClient *BinanceExchangeClient) writeToChannel(conn *websocket.Conn, channel chan<- domain.Depth) {
	for {
		_, message, err := conn.ReadMessage()

		response := external.DepthWSResponse{}

		if err != nil {
			log.Println("read:", err)
			break
		}

		jsonErr := json.Unmarshal(message, &response)
		if jsonErr != nil {
			log.Println("jsonErr:", jsonErr)
		}

		bids, convertErr := domain.ConvertArrayToSide(response.Bids)
		asks, convertErr2 := domain.ConvertArrayToSide(response.Asks)

		if convertErr != nil {
			log.Println("ConvertArrayToSide:", err)
		}
		if convertErr2 != nil {
			log.Println("ConvertArrayToSide:", err)
		}

		channel <- domain.Depth{
			EventType:     response.EventType,
			EventTime:     response.EventTime,
			Symbol:        response.Symbol,
			FirstUpdateId: response.FirstUpdateId,
			FinalUpdateId: response.FinalUpdateId,
			Bids:          bids,
			Asks:          asks,
		}
	}
}
