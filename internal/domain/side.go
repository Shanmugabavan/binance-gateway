package domain

import "strconv"

type Side struct {
	Price    float64
	Quantity float64
}

func ConvertArrayToSide(sidesArray [][2]string) ([]Side, error) {
	var sides []Side

	for _, side := range sidesArray {
		price, err := strconv.ParseFloat(side[0], 64)

		if err != nil {
			return nil, err
		}

		quantity, err := strconv.ParseFloat(side[1], 64)
		if err != nil {
			return nil, err
		}

		side := Side{
			price,
			quantity,
		}

		sides = append(sides, side)
	}
	return sides, nil
}
