package service

import "github.com/gorilla/websocket"

type IWebSocketClient interface {
	EstablishConnection(url string) (*websocket.Conn, error)
	WriteMessageToSocketConn(conn *websocket.Conn, data []byte) error
}

const (
	// Channels
	Matches = "matches"

	// Types
	Subscribe = "subscribe"
	Match     = "match"
)

type SubscribeMessage struct {
	Type       string        `json:"type"`
	ProductIDs []string      `json:"product_ids"`
	Channels   []interface{} `json:"channels"`
}

type TickerMessage struct {
	Name       string   `json:"name"`
	ProductIDs []string `json:"product_ids"`
}

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
	MatchQueue []Response
	TotalCount int
	TotalSum   float64
	Average    float64
}
