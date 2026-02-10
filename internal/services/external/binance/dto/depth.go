package external

type Depth struct {
	EventType     string      `json:"e"`
	EventTime     int64       `json:"E"`
	Symbol        string      `json:"s"`
	FirstUpdateId int64       `json:"U"`
	FinalUpdateId int64       `json:"u"`
	Bids          [][2]string `json:"b"`
	Asks          [][2]string `json:"a"`
}
