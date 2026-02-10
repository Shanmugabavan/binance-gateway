package external

type Snapshot struct {
	LastUpdateId      int64       `json:"lastUpdateId"`
	MessageOutputTime int64       `json:"E"`
	TransactionTime   int64       `json:"T"`
	Bids              [][2]string `json:"bids"`
	Asks              [][2]string `json:"asks"`
}
