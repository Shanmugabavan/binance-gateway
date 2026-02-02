package domain

import "strconv"

type BidAsk struct {
	Price    float64
	Quantity float64
}

func ConvertArrayToBidAsk(bidAskArray [][2]string) ([]BidAsk, error) {
	var bidAsks []BidAsk

	for _, bidAsk := range bidAskArray {
		price, err := strconv.ParseFloat(bidAsk[0], 64)

		if err != nil {
			return nil, err
		}

		quantity, err := strconv.ParseFloat(bidAsk[1], 64)
		if err != nil {
			return nil, err
		}

		bidAsk := BidAsk{
			price,
			quantity,
		}

		bidAsks = append(bidAsks, bidAsk)
	}
	return bidAsks, nil
}
