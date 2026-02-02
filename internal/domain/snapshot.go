package domain

type SnapShot struct {
	LastUpdateId      int64    `json:"lastUpdateId"`
	MessageOutputTime int64    `json:"E"`
	TransactionTime   int64    `json:"T"`
	Bids              []BidAsk `json:"bids"`
	Asks              []BidAsk `json:"asks"`
}
