package fetcher

import (
	"encoding/json"
	"log"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"

	"binance-gateway/bootstrap"
	"binance-gateway/internal/domain"
	"binance-gateway/internal/services/depth/fetcher/dto"
)

func FetchDepthStream(symbol string, channel chan<- domain.Depth) {
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

	writeToChannel(c, channel)
}

func writeToChannel(conn *websocket.Conn, channel chan<- domain.Depth) {
	for {
		_, message, err := conn.ReadMessage()

		response := dto.DepthWSResponse{}

		if err != nil {
			log.Println("read:", err)
			break
		}

		jsonErr := json.Unmarshal(message, &response)
		if jsonErr != nil {
			log.Println("jsonErr:", jsonErr)
		}

		bids, convertErr := domain.ConvertArrayToBidAsk(response.Bids)
		asks, convertErr2 := domain.ConvertArrayToBidAsk(response.Asks)

		if convertErr != nil {
			log.Println("ConvertArrayToBidAsk:", err)
		}
		if convertErr2 != nil {
			log.Println("ConvertArrayToBidAsk:", err)
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
