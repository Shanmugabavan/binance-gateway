package domain

import "sync"

type OrderBook struct {
	Snapshot    SnapShot
	Subscribers []*chan Depth
	DepthChan   chan Depth
	Symbol      string
	Mutex       sync.Mutex
}
