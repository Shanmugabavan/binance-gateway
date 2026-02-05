package domain

type Depth struct {
	EventType     string `json:"e"`
	EventTime     int64  `json:"E"`
	Symbol        string `json:"s"`
	FirstUpdateId int64  `json:"U"`
	FinalUpdateId int64  `json:"u"`
	Bids          []Side `json:"b"`
	Asks          []Side `json:"a"`
}
