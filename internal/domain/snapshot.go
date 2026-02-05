package domain

type SnapShot struct {
	LastUpdateId      int64  `json:"lastUpdateId"`
	MessageOutputTime int64  `json:"E"`
	TransactionTime   int64  `json:"T"`
	Bids              []Side `json:"bids"`
	Asks              []Side `json:"asks"`
}
