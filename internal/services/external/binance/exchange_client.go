package binance

import (
	"binance-gateway/bootstrap"
	"binance-gateway/internal/domain"
	external2 "binance-gateway/internal/services/external/binance/dto"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

type ExchangeClient struct {
}

func (binanceExchangeClient *ExchangeClient) Feed(symbol string, channel chan<- domain.Depth) error {
	lowerSymbol := strings.ToLower(symbol)
	err := binanceExchangeClient.feed(lowerSymbol, channel)
	if err != nil {
		return err
	}
	return nil
}

func (binanceExchangeClient *ExchangeClient) GetSnapShot(symbol string) (domain.SnapShot, error) {
	response, err := binanceExchangeClient.getSnapShot(symbol)

	if err != nil {
		return domain.SnapShot{}, err
	}

	snapshot, convError := binanceExchangeClient.toSnapShot(response)

	if convError != nil {
		return domain.SnapShot{}, convError
	}

	return snapshot, nil
}

func (binanceExchangeClient *ExchangeClient) toSnapShot(response external2.Snapshot) (domain.SnapShot, error) {
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

func (binanceExchangeClient *ExchangeClient) getSnapShot(symbol string) (external2.Snapshot, error) {
	snapshotResponse := external2.Snapshot{}
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

	return snapshotResponse, nil
}

func (binanceExchangeClient *ExchangeClient) feed(symbol string, channel chan<- domain.Depth) error {
	env := bootstrap.EnvironmentSingleton
	u := url.URL{Scheme: "wss", Host: env.BinanceBaseWebsocketURL, Path: "/ws/" + strings.ToLower(symbol) + "@depth"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

	if err != nil {

		return err
	}

	defer func(c *websocket.Conn) {
		err := c.Close()
		if err != nil {
			log.Println("close:", err)
			return
		}
	}(c)

	binanceExchangeClient.push(c, channel)
	return nil
}

func (binanceExchangeClient *ExchangeClient) push(conn *websocket.Conn, channel chan<- domain.Depth) {
	for {
		_, message, err := conn.ReadMessage()

		response := external2.Depth{}

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

func NewExchangeClient() *ExchangeClient {
	return &ExchangeClient{}
}
