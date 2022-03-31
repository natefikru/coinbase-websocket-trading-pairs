package websocketClient

import (
	"coinbase-websocket-trading-pairs/util"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type WebSocketClient struct {
}

func NewWebSocketClient(config *util.Config) *WebSocketClient {
	return &WebSocketClient{}
}

func (ws *WebSocketClient) EstablishConnection(url string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		err = errors.Wrap(err, "Error connecting to Websocket Server")
		return nil, err
	}
	log.Printf("Established Socket Connection to: %v", url)
	return conn, nil
}

func (ws *WebSocketClient) WriteMessageToSocketConn(conn *websocket.Conn, data []byte) error {
	for {
		select {
		case <-time.After(time.Second * 4):
			// Send a packet every 4 seconds
			err := conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				err = errors.Wrap(err, "Error during writing to websocket")
				return err
			}
		}
	}
}
