package service

import (
	"time"

	"github.com/gorilla/websocket"
)

type IWebSocketClient interface {
	EstablishConnection(url string) (*websocket.Conn, error)
	WriteMessageToSocketConnInterval(conn *websocket.Conn, data []byte, seconds time.Duration) error
	ReadMessageFromSockenConn(conn *websocket.Conn) ([]byte, error)
}

type IFileClient interface {
	InitFileConn() error
	WriteToFile(str string)
}

const (
	// Channels
	matches = "matches"

	// Types
	subscribe = "subscribe"
	match     = "match"

	// Options
	subscribeIntervalSeconds = 4
)

// getSupportedProductIDs: define trading pair product ID's that web socket listens to
func (s *Service) getSupportedProductIDs() []string {
	return []string{
		"BTC-USD",
		"ETH-USD",
		"ETH-BTC",
	}
}

// SubscribeMessage: Coinbase Websocket request struct that is sent every 4 seconds to keep the connection alive.
type SubscribeMessage struct {
	Type       string        `json:"type"`
	ProductIDs []string      `json:"product_ids"`
	Channels   []interface{} `json:"channels"`
}

// Response: Coinbase Websocket response struct
type Response struct {
	Type         string `json:"type"`
	TradeID      int    `json:"trade_id"`
	Sequence     int    `json:"sequence"`
	MakerOrderID string `json:"maker_order_id"`
	TakerOrderID string `json:"taker_order_id"`
	Time         string `json:"time"`
	ProductID    string `json:"product_id"`
	Size         string `json:"size"`
	Price        string `json:"price"`
	Side         string `json:"side"`
}

type PairTotalValue struct {
	// MatchQueue: overall queue that keeps track of total responses in FIFO format
	MatchQueue []Response

	// TotalSum: attribute that keeps track of the total sum of all the priceses in the match queue
	TotalSum float64

	// VolumeWeightedMovingAverage: attribute that reflects TotalSum/len(MatchQueue)
	VolumeWeightedMovingAverage float64
}
